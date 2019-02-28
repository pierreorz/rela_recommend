
package algo

import (
	"math"
	"fmt"
	// "reflect"
	"bytes"
)

//********************************* 服务端日志
type RecommendLog struct {
	RankId string  
	Index int64
	UserId int64
	DataId int64
	Algo string
	AlgoScore float32
	Score float32
	Features string
	AbMap string
}

//********************************* 特征
type Feature struct {
	Index int
	Value float32
}

func (feature *Feature) ToString() string {
	return fmt.Sprintf("%d:%g", feature.Index, feature.Value)
}

//********************************* 特征列表
// type Features struct {
// 	featuresMap map[int]float32
// }

// func (self *Features) checkInit() {
// 	if self.featuresMap == nil {
// 		self.featuresMap = make(map[int]float32)
// 	}
// }

// func (self *Features) ToString() string {
// 	self.checkInit()
// 	var buffer bytes.Buffer
// 	var i int = 0
// 	for key, val := range self.featuresMap {
// 		if i != 0 {
// 			buffer.WriteString(",")
// 		}
// 		str := fmt.Sprintf("%d:%f", key, val)
// 		buffer.WriteString(str)
// 		i++
// 	}
// 	return buffer.String()
// }

// func (self *Features) ToMap() map[int]float32 {
// 	self.checkInit()
// 	return self.featuresMap
// }

// func (self *Features) Add(key int, val float32) bool {
// 	self.checkInit()
// 	if key >= 0 && math.Abs(float64(val)) >= 0.000001 {
// 		self.featuresMap[key] = val
// 		return true
// 	}
// 	return false
// }

// func (self *Features) Get(key int) float32 {
// 	self.checkInit()
// 	if val, ok := self.featuresMap[key]; ok {
// 		return val
// 	}
// 	return 0.0
// }





func Features2String(features []Feature) string {
	var buffer bytes.Buffer
	for i, feature := range features {
		if i != 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(feature.ToString())
	}
	return buffer.String()
}

func FeaturesMap2String(features map[int]float32) string {
	var buffer bytes.Buffer
	var i int = 0
	for key, val := range features {
		if i > 0 {
			buffer.WriteString(",")
		}
		str := fmt.Sprintf("%d:%f", key, val)
		buffer.WriteString(str)

		i++
	}
	return buffer.String()
}

// list to Features
func List2Features(arr []float32) []Feature {
	fts := make([]Feature, 0)
	for i, v := range arr {
		if math.Abs(float64(v)) >= 0.00001 {
			fts = append(fts, Feature{i, v})
		}
	}
	return fts
}

// 模型接口
type IModel interface {
	Init(string)
	PredictSingle([]float32) float32
}

// 算法接口
type IAlgorithm interface {
	Name() string
	Features() []float32
	PredictSingle() float32
	Predict() []float32
	Init()
}

// 算法基础类
// type BaseAlgorithm struct {
// 	FilePath string
// }

// // 算法名称
// func (model *BaseAlgorithm) Name() string {
// 	return reflect.TypeOf(model).String()
// }

// // 计算一条纪录的特征
// func (model *BaseAlgorithm) Features() Features {
// 	features := Features{}
// 	return features
// }

// // 计算一条纪录
// func (model *BaseAlgorithm) PredictSingle(features Features) float32 {
// 	maps := features.ToMap()
// 	value, ok := maps[0]
// 	if !ok {
// 		return 0.0
// 	} else {
// 		return value
// 	}
// }

// // 计算多条纪录
// func (model *BaseAlgorithm) Predict(features []Features) []float32 {
// 	scores := make([]float32, len(features))
// 	for i, features := range features {
// 		scores[i] = model.PredictSingle(features)
// 	}
// 	return scores
// }
