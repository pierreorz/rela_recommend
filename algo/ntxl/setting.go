package ntxl

import (
	"rela_recommend/algo"
	"rela_recommend/algo/base/sort"
)

var workDir = algo.GetWorkDir("/algo_files/ntxl/")

var builderMap = map[string]algo.IBuilder{
	"base": &algo.BuilderBase{DoBuild: DoBuildData},
}
var strategyMap = map[string]algo.IStrategy{}
var sorterMap = map[string]algo.ISorter{
	"base":     &sort.SorterBase{},
	"interval": &sort.SorterWithInterval{},
}
var pagerMap = map[string]algo.IPager{
	"base": &algo.PagerBase{},
}
var loggerMap = map[string]algo.ILogger{
	"features": &algo.LoggerBase{},
	"performs": &algo.LoggerPerforms{},
}

var richStrategyMap = map[string]algo.IRichStrategy{}

var _ = algo.AlgoListInitToMap([]algo.IAlgo{})

// 女通讯录
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "ntxl", Module: "ntxl", Path: workDir,
	AlgoKey: "model", AlgoDefault: "base", AlgoMap: nil,
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap,
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: richStrategyMap})
