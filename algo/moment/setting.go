package moment

import(
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
)

var appName = "moment"
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

// 精排算法
var algosMap = algo.AlgoListInitToMap([]algo.IAlgo{
	&algo.AlgoBase{AlgoName: "model_base", FilePath: workDir + "moment_xg_v1.1.model", 
				   Model: &utils.XgboostClassifier{}, FeaturesFunc: GetMomentFeatures },
})

var appInfo = &algo.AppInfo{
	Name: appName, Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: algosMap, 
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	StrategyKey: "strategies", StrategyDefault: "time_level", StrategyMap: strategyMap, 
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	LoggerKey: "loggers", LoggerDefault: "features,performs", LoggerMap: loggerMap}
var _ = algo.AddAppInfo(appInfo)
