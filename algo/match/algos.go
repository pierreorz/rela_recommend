package match

import (
	"os"
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
)

var Work_dir string = ""
var MatchAlgosMap = map[string]algo.IAlgo{}
func init() {
	Work_dir, _ = os.Getwd()
	Work_dir = Work_dir + "/algo_files/match/"

	modelList := [...]algo.IAlgo{
		&algo.AlgoBase{ AlgoName: "LiveModelV1_0", FilePath: Work_dir + "gbdtlr_6_200_v1.0.model", Model: &utils.GradientBoostingLRClassifier{} },
	}

	for index, _ := range modelList {
		modelList[index].Init()
		MatchAlgosMap[modelList[index].Name()] = modelList[index]
	}
}
