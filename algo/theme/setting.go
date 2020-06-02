package theme

import (
	"rela_recommend/algo"
	"rela_recommend/algo/base/strategy"
	"rela_recommend/algo/utils"
)

var appName = "theme"
var workDir = algo.GetWorkDir("/algo_files/theme/")

var builderMap = map[string]algo.IBuilder{
	"base":      &algo.BuilderBase{DoBuild: DoBuildData},
	"maybelike": &algo.BuilderBase{DoBuild: DoBuildMayBeLikeData},
	"quick":     &algo.BuilderBase{DoBuild: DoBuildData},
}
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
		"behavior": &strategy.BaseBehaviorRichStrategy{
			UserStrategyItemFunc: UserBehaviorStrategyFunc,
			ItemStrategyItemFunc: ItemBehaviorStrategyFunc},
		"text_down": &strategy.BaseRichStrategy{ StrategyItemFunc: TextDownStrategyItem },
	}

var algosMap = algo.AlgoListInitToMap([]algo.IAlgo{
	&algo.AlgoBase{AlgoName: "model_base", FilePath: workDir + "mods_1.0.dumps.gz",
		Model: &utils.XgboostClassifier{}, FeaturesFunc: GetThemeFeatures },
	&algo.AlgoBase{AlgoName: "model_theme_v2.0", FilePath: workDir + "mods_2.0.dumps.gz",
		Model: &utils.XgboostClassifier{}, FeaturesFunc: GetThemeFeaturesv0 },
	&algo.AlgoBase{AlgoName: "model_theme_v2.1", FilePath: workDir + "mods_2.1.dumps.gz",
		Model: &utils.XgboostClassifier{}, FeaturesFunc: GetThemeFeaturesv0 },
})
var algosQuickMap = algo.AlgoListInitToMap([]algo.IAlgo{
	&algo.AlgoBase{AlgoName: "model_quick", FilePath: workDir + "mods_1.0.dumps.gz",
		Model: &utils.XgboostClassifier{}, FeaturesFunc: GetThemeFeatures },
	&algo.AlgoBase{AlgoName: "model_quick_v1.0", FilePath: workDir + "mods_quick_2.1.dumps.gz",
		Model: &utils.XgboostClassifier{}, FeaturesFunc: GetThemeQuickFeatures },
})

// 话题推荐列表
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "theme", Module: "theme", Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: algosMap,
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap, 
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: richStrategyMap})

// 新话题列表
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "theme.news", Module: "theme", Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: algosMap,
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap, 
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: richStrategyMap})

// 一周精选话题列表
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "theme.hotweek", Module: "theme", Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: algosMap,
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap, 
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: richStrategyMap})

// 相关话题
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "theme.maybelike", Module: "theme", Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: nil,
	BuilderKey: "build", BuilderDefault: "maybelike", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap, 
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: richStrategyMap})

// 话题快捷列表
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "theme.quick", Module: "theme", Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_quick", AlgoMap: algosQuickMap,
	BuilderKey: "build", BuilderDefault: "quick", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap,
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: richStrategyMap})