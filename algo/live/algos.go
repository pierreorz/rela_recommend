package live

import (
	"os"
)

var Work_dir string = ""
var LiveAlgosMap = map[string]ILiveAlgo{}
func init() {
	Work_dir, _ = os.Getwd()
	Work_dir = Work_dir + "/algo_files/live/"

	modelList := [...]ILiveAlgo{
		// todo models
		&LiveGbdtLrV0{LiveAlgoBase{ AlgoName: "LiveModelV1_0", FilePath: Work_dir + "gbdtlr_6_200_v1.0.model" }},
		&LiveGbdtLrV0{LiveAlgoBase{ AlgoName: "LiveModelV1_1", FilePath: Work_dir + "gbdtlr_6_200_v1.1.model" }},
		&LiveGbdtLrV0{LiveAlgoBase{ AlgoName: "LiveModelV1_2", FilePath: Work_dir + "gbdtlr_6_200_v1.2.model" }},
		&LiveGbdtLrV0{LiveAlgoBase{ AlgoName: "LiveModelV1_3", FilePath: Work_dir + "gbdtlr_6_200_v1.3.model" }},
	}

	for index, _ := range modelList {
		modelList[index].Init()
		LiveAlgosMap[modelList[index].Name()] = modelList[index]
	}
}
