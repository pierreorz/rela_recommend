package moment

import (
	"os"
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
)

var Work_dir string = ""
var AlgosMap = map[string]IMomentAlgo{}
var AlgosMap2 = map[string]algo.IAlgo{}
var AlgosCoarseMap = map[string]IMomentAlgo{}
func init() {
	Work_dir, _ = os.Getwd()
	Work_dir = Work_dir + "/algo_files/moment/"

	// 精排算法
	modelList := [...]IMomentAlgo{
		&MomentAlgoV0{MomentAlgoBase: MomentAlgoBase{AlgoName: "MomentModelV1_0", FilePath: Work_dir + "moment_xg_v1.0.model" }},
	}

	for index, _ := range modelList {
		modelList[index].Init()
		AlgosMap[modelList[index].Name()] = modelList[index]
	}




	// 精排算法
	modelList2 := [...]algo.IAlgo{
		&algo.AlgoBase{AlgoName: "model_base", FilePath: Work_dir + "moment_xg_v1.0.model", 
					   Model: &utils.XgboostClassifier{}, FeaturesFunc: GetMomentFeatures2 },
	}

	for index, _ := range modelList2 {
		modelList2[index].Init()
		AlgosMap2[modelList2[index].Name()] = modelList2[index]
	}















	// 粗排算法
	workDir := Work_dir + "coarse/"
	modelCoarseList := [...]IMomentAlgo{
		&MomentAlgoCoarse{MomentAlgoBase: MomentAlgoBase{AlgoName: "MomentCoarseModelV1_0", FilePath: workDir + "moment_coarse_xg_v1.0.model" }},
	}
	for index, _ := range modelCoarseList {
		modelCoarseList[index].Init()
		AlgosCoarseMap[modelCoarseList[index].Name()] = modelCoarseList[index]
	}
}
