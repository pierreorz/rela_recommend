package theme

import(
	"math"
	"rela_recommend/algo"
	"rela_recommend/models/redis"
	"rela_recommend/algo/utils"
	rutils "rela_recommend/utils"
)

// 计算曝光与操作的分数
func countRateTimeScore(expBe *redis.Behavior, actBe *redis.Behavior, avgExp float64, 
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

// 热门提升权重
func DoHotBehaviorUpper(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	var avgExpCount float64 = 1000
	var avgInfCount float64 = 50
	var upperRate float32 = 0.0
	behavior := dataInfo.ThemeBehavior
	if behavior != nil {
		listCountScore, listRateScore, listTimeScore := countRateTimeScore(
			behavior.GetTotalListExposure(), behavior.GetTotalListClick(), avgExpCount, 0, 0, 0)
		infoCountScore, infoRateScore, infoTimeScore := countRateTimeScore(
			behavior.DetailExposure, behavior.GetTotalInteract(), avgInfCount, 0, 0, 0)

		upperRate = float32(0.4 * listCountScore * listRateScore * listTimeScore + 
							0.6 * infoCountScore * infoRateScore * infoTimeScore)
	}
	if upperRate != 0.0 {
		rankInfo.AddRecommend("ThemeBehavior", 1.0 + upperRate)
	}
	
	return nil
}

// 对自己的行为进行权重处理
type UserBehaviorStrategy struct { }
func (self *UserBehaviorStrategy) Do(ctx algo.IContext) error {
	var err error
	var avgExpCount float64 = 5
	var avgInfCount float64 = 1
	var upperRate float32 = 0.0
	var currTime = float64(ctx.GetCreateTime().Unix())
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		upperRate = 0.0

		behavior := dataInfo.UserBehavior
		if behavior != nil {
			listCountScore, listRateScore, listTimeScore := countRateTimeScore(
				behavior.GetTotalListExposure(), behavior.GetTotalListClick(), 
				avgExpCount, currTime, 3600, 36000)
			infoCountScore, infoRateScore, infoTimeScore := countRateTimeScore(
				behavior.DetailExposure, behavior.GetTotalInteract(), 
				avgInfCount, currTime, 3600, 36000)
			
			listRateScore = 2 * (listRateScore - 0.5)
			upperRate = float32(0.4 * listCountScore * listRateScore * listTimeScore + 
								0.6 * infoCountScore * infoRateScore * infoTimeScore)
		}
		if upperRate != 0.0 {
			rankInfo.AddRecommend("UserBehavior", 1.0 + upperRate)
		}
	}
	return err
}
