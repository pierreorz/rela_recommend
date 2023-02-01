package algo

import (
	"bytes"
	"fmt"
	"math"
	"rela_recommend/algo/utils"
	rutils "rela_recommend/utils"
	"sort"
)

const (
	TypeYouFollow =iota
	TypeNearby
	TypeHot
	TypeYouMayLike
	TypeEmpty  // 推荐理由等级，默认情况下值越小越优先展示

)

type ReasonType int8

var allReasonTypes = map[ReasonType]clientReason{
	TypeEmpty: {
		Type: TypeEmpty,
		Text: nil,
	},
	TypeYouMayLike: {
		Type: TypeYouMayLike,
		Text: &MultiLanguage{
			Chs: "你可能感兴趣",

		},
	},
	TypeHot: {
		Type: TypeHot,
		Text: &MultiLanguage{
			Chs: "超多点赞",
			Cht: "超多按贊",
			En: "From trending posts",
		},
	},
	TypeYouFollow: {
		Type: TypeYouFollow,
		Text: &MultiLanguage{
			Chs: "你的关注",
			Cht: "你的關注",
			En: "From user you're following",
		},
	},
	TypeNearby: {
		Type: TypeNearby,
		Text: &MultiLanguage{
			Chs: "在你附近",
			Cht: "在你附近",
			En: "From nearby posts",
		},
	},
}

type AppInfo struct {
	// app名称，用于标识app，读取app的abtest
	Name string
	// module, 属于某个模块
	Module string
	Path   string
	// 算法的abtest key
	AlgoKey     string
	AlgoDefault string
	AlgoMap     map[string]IAlgo
	// 构造数据的abtest key
	BuilderKey     string
	BuilderDefault string
	BuilderMap     map[string]IBuilder
	// 排序的abtest key
	SorterKey     string
	SorterDefault string
	SorterMap     map[string]ISorter
	// 分页的abtest key
	PagerKey     string
	PagerDefault string
	PagerMap     map[string]IPager
	// 策略的abtest key
	StrategyKeyFormatter string
	StrategyMap          map[string]IStrategy
	// 日志的abtest key
	LoggerKeyFormatter string
	LoggerMap          map[string]ILogger
	// 富策略，同时包含加载数据/执行策略/记录内容
	RichStrategyKeyFormatter string
	RichStrategyMap          map[string]IRichStrategy
}

// 请求参数
// swagger:model order
type RecommendRequest struct {
	// 功能场景名
	//
	// required: true
	// enum: moment,theme,user,live,match
	App string `json:"app" form:"app"`

	Addr string `json:"addr" form:"addr"`
	// 子功能名
	//
	// required: false
	// enum: nearby,reply,detail_reply
	Type string `json:"type" form:"type"` // 是推荐/热门/
	// 分页每页数量
	//
	// required: false
	// example: 10
	Limit int64 `json:"limit" form:"limit"`
	// 分页起始位置
	//
	// required: false
	// example: 0
	Offset int64 `json:"offset" form:"offset"`
	// 浏览器UA
	//
	// required: false
	Ua string `json:"ua" form:"ua"`
	// 手机系统
	//
	// required: false
	MobileOS string `json:"mobileOS" form:"mobileOS"`
	// 客户端版本
	//
	// required: false
	// example: 050303
	ClientVersion int `json:"clientVersion" form:"clientVersion"`
	// 经度
	//
	// required: false
	// example: 33.0
	Lat float32 `json:"lat" form:"lat"`
	// 纬度
	//
	// required: false
	// example: 121.0
	Lng float32 `json:"lng" form:"lng"`
	// 用户ID
	//
	// required: false
	// example: 3567
	UserId  int64   `json:"userId" form:"userId"`
	DataIds []int64 `json:"dataIds" form:"dataIds"`
	// AB配置信息
	//
	// required: false
	AbMap map[string]string `json:"abMap" form:"abMap"`
	// 其他参数，比如搜索筛选项
	//
	// required: false
	Params map[string]string `json:"params" form:"params"`

	// 内部缓存变量
	osName  string
	version int
}

func (self *RecommendRequest) GetOS() string {
	if self.osName == "" {
		if os := rutils.GetPlatformName(self.MobileOS); os == "other" || os == "" {
			self.osName = rutils.GetPlatformName(self.Ua)
		} else {
			self.osName = os
		}
	}
	return self.osName
}

func (self *RecommendRequest) GetUa() string {
	if self.osName == "" {
		if os := rutils.GetPlatformName(self.Ua); os == "other" || os == "" {
			self.osName = rutils.GetPlatformName(self.Ua)
		} else {
			self.osName = os
		}
	}
	return self.osName
}

func (self *RecommendRequest) GetVersion() int {
	if self.version == 0 {
		if self.ClientVersion > 0 {
			self.version = self.ClientVersion
		} else {
			self.version = rutils.GetVersion(self.Ua)
		}
	}
	return self.version
}

type RecommendResponseItem struct {
	DataId         int64          `json:"dataId" form:"dataId"`
	Data           interface{}    `json:"data" form:"data"`
	Index          int            `json:"index" form:"index"`
	Reason         string         `json:"reason" form:"reason"`
	ReasonMultiple *MultiLanguage `json:"reason_multiple"`
	Score          float32        `json:"score" form:"score"`
}

// 返回参数
// swagger:model recommendResponseInner
type RecommendResponse struct {
	Status   string                  `json:"status" form:"status"`
	Message  string                  `json:"message" form:"message"`
	RankId   string                  `json:"rankId" form:"rankId"`
	DataIds  []int64                 `json:"dataIds" form:"dataIds"`
	DataList []RecommendResponseItem `json:"dataList" form:"dataList"`
}

type RecommendItem struct {
	Reason       string     // 推荐理由
	Score        float32    // 推荐分数
	NeedReturn   bool       // 是否返回给前端
	ClientReason ReasonType // 客户端显示的推荐理由
}

type RankInfo struct {
	Features    *utils.Features // 特征
	IsTop       int             // 1: 置顶， 0: 默认， -1:置底
	PagedIndex  int             // 分页展示过的index
	Level       int             // 推荐优先级
	Recommends  []RecommendItem // 推荐系数
	Punish      float32         // 惩罚系数
	AlgoName    string          // 算法名称
	AlgoScore   float32         // 算法得分
	PaiScore    float64         //Pai 算法得分
	Score       float32         // 最终得分
	Index       int             // 排在第几
	LiveIndex   int             //热门直播日志的排序
	TopLive     int             //是否是头部主播的直播日志
	HopeIndex   int             // 期望排在第几，排序结束后调整
	IsBussiness int             //是否是业务日志（用户关注日志、点击头像多次未看过日志）
	IsBlindMom  int             //是否是多人语音相遇日志
	IsTagMom    int
	IsSoftTop   int             //是否软置顶日志   1:是  0：默认
	ExpId       string          //Pai实验Id
	RequestId   string          //Pai请求id
	OffTime     int             //超时标记位
	IsHourTop   int				//小时top3
}

type MultiLanguage struct {
	Chs string `json:"chs"`
	Cht string `json:"cht"`
	En  string `json:"en"`
}

type clientReason struct {
	Type int8           `json:"type"`
	Text *MultiLanguage `json:"text"`
}

// 获取Features的字符串形式：1:1.0,1000:1.0,99:1.0
func (self *RankInfo) GetFeaturesString() string {
	if self.Features == nil {
		return ""
	} else {
		return self.Features.ToString()
	}
}

func (self *RankInfo) AddRecommend(reason string, score float32) {
	item := RecommendItem{Reason: reason, Score: score, NeedReturn: false,ClientReason:TypeEmpty}
	self.Recommends = append(self.Recommends, item)
}

func (self *RankInfo) AddRecommendWithType(reason string, score float32, reasonType ReasonType) {
	item := RecommendItem{Reason: reason, Score: score, NeedReturn: true, ClientReason: reasonType}
	self.Recommends = append(self.Recommends, item)
}

func (self *RankInfo) AddRecommendNeedReturn(reason string, score float32) {
	item := RecommendItem{Reason: reason, Score: score, NeedReturn: true}
	self.Recommends = append(self.Recommends, item)
}

func (self *RankInfo) ClientReasonString() *MultiLanguage {
	var reason *clientReason
	for _, rd := range self.Recommends {
		current, ok := allReasonTypes[rd.ClientReason]
		if ok {
			if reason == nil {
				reason = &current
			}
			if reason.Type > current.Type {
				reason = &current
			}
		}
	}

	if reason == nil {
		return nil
	}

	return reason.Text
}

// 增加推荐理由，以,隔开：TOP,RECOMMEND
func (self *RankInfo) ReasonString() string {
	return self.getRecommendsString(false, func(reason string, score float32) string {
		return fmt.Sprintf(",%s", reason)
	})
}

// 将推荐理由转化为字符串
func (self *RankInfo) RecommendsString() string {
	return self.getRecommendsString(true, func(reason string, score float32) string {
		return fmt.Sprintf(",%s:%g", reason, score)
	})
}

// 将推荐理由转化为字符串, returnAll: 是否返回所有，false只返回客户端需要的内容
func (self *RankInfo) getRecommendsString(returnAll bool, f func(string, float32) string) string {
	var buffer bytes.Buffer
	if self.IsTop > 0 && self.IsSoftTop != 1 { //置顶且非软置顶
		buffer.WriteString(f("TOP", 1))
	} else if self.IsTop < 0 {
		buffer.WriteString(f("BOTTOM", 1))
	}

	//if self.HopeIndex > 0 {
	//	buffer.WriteString(f("HOPE", float32(self.HopeIndex)))
	//}

	if self.Level > 0 {
		buffer.WriteString(f("LEVEL", float32(self.Level)))
	}

	for _, recommend := range self.Recommends {
		if returnAll || recommend.NeedReturn {
			buffer.WriteString(f(recommend.Reason, recommend.Score))
		}
	}
	res := buffer.String()
	return res[rutils.GetInt(len(res) > 0):]
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
	var i = 0
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

// 权重进行升序排序
type KeyWeight struct {
	Key    string
	Value  interface{}
	Weight int
}

type KeyWeightSorter struct {
	list   []KeyWeight
	sorted bool
}

func (self *KeyWeightSorter) Swap(i, j int) {
	self.list[i], self.list[j] = self.list[j], self.list[i]
}
func (self *KeyWeightSorter) Len() int { return len(self.list) }
func (self *KeyWeightSorter) Less(i, j int) bool { // 权重正序
	if self.list[i].Weight != self.list[j].Weight {
		return self.list[i].Weight < self.list[j].Weight
	} else {
		return self.list[i].Key < self.list[j].Key
	}
}
func (self *KeyWeightSorter) Append(key string, value interface{}, weight int) bool {
	self.sorted = false
	self.list = append(self.list, KeyWeight{Key: key, Value: value, Weight: weight})
	return true
}
func (self *KeyWeightSorter) Sort() []KeyWeight {
	if !self.sorted {
		self.sorted = true
		sort.Sort(self)
	}
	return self.list
}
func (self *KeyWeightSorter) Get(key string) *KeyWeight {
	for _, item := range self.list {
		if item.Key == key {
			return &item
		}
	}
	return nil
}
func (self *KeyWeightSorter) Foreach(itemFunc func(string, interface{}) error) error {
	var err error
	for _, item := range self.Sort() {
		if partErr := itemFunc(item.Key, item.Value); partErr != nil {
			err = partErr
		}
	}
	return err
}
