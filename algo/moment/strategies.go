package moment

import(
	"rela_recommend/algo"
	"rela_recommend/models/behavior"
	"rela_recommend/algo/base/strategy"
)

// 按照6小时优先策略
func DoTimeLevel(ctx algo.IContext, index int) error {
	abtest := ctx.GetAbTest()
	hourStrategy:=abtest.GetInt("time_interval",3)
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	hours := int(ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Hours()) / hourStrategy
	rankInfo.Level = -hours
	return nil
}

// 按照秒级时间优先策略
func DoTimeFirstLevel(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	rankInfo.Level = int(dataInfo.MomentCache.InsertTime.Unix())
	return nil
}

// 按数据被访问行为进行策略提降权
func ItemBehaviorStrategyFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, itembehavior *behavior.UserBehavior, rankInfo *algo.RankInfo) error {
	var err error
	var avgExpCount float64 = 50

	listCountScore, listRateScore, listTimeScore := strategy.BehaviorCountRateTimeScore(
		itembehavior.GetMomentListExposure(), itembehavior.GetMomentListInteract(), avgExpCount, 0, 0, 0)

	upperRate := float32(listCountScore * listRateScore * listTimeScore)

	if upperRate != 0.0 {
		rankInfo.AddRecommend("ItemBehavior", 1.0 + upperRate)
	}
	return err
}


// 按用户访问行为进行策略提降权
func UserBehaviorStrategyFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, userbehavior *behavior.UserBehavior, rankInfo *algo.RankInfo) error {
	var err error
	var currTime = float64(ctx.GetCreateTime().Unix())
	
	if userbehavior != nil {
		var avgExpCount float64 = 2

		listCountScore, _, listTimeScore := strategy.BehaviorCountRateTimeScore(
			userbehavior.GetMomentListExposure(), userbehavior.GetMomentListInteract(), 
			avgExpCount, currTime, 36000, 18000)
		
		upperRate := -float32(listCountScore * listTimeScore)
		
		if upperRate != 0.0 {
			rankInfo.AddRecommend("UserBehavior", 1.0 + upperRate)
		}
	}
	return err
}
