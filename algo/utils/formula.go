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

//将得分映射到0-max之间
func Norm(score float64,max float64) float64{
	return float64(Expit(float32(score))-0.5)*max*2
}
type ArrayMetric = func(arr1, arr2 []float32) float32

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

// 向量 cosine
func ArrayCosineInvert(arr1, arr2 []float32) float32 {
	var sum, absArr1, absArr2 float32
	if arr1 != nil && arr2 != nil && len(arr1) == len(arr2) {
		for i, arr1i := range arr1 {
			sum += arr1i * arr2[i]
			absArr1 += arr1[i] * arr1[i]
			absArr2 += arr2[i] * arr2[i]
		}

	}
	return float32(math.Sqrt(float64(absArr1*absArr2))+0.01) / (sum + 0.01) // cosine是越小越相似，因此此处分子分母反过来
}

// 向量距离
func ArrayDistance(arr1, arr2 []float32) float32 {
	var sum float32
	if arr1 != nil && arr2 != nil && len(arr1) == len(arr2) {
		for i, arr1i := range arr1 {
			sum += (arr1i - arr2[i]) * (arr1i - arr2[i])
		}

	}
	return float32(math.Sqrt(float64(sum)))
}
