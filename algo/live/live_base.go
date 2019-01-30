package live

import(
	// "time"
	// "rela_recommend/log"
	"rela_recommend/models/pika"
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/service/abtest"
)

// 用户信息
type UserInfo struct {
	UserId int64
	UserCache *pika.UserProfile
}

// 主播信息
type LiveInfo struct {
	UserId int64
	UserCache *pika.UserProfile
	AlgoScore float32
	Score float32
	Features []algo.Feature
}

// 直播推荐算法上下文
type LiveAlgoContext struct {
	RankId string
	Ua string
	AbTest *abtest.AbTest
	User *UserInfo
	LiveList []LiveInfo
}

type ILiveAlgo interface {
	Name() string
	Init()
	Features(*LiveAlgoContext, *LiveInfo) map[int]float32
	PredictSingle([]float32) float32
	Predict(*LiveAlgoContext)
}

type LiveAlgoBase struct {
	FilePath string
	AlgoName string
	model algo.IModel
}

func (self *LiveAlgoBase) Name() string {
	return self.AlgoName
}

func (self *LiveAlgoBase) Init() {
	model := &utils.LR{}
	model.Init(self.FilePath)
	self.model = model
}

func (self *LiveAlgoBase) PredictSingle(features []float32) float32 {
	return self.model.PredictSingle(features)
}

func (self *LiveAlgoBase) Predict(ctx *LiveAlgoContext) {
	for i := 0; i < len(ctx.LiveList); i++ {
		features := self.Features(ctx, &ctx.LiveList[i])
		ctx.LiveList[i].AlgoScore = self.PredictSingle(features)
		ctx.LiveList[i].Score = ctx.LiveList[i].AlgoScore
		ctx.LiveList[i].Features = algo.List2Features(features)
	}
}

func (self *LiveAlgoBase) Features(ctx *LiveAlgoContext, user *LiveInfo) []float32 {
	var res []float32
	return res
}
