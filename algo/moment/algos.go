package moment

import (
	"os"
)

var Work_dir string = ""
var AlgosMap = map[string]IMomentAlgo{}
func init() {
	Work_dir, _ = os.Getwd()
	Work_dir = Work_dir + "/algo_files/moment/"

	modelList := [...]IMomentAlgo{
		&MomentAlgoV0{MomentAlgoBase: MomentAlgoBase{AlgoName: "MomentModelV1_0", FilePath: Work_dir + "moment_xg_v1.0.model" }},
	}

	for index, _ := range modelList {
		modelList[index].Init()
		AlgosMap[modelList[index].Name()] = modelList[index]
	}
}
