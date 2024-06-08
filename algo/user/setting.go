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
	"base":     &sort.SorterBase{},
	"origin":   &sort.SorterOrigin{},
	"interval": &sort.SorterWithInterval{},
}
var pagerMap = map[string]algo.IPager{
	"base":   &algo.PagerBase{},
	"origin": &algo.PagerOrigin{}}
var loggerMap = map[string]algo.ILogger{
	"features": &algo.LoggerBase{},
	"performs": &algo.LoggerPerforms{},
	"seen":     &DoNearbySeenSearchLogger{},
}

var richStrategyMap = map[string]algo.IRichStrategy{
	"paged": &strategy.PagedRichStrategy{},
	// 根据距离排序
	"distance_sort":        &strategy.BaseRichStrategy{StrategyItemFunc: SortWithDistanceItem},
	"wilson_behavior":      &strategy.BaseRichStrategy{StrategyItemFunc: ItemBehaviorWilsonItemFunc},
	"clicked_down":         &strategy.BaseRichStrategy{StrategyItemFunc: BehaviorClickedDownItemFunc},
	"simple_upper":         &strategy.BaseRichStrategy{StrategyItemFunc: SimpleUpperItemFunc, DefaultWeight: 2},
	"exposure_increase":    &strategy.BaseRichStrategy{StrategyFunc: strategy.ExposureIncreaseFunc, DefaultWeight: 3},
	"no_interact_decrease": &strategy.BaseRichStrategy{StrategyFunc: strategy.NoInteractDecreaseFunc, DefaultWeight: 3},
	"exposure_bottom":      &strategy.BaseRichStrategy{StrategyFunc: strategy.ExposureBottomFunc},
	"week_no_interact":     &strategy.BaseRichStrategy{StrategyFunc: WeekExposureNoInteractFunc},
	"single_user_exposure": &strategy.BaseRichStrategy{StrategyItemFunc: ExpoTooMuchDownItemFunc},
	"cover_face":           &strategy.BaseRichStrategy{StrategyItemFunc: CoverFaceUpperItem},
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
var searchRichStrategyMap = map[string]algo.IRichStrategy{
	"on_live": &strategy.BaseRichStrategy{StrategyItemFunc: NtxOnLiveWeightFunc, DefaultWeight: 2},
}

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

var ntxlBuilderMap = map[string]algo.IBuilder{
	"base": &algo.BuilderBase{DoBuild: DoBuildNtxlData},
}
var ntxlStrategyMap = map[string]algo.IStrategy{}
var ntxlRichStrategyMap = map[string]algo.IRichStrategy{
	"paged":           &strategy.PagedRichStrategy{DefaultWeight: 1},
	"on_live":         &strategy.BaseRichStrategy{StrategyItemFunc: NtxOnLiveWeightFunc, DefaultWeight: 2},
	"nearby":          &strategy.BaseRichStrategy{StrategyItemFunc: NtxNearbyDecayWeightFunc, DefaultWeight: 2},
	"active":          &strategy.BaseRichStrategy{StrategyItemFunc: NtxlActiveDecayWeightFunc, DefaultWeight: 1},
	"user_visit":      &strategy.BaseRichStrategy{StrategyItemFunc: NtxlUserPageViewItemFunc, DefaultWeight: 3},
	"user_wink":       &strategy.BaseRichStrategy{StrategyItemFunc: NtxlUserWinkItemFunc, DefaultWeight: 3},
	"moment_interact": &strategy.BaseRichStrategy{StrategyItemFunc: NtxlMomentInteractItemFunc, DefaultWeight: 3},
	"send_message":    &strategy.BaseRichStrategy{StrategyItemFunc: NtxlSendMessageItemFunc, DefaultWeight: 3},
	"not_single":      &strategy.BaseRichStrategy{StrategyItemFunc: NtxNotSingleDecayFunc, DefaultWeight: 1},
}

// 女通讯录
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "user.ntxl", Module: "user", Path: workDir,
	AlgoKey: "model", AlgoDefault: "base", AlgoMap: nil,
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: ntxlBuilderMap,
	SorterKey: "sorter", SorterDefault: "origin", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "origin", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: ntxlStrategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap,
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: ntxlRichStrategyMap})
