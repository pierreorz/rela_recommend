package theme

import (
	"rela_recommend/algo"
)

var appName = "theme"
var workDir = algo.GetWorkDir("/algo_files/theme/")

var builderMap = map[string]algo.IBuilder{"base": &algo.BuilderBase{DoBuild: DoBuildData}}
var strategyMap = map[string]algo.IStrategy{}
var sorterMap = map[string]algo.ISorter{
	"base": &algo.SorterBase{}}
var pagerMap = map[string]algo.IPager{
	"base": &algo.PagerBase{}}
var loggerMap = map[string]algo.ILogger{
	"features": &algo.LoggerBase{},
	"performs": &algo.LoggerPerforms{}}

var algosMap = algo.AlgoListInitToMap([]algo.IAlgo{ 

})
var appInfo = &algo.AppInfo{
	Name: appName, Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: nil,
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	StrategyKey: "strategies", StrategyDefault: "", StrategyMap: nil,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: nil,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	LoggerKey: "loggers", LoggerDefault: "features,performs", LoggerMap: loggerMap}
var _ = algo.AddAppInfo(appInfo)
