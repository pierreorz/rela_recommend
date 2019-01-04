package quick_match

import(
	// "time"
	"rela_recommend/log"
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/service"
)

type IQuickMatch interface {
	Name() string
	Init()
	Features(*QuickMatchContext, *UserInfo) []float32
	PredictSingle([]float32) float32
	Predict(*QuickMatchContext)
}

type QuickMatchBase struct {
	FilePath string
	AlgoName string
	model algo.IModel
}

func (self *QuickMatchBase) Name() string {
	return self.AlgoName
}

func (self *QuickMatchBase) Init() {
	tree := &utils.DecisionTree{}
	tree.Init(self.FilePath)
	self.model = tree
}

func (self *QuickMatchBase) Features(ctx *QuickMatchContext, user *UserInfo) []float32 {
	return service.UserRow(ctx.User.UserCache, user.UserCache)
}

func (self *QuickMatchBase) PredictSingle(features []float32) float32 {
	return self.model.PredictSingle(features)
}

type goResult struct {
	Index int
	Score float32
	Features []algo.Feature
}

func (self *QuickMatchBase) goPredict(r chan goResult, countChan chan int, index int, ctx *QuickMatchContext) {
	features := self.Features(ctx, &ctx.UserList[index])
	score := self.PredictSingle(features) 
	featureList := algo.List2Features(features)
	r <- goResult{Index: index, Score: score, Features: featureList}
	<- countChan
}

func (self *QuickMatchBase) Predict(ctx *QuickMatchContext) {
	var resultChan chan goResult = make(chan goResult)
	var countChan chan int = make(chan int, 8)
	
	for i := 0; i < len(ctx.UserList); i++ {
		countChan <- 1
		go self.goPredict(resultChan, countChan, i, ctx)
		log.Infof("i: %d", i)
	}

	for j := 0; j < len(ctx.UserList); j++ {
		log.Infof("j: %d", j)
		select {
			case result := <- resultChan:
				ctx.UserList[result.Index].AlgoScore = result.Score
				ctx.UserList[result.Index].Score = result.Score
				ctx.UserList[result.Index].Features = result.Features
			}
	}
	
	// for i := 0; i < len(ctx.UserList); i++ {
	// 	features := self.Features(ctx, &ctx.UserList[i])
	// 	ctx.UserList[i].AlgoScore = self.PredictSingle(features)
	// 	ctx.UserList[i].Score = ctx.UserList[i].AlgoScore
	// 	ctx.UserList[i].Features = algo.List2Features(features)
	// }
}