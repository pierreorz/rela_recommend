package live


import(
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
)

var workDir = algo.GetWorkDir("/algo_files/live/")

var builderMap = map[string]algo.IBuilder{
	"base": &algo.BuilderBase{DoBuild: DoBuildData},
}
var strategyMap = map[string]algo.IStrategy{
	"top_recommend_level": &algo.StrategyBase{ DoSingle: LiveTopRecommandStrategyFunc },
	"old_score": &OldScoreStrategy{},
}
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
	&algo.AlgoBase{AlgoName: "base", FilePath: workDir + "gbdtlr_6_200_v1.3.gz", 
				   Model: &utils.GradientBoostingLRClassifier{}, FeaturesFunc: GetLiveFeaturesV2 },
	&algo.AlgoBase{AlgoName: "xgb_1.0", FilePath: workDir + "xgb_1.0.gz", 
				   Model: &utils.XgboostClassifier{}, FeaturesFunc: GetLiveFeaturesV2 },
})

// 推荐栏目
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "live", Path: workDir,
	AlgoKey: "model", AlgoDefault: "base", AlgoMap: algosMap, 
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	StrategyKey: "strategies", StrategyDefault: "top_recommend_level,old_score", StrategyMap: strategyMap, 
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	LoggerKey: "loggers", LoggerDefault: "features,performs", LoggerMap: loggerMap})