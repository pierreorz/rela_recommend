package moment

import(
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/algo/base/strategy"
)

var workDir = algo.GetWorkDir("/algo_files/moment/")

var builderMap = map[string]algo.IBuilder{
	"base": &algo.BuilderBase{DoBuild: DoBuildData},
	"arounddetail":&algo.BuilderBase{DoBuild:DoBuildMomentAroundDetailSimData},
	"followdetail":&algo.BuilderBase{DoBuild:DoBuildMomentFriendDetailSimData},
	"recdetail":&algo.BuilderBase{DoBuild:DoBuildMomentRecommendDetailSimData}}
var strategyMap = map[string]algo.IStrategy{
	"time_level": &algo.StrategyBase{ DoSingle: DoTimeLevel },
	"time_weight": &algo.StrategyBase{ DoSingle: DoTimeWeightLevel },
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
var richStrategyMap = map[string]algo.IRichStrategy {
	"paged": &strategy.PagedRichStrategy{},
	"behavior": &strategy.BaseBehaviorRichStrategy{
		UserStrategyItemFunc: UserBehaviorStrategyFunc,
		ItemStrategyItemFunc: ItemBehaviorStrategyFunc},
	"detail_rec":&strategy.BaseBehaviorRichStrategy{
		UserStrategyItemFunc: DetailRecommendStrategyFunc,
		ItemStrategyItemFunc: ItemBehaviorStrategyFunc},
}


// 精排算法
var algosMap = algo.AlgoListInitToMap([]algo.IAlgo{
	&algo.AlgoBase{AlgoName: "model_base", FilePath: workDir + "moment_xg_v1.1.model", 
				   Model: &utils.XgboostClassifier{}, FeaturesFunc: GetMomentFeatures },
	&algo.AlgoBase{AlgoName: "model_v2", FilePath: workDir + "mods_1.2.dumps.gz",
		Model: &utils.XgboostClassifier{}, FeaturesFunc: GetMomentFeatures },
	&algo.AlgoBase{AlgoName: "model_around", FilePath: workDir + "around_moments.dumps.gz",
		Model: &utils.GradientBoostingLRClassifier{}, FeaturesFunc: GetMomentFeatures },
	&algo.AlgoBase{AlgoName: "model_embedding", FilePath: workDir + "mods_2.0.dumps.gz",
		Model: &utils.XgboostClassifier{}, FeaturesFunc: GetMomentFeatures },
	&algo.AlgoBase{AlgoName: "model_momemb", FilePath: workDir + "mods_2.2.dumps.gz",
		Model: &utils.XgboostClassifier{}, FeaturesFunc: GetMomentFeatures },
})


// 推荐日志
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "moment", Module: "moment", Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: algosMap, 
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap, 
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: richStrategyMap})

// 日志附近的人
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "moment.near", Module: "moment", Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: algosMap, 
	BuilderKey: "build", BuilderDefault: "base", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap, 
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: richStrategyMap})

//推荐日志详情页
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "moment.recdetail", Module: "moment", Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: nil,
	BuilderKey: "build", BuilderDefault: "recdetail", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap,
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: richStrategyMap})

//附近日志详情页
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "moment.arounddetail", Module: "moment", Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_embedding", AlgoMap: algosMap,
	BuilderKey: "build", BuilderDefault: "arounddetail", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap,
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: richStrategyMap})

//关注日志详情页
var _ = algo.AddAppInfo(&algo.AppInfo{
	Name: "moment.followdetail", Module: "moment", Path: workDir,
	AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: nil,
	BuilderKey: "build", BuilderDefault: "followdetail", BuilderMap: builderMap,
	SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
	PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
	StrategyKeyFormatter: "strategy:%s:weight", StrategyMap: strategyMap,
	LoggerKeyFormatter: "logger:%s:weight", LoggerMap: loggerMap,
	RichStrategyKeyFormatter: "rich_strategy:%s:weight", RichStrategyMap: richStrategyMap})