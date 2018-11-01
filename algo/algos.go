package algo

import (
	"os"
	"rela_recommend/algo/quick_match"
)


work_dir := os.Getwd()

matchAlgo = quick_match.QuickMatchTree{FilePath=work_dir + "/algo_files/quick_match_tree.model"}
matchAlgo.Init()