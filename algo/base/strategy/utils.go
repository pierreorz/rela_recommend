package strategy

import (
	"math"
	"rela_recommend/algo/utils"
	rutils "rela_recommend/utils"
	"rela_recommend/models/behavior"
)


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


