package algo

import (
	"math"
	"rela_recommend/log"
	rutils "rela_recommend/utils"
)

// 策略组件
type IBuilder interface {
	Do(ctx IContext) error
}

type BuilderBase struct {
	DoBuild func(IContext) error
}

func (self *BuilderBase) Do(ctx IContext) error {
	return self.DoBuild(ctx)
}

// 策略组件
type IStrategy interface {
	Do(ctx IContext) error
}

type StrategyBase struct {
	DoSingle func(IContext, int) error
}

func (self *StrategyBase) Do(ctx IContext) error {
	var err error
	for i := 0; i < ctx.GetDataLength(); i++ {
		err = self.DoSingle(ctx, i)
		if err != nil {
			break
		}
	}
	return err
}

// 计算分数
func StrategyScoreFunc(ctx IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index)
	rankInfo := dataInfo.GetRankInfo()
	for _, item := range rankInfo.Recommends {
		if item.Score > 0 {
			rankInfo.Score *= item.Score
		}
	}
	// 防止score正负无穷导致返回无法序列化, 最大值10000
	score64 := float64(rankInfo.Score)
	if math.IsInf(score64, 0) {
		rankInfo.Score = float32(math.Copysign(100000000, score64))
	}
	return nil
}

// 排序组件
type ISorter interface {
	Do(ctx IContext) error
}

// 分页组件
type IPager interface {
	Do(ctx IContext) error
}

type PagerBase struct{}

func (self *PagerBase) Do(ctx IContext) error {
	params := ctx.GetRequest()
	index := math.Min(float64(ctx.GetDataLength()), float64(params.Offset+params.Limit))
	minIndex := int(math.Max(float64(params.Offset), 0.0))
	maxIndex := int(math.Max(index, 0.0))

	response, err := self.BuildResponse(ctx, minIndex, maxIndex)
	ctx.SetResponse(response)
	return err
}

func (self *PagerBase) BuildResponse(ctx IContext, minIndex int, maxIndex int) (*RecommendResponse, error) {
	returnIds, returnObjs := make([]int64, 0), make([]RecommendResponseItem, 0)
	for i := minIndex; i < maxIndex; i++ {
		currData := ctx.GetDataByIndex(i)
		returnIds = append(returnIds, currData.GetDataId())
		rankInfo := currData.GetRankInfo()
		rankInfo.Index = i
		returnObjs = append(returnObjs, RecommendResponseItem{
			DataId:         currData.GetDataId(),
			PlanId:         rankInfo.PlanId,
			Data:           currData.GetResponseData(ctx),
			Index:          rankInfo.Index,
			Score:          rankInfo.Score,
			Reason:         rankInfo.ReasonString(),
			ReasonMultiple: rankInfo.ClientReasonString(),
		})
	}
	response := &RecommendResponse{RankId: ctx.GetRankId(), DataIds: returnIds, DataList: returnObjs, Status: "ok"}
	return response, nil
}

// 不分页返回原版数据
type PagerOrigin struct {
	PagerBase
}

func (self *PagerOrigin) Do(ctx IContext) error {
	response, err := self.BuildResponse(ctx, 0, ctx.GetDataLength())
	ctx.SetResponse(response)
	return err
}

type ILogger interface {
	Do(ctx IContext) error
}

type LoggerBase struct{}

func (self *LoggerBase) Do(ctx IContext) error {
	response := ctx.GetResponse()
	if response != nil {
		for _, item := range response.DataList {
			currData := ctx.GetDataByIndex(item.Index)
			rankInfo := currData.GetRankInfo()

			logStr := RecommendLog{
				Module:          ctx.GetAppInfo().Name,
				RankId:          ctx.GetRankId(),
				Index:           int64(item.Index),
				DataId:          currData.GetDataId(),
				UserId:          ctx.GetRequest().UserId,
				Algo:            rankInfo.AlgoName,
				AlgoScore:       rankInfo.AlgoScore,
				PlanId:          rankInfo.PlanId,
				Score:           rankInfo.Score,
				RecommendScores: rankInfo.RecommendsString(),
				Features:        rankInfo.GetFeaturesString(),
				AbMap:           ctx.GetAbTest().GetTestings(rankInfo.ExpId, rankInfo.RequestId),
				PagedIndex:      rankInfo.PagedIndex,
			}
			log.Infof("%+v\n", logStr) // 此日志格式会有实时任务解析，谨慎更改
		}
	}
	return nil
}

type LoggerPerforms struct{}

func (self *LoggerPerforms) Do(ctx IContext) error {
	app := ctx.GetAppInfo()
	pfm := ctx.GetPerforms()
	params := ctx.GetRequest()
	response := ctx.GetResponse()
	abtest := ctx.GetAbTest()
	returnLen := 0
	if response != nil {
		returnLen = len(response.DataIds)
	}
	version := params.GetVersion()

	requestLog := RecommendRequestLog{
		Module:  app.Name,
		Limit:   params.Limit,
		Offset:  params.Offset,
		Ua:      params.Ua,
		Os:      params.GetOS(),
		Version: version,
		Lat:     params.Lat,
		Lng:     params.Lng,
		UserId:  params.UserId,
		// DataIds
		AbMap:  params.AbMap,
		Params: params.Params,
		//
		CreateTime: ctx.GetCreateTime(),
		RankId:     ctx.GetRankId(),
		Returns:    returnLen,
		Performs:   pfm.ToJson(),
	}
	log.Infof("performs %s\n", requestLog.ToJson())

	if abtest.GetBool("logger_performs_writer_switched", true) {
		pfm.ToWriteChan("algo", map[string]string{
			"app":     app.Name,
			"os":      params.GetOS(),
			"version": rutils.GetString(version),
		}, map[string]interface{}{
			"request.user_id": params.UserId,
			"request.offset":  params.Offset,
			"request.limit":   params.Limit,
			"response.len":    returnLen,
		}, ctx.GetCreateTime())
	}
	return nil
}

type IRichStrategy interface {
	New(ctx IContext) IRichStrategy
	GetDefaultWeight() int
	BuildData() error // 加载数据
	Strategy() error  // 执行策略
	Logger() error    // 记录结果
}
