package algo

import(
	"rela_recommend/algo"
	"rela_recommend/service"
)

type QuickMatchTree struct {
	algo.BaseAlgorithm
	tree *algo.DecisionTree
}

func (model *QuickMatchTree) Name() string {
	return "QuickMatchTree"
}

func (model *QuickMatchTree) Init() {
	tree := &DecisionTree{}
	tree.Load(model.FilePath)
	model.tree = tree
}

func (model *QuickMatchTree) Features(cxt *QuickMatchContext, user *UserInfo) []float32 {
	return []float32{}
}

func (model *QuickMatchTree) PredictSingle(features []float32) float32 {
	return model.tree.PredictSingle(features)
}

func (model *QuickMatchTree) Predict(cxt *QuickMatchContext) {
	for _, user := range cxt.UserList {
		features := model.Features(ctx, user)
		user.Score = model.PredictSingle(features)
		user.Features = feature
	}
}