package live

import (
	"math"
	"time"
	"rela_recommend/algo"
	"rela_recommend/algo/live"
	"rela_recommend/log"
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/service/abtest"
	"rela_recommend/utils"
	"sort"
	"rela_recommend/utils/response"
	"rela_recommend/utils/request"
)

type Live struct {
	LiveId  int64 `json:"liveId" form:"liveId"`
	ViewCount int `json:"views" form:"views"`
}

type LiveRecommendRequest struct {
	Limit   int64  `json:"limit" form:"limit"`
	Offset  int64  `json:"offset" form:"offset"`
	Ua      string `json:"ua" form:"ua"`
	UserId  int64  `json:"userId" form:"userId"`
	Lives   []Live `json:"lives" form:"lives"`
}

type LiveRecommendResponse struct {
	Status  string
	RankId	string
	LiveIds []int64
}

func LiveRecommendListHTTP(c *routers.Context) {
	var params LiveRecommendRequest
	if err := request.Bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}
	res := DoRecommend(&params)
	c.JSON(response.FormatResponse(res, service.WarpError(nil, "", "")))
}

func DoRecommend(params *LiveRecommendRequest) LiveRecommendResponse {
	var startTime = time.Now()
	rank_id := utils.UniqueId()
	// 加载缓存
	var startCacheTime = time.Now()
	// 构建上下文
	var startCtxTime = time.Now()
	ctx := live.LiveAlgoContext{
		RankId: rank_id, Ua: params.Ua,
		AbTest: abtest.GetAbTest("live", params.UserId),
		User: nil, LiveList: nil}
	
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

	// 结果排序
	var startSortTime = time.Now()
	sort.Sort(live.LiveInfoListSorter{ctx.LiveList, &ctx})

	// 分页结果
	var startPageTime = time.Now()
	maxIndex := int64(math.Min(float64(len(params.Lives)), float64(params.Offset + params.Limit)))
	returnIds := make([]int64, maxIndex - params.Offset)
	for i := params.Offset; i < maxIndex; i++ {
		j := i // - params.Offset
		currData := ctx.LiveList[i]
		returnIds[j] = currData.UserId
		// 记录日志
		logStr := algo.RecommendLog{RankId: rank_id, Index: j,
									UserId: ctx.User.UserId,
									DataId: currData.UserId,
									Algo: model.Name(),
									AlgoScore: currData.AlgoScore,
									Score: currData.Score,
									Features: algo.Features2String(currData.Features),
									AbMap: ctx.AbTest.GetTestings() }
		log.Infof("%+v\n", logStr)
	}
	var startLogTime = time.Now()
	log.Infof("paramuser %d,user %d,paramlen %d,len %d,return %d,max %g,min %g;total:%.3f,init:%.3f,cache:%.3f,ctx:%.3f,predict:%.3f,sort:%.3f,page:%.3f\n",
			  params.UserId, ctx.User.UserId, len(params.Lives), dataLen, len(returnIds),
			  ctx.LiveList[0].Score, ctx.LiveList[dataLen-1].Score,
			  startLogTime.Sub(startTime).Seconds(), startCacheTime.Sub(startTime).Seconds(),
			  startCtxTime.Sub(startCacheTime).Seconds(), startPredictTime.Sub(startCtxTime).Seconds(),
			  startSortTime.Sub(startPredictTime).Seconds(), startPageTime.Sub(startSortTime).Seconds(),
			  startLogTime.Sub(startPageTime).Seconds())
	// 返回
	res := LiveRecommendResponse{RankId: rank_id, LiveIds: returnIds, Status: "ok"}
	return res
}
