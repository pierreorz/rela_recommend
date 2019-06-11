package moment

import(
	"time"
	"sync"
	"rela_recommend/algo"
	rutils "rela_recommend/utils"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	"rela_recommend/algo/utils"
	"rela_recommend/service/abtest"
)

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
	RankInfo			*algo.RankInfo
	Features 			*utils.Features
}

func (self *DataInfo) GetDataId() int64 {
	return self.DataId
}

func(self *DataInfo) SetRankInfo(rankInfo *algo.RankInfo) {
	self.RankInfo = rankInfo
}

func(self *DataInfo) GetRankInfo() *algo.RankInfo {
	return self.RankInfo
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
	CheckWords([]string) []string
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

// 使用简单计算单个
func (self *MomentAlgoBase) doPredictSingle(ctx *AlgoContext, index int) {
	features := self.Features(ctx, &ctx.DataList[index])
	ctx.DataList[index].Features = features
	ctx.DataList[index].RankInfo.AlgoScore = self.PredictSingle(features)
	ctx.DataList[index].RankInfo.Score = ctx.DataList[index].RankInfo.AlgoScore
}

// 使用简单计算
func (self *MomentAlgoBase) doPredict(ctx *AlgoContext) {
	for i := 0; i < len(ctx.DataList); i++ {
		self.doPredictSingle(ctx, i)
	}
}
// 使用goroutine多线程并行计算
func (self *MomentAlgoBase) goPredict(ctx *AlgoContext, batch int) {
	parts := utils.SplitIndexs(len(ctx.DataList), batch)
	wg := new(sync.WaitGroup)
	for _, part := range parts {
		wg.Add(1)
		go func(part []int) {
			defer wg.Done()
			for _, indx := range part {
				self.doPredictSingle(ctx, indx)
			}
        }(part)
	}
	wg.Wait()
}


func (self *MomentAlgoBase) Predict(ctx *AlgoContext) {
	if len(ctx.DataList) < 100 {
		self.doPredict(ctx)
	} else {
		self.goPredict(ctx, 6)
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
