package quick_match

import (
	"os"
)

var Work_dir string = ""
var MatchAlgoV1_0 *QuickMatchTreeV1_0
var MatchAlgosMap = map[string]IQuickMatch{}
func init() {
	Work_dir, _ = os.Getwd()
	Work_dir = Work_dir + "/algo_files"

	MatchAlgoV1_0 = &QuickMatchTreeV1_0{QuickMatchBase{FilePath: Work_dir + "/quick_match_tree.model", 
													  AlgoName: "QuickMatchTreeV1_0"}}
	modelList := [...]IQuickMatch{
		&QuickMatchTreeV1_0{QuickMatchBase{FilePath: Work_dir + "/quick_match_tree_v1.1.model",
										   AlgoName: "QuickMatchTreeV1_1"}},
		&QuickMatchTreeV1_0{QuickMatchBase{FilePath: Work_dir + "/quick_match_tree_v1.2.model", 
										   AlgoName: "QuickMatchTreeV1_2"}},
		&QuickMatchTreeV1_0{QuickMatchBase{FilePath: Work_dir + "/quick_match_tree_v1.3.model", 
										   AlgoName: "QuickMatchTreeV1_3"}},
		MatchAlgoV1_0}

	for index, _ := range modelList {
		modelList[index].Init()
		MatchAlgosMap[modelList[index].Name()] = modelList[index]
	}
}