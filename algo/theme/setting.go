package theme

import (
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
)

var appName = "theme"
var workDir = algo.GetWorkDir("/algo_files/theme/")

var builderMap = map[string]algo.IBuilder{"base": &algo.BuilderBase{DoBuild: DoBuildData}}
var strategyMap = map[string]algo.IStrategy{
	"hot": &algo.StrategyBase{ DoSingle: DoHotBehaviorUpper },
	"user_behavior": &UserBehaviorStrategy{},
}
var sorterMap = map[string]algo.ISorter{
	"base": &algo.SorterBase{}}
var pagerMap = map[string]algo.IPager{
	"base": &algo.PagerBase{}}
var loggerMap = map[string]algo.ILogger{
	"features": &algo.LoggerBase{},
	"performs": &algo.LoggerPerforms{}}

var algosMap = algo.AlgoListInitToMap([]algo.IAlgo{
	&algo.AlgoBase{AlgoName: "model_base", FilePath: workDir + "mods_1.0.dumps.gz",
		Model: &utils.XgboostClassifier{}, FeaturesFunc: GetThemeFeatures },

})

// 话题推荐列表
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "theme", Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: algosMap,
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	StrategyKey: "strategies", StrategyDefault: "hot,user_behavior", StrategyMap: strategyMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	LoggerKey: "loggers", LoggerDefault: "features,performs", LoggerMap: loggerMap})
