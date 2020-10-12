package algo

import (
	"math"
	"rela_recommend/log"
	rutils "rela_recommend/utils"
	"sort"
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

type SorterBase struct {
	Context IContext
}

func (self SorterBase) Swap(i, j int) {
	list := self.Context.GetDataList()
	list[i], list[j] = list[j], list[i]
}
func (self SorterBase) Len() int {
	return self.Context.GetDataLength()
}

// 以此按照：打分，最后登陆时间
func (self SorterBase) Less(i, j int) bool {
	listi, listj := self.Context.GetDataByIndex(i), self.Context.GetDataByIndex(j)
	ranki, rankj := listi.GetRankInfo(), listj.GetRankInfo()

	if ranki.IsTop != rankj.IsTop {
		return ranki.IsTop > rankj.IsTop // IsTop ： 倒序， 是否置顶
	} else {
		if ranki.PagedIndex != rankj.PagedIndex { // PagedIndex: 已经被分页展示过的index, 升序排列
			return ranki.PagedIndex < rankj.PagedIndex
		} else {
			if ranki.Level != rankj.Level {
				return ranki.Level > rankj.Level // Level : 倒序， 推荐星数
			} else {
				if ranki.Score != rankj.Score {
					return ranki.Score > rankj.Score // Score : 倒序， 推荐分数
				} else {
					return listi.GetDataId() < listj.GetDataId() // UserId : 正序
				}
			}
		}
	}
}

func (self *SorterBase) Do(ctx IContext) error {
	sorter := &SorterBase{Context: ctx}
	sort.Sort(sorter)
	return nil
}

// 返回，不做排序
type SorterOrigin struct {
	Context IContext
}

func (self *SorterOrigin) Do(ctx IContext) error {
	return nil
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
			DataId: currData.GetDataId(),
			Data:   currData.GetResponseData(ctx),
			Index:  rankInfo.Index,
			Score:  rankInfo.Score,
			Reason: rankInfo.ReasonString()})
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
			logStr := RecommendLog{Module: ctx.GetAppInfo().Name,
				RankId: ctx.GetRankId(), Index: int64(item.Index),
				DataId:          currData.GetDataId(),
				UserId:          ctx.GetRequest().UserId,
				Algo:            rankInfo.AlgoName,
				AlgoScore:       rankInfo.AlgoScore,
				Score:           rankInfo.Score,
				RecommendScores: rankInfo.RecommendsString(),
				Features:        rankInfo.GetFeaturesString(),
				AbMap:           ctx.GetAbTest().GetTestings()}
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
		RankId:   ctx.GetRankId(),
		Returns:  returnLen,
		performs: pfm.ToJson(),
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
