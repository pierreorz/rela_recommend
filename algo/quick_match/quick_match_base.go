package quick_match

import(
	// "time"
	// "rela_recommend/log"
	"sync"
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

// 使用简单计算单个
func (self *QuickMatchBase) doPredictSingle(ctx *QuickMatchContext, index int) {
	features := self.Features(ctx, &ctx.UserList[index])
	ctx.UserList[index].AlgoScore = self.PredictSingle(features)
	ctx.UserList[index].Score = ctx.UserList[index].AlgoScore
	ctx.UserList[index].Features = algo.List2Features(features)
}

// 使用简单计算
func (self *QuickMatchBase) doPredict(ctx *QuickMatchContext) {
	for i := 0; i < len(ctx.UserList); i++ {
		self.doPredictSingle(ctx, i)
	}
}
// 使用goroutine多线程并行计算
func (self *QuickMatchBase) goPredict(ctx *QuickMatchContext, batch int) {
	parts := utils.SplitIndexs(len(ctx.UserList), batch)
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

func (self *QuickMatchBase) Predict(ctx *QuickMatchContext) {
	if len(ctx.UserList) < 100 {
		self.doPredict(ctx)
	} else {
		self.goPredict(ctx, 6)
	}
}