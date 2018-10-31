
package algorithm

import (
	"fmt"
	"reflect"
	"bytes"
)
// 特征
type Feature struct {
	Index int
	Value float32
}

func (feature *Feature) ToString() string {
	return fmt.Sprintf("%d:%f", feature.Index, feature.Value)
}

// 特征列表
type Features struct {
	Features []Feature
}

func (features *Features) ToString() string {
	var buffer bytes.Buffer
	for _, feature := range features.Features {
		buffer.WriteString(feature.ToString())
	}
	return buffer.String()
}

func (features *Features) ToMap() map[int]float32 {
	maps := map[int]float32{}
	for _, feature := range features.Features {
		maps[feature.Index] = feature.Value
	}
	return maps
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
type BaseAlgorithm struct {

}

// 算法名称
func (model *BaseAlgorithm) Name() string {
	return reflect.TypeOf(model).String()
}

// 计算一条纪录的特征
func (model *BaseAlgorithm) Features() Features {
	features := Features{}
	return features
}

// 计算一条纪录
func (model *BaseAlgorithm) PredictSingle(features Features) float32 {
	maps := features.ToMap()
	value, ok := maps[0]
	if !ok {
		return 0.0
	} else {
		return value
	}
}

// 计算多条纪录
func (model *BaseAlgorithm) Predict(features []Features) []float32 {
	scores := make([]float32, len(features))
	for i, features := range features {
		scores[i] = model.PredictSingle(features)
	}
	return scores
}
