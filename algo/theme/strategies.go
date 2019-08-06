package theme

import(
	"rela_recommend/algo"
	"rela_recommend/models/behavior"
	"rela_recommend/algo/base/strategy"
)

func ItemBehaviorStrategyFunc(ctx algo.IContext, itembehavior *behavior.UserBehavior, rankInfo *algo.RankInfo) error {
	var err error
	var avgExpCount float64 = 1000
	var avgInfCount float64 = 50

	listCountScore, listRateScore, listTimeScore := strategy.BehaviorCountRateTimeScore(
		itembehavior.GetThemeListExposure(), itembehavior.GetThemeListInteract(), avgExpCount, 0, 0, 0)
	infoCountScore, infoRateScore, infoTimeScore := strategy.BehaviorCountRateTimeScore(
		itembehavior.GetThemeDetailExposure(), itembehavior.GetThemeDetailInteract(), avgInfCount, 0, 0, 0)

	upperRate := float32(0.4 * listCountScore * listRateScore * listTimeScore + 
						 0.6 * infoCountScore * infoRateScore * infoTimeScore)

	if upperRate != 0.0 {
		rankInfo.AddRecommend("ItemBehavior", 1.0 + upperRate)
	}
	return err
}


func UserBehaviorStrategyFunc(ctx algo.IContext, userbehavior *behavior.UserBehavior, rankInfo *algo.RankInfo) error {
	var err error
	var avgExpCount float64 = 5
	var avgInfCount float64 = 1
	var currTime = float64(ctx.GetCreateTime().Unix())

	listCountScore, listRateScore, listTimeScore := strategy.BehaviorCountRateTimeScore(
		userbehavior.GetThemeListExposure(), userbehavior.GetThemeListInteract(), 
		avgExpCount, currTime, 3600, 36000)
	infoCountScore, infoRateScore, infoTimeScore := strategy.BehaviorCountRateTimeScore(
		userbehavior.GetThemeDetailExposure(), userbehavior.GetThemeDetailInteract(), 
		avgInfCount, currTime, 3600, 36000)
	
	listRateScore = 2 * (listRateScore - 0.5)
	upperRate := float32(0.4 * listCountScore * listRateScore * listTimeScore + 
						0.6 * infoCountScore * infoRateScore * infoTimeScore)
	
	if upperRate != 0.0 {
		rankInfo.AddRecommend("UserBehavior", 1.0 + upperRate)
	}
	
	return err
}
