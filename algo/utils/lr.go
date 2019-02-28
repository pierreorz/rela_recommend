package utils

import (
)

// {"n_features": 10, "n_classes": 2, "coef": {1: 1.2, 10: 5.2}, "intercept": -2.111}
type LogisticRegression struct {
	ModelAlgoBase
	CoefMap []float32 		`json:"coef"`
	Intercept float32 		`json:"intercept"`
}

func (self *LogisticRegression) Init(path string)  {
	LoadModel(path, self)
}

func (self *LogisticRegression) PredictSingle(features *Features) float32 {
	var score float32 = self.Intercept
	for i, feature := range features.ToMap() {
		score += self.CoefMap[i] * feature
	}
	return Expit(score)
}
