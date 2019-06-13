package moment

import (
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
)

var algosCoarseMap = map[string]algo.IAlgo{}
func init() {
	workDir := algo.GetWorkDir("/algo_files/moment/coarse/")

	// 粗排算法
	modelCoarseList := []algo.IAlgo{
		&algo.AlgoBase{AlgoName: "model_base", FilePath: workDir + "moment_coarse_xg_v1.0.model",
						Model: &utils.XgboostClassifier{}, FeaturesFunc: GetMomentFeatures },
	}
	algosCoarseMap = algo.AlgoListInitToMap(modelCoarseList)
	
	appInfo := &algo.AppInfo{
		Name: "moment_coarse", Path: workDir,
		AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: algosCoarseMap,
		StrategyKey: "strategies", StrategyDefault: "", StrategyMap: nil,
		SorterKey: "sorter", SorterDefault: "base", SorterMap: nil,
		PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
		LoggerKey: "loggers", LoggerDefault: "features,performs", LoggerMap: loggerMap}
	algo.SetAppInfo("moment_coarse", appInfo)
}
