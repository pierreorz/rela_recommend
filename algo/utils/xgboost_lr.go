package utils

import (
	
)
 
type XgboostLRClassifier struct {
	ModelAlgoBase
	GBDT XgboostClassifier 				`json:"gbdt"`
	OneHot OneHotEncoder 				`json:"one_hot"`
	LR LogisticRegression 				`json:"lr"`
	FeatureCount int					`json:"n_features"`
}

func (self *XgboostLRClassifier) Init(path string)  {
	LoadModel(path, self)
}

func (self *XgboostLRClassifier) PredictSingle(features *Features) float32 {
	leafs := self.GBDT.PredictSingleLeafs(features)
	new_features := self.OneHot.Transform(leafs)
	return self.LR.PredictSingle(new_features)
}
