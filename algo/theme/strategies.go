package theme

import(
	"rela_recommend/algo"
	"rela_recommend/models/behavior"
	"rela_recommend/algo/base/strategy"
)

func ItemBehaviorStrategyFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, itembehavior *behavior.UserBehavior, rankInfo *algo.RankInfo) error {
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


func UserBehaviorStrategyFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, userbehavior *behavior.UserBehavior, rankInfo *algo.RankInfo) error {
	var err error
	var abTest = ctx.GetAbTest()
	var currTime = float64(ctx.GetCreateTime().Unix())

	if userbehavior != nil {
		var upperRate float32
		var avgExpCount float64 = 3
		var avgInfCount float64 = 1
	
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
	}
	
	// 首次在一定时间内看到置顶，后续不置顶; 0 关闭，大于等于1 为打开多久时间内会置顶一次
	if selfTopTime := abTest.GetFloat64("rich_strategy:behavior:self_top_time", 0); selfTopTime >= 1.0 {
		dataInfo := iDataInfo.(*DataInfo)
		if dataInfo != nil && dataInfo.MomentCache != nil && ctx.GetUserInfo() != nil {
			user := ctx.GetUserInfo().(*UserInfo)
			if dataInfo.MomentCache.UserId == user.UserId {
				lastBehaviorTime := 0.0
				if userbehavior != nil {
					behaviors := userbehavior.Gets("theme.hotweek:exposure", "theme.hotweek:click")
					lastBehaviorTime = behaviors.LastTime
				}

				if currTime - lastBehaviorTime <= selfTopTime {
					rankInfo.IsTop = 1
				}
			}
		}
	}
	return err
}

// 用户自己的内容提权
// func SelfUpperStrategyItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
// 	dataInfo := iDataInfo.(*DataInfo)
// 	if dataInfo != nil && dataInfo.MomentCache != nil && ctx.GetUserInfo() != nil {
// 		user := ctx.GetUserInfo().(*UserInfo)
// 		if dataInfo.MomentCache.UserId == user.UserId {
// 			upperRate := ctx.GetAbTest().GetFloat("rich_strategy:self_upper:score", 0.3)
// 			rankInfo.AddRecommend("SelfUpper", 1.0 + upperRate)
// 		}
// 	}
// 	return nil
// }
