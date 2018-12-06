package quick_match

import(
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/service"
)

type QuickMatchTree struct {
	algo.BaseAlgorithm
	FilePath string
	tree *utils.DecisionTree
}

func (model *QuickMatchTree) Init() {
	tree := &utils.DecisionTree{}
	tree.Init(model.FilePath)
	model.tree = tree
}

func (model *QuickMatchTree) Features(ctx *QuickMatchContext, user *UserInfo) []float32 {
	return service.UserRow(ctx.User.UserCache, user.UserCache)
}

func (model *QuickMatchTree) PredictSingle(features []float32) float32 {
	return model.tree.PredictSingle(features)
}

func (model *QuickMatchTree) Predict(ctx *QuickMatchContext) {
	for i := 0; i < len(ctx.UserList); i++ {
		features := model.Features(ctx, &ctx.UserList[i])
		ctx.UserList[i].Score = model.PredictSingle(features)
		ctx.UserList[i].Features = algo.List2Features(features)
	}
}


type QuickMatchTreeV1_0 struct {
	QuickMatchBase
}

func (self *QuickMatchTreeV1_0) Name() string {
	return "QuickMatchTreeV1_0"
}
