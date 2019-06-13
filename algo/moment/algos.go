package moment

import (
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
)

var algosMap = map[string]algo.IAlgo{}
func init() {
	workDir := algo.GetWorkDir("/algo_files/moment/")
	// 精排算法
	modelList := []algo.IAlgo{
		&algo.AlgoBase{AlgoName: "model_base", FilePath: workDir + "moment_xg_v1.1.model", 
					   Model: &utils.XgboostClassifier{}, FeaturesFunc: GetMomentFeatures },
	}




	// 初始化app内容
	algosMap = algo.AlgoListInitToMap(modelList)
	appInfo := &algo.AppInfo{
		Name: "moment", Path: workDir,
		AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: algosMap, 
		StrategyKey: "strategies", StrategyDefault: "time_level", StrategyMap: strategyMap, 
		SorterKey: "sorter", SorterDefault: "base", SorterMap: sorterMap,
		PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
		LoggerKey: "loggers", LoggerDefault: "features,performs", LoggerMap: loggerMap}
	algo.SetAppInfo("moment", appInfo)
}
