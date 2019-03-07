package utils

import (
)

// tree: 
type GradientBoostingClassifier struct {
	ModelAlgoBase
	MaxDepth int `json:"max_depth"`
	FeatureCount int `json:"n_features"`
	InitValue float32 `json:"init_value"`
	ClassCount int `json:"n_classes"`
	EstimatorCount int `json:"n_estimators"`
	LearningRate float32 `json:"learning_rate"`
	Estimators []DecisionTreeRegressor `json:"estimators"`
}

func (self *GradientBoostingClassifier) Init(path string)  {
	LoadModel(path, self)
}

func (self *GradientBoostingClassifier) PredictSingle(features *Features) float32 {
	score := self.InitValue
	for i := 0; i < self.EstimatorCount; i++ {
		score += self.LearningRate * self.Estimators[i].PredictSingle(features)
	}
	return Expit(score)
}

// 获取命中的叶子节点 [EstimatorCount]int
// 如每棵树为 [[0,1,2], [0,1,2,3], [0,1,2,3,4]] 返回每棵树命中节点 [1, 1, 2]
func (self *GradientBoostingClassifier) PredictSingleLeafs(features *Features) []int {
	leafs := make([]int, self.EstimatorCount)
	for i := 0; i < self.EstimatorCount; i++ {
		leafs[i] = self.Estimators[i].PredictSingleLeaf(features)
	}
	return leafs
}
