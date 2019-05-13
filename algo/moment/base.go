package moment

import(
	"time"
	"rela_recommend/algo"
	rutils "rela_recommend/utils"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	"rela_recommend/algo/utils"
	"rela_recommend/service/abtest"
)

type RankInfo struct {
	IsTop		int 		// 1: 置顶， 0: 默认， -1:置底
	Level		int			// 推荐等级
	Punish		float32		// 惩罚系数
	AlgoName	string		// 算法名称
	AlgoScore 	float32		// 算法得分
	Score 		float32		// 最终得分
}

// 用户信息
type UserInfo struct {
	UserId int64
	UserCache *pika.UserProfile
	UserConcerns *rutils.SetInt64
}

// 主播信息
type DataInfo struct {
	DataId 				int64
	UserCache 			*pika.UserProfile
	MomentCache 		*redis.Moments
	MomentExtendCache 	*redis.MomentsExtend
	MomentProfile		*redis.MomentsProfile
	RankInfo			*RankInfo
	Features 			*utils.Features
}

// 直播推荐算法上下文
type AlgoContext struct {
	RankId string
	CreateTime time.Time
	Platform int
	Request *algo.RecommendRequest
	AbTest *abtest.AbTest
	User *UserInfo
	DataIds []int64
	DataList []DataInfo
}

type IMomentAlgo interface {
	Name() string
	Init()
	Features(*AlgoContext, *DataInfo) *utils.Features
	PredictSingle(*utils.Features) float32
	Predict(*AlgoContext)
}

type MomentAlgoBase struct {
	FilePath string
	AlgoName string
	Model utils.IModelAlgo		`json:"model"`
	Words	map[string]int		`json:"words"`
}

func (self *MomentAlgoBase) Name() string {
	return self.AlgoName
}

func (self *MomentAlgoBase) Init() {
	model := &utils.LogisticRegression{}
	model.Init(self.FilePath)
	self.Model = model
}

func (self *MomentAlgoBase) PredictSingle(features *utils.Features) float32 {
	new_features := self.Model.TransformSingle(features)
	return self.Model.PredictSingle(new_features)
}

func (self *MomentAlgoBase) Predict(ctx *AlgoContext) {
	for i := 0; i < len(ctx.DataList); i++ {
		features := self.Features(ctx, &ctx.DataList[i])
		ctx.DataList[i].Features = features
		ctx.DataList[i].RankInfo.AlgoName = self.Name()
		ctx.DataList[i].RankInfo.AlgoScore = self.PredictSingle(features)
		ctx.DataList[i].RankInfo.Score = ctx.DataList[i].RankInfo.AlgoScore
	}
}

func (self *MomentAlgoBase) Features(ctx *AlgoContext, data *DataInfo) *utils.Features {
	return GetMomentFeatures(self, ctx, data)
}

// 检查词是否被允许
func (self *MomentAlgoBase) CheckWords(words []string) []string {
	res := make([]string, 0)
	for _, word := range words {
		if _, ok := self.Words[word]; ok {
			res = append(res, word)
		}
	}
	return res
}
