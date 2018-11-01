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

func (model *QuickMatchTree) Name() string {
	return "QuickMatchTree"
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
	for _, user := range ctx.UserList {
		features := model.Features(ctx, &user)
		user.Score = model.PredictSingle(features)
		user.Features = algo.List2Features(features)
	}
}