package quick_match

import (
	"fmt"
	"os"
)

var Work_dir = ""
var MatchAlgo QuickMatchTree
func init() {
	Work_dir, _ = os.Getwd()
	model_file := Work_dir + "/algo_files/quick_match_tree.model"
	fmt.Println(model_file)
	MatchAlgo = QuickMatchTree{FilePath: model_file}
	MatchAlgo.Init()
}
