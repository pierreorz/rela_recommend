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

	modelList := [...]algo.IAlgo{}

	for index, _ := range modelList {
		modelList[index].Init()
		MatchAlgosMap[modelList[index].Name()] = modelList[index]
	}
}
