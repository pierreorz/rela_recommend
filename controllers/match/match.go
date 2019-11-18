package match

import (
	"math"
	"time"
	"rela_recommend/algo"
	"rela_recommend/algo/base"
	"rela_recommend/algo/quick_match"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/pika"
	"rela_recommend/utils/routers"
	"rela_recommend/service"
	"rela_recommend/service/abtest"
	"rela_recommend/utils"
	"sort"
	"strings"
	"rela_recommend/utils/response"
	"rela_recommend/utils/request"
)

type MatchRecommendReqParams struct {
	Limit   int64  `json:"limit" form:"limit"`
	Offset  int64  `json:"offset" form:"offset"`
	Ua      string `json:"ua" form:"ua"`
	UserId  int64  `json:"userId" form:"userId"`
	UserIds string `json:"userIds" form:"userIds"`
	AbMap	map[string]string	`json:"abMap" form:"abMap"`
}

type MatchRecommendResponse struct {
	Status  string
	RankId	string
	UserIds []int64
}

type MatchRecommendLog struct {
	RankId string
	Index int64
	UserId int64
	ReceiverId int64
	Algo string
	AlgoScore float32
	Score float32
	Features string
	AbMap string
}

func MatchRecommendListHTTP(c *routers.Context) {
	var params MatchRecommendReqParams
	if err := request.Bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}
	var userIds = make([]int64, 0)
	var userIdsStrs = strings.Split(params.UserIds, ",")
	for _, uid := range userIdsStrs {
		userIds = append(userIds, utils.GetInt64(uid))
	}
	matchAb := abtest.GetAbTestWithSetting("match", params.UserId, params.AbMap)

	old := matchAb.GetBool("use_old_algo", true)
	if old {
		res := DoRecommend(&params, userIds)
		c.JSON(response.FormatResponse(res, service.WarpError(nil, "", "")))
	} else {
		var params2 = &algo.RecommendRequest{
			Limit: params.Limit,
			Offset: params.Offset,
			Ua: params.Ua,
			Lat: 0.0,
			Lng: 0.0,
			UserId: params.UserId,
			DataIds: userIds,
			AbMap: params.AbMap,
		}
		ctx := &base.ContextBase{}
		err := ctx.Do(algo.GetAppInfo("match"), params2)
		res2 := ctx.GetResponse()
		res := MatchRecommendResponse{
			Status: res2.Status,
			RankId: res2.RankId,
			UserIds: res2.DataIds,
		}
		c.JSON(response.FormatResponse(res, service.WarpError(err, "", "")))
	}
}

func DoRecommend(params *MatchRecommendReqParams, userIds []int64) MatchRecommendResponse {
	var startTime = time.Now()
	matchAb := abtest.GetAbTest("match", params.UserId)
	rank_id := matchAb.RankId
	// 加载用户缓存
	var startCacheTime = time.Now()
	aulm := pika.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
	user, users, err := aulm.QueryByUserAndUsers(params.UserId, userIds)
	if err != nil {
		log.Error(err.Error())
	}
	userLen := len(users)
	// 构建上下文
	var startCtxTime = time.Now()
	userInfo := &quick_match.UserInfo{UserId: user.UserId, UserCache: &user}
	usersInfo := make([]quick_match.UserInfo, userLen)
	for i, u := range users {
		usersInfo[i].UserId = u.UserId
		usersInfo[i].UserCache = &users[i]
	}
	ctx := quick_match.QuickMatchContext{
		RankId: rank_id, Ua: params.Ua,
		AbTest: matchAb,
		User: userInfo, UserList: usersInfo}
	dataLen := len(ctx.UserList)
	// 算法预测打分
	var startPredictTime = time.Now()

	var modelName = ctx.AbTest.GetString("match_model", "QuickMatchTreeV1_0")
	model, ok := quick_match.MatchAlgosMap[modelName]
	// log.Infof("%s,%s,%s,%s", modelName, model, ok, ctx.AbTest.FactorMap)
	if !ok {
		log.Errorf("model not find: %s\n", modelName)
		model = quick_match.MatchAlgoV1_0
	}
	
	model.Predict(&ctx)
	// 提升活跃用户权重
	ActiveUserUpper(&ctx)
	// 结果排序
	var startSortTime = time.Now()
	sort.Sort(quick_match.UserInfoListSort(ctx.UserList))
	// 分页结果
	var startPageTime = time.Now()
	maxIndex := int64(math.Min(float64(len(ctx.UserList)), float64(params.Offset+params.Limit)))
	returnIds := make([]int64, maxIndex-params.Offset)
	for i := params.Offset; i < maxIndex; i++ {
		j := i - params.Offset
		currUser := ctx.UserList[i]
		returnIds[j] = currUser.UserId
		// 记录日志
		logStr := MatchRecommendLog{RankId: rank_id, Index: j,
									UserId: ctx.User.UserId,
									ReceiverId: currUser.UserId,
									Algo: model.Name(),
									AlgoScore: currUser.AlgoScore,
									Score: currUser.Score,
									Features: algo.Features2String(currUser.Features),
									AbMap: ctx.AbTest.GetTestings() }
		log.Infof("%+v\n", logStr)
	}
	var startLogTime = time.Now()
	var minScore, maxScore float32
	if dataLen > 0 {
		minScore, maxScore = ctx.UserList[0].Score, ctx.UserList[userLen-1].Score
	}
	log.Infof("paramuser %d,user %d,paramlen %d,len %d,return %d,max %g,min %g;total:%.3f,init:%.3f,cache:%.3f,ctx:%.3f,predict:%.3f,sort:%.3f,page:%.3f\n",
			  params.UserId, ctx.User.UserId, len(userIds), userLen, len(returnIds), minScore, maxScore,
			  startLogTime.Sub(startTime).Seconds(), startCacheTime.Sub(startTime).Seconds(),
			  startCtxTime.Sub(startCacheTime).Seconds(), startPredictTime.Sub(startCtxTime).Seconds(),
			  startSortTime.Sub(startPredictTime).Seconds(), startPageTime.Sub(startSortTime).Seconds(),
			  startLogTime.Sub(startPageTime).Seconds())
	// 返回
	res := MatchRecommendResponse{RankId: rank_id, UserIds: returnIds, Status: "ok"}
	return res
}

func ActiveUserUpper(ctx *quick_match.QuickMatchContext) {
	var upperRate float32 = ctx.AbTest.GetFloat("match_active_user_upper", 0.1)
	var offsetTime int64 = 1 * 60 * 60
	nowTime := time.Now().Unix()
	before24HourTime := nowTime - offsetTime
	for i, user := range ctx.UserList {
		if user.UserCache.LastUpdateTime >= before24HourTime {
			var addRate = float32(user.UserCache.LastUpdateTime - before24HourTime) / float32(offsetTime) * upperRate
			ctx.UserList[i].Score = ctx.UserList[i].Score * (1.0 + addRate)
			// log.Infof("ActiveUserUpper before:%d, last:%d, add:%.3f old:%.3f new:%.3f", 
			// before24HourTime, user.UserCache.LastUpdateTime, addRate, ctx.UserList[i].AlgoScore, ctx.UserList[i].Score)
		}
	}
}
