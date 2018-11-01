package utils

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"compress/gzip"
)

type ITree interface {
	Load()
	PredictSingle([]float32) []float32
}

type Node struct {
	Id int `json:"id"`
	Feature int `json:"feature"`
	Column string `json:"column"`
	Threshold float32 `json:"threshold"`
	Impurity float32 `json:"impurity"`
	Sample float32 `json:"sample"`
	Value []float32 `json:"value"`
	Left *Node `json:"left"`
	Right *Node `json:"right"`
}

// tree: 
type DecisionTree struct {
	MaxDepth int `json:"max_depth"`
	NodeCount int `json:"node_count"`
	FeatureCount int `json:"n_features"`
	ClassCount int `json:"n_classes"`
	RootNode *Node `json:"node"`
}

func (tree *DecisionTree) Init(path string)  {
	fr, _ := os.Open(path)
	defer fr.Close()

	gzf, _ := gzip.NewReader(fr)
	defer gzf.Close()

	data, _ := ioutil.ReadAll(gzf)
	jerr := json.Unmarshal(data, tree)
	fmt.Println(jerr)
}

func (tree *DecisionTree) PredictSingle(features []float32) float32 {
	if tree.RootNode != nil {
		node := tree.RootNode
		for node != nil && node.Feature >= 0 {
			if features[node.Feature] <= node.Threshold {
				node = node.Left
			} else {
				node = node.Right
			}
		}
		return node.Value[1] / node.Sample
	}
	return 0.0
}


func main() {
	work_dir, _ := os.Getwd()
	nodes := make([]Node, 5)
	for i, _ := range nodes {
		nodes[i].Id = 100
	}
	println(len(nodes))
	tree := DecisionTree{}
	tree.Init(work_dir + "/../../algo_files/quick_match_tree.model")
	target := tree.PredictSingle([]float32{5.1, 3.5, 1.4, 0.2})
	println("target:", target)
}