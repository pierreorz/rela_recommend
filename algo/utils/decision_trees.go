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
	
	Threshold []float32 `json:"threshold"`
	Impurity []float32 `json:"impurity"`
	Value [][]float32 `json:"value"`
	Feature []int `json:"feature"`
	ChildrenLeft []int `json:"children_left"`
	ChildrenRight []int `json:"children_right"`
	ChildrenMiss []int `json:"children_miss"`
}

func (tree *DecisionTreeBase) PredictSingleLeaf(features *Features) int {
	var node_id int = 0
	for {
		feature_id := tree.Feature[node_id]
		if feature_id < 0 {
			break
		}
		feature_val, ok := features.Get(feature_id)
		if ok || len(tree.ChildrenMiss) == 0 {
			if feature_val <= tree.Threshold[node_id] {
				node_id = tree.ChildrenLeft[node_id]
			} else {
				node_id = tree.ChildrenRight[node_id]
			}
		} else {
			node_id = tree.ChildrenMiss[node_id]
		}
	}
	return node_id
}


// ********************************************** DecisionTreeRegressor
type DecisionTreeRegressor struct {
	DecisionTreeBase
}

func (self *DecisionTreeRegressor) Init(path string)  {
	LoadModel(path, self)
}

func (tree *DecisionTreeRegressor) PredictSingle(features *Features) float32 {
	node_id := tree.PredictSingleLeaf(features)
	return tree.Value[node_id][0]
}


// ********************************************** DecisionTreeClassifier
type DecisionTreeClassifier struct {
	DecisionTreeBase
}

func (self *DecisionTreeClassifier) Init(path string)  {
	LoadModel(path, self)
}

func (tree *DecisionTreeClassifier) PredictSingle(features *Features) float32 {
	node_id := tree.PredictSingleLeaf(features)
	node_values := tree.Value[node_id]
	return node_values[1] / (node_values[0] + node_values[1])
}
