package moment_tag

import (
	"rela_recommend/algo"
	"rela_recommend/algo/base/sort"
)

var appName = "moment_tag"
var workDir = algo.GetWorkDir("/algo_files/moment_tag/")

var searchBuilderMap = map[string]algo.IBuilder{
	"base": &algo.BuilderBase{DoBuild: DoBuildData},
}
var strategyMap = map[string]algo.IStrategy{}
var sorterMap = map[string]algo.ISorter{
	"base":     &sort.SorterBase{},
	"origin":   &sort.SorterOrigin{},
	"interval": &sort.SorterWithInterval{},
}
var pagerMap = map[string]algo.IPager{
	"base":   &algo.PagerBase{},
	"origin": &algo.PagerOrigin{}}
var loggerMap = map[string]algo.ILogger{
	"features": &algo.LoggerBase{},
	"performs": &algo.LoggerPerforms{}}

var searchStrategyMap = map[string]algo.IStrategy{}
var searchRichStrategyMap = map[string]algo.IRichStrategy{}

// 用户搜索
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "moment_tag.search", Module: "moment_tag", Path: workDir,
	AlgoKey: "model", AlgoDefault: "base", AlgoMap: nil,
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: searchBuilderMap,
	SorterKey: "sorter", SorterDefault: "origin", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "origin", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: searchStrategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap,
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: searchRichStrategyMap})
