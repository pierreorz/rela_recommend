package moment

import (
	"rela_recommend/algo"
	"rela_recommend/models/behavior"
	"rela_recommend/algo/base/strategy"
	"math"
)

// 按照6小时优先策略
func DoTimeLevel(ctx algo.IContext, index int) error {
	abtest := ctx.GetAbTest()
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	if hourStrategy := abtest.GetInt("DoTimeLevel:time_interval", 3); hourStrategy > 0 {
		hours := int(ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Hours()) / hourStrategy
		rankInfo.Level = -hours
		}
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
	var abTest = ctx.GetAbTest()

	if abTest.GetBool("rich_strategy:behavior:moment_item_new", false) {
		listRate := strategy.WilsonScore(itembehavior.GetMomentListExposure(), itembehavior.GetMomentListInteract(), 3)
		upperRate := float32(listRate)
		if upperRate != 0.0 {
			rankInfo.AddRecommend("ItemBehaviorV1", 1.0 + upperRate)
		}
	} else{
	var avgExpCount float64 = 50

	listCountScore, listRateScore, listTimeScore := strategy.BehaviorCountRateTimeScore(
	itembehavior.GetMomentListExposure(), itembehavior.GetMomentListInteract(), avgExpCount, 0, 0, 0)

	upperRate := float32(listCountScore * listRateScore * listTimeScore)

	if upperRate != 0.0 {
	rankInfo.AddRecommend("ItemBehavior", 1.0 + upperRate)
		}
	}
	return err
}

// 按用户访问行为进行策略提降权
func UserBehaviorStrategyFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, userbehavior *behavior.UserBehavior, rankInfo *algo.RankInfo) error {
	var err error
	var abtest=ctx.GetAbTest()
	if abtest.GetBool("rich_strategy:behavior:moment_item_new", false){
		if userbehavior != nil {
			// 浏览过的内容使用浏览次数反序排列，3:未浏览过，2：浏览一次，1：浏览2次，0：浏览3次以上
			allBehavior := behavior.MergeBehaviors(userbehavior.GetMomentListExposure(), userbehavior.GetMomentListInteract())
			if allBehavior != nil {
				rankInfo.Level = int(-math.Min(allBehavior.Count, 5))
				}
		}
	}else{
		var currTime = float64(ctx.GetCreateTime().Unix())
		ifSort:=abtest.GetBool("sort_with_time",false)
		//若按时间策略，会将日志分级别展示，0-6 6-12 12-18 18-24 每6h一级别
		if ifSort{
			dataInfo := iDataInfo.(*DataInfo)
			rankInfo.Level=int(ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Hours()) / 6
		}
		if userbehavior != nil {
			if ifSort {
				nearBehavior := behavior.MergeBehaviors(userbehavior.GetMomentNearListInteract(), userbehavior.GetMomentNearListExposure())
				if nearBehavior != nil {
					if nearBehaviorNum := int(math.Max(nearBehavior.Count, 1)); nearBehaviorNum%2 == 0 {
						rankInfo.Level -= 4
					}
				}
			}
			var avgExpCount float64 = 2

			listCountScore, _, listTimeScore := strategy.BehaviorCountRateTimeScore(
				userbehavior.GetMomentListExposure(), userbehavior.GetMomentListInteract(),
				avgExpCount, currTime, 36000, 18000)

			upperRate := -float32(listCountScore * listTimeScore)

			if upperRate != 0.0 {rankInfo.AddRecommend("UserBehavior", 1.0+upperRate)
			}
		}
	}
	return err
}
