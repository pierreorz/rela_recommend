package algorithms

import(
	"rela_recommend/algorithms"
)

type QuickMatchTree struct {
	algorithms.BaseAlgorithm
}

func (model *QuickMatchTree) Name() string {
	return "QuickMatchTree"
}

func (model *QuickMatchTree) Init() {

}

func (model *QuickMatchTree) Features() []algorithms.Feature {
	return []float32{}
}

func (model *QuickMatchTree) PredictSingle(features []float32) float32 {
	
}

func (model *QuickMatchTree) Predict(featuresList [][]float32) []float32 {
	scores := [len(featuresList)]float32
	for i, features := range featuresList {
		scores[i] = model.PredictSingle(features)
	}
	return scores
}