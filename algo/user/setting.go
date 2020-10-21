package user

import (
	"rela_recommend/algo"
	"rela_recommend/algo/base/sort"
	"rela_recommend/algo/base/strategy"
	"rela_recommend/algo/utils"
)

var appName = "user"
var workDir = algo.GetWorkDir("/algo_files/user/")

var builderMap = map[string]algo.IBuilder{
	"v1": &algo.BuilderBase{DoBuild: DoBuildDataV1},
}
var strategyMap = map[string]algo.IStrategy{}
var sorterMap = map[string]algo.ISorter{
	"base":   &sort.SorterBase{},
	"origin": &sort.SorterOrigin{}}
var pagerMap = map[string]algo.IPager{
	"base":   &algo.PagerBase{},
	"origin": &algo.PagerOrigin{}}
var loggerMap = map[string]algo.ILogger{
	"features": &algo.LoggerBase{},
	"performs": &algo.LoggerPerforms{}}

var richStrategyMap = map[string]algo.IRichStrategy{
	"paged": &strategy.PagedRichStrategy{},
	// 根据距离排序
	"distance_sort":   &strategy.BaseRichStrategy{StrategyItemFunc: SortWithDistanceItem},
	"wilson_behavior": &strategy.BaseRichStrategy{StrategyItemFunc: ItemBehaviorWilsonItemFunc},
	"clicked_down":    &strategy.BaseRichStrategy{StrategyItemFunc: UserBehaviorClickedDownItemFunc},
	"simple_upper":    &strategy.BaseRichStrategy{StrategyItemFunc: SimpleUpperItemFunc, DefaultWeight: 2},
}

var algosMap = algo.AlgoListInitToMap([]algo.IAlgo{
	&algo.AlgoBase{AlgoName: "base", FilePath: workDir + "nearby/mods_1.0.model.gz",
		Model: &utils.XgboostClassifier{}, FeaturesFunc: GetFeaturesV0},
	&algo.AlgoBase{AlgoName: "v1.1", FilePath: workDir + "nearby/mods_1.1.model.gz",
		Model: &utils.XgboostClassifier{}, FeaturesFunc: GetFeaturesV0},
	&algo.AlgoBase{AlgoName: "v1.2", FilePath: workDir + "nearby/mods_1.2_2.model.gz",
		Model: &utils.XgboostClassifier{}, FeaturesFunc: GetFeaturesV0},
	&algo.AlgoBase{AlgoName: "v1.3", FilePath: workDir + "nearby/mods_1.3_2.model.gz",
		Model: &utils.XgboostClassifier{}, FeaturesFunc: GetFeaturesV0},
})

// 附近的人
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "user.nearby", Module: "user", Path: workDir,
	AlgoKey: "model", AlgoDefault: "base", AlgoMap: algosMap,
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap,
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: richStrategyMap})

// ***************************************************************** 用户搜索
var searchBuilderMap = map[string]algo.IBuilder{
	"base": &algo.BuilderBase{DoBuild: DoBuildSearchData},
}
var searchStrategyMap = map[string]algo.IStrategy{}
var searchRichStrategyMap = map[string]algo.IRichStrategy{}

// 用户搜索
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "user.search", Module: "user", Path: workDir,
	AlgoKey: "model", AlgoDefault: "base", AlgoMap: nil,
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: searchBuilderMap,
	SorterKey: "sorter", SorterDefault: "origin", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "origin", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: searchStrategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap,
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: searchRichStrategyMap})
