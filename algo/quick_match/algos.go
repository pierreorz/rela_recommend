package quick_match

import (
	"os"
)

var Work_dir = ""
var MatchAlgo QuickMatchTree
var MatchAlgoV1_1 QuickMatchTree
func init() {
	Work_dir, _ = os.Getwd()
	// v1.0
	model_file := Work_dir + "/algo_files/quick_match_tree.model"
	MatchAlgo = QuickMatchTree{FilePath: model_file}
	MatchAlgo.Init()
	// v1.1
	modelFileV1_1 := Work_dir + "/algo_files/quick_match_tree_v1.1.model"
	MatchAlgoV1_1 = QuickMatchTree{FilePath: modelFileV1_1}
	MatchAlgoV1_1.Init()
}
