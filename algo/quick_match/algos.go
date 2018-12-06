package quick_match

import (
	"os"
)

var Work_dir = ""
var MatchAlgo QuickMatchTree
var MatchAlgoV1_0 QuickMatchTreeV1_0
var MatchAlgoV1_1 QuickMatchTreeV1_1
func init() {
	Work_dir, _ = os.Getwd()
	// v1.0
	model_file := Work_dir + "/algo_files/quick_match_tree.model"
	MatchAlgo = QuickMatchTree{FilePath: model_file}
	MatchAlgo.Init()

	modelFileV1_0 := Work_dir + "/algo_files/quick_match_tree.model"
	MatchAlgoV1_0 = QuickMatchTreeV1_0{QuickMatchBase{FilePath: modelFileV1_0}}
	MatchAlgoV1_0.Init()
	// v1.1
	modelFileV1_1 := Work_dir + "/algo_files/quick_match_tree_v1.1.model"
	MatchAlgoV1_1 = QuickMatchTreeV1_1{QuickMatchBase{FilePath: modelFileV1_1}}
	MatchAlgoV1_1.Init()
}
