package redis

import (
	"math"
)

type Behavior struct {
	Count    float64 `json:"count"`
	LastTime float64 `json:"last_time"`
}

// 合并行为
func MergeBehaviors(behaviors ...*Behavior) *Behavior {
	res := &Behavior{}
	for _, behavior := range behaviors {
		if behavior != nil {
			res.Count += behavior.Count
			res.LastTime = math.Max(res.LastTime, behavior.LastTime)
		}
	}
	return res
}
