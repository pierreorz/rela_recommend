package utils

import (
	
)
 
type GradientBoostingLRClassifier struct {
	ModelAlgoBase
	GBDT GradientBoostingClassifier 	`json:"gbdt"`
	OneHot OneHotEncoder 				`json:"one_hot"`
	LR LogisticRegression 				`json:"lr"`
	FeatureCount int					`json:"n_features"`
}

func (self *GradientBoostingLRClassifier) Init(path string)  {
	LoadModel(path, self)
}

func (self *GradientBoostingLRClassifier) PredictSingle(features *Features) float32 {
	leafs := self.GBDT.PredictSingleLeafs(features)
	new_features := self.OneHot.Transform(leafs)
	return self.LR.PredictSingle(new_features)
}
