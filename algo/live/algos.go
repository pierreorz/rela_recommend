package live

import (
	"os"
)

var Work_dir string = ""
var LiveAlgosMap = map[string]ILiveAlgo{}
func init() {
	Work_dir, _ = os.Getwd()
	Work_dir = Work_dir + "/algo_files/live"

	modelList := [...]ILiveAlgo{
		// todo models
	}

	for index, _ := range modelList {
		modelList[index].Init()
		LiveAlgosMap[modelList[index].Name()] = modelList[index]
	}
}