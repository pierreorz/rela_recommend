package live

import (
	"fmt"
	"math"
	"time"
	"rela_recommend/algo"
	"rela_recommend/algo/live"
	"rela_recommend/log"
	"rela_recommend/factory"
	"rela_recommend/utils/routers"
	"rela_recommend/service"
	"rela_recommend/service/abtest"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	"rela_recommend/utils"
	"rela_recommend/utils/response"
	"rela_recommend/utils/request"
)

type LiveRecommendRequest struct {
	Limit   int64  `json:"limit" form:"limit"`
	Offset  int64  `json:"offset" form:"offset"`
	Ua      string `json:"ua" form:"ua"`
	UserId  int64  `json:"userId" form:"userId"`
	LiveIdStr string `json:"liveIds" form:"liveIds"`
	LiveIds []int64
}

type LiveRecommendResponse struct {
	Status  string		`json:"status" form:"status"`
	Message string		`json:"message" form:"message"`
	RankId	string		`json:"rankId" form:"rankId"`
	LiveIds []int64		`json:"liveIds" form:"liveIds"`
}

func LiveRecommendListHTTP(c *routers.Context) {
	var params LiveRecommendRequest
	if err := request.Bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}
	params.LiveIds = utils.GetInt64s(params.LiveIdStr)

	old := true
	if old {
		res := DoRecommend(&params)
		c.JSON(response.FormatResponse(res, service.WarpError(nil, "", "")))
	} else {
		var params2 = &algo.RecommendRequest{
			Limit: params.Limit,
			Offset: params.Offset,
			Ua: params.Ua,
			Lat: 0.0,
			Lng: 0.0,
			UserId: params.UserId,
			DataIds: params.LiveIds,
		}
		ctx := &algo.ContextBase{}
		err := ctx.Do(algo.GetAppInfo("live"), params2)
		res2 := ctx.GetResponse()
		res := LiveRecommendResponse{
			Status: res2.Status,
			Message: res2.Message,
			RankId: res2.RankId,
			LiveIds: res2.DataIds,
		}
		c.JSON(response.FormatResponse(res, service.WarpError(err, "", "")))
	}
}

// 构建上下文
func BuildContext(params *LiveRecommendRequest) (*live.LiveAlgoContext, error) {
	rank_id := utils.UniqueId()
	userCache := pika.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
	liveCache := pika.NewLiveCacheModule(&factory.CacheLiveRds)

	rdsPikaCache := redis.NewLiveCacheModule(nil, &factory.CacheCluster, &factory.PikaCluster)

	// 获取主播列表
	allLives := live.GetCachedLiveList()
	if allLives == nil || len(allLives) == 0 {
		var err error
		allLives, err = liveCache.QueryLiveList()
		log.Warnf("cached live list is nil, %s\n", err)
	}
	lives := liveCache.MgetByLiveIds(allLives, params.LiveIds)
	liveIds := make([]int64, len(lives))
	for i, _ := range lives {
		liveIds[i] = lives[i].Live.UserId
	}
	// 获取基础用户画像
	user, users, err := userCache.QueryByUserAndUsers(params.UserId, liveIds)
	if err != nil {
		log.Errorf("QueryByUserAndUsers err: %s\n", err)
		return nil, err
	}
	usersMap := make(map[int64]pika.UserProfile)
	for i, _ := range users {
		usersMap[users[i].UserId] = users[i]
	}
	// 获取刷新用户画像
	user2, users2, err2 := rdsPikaCache.QueryLiveProfileByUserAndUsers(params.UserId, liveIds)
	if err2 != nil {
		log.Warnf("redis QueryLiveProfileByUserAndUsers err: %s\n", err2)
	}
	usersMap2 := make(map[int64]redis.LiveProfile)
	for i, _ := range users2 {
		usersMap2[users2[i].UserId] = users2[i]
	}

	// 获取关注信息
	// concerns := make([]int64, 0)
	concerns, err := userCache.QueryConcernsByUser(params.UserId)
	if err != nil {
		log.Warnf("QueryConcernsByUser err: %s\n", err)
	}


	livesInfo := make([]live.LiveInfo, 0)
	for i, _ := range lives {
		liveInfo := live.LiveInfo{ 
			UserId: lives[i].Live.UserId, 
			LiveCache: &lives[i], 
			UserCache: nil, LiveProfile: nil,
			RankInfo: &algo.RankInfo{} }
		if liveUser, ok := usersMap[lives[i].Live.UserId]; ok {
			liveInfo.UserCache = &liveUser
		}
		if liveUser2, ok := usersMap2[lives[i].Live.UserId]; ok {
			liveInfo.LiveProfile = &liveUser2
		}
		livesInfo = append(livesInfo, liveInfo)
	}

	userInfo := &live.UserInfo{
		UserId: user.UserId, UserCache: &user, 
		LiveProfile: &user2,
		UserConcerns: utils.NewSetInt64FromArray(concerns)}

	ctx := live.LiveAlgoContext{
		RankId: rank_id, Ua: params.Ua, Platform: utils.GetPlatform(params.Ua),
		CreateTime: time.Now(), AbTest: abtest.GetAbTest("live", params.UserId),
		User: userInfo, LiveList: livesInfo}

	return &ctx, nil
}

func DoRecommend(params *LiveRecommendRequest) LiveRecommendResponse {
	var startTime = time.Now()
	// 加载缓存
	var startCacheTime = time.Now()
	// 构建上下文
	var startCtxTime = time.Now()
	ctx, err := BuildContext(params)
	if err != nil || ctx == nil || ctx.LiveList == nil || ctx.User == nil {
		log.Infof("not list or user,paramuser %d,offset %d,limit %d,paramlen %d,err %d\n", 
				  params.UserId, params.Offset, params.Limit, len(params.LiveIds), err)
		return LiveRecommendResponse{Status: "error", Message: fmt.Sprintf("not list or user; %s", err)}
	}

	dataLen := len(ctx.LiveList)
	// 算法预测打分
	var startPredictTime = time.Now()
	var modelName = ctx.AbTest.GetString("live_model", "LiveModelV1_0")
	model, ok := live.LiveAlgosMap[modelName]
	// log.Infof("%s,%s,%s,%s", modelName, model, ok, ctx.AbTest.FactorMap)
	if !ok {
		log.Errorf("model not find: %s\n", modelName)
		model = nil
	}
	model.Predict(ctx)
	// 结果排序
	var startSortTime = time.Now()
	sorter := &live.LiveInfoListSorter{List: ctx.LiveList, Context: ctx}
	sorter.DoStrategies()
	sorter.Sort()

	// 分页结果
	var startPageTime = time.Now()
	maxIndex := int64(math.Min(float64(dataLen), float64(params.Offset + params.Limit)))
	returnIds := make([]int64, 0)
	for i := params.Offset; i < maxIndex; i++ {
		j := i // - params.Offset
		currData := ctx.LiveList[i]
		returnIds = append(returnIds, currData.UserId)
		// 记录日志
		logStr := algo.RecommendLog{RankId: ctx.RankId, Index: j,
									UserId: ctx.User.UserId,
									DataId: currData.UserId,
									Algo: model.Name(),
									AlgoScore: currData.RankInfo.AlgoScore,
									Score: currData.RankInfo.Score,
									Features: currData.Features.ToString(),
									AbMap: ctx.AbTest.GetTestings() }
		log.Infof("%+v\n", logStr)
	}
	var startLogTime = time.Now()
	var minScore, maxScore float32
	if dataLen > 0 {
		minScore, maxScore = ctx.LiveList[0].RankInfo.Score, ctx.LiveList[dataLen-1].RankInfo.Score
	}
	log.Infof("rankid %s,paramuser %d,offset %d,limit %d,user %d,paramlen %d,len %d,return %d,max %g,min %g;total:%.3f,init:%.3f,cache:%.3f,ctx:%.3f,predict:%.3f,sort:%.3f,page:%.3f\n",
			  ctx.RankId, params.UserId, params.Offset, params.Limit, ctx.User.UserId, len(params.LiveIds), dataLen, len(returnIds), minScore, maxScore,
			  startLogTime.Sub(startTime).Seconds(), startCacheTime.Sub(startTime).Seconds(),
			  startCtxTime.Sub(startCacheTime).Seconds(), startPredictTime.Sub(startCtxTime).Seconds(),
			  startSortTime.Sub(startPredictTime).Seconds(), startPageTime.Sub(startSortTime).Seconds(),
			  startLogTime.Sub(startPageTime).Seconds())
	// 返回
	res := LiveRecommendResponse{RankId: ctx.RankId, LiveIds: returnIds, Status: "ok"}
	return res
}
