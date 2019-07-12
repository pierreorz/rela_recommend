package live

import(
	"time"
	"rela_recommend/algo/utils"
	"rela_recommend/service/abtest"
)

// 直播推荐算法上下文
type LiveAlgoContext struct {
	RankId string
	CreateTime time.Time
	Ua string
	Platform int
	AbTest *abtest.AbTest
	User *UserInfo
	LiveList []LiveInfo
}

type ILiveAlgo interface {
	Name() string
	Init()
	Features(*LiveAlgoContext, *LiveInfo) *utils.Features
	PredictSingle(*utils.Features) float32
	Predict(*LiveAlgoContext)
}

type LiveAlgoBase struct {
	FilePath string
	AlgoName string
	model utils.IModelAlgo
}

func (self *LiveAlgoBase) Name() string {
	return self.AlgoName
}

func (self *LiveAlgoBase) Init() {
	model := &utils.LogisticRegression{}
	model.Init(self.FilePath)
	self.model = model
}

func (self *LiveAlgoBase) PredictSingle(features *utils.Features) float32 {
	new_features := self.model.TransformSingle(features)
	return self.model.PredictSingle(new_features)
}

func (self *LiveAlgoBase) Predict(ctx *LiveAlgoContext) {
	for i := 0; i < len(ctx.LiveList); i++ {
		features := self.Features(ctx, &ctx.LiveList[i])
		ctx.LiveList[i].Features = features
		ctx.LiveList[i].RankInfo.AlgoScore = self.PredictSingle(features)
		ctx.LiveList[i].RankInfo.Score = ctx.LiveList[i].RankInfo.AlgoScore
	}
}

func (self *LiveAlgoBase) Features(ctx *LiveAlgoContext, user *LiveInfo) *utils.Features {
	return GetLiveFeatures(ctx, user)
}
