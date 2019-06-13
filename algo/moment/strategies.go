package moment

import(
	"rela_recommend/algo"
)

var strategyMap = map[string]algo.IStrategy{
	"time_level": &algo.StrategyBase{ DoSingle: DoTimeLevel },
}
var sorterMap = map[string]algo.ISorter{
	"base": &algo.SorterBase{},
}
var pagerMap = map[string]algo.IPager{
	"base": &algo.PagerBase{},
}
var loggerMap = map[string]algo.ILogger{
	"features": &algo.LoggerBase{},
	"performs": &algo.LoggerPerforms{},
}

// 按照6小时优先策略
func DoTimeLevel(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	hours := int(ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Hours()) / 6
	rankInfo.Level = -hours
	return nil
}
