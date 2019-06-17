
package algo

import (
	"math"
	"fmt"
	// "reflect"
	"bytes"
	"rela_recommend/algo/utils"
)

type AppInfo struct {
	// app名称，用于标识app，读取app的abtest
	Name string
	Path string
	// 算法的abtest key
	AlgoKey string
	AlgoDefault string
	AlgoMap map[string]IAlgo
	// 构造数据的abtest key
	BuilderKey string
	BuilderDefault string
	BuilderMap map[string]IBuilder
	// 策略的abtest key
	StrategyKey string
	StrategyDefault string
	StrategyMap map[string]IStrategy
	// 排序的abtest key
	SorterKey string
	SorterDefault string
	SorterMap map[string]ISorter
	// 分页的abtest key
	PagerKey string
	PagerDefault string
	PagerMap map[string]IPager
	// 日志的abtest key
	LoggerKey string
	LoggerDefault string
	LoggerMap map[string]ILogger
}

//********************************* 服务端日志
type RecommendLog struct {
	Module string
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

// 请求参数
type RecommendRequest struct {
	App		string 				`json:"app" form:"app"`
	Limit   int64  				`json:"limit" form:"limit"`
	Offset  int64  				`json:"offset" form:"offset"`
	Ua      string 				`json:"ua" form:"ua"`
	Lat		float32 			`json:"lat" form:"lat"`
	Lng		float32 			`json:"lng" form:"lng"`
	UserId  int64  				`json:"userId" form:"userId"`
	Type	string				`json:"type" form:"type"`	// 是推荐/热门/
	DataIds []int64				`json:"dataIds" form:"dataIds"`
	AbMap	map[string]string	`json:"abMap" form:"abMap"`
}

type RecommendResponseItem struct {
	DataId int64	`json:"dataId" form:"dataId"`
	Index 	int		`json:"index" form:"index"`
	Reason string	`json:"reason" form:"reason"`
	Score float32	`json:"score" form:"score"`
}

// 返回参数
type RecommendResponse struct {
	Status  string		`json:"status" form:"status"`
	Message string		`json:"message" form:"message"`
	RankId	string		`json:"rankId" form:"rankId"`
	DataIds []int64		`json:"dataIds" form:"dataIds"`
	DataList []RecommendResponseItem	`json:"dataList" form:"dataList"`
}

type RecommendItem struct {
	Reason 		string				// 推荐理由
	Score		float32				// 推荐分数
}

type RankInfo struct {
	Features 	*utils.Features			// 特征
	IsTop		int 					// 1: 置顶， 0: 默认， -1:置底
	Level		int						// 推荐优先级
	Recommends	[]RecommendItem	// 推荐系数
	Punish		float32					// 惩罚系数
	AlgoName	string					// 算法名称
	AlgoScore 	float32					// 算法得分
	Score 		float32					// 最终得分
	Index 		int						// 排在第几
	Reason		string					// 推荐理由
}

// 获取Features的字符串形式：1:1.0,1000:1.0,99:1.0
func (self *RankInfo) GetFeaturesString() string {
	if self.Features == nil {
		return ""
	} else {
		return self.Features.ToString()
	}
}

// 增加推荐理由，以,隔开：TOP,RECOMMEND
func (self *RankInfo) AddReason(reason string) string {
	if len(self.Reason) > 0 {
		self.Reason = self.Reason + "," + reason
	} else {
		self.Reason = reason
	}
	return self.Reason
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
