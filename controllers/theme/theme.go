package theme

import (
	"fmt"
	"math"
	"time"
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/factory"
	"rela_recommend/routers"
	"rela_recommend/algo/theme"
	"rela_recommend/service"
	"rela_recommend/service/abtest"
	// "rela_recommend/models/pika"
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
func BuildContext(params *algo.RecommendRequest) (*theme.ThemeAlgoContext, error) {
	rank_id := utils.UniqueId()
	rdsPikaCache := redis.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)

	dataList, err := rdsPikaCache.GetInt64List(params.UserId, "theme_recommend_list:%d")
	if err == nil {
		log.Warnf("theme recommend list is nil, %s\n", err)
	}
	if len(dataList) == 0{
		dataList, _ = rdsPikaCache.GetInt64List(-999999999, "theme_recommend_list:%d")
	}
	ctx := theme.ThemeAlgoContext{
		Request: params,
		RankId: rank_id, Platform: utils.GetPlatform(params.Ua),
		CreateTime: time.Now(), AbTest: abtest.GetAbTest("theme", params.UserId),
		ThemeIds: dataList}

	return &ctx, nil
}

func DoRecommend(params *algo.RecommendRequest) algo.RecommendResponse {
	var startTime = time.Now()
	// 加载缓存
	var startCacheTime = time.Now()
	// 构建上下文
	var startCtxTime = time.Now()
	ctx, err := BuildContext(params)
	if err != nil || ctx == nil || ctx.ThemeIds == nil {
		log.Infof("not list or user,paramuser %d,offset %d,limit %d,err %d\n", 
				  params.UserId, params.Offset, params.Limit, err)
		return algo.RecommendResponse{Status: "error", Message: fmt.Sprintf("not list or user; %s", err)}
	}

	dataLen := len(ctx.ThemeIds)
	// 算法预测打分
	var startPredictTime = time.Now()
	modelName := "v0"
	// 结果排序
	var startSortTime = time.Now()

	// 分页结果
	var startPageTime = time.Now()
	maxIndex := int64(math.Min(float64(dataLen), float64(params.Offset + params.Limit)))
	returnIds := make([]int64, 0)
	for i := params.Offset; i < maxIndex; i++ {
		j := i // - params.Offset
		currData := ctx.ThemeIds[i]
		returnIds = append(returnIds, currData)
		// 记录日志
		logStr := algo.RecommendLog{RankId: ctx.RankId, Index: j,
									UserId: ctx.Request.UserId,
									DataId: currData,
									Algo: modelName,
									AlgoScore: 0.0,
									Score: 0.0,
									Features: "",
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
