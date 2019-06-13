package theme

import (
	"rela_recommend/algo"
)

var algosMap = map[string]algo.IAlgo{}
func init() {
	workDir := algo.GetWorkDir("/algo_files/theme/")
	// 精排算法
	modelList := []algo.IAlgo{ }




	// 初始化app内容
	algosMap = algo.AlgoListInitToMap(modelList)
	appInfo := &algo.AppInfo{
		Name: "theme", Path: workDir,
		AlgoKey: "model", AlgoDefault: "model_base", AlgoMap: nil,
		StrategyKey: "strategies", StrategyDefault: "time_level", StrategyMap: nil,
		SorterKey: "sorter", SorterDefault: "base", SorterMap: nil,
		PagerKey: "pager", PagerDefault: "base", PagerMap: pagerMap,
		LoggerKey: "loggers", LoggerDefault: "features,performs", LoggerMap: loggerMap}
	algo.SetAppInfo("theme", appInfo)
}
