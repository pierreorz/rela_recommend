package theme

import(
	"rela_recommend/algo"
	"rela_recommend/models/behavior"
	"rela_recommend/algo/base/strategy"
)

func ItemBehaviorStrategyFunc(ctx algo.IContext, itembehavior *behavior.UserBehavior, rankInfo *algo.RankInfo) error {
	var err error
	var avgExpCount float64 = 20000
	var avgInfCount float64 = 500

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
	var upperRate float32
	// var abTest = ctx.GetAbTest()
	var avgExpCount float64 = 3
	var avgInfCount float64 = 1
	var currTime = float64(ctx.GetCreateTime().Unix())

	listCountScore, _, listTimeScore := strategy.BehaviorCountRateTimeScore(
		userbehavior.GetThemeListExposure(), userbehavior.GetThemeListInteract(), 
		avgExpCount, currTime, 18000, 18000)
	infoCountScore, _, infoTimeScore := strategy.BehaviorCountRateTimeScore(
		userbehavior.GetThemeDetailExposure(), userbehavior.GetThemeDetailInteract(), 
		avgInfCount, currTime, 36000, 18000)

	upperRate = - float32(0.3 * listCountScore * listTimeScore + 0.7 * infoCountScore * infoTimeScore)
	if upperRate != 0.0 {
		rankInfo.AddRecommend("UserBehavior", 1.0 + upperRate)
	}
	
	return err
}

// 用户自己的内容提权
func SelfUpperStrategyItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	dataInfo := iDataInfo.(*DataInfo)
	if dataInfo != nil && dataInfo.MomentCache != nil && ctx.GetUserInfo() != nil {
		user := ctx.GetUserInfo().(*UserInfo)
		if dataInfo.MomentCache.UserId == user.UserId {
			upperRate := ctx.GetAbTest().GetFloat("rich_strategy:self_upper:score", 0.3)
			rankInfo.AddRecommend("SelfUpper", 1.0 + upperRate)
		}
	}
	return nil
}
