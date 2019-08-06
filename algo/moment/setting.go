package moment

import(
	"rela_recommend/algo"
	"rela_recommend/algo/base"
	"rela_recommend/algo/utils"
	"rela_recommend/algo/base/strategy"
)

var workDir = algo.GetWorkDir("/algo_files/moment/")

var builderMap = map[string]algo.IBuilder{"base": &algo.BuilderBase{DoBuild: DoBuildData}}
var strategyMap = map[string]algo.IStrategy{
	"time_level": &algo.StrategyBase{ DoSingle: DoTimeLevel },
	"time_frist": &algo.StrategyBase{ DoSingle: DoTimeFirstLevel },
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
var richStrategyMap = map[string]algo.IRichStrategy {
	"paged": &base.PagedRichStrategy{},
	"behavior": &strategy.BaseBehaviorRichStrategy{
		UserStrategyItemFunc: UserBehaviorStrategyFunc,
		ItemStrategyItemFunc: ItemBehaviorStrategyFunc},
}


// 精排算法
var algosMap = algo.AlgoListInitToMap([]algo.IAlgo{
	&algo.AlgoBase{AlgoName: "model_base", FilePath: workDir + "moment_xg_v1.1.model", 
				   Model: &utils.XgboostClassifier{}, FeaturesFunc: GetMomentFeatures },
})

// 推荐日志
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "moment", Module: "moment", Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: algosMap, 
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	StrategyKey: "strategies", StrategyDefault: "time_level", StrategyMap: strategyMap, 
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	LoggerKey: "loggers", LoggerDefault: "features,performs", LoggerMap: loggerMap,
	RichStrategyKey: "rich_strategies", RichStrategyDefault:"paged", RichStrategyMap: richStrategyMap})

// 日志附近的人
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "moment.near", Module: "moment", Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: algosMap, 
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	StrategyKey: "strategies", StrategyDefault: "time_level", StrategyMap: strategyMap, 
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	LoggerKey: "loggers", LoggerDefault: "features,performs", LoggerMap: loggerMap,
	RichStrategyKey: "rich_strategies", RichStrategyDefault:"paged", RichStrategyMap: richStrategyMap})
