package live

import (
	"rela_recommend/algo"
	"rela_recommend/algo/base/sort"
	"rela_recommend/algo/base/strategy"
	"rela_recommend/algo/utils"
)

var workDir = algo.GetWorkDir("/algo_files/live/")

var builderMap = map[string]algo.IBuilder{
	"base": &algo.BuilderBase{DoBuild: DoBuildData},
}
var strategyMap = map[string]algo.IStrategy{
	"top_recommend_level": &algo.StrategyBase{DoSingle: LiveTopRecommandStrategyFunc},
	"old_score":           &OldScoreStrategy{},
	"algo_v2":                &NewScoreStrategyV2{},
	"algo_score":         &NewLiveStrategy{},
}
var richStrategyMap = map[string]algo.IRichStrategy{
	"per_hour_top": &strategy.BaseRichStrategy{StrategyFunc: HourRankRecommendFunc, DefaultWeight: 1}, // 执行优先级在top_recommend之后，避免覆盖
	"exposure_down": &strategy.BaseRichStrategy{StrategyItemFunc: UserBehaviorExposureDownItemFunc},
	"interest": &strategy.BaseRichStrategy{StrategyFunc: StrategyRecommendFunc},
	"live_add_exposure": &strategy.BaseRichStrategy{StrategyFunc: LiveExposureFunc,DefaultWeight:2},

}

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

// 精排算法
var algosMap = algo.AlgoListInitToMap([]algo.IAlgo{
	&algo.AlgoBase{AlgoName: "base", FilePath: workDir + "gbdtlr_6_200_v1.3.gz",
		Model: &utils.GradientBoostingLRClassifier{}, FeaturesFunc: GetLiveFeaturesV2},
	&algo.AlgoBase{AlgoName: "xgb_1.0", FilePath: workDir + "xgb_1.0.gz",
		Model: &utils.XgboostClassifier{}, FeaturesFunc: GetLiveFeaturesV2},
})

// 推荐栏目
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "live", Module: "live", Path: workDir,
	AlgoKey: "model", AlgoDefault: "base", AlgoMap: algosMap,
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap,
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: richStrategyMap,
})
