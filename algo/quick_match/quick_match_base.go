package quick_match

import(
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
	model algo.IModel
}

func (self *QuickMatchBase) Name() string {
	return "QuickMatchBase"
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

func (self *QuickMatchBase) Predict(ctx *QuickMatchContext) {
	for i := 0; i < len(ctx.UserList); i++ {
		features := self.Features(ctx, &ctx.UserList[i])
		ctx.UserList[i].Score = self.PredictSingle(features)
		ctx.UserList[i].Features = algo.List2Features(features)
	}
}