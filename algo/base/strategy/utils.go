package strategy

import (
	"math"
	"rela_recommend/algo/utils"
	rutils "rela_recommend/utils"
	"rela_recommend/models/behavior"
)

// 威尔逊得分 伯努利版
// pos: 正例数；total 总量； z 一般取2，95%置信度
func wilsonScoreForBernoulli(pos, total, z float64) float64 {
	if total > 0 {
		p := pos / total
		zz := math.Pow(z, 2.0)
		return (p + 0.5 * zz / total - 0.5 * z / total * math.Sqrt(4 * total * (1 - p) * p + zz)) / (1 + zz / total)
	}
	return 0.0
}

// 威尔逊得分计算互动/浏览分数
func WilsonScore(expBe *behavior.Behavior, actBe *behavior.Behavior, scale float64) float64 {
	if expBe != nil && expBe.Count > 0 && actBe != nil {
		s := wilsonScoreForBernoulli(actBe.Count, expBe.Count, 2.0)
		return math.Min(s * scale, 1.0)
	}
	return 0.0
}

// 计算曝光与操作的分数，根据互动提权
// 返回 2个分数： 0-1递增的浏览次数分数，0-1递增的互动概率分数，1-0递减的时间衰减分数
func BehaviorCountRateTimeScore(expBe *behavior.Behavior, actBe *behavior.Behavior, avgExp float64, 
	curTime float64, expTimeBase float64, actTimeBase float64) (float64, float64, float64) {
	var countScore float64 = 0.0
	var rateScore float64 = 0.0
	var timeScore float64 = 1.0
	if expBe != nil && expBe.Count > 0 && actBe != nil {
		hasAct := actBe.Count > 0
		countScore = 1.0 - math.Exp(- expBe.Count / avgExp)
		rate := math.Max(0.000001, math.Min(1.0, actBe.Count / expBe.Count))
		rateScore = utils.ExpLogit(rate)
		if curTime > 0 && expTimeBase > 0 && actTimeBase > 0 { // 时间衰减
			lastTime := math.Max(actBe.LastTime, expBe.LastTime)
			//避免脏数据影响
			lastTime =math.Min(lastTime,curTime)
			timeBase := rutils.IfElse(hasAct, actTimeBase, expTimeBase)
			timeScore = math.Exp(- (curTime - lastTime) / timeBase)
		}
	}
	return countScore, rateScore, timeScore
}
// 计算曝光与操作的分数，根据互动降权
func BehaviorCountRateTimeLowerScore(expBe *behavior.Behavior, actBe *behavior.Behavior, avgExp float64, 
	curTime float64, expTimeBase float64, actTimeBase float64) (float64, float64, float64) {
	var countScore float64 = 0.0
	var rateScore float64 = 0.0
	var timeScore float64 = 1.0
	if expBe != nil && expBe.Count > 0 && actBe != nil {
		hasAct := actBe.Count > 0
		countScore = 1.0 - math.Exp(- expBe.Count / avgExp)
		rate := math.Max(0.000001, math.Min(1.0, actBe.Count / expBe.Count))
		rateScore = utils.ExpLogit(rate)
		if curTime > 0 && expTimeBase > 0 && actTimeBase > 0 { // 时间衰减
			lastTime := rutils.IfElse(hasAct, actBe.LastTime, expBe.LastTime)
			timeBase := rutils.IfElse(hasAct, actTimeBase, expTimeBase)
			timeScore = math.Exp(- (curTime - lastTime) / timeBase)
		}
	}
	return countScore, rateScore, timeScore
}


