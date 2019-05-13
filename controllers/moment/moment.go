package moment

import (
	"fmt"
	"math"
	"time"
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/factory"
	"rela_recommend/rpc/search"
	"rela_recommend/routers"
	"rela_recommend/algo/moment"
	"rela_recommend/service"
	"rela_recommend/service/abtest"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	"rela_recommend/utils"
	"rela_recommend/utils/response"
	"rela_recommend/utils/request"
)

func RecommendListHTTP(c *routers.Context) {
	var params algo.RecommendRequest
	if err := request.Bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}

	res := DoRecommend(&params)
	c.JSON(response.FormatResponse(res, service.WarpError(nil, "", "")))
}

// 构建上下文
func BuildContext(params *algo.RecommendRequest) (*moment.AlgoContext, error) {
	rank_id := utils.UniqueId()
	userCache := pika.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
	momentCache := redis.NewMomentCacheModule(&factory.CacheCluster, &factory.PikaCluster)
	// search list
	var err error
	dataIds := params.DataIds
	if dataIds == nil || len(dataIds) == 0 {
		dataIds, err = search.CallNearMomentList(params.UserId, params.Lat, params.Lng, 0, 1000)
		if err != nil {
			return nil, err
		}
	}
	// 获取日志内容
	moms, err := momentCache.QueryMomentsByIds(dataIds)
	userIds := make([]int64, 0)
	if err != nil {
		log.Warnf("moment list is err, %s\n", err)
	} else {
		for _, mom := range moms {
			if mom.Moments != nil {
				userIds = append(userIds, mom.Moments.UserId)
			}
		}
		userIds = utils.NewSetInt64FromArray(userIds).ToList()
	}
	// 获取用户信息
	user, usersMap, err := userCache.QueryByUserAndUsersMap(params.UserId, userIds)
	if err != nil {
		log.Warnf("users list is err, %s\n", err)
	}

	userInfo := &moment.UserInfo{
		UserId: params.UserId,
		UserCache: user}

	dataList := make([]moment.DataInfo, 0)
	for _, mom := range moms {
		if mom.Moments != nil && mom.Moments.Id > 0 {
			momUser, _ := usersMap[mom.Moments.UserId]
			info := moment.DataInfo{
				DataId: mom.Moments.Id,
				UserCache: momUser,
				MomentCache: mom.Moments,
				MomentExtendCache: mom.MomentsExtend,
				MomentProfile: mom.MomentsProfile,
				RankInfo: &moment.RankInfo{}}
			dataList = append(dataList, info)
		}
	}

	ctx := moment.AlgoContext{
		Request: params, User: userInfo,
		RankId: rank_id, Platform: utils.GetPlatform(params.Ua),
		CreateTime: time.Now(), AbTest: abtest.GetAbTest("theme", params.UserId),
		DataIds: dataIds, DataList: dataList}

	return &ctx, nil
}

func DoRecommend(params *algo.RecommendRequest) algo.RecommendResponse {
	var startTime = time.Now()
	// 加载缓存
	var startCacheTime = time.Now()
	// 构建上下文
	var startCtxTime = time.Now()
	ctx, err := BuildContext(params)
	if err != nil || ctx == nil || ctx.DataList == nil || len(ctx.DataList) == 0 {
		log.Infof("not list or user,paramuser %d,offset %d,limit %d,err %d\n", 
				  params.UserId, params.Offset, params.Limit, err)
		return algo.RecommendResponse{Status: "error", Message: fmt.Sprintf("not list or user; %s", err)}
	}

	dataLen := len(ctx.DataList)
	// 算法预测打分
	var startPredictTime = time.Now()
	sorter := &moment.DataListSorter{List: ctx.DataList, Context: ctx}
	if err = sorter.DoAlgo(); err != nil {
		log.Errorf("%s\n", err)
	}
	// 结果排序
	var startSortTime = time.Now()
	sorter.DoStrategies()
	sorter.Sort()

	// 分页结果
	var startPageTime = time.Now()
	maxIndex := int64(math.Min(float64(dataLen), float64(params.Offset + params.Limit)))
	returnIds := make([]int64, 0)
	for i := params.Offset; i < maxIndex; i++ {
		j := i // - params.Offset
		currData := ctx.DataList[i]
		returnIds = append(returnIds, currData.DataId)
		// 记录日志
		logStr := algo.RecommendLog{RankId: ctx.RankId, Index: j,
									UserId: ctx.User.UserId,
									DataId: currData.DataId,
									Algo: currData.RankInfo.AlgoName,
									AlgoScore: currData.RankInfo.AlgoScore,
									Score: currData.RankInfo.Score,
									Features: currData.Features.ToString(),
									AbMap: ctx.AbTest.GetTestings() }
		log.Infof("%+v\n", logStr)
	}
	var startLogTime = time.Now()
	log.Infof("rankid %s,paramuser %d,offset %d,limit %d,user %d,paramlen %d,len %d,return %d,max %g,min %g;total:%.3f,init:%.3f,cache:%.3f,ctx:%.3f,predict:%.3f,sort:%.3f,page:%.3f\n",
			  ctx.RankId, params.UserId, params.Offset, params.Limit, params.UserId, dataLen, dataLen, len(returnIds), 0.0, 0.0,
			  startLogTime.Sub(startTime).Seconds(), startCacheTime.Sub(startTime).Seconds(),
			  startCtxTime.Sub(startCacheTime).Seconds(), startPredictTime.Sub(startCtxTime).Seconds(),
			  startSortTime.Sub(startPredictTime).Seconds(), startPageTime.Sub(startSortTime).Seconds(),
			  startLogTime.Sub(startPageTime).Seconds())
	// 返回
	res := algo.RecommendResponse{RankId: ctx.RankId, DataIds: returnIds, Status: "ok"}
	return res
}
