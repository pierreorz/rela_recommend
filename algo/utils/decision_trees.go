package utils

import (
)

// ********************************************** DecisionTreeBase
type DecisionTreeBase struct {
	ModelAlgoBase
	MaxDepth int `json:"max_depth"`
	NodeCount int `json:"node_count"`
	FeatureCount int `json:"n_features"`
	ClassCount int `json:"n_classes"`
	MissingType int `json:"missing_type"`  	// 缺失值处理方式：0:按照0处理，1:按照左分支处理, 2:按照右分支处理
	
	Threshold []float32 `json:"threshold"`
	Impurity []float32 `json:"impurity"`
	Value [][]float32 `json:"value"`
	Feature []int `json:"feature"`
	ChildrenLeft []int `json:"children_left"`
	ChildrenRight []int `json:"children_right"`
}

func (tree *DecisionTreeBase) PredictSingleLeaf(features *Features) int {
	var node_id int = 0
	for {
		feature_id := tree.Feature[node_id]
		if feature_id < 0 {
			break
		}
		feature_val, ok := features.Get(feature_id)

		var toLeft = false
		if !ok && tree.MissingType > 0 {
			toLeft = tree.MissingType == 1
		} else {
			toLeft = feature_val <= tree.Threshold[node_id]
		}

		if toLeft {
			node_id = tree.ChildrenLeft[node_id]
		} else {
			node_id = tree.ChildrenRight[node_id]
		}
	}
	return node_id
}


// ********************************************** DecisionTreeRegressor
type DecisionTreeRegressor struct {
	DecisionTreeBase
}

func (tree *DecisionTreeRegressor) PredictSingle(features *Features) float32 {
	node_id := tree.PredictSingleLeaf(features)
	return tree.Value[node_id][0]
}


// ********************************************** DecisionTreeClassifier
type DecisionTreeClassifier struct {
	DecisionTreeBase
}

func (tree *DecisionTreeClassifier) PredictSingle(features *Features) float32 {
	node_id := tree.PredictSingleLeaf(features)
	node_values := tree.Value[node_id]
	return node_values[1] / (node_values[0] + node_values[1])
}
