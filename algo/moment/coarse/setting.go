package coarse

import (
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/algo/moment"
)

var appName = "moment_coarse"
var workDir = algo.GetWorkDir("/algo_files/moment/coarse/")

var builderMap = map[string]algo.IBuilder{"base": &algo.BuilderBase{DoBuild: DoBuildCoarseData}}
var strategyMap = map[string]algo.IStrategy{ }
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
	&algo.AlgoBase{AlgoName: "model_base", FilePath: workDir + "moment_coarse_xg_v1.0.model",
				   Model: &utils.XgboostClassifier{}, FeaturesFunc: moment.GetMomentFeatures },
})


var appInfo = &algo.AppInfo{
	Name: appName, Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: algosMap,
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: nil,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap}
var _ = algo.AddAppInfo(appInfo)
