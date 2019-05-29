package moment

import (
	"fmt"
	"math"
	"time"
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/factory"
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

func CoarseRecommendListHTTP(c *routers.Context) {
	var params algo.RecommendRequest
	if err := request.Bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}

	res := DoCoarseRecommend(&params)
	c.JSON(response.FormatResponse(res, service.WarpError(nil, "", "")))
}

// 构建上下文
func BuildCoarseContext(params *algo.RecommendRequest) (*moment.AlgoContext, error) {
	var startTime = time.Now()
	abTest := abtest.GetAbTest("theme", params.UserId)
	rank_id := utils.UniqueId()
	userCache := pika.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
	momentCache := redis.NewMomentCacheModule(&factory.CacheCluster, &factory.PikaCluster)

	// search list
	var err error
	dataIds := params.DataIds
	// 获取日志内容
	var startMomentTime = time.Now()
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
	var startUserTime = time.Now()
	_, usersMap, err := userCache.QueryByUserAndUsersMap(params.UserId, userIds)
	if len(usersMap) > 0 {
		log.Warnf("users list is err, %s\n", err)
	}

	var startBuildTime = time.Now()
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
		Request: params, User: nil,
		RankId: rank_id, Platform: utils.GetPlatform(params.Ua),
		CreateTime: time.Now(), AbTest: abTest,
		DataIds: dataIds, DataList: dataList}

	var endTime = time.Now()
	log.Infof("rankid %s,searchlen:%d;total:%.3f,other:%.3f,moment:%.3f,user:%.3f,build:%.3f\n",
			  ctx.RankId, len(dataIds),
			  endTime.Sub(startTime).Seconds(), startMomentTime.Sub(startTime).Seconds(),
			  startUserTime.Sub(startMomentTime).Seconds(), startBuildTime.Sub(startUserTime).Seconds(),
			  endTime.Sub(startBuildTime).Seconds() )
	return &ctx, nil
}

func DoCoarseRecommend(params *algo.RecommendRequest) algo.RecommendResponse {
	var startTime = time.Now()
	// 加载缓存
	var startCacheTime = time.Now()
	// 构建上下文
	var startCtxTime = time.Now()
	ctx, err := BuildCoarseContext(params)
	if err != nil || ctx == nil || ctx.DataList == nil || len(ctx.DataList) == 0 {
		log.Infof("not list or user,paramuser %d,offset %d,limit %d,err %d\n", 
				  params.UserId, params.Offset, params.Limit, err)
		return algo.RecommendResponse{Status: "error", Message: fmt.Sprintf("not list or user; %s", err)}
	}

	dataLen := len(ctx.DataList)
	// 算法预测打分
	var startPredictTime = time.Now()

	var modelName = ctx.AbTest.GetString("moment_coarse_model", "MomentCoarseModelV1_0")
	if model, ok := moment.AlgosCoarseMap[modelName]; ok {
		model.Predict(ctx)
	} else {
		log.Errorf("algo not found:%s\n", modelName)
	}
	// 结果排序
	var startSortTime = time.Now()

	// 分页结果
	var startPageTime = time.Now()
	maxIndex := int64(math.Min(float64(dataLen), float64(params.Offset + params.Limit)))
	returnIds, returnObjs := make([]int64, 0), make([]algo.RecommendResponseItem, 0)
	for i := params.Offset; i < maxIndex; i++ {
		j := i // - params.Offset
		currData := ctx.DataList[i]
		returnIds = append(returnIds, currData.DataId)
		returnObjs = append(returnObjs, algo.RecommendResponseItem{
			DataId: currData.DataId, 
			Score: currData.RankInfo.AlgoScore })
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
	res := algo.RecommendResponse{RankId: ctx.RankId, DataIds: returnIds, DataList: returnObjs, Status: "ok"}
	return res
}
