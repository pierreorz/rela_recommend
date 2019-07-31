package utils

import (
	"math"
)

// Sigmoid 函数 1 / (1 + exp(-x))
func Expit(score float32) float32 {
	return 1.0 / (1.0 + float32(math.Exp(-float64(score))))
}

// 0-1的数 快速变化为 0-1 之间： 2/(1+exp(-log(x)))
func ExpLogit(score float64) float64 {
	return 2.0 / (1.0 + math.Exp(-math.Log(score)))
}

// 数组相乘的和
func ArrayMultSum(arr1, arr2 []float32) float32 {
	var sum float32 = 0.0
	if arr1 != nil && arr2 != nil && len(arr1) == len(arr2) {
		for i, arr1i := range arr1 {
			sum += arr1i * arr2[i]
		}
	}
	return sum
}
