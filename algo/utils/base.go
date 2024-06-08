package utils

import (
	"os"
	"time"
	"compress/gzip"
	"io/ioutil"
	"encoding/json"
	"reflect"
	"rela_recommend/log"
)

//********************************* 算法接口
type IModelAlgo interface {
	Init(string)
	PredictSingle(*Features) float32
	TransformSingle(*Features) *Features
}

//********************************* 算法基类
type ModelAlgoBase struct {
	FeaturesMap 	FeaturesMapEncoder		`json:"features_map"`
	Features		[]string				`json:"features"`
	Description		string					`json:"description"`	// 模型描述
}
func (self *ModelAlgoBase) PredictSingle(features *Features) float32 {
	return 0
}
func (self *ModelAlgoBase) TransformSingle(features *Features) *Features {
	features = self.FeaturesMap.Transform(features)
	return features
}

// 分割索引
func SplitIndexs(lens int, batch int) [][]int {
	arrs := make([][]int, batch)
	for i := 0; i < lens; i++ {
		index := i % batch
		if arrs[index] == nil {
			arrs[index] = make([]int, 0)
		}
		arrs[index] = append(arrs[index], i)
	}
	return arrs
}

// 模型加载 json -> gzip 
func LoadModel(path string, model interface{}) bool {
	var startTime = time.Now()
	fr, oerr := os.Open(path)
	name := reflect.TypeOf(model).String()
	defer fr.Close()
	if oerr != nil {
		log.Infof("%s:open tree file err, %s, %s", name, path, oerr.Error())
		return false
	}
	gzf, gerr := gzip.NewReader(fr)
	defer gzf.Close()
	if gerr != nil {
		log.Infof("%s:read gzip file err, %s, %s", name, path, gerr.Error())
		return false
	}
	data, rerr := ioutil.ReadAll(gzf)
	if rerr != nil {
		log.Infof("%s:read all err, %s, %s", name, path, rerr.Error())
		return false
	}
	jerr := json.Unmarshal(data, model)
	if jerr != nil {
		log.Infof("%s:load json err, %s, %s", path, name, jerr)
		return false
	}
	var endTime = time.Now()
	log.Infof("%s:init %.2f ok %s", name, endTime.Sub(startTime).Seconds(), path)
	return true
}