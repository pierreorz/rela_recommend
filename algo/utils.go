package algo
import (
	"os"
)

// 获取当前工作目录
func GetWorkDir(path string) string {
	work_dir, _ := os.Getwd()
	return work_dir + path
}

// 将算法列表初始化，并且生成到指定Map
func AlgoListInitToMap(algoList []IAlgo) map[string]IAlgo {
	algoMap := map[string]IAlgo{}
	for index, _ := range algoList {
		algoList[index].Init()
		algoMap[algoList[index].Name()] = algoList[index]
	}
	return algoMap
}