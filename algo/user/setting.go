package user

import (
	"rela_recommend/algo"
	"rela_recommend/algo/base/strategy"
	// "rela_recommend/algo/utils"
)

var appName = "user"
var workDir = algo.GetWorkDir("/algo_files/user/")

var builderMap = map[string]algo.IBuilder{"base": &algo.BuilderBase{DoBuild: DoBuildData}}
var strategyMap = map[string]algo.IStrategy{}
var sorterMap = map[string]algo.ISorter{
	"base": &algo.SorterBase{}}
var pagerMap = map[string]algo.IPager{
	"base": &algo.PagerBase{}}
var loggerMap = map[string]algo.ILogger{
	"features": &algo.LoggerBase{},
	"performs": &algo.LoggerPerforms{}}

var richStrategyMap = map[string]algo.IRichStrategy {
		"paged": &strategy.PagedRichStrategy{},
	}

var algosMap = algo.AlgoListInitToMap([]algo.IAlgo{
	// &algo.AlgoBase{AlgoName: "base", FilePath: workDir + "mods_1.0.dumps.gz",
	// 	Model: &utils.XgboostClassifier{}, FeaturesFunc: GetFeaturesV0 },

})

// 话题推荐列表
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "user.near", Module: "user", Path: workDir,
	AlgoKey: "model", AlgoDefault: "base", AlgoMap: algosMap,
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap, 
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: richStrategyMap})
