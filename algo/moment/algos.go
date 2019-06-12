package moment

import (
	"os"
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
)

var Work_dir string = ""
var AlgosMap = map[string]algo.IAlgo{}
var AlgosCoarseMap = map[string]algo.IAlgo{}
func init() {
	Work_dir, _ = os.Getwd()
	Work_dir = Work_dir + "/algo_files/moment/"

	// 精排算法
	modelList := [...]algo.IAlgo{
		&algo.AlgoBase{AlgoName: "model_base", FilePath: Work_dir + "moment_xg_v1.1.model", 
					   Model: &utils.XgboostClassifier{}, FeaturesFunc: GetMomentFeatures },
	}

	for index, _ := range modelList {
		modelList[index].Init()
		AlgosMap[modelList[index].Name()] = modelList[index]
	}















	// 粗排算法
	workDir := Work_dir + "coarse/"
	modelCoarseList := [...]algo.IAlgo{
		&algo.AlgoBase{AlgoName: "model_base", FilePath: workDir + "moment_coarse_xg_v1.0.model",
						Model: &utils.XgboostClassifier{}, FeaturesFunc: GetMomentFeatures },
	}

	for index, _ := range modelCoarseList {
		modelCoarseList[index].Init()
		AlgosCoarseMap[modelCoarseList[index].Name()] = modelCoarseList[index]
	}
}
