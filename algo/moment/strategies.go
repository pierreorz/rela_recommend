package moment

import (
	"rela_recommend/algo"
	"rela_recommend/models/behavior"
	"rela_recommend/algo/base/strategy"
	"math"
	"rela_recommend/utils"
)

// 按照6小时优先策略
func DoTimeLevel(ctx algo.IContext, index int) error {
	abtest := ctx.GetAbTest()
	if hourStrategy := abtest.GetInt("DoTimeLevel:time_interval", 3); hourStrategy > 0 {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		hours := int(ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Hours()) / hourStrategy
		rankInfo.Level = -hours
	}
	return nil
}

//日志提权策略
func DoTimeWeightLevel(ctx algo.IContext, index int) error{
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	timeLevel := int(ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Hours()) / 3
	if timeLevel <= 3 {
		rankInfo.AddRecommend("momentNearTimeWeight", 1.0+float32(1.0/(2.0+float32(timeLevel))))
	} else {
		rankInfo.AddRecommend("momentNearTimeWeight", float32(1.0/(float32(timeLevel)-3.0)))
	}
	return nil
}

//标签日志提权
func MomLabelAddWeight(ctx algo.IContext, index int) error{
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	if utils.IfLabel(dataInfo.MomentCache.MomentsText){
		rankInfo.AddRecommend("labelMomWeight",1.2)
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

//附近日志新用户提权策略
func AroundNewUserAddWeightFunc(ctx algo.IContext, index int) error{
	abtest := ctx.GetAbTest()
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	if newUserDefine:=abtest.GetInt("new_user_define",0);newUserDefine>0{
		hourInterval:=int(ctx.GetCreateTime().Sub(dataInfo.UserCache.CreateTime.Time).Hours())/24
		if hourInterval<newUserDefine{
			rankInfo.AddRecommend("newUserWeight",1.2)
		}
	}
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

//日志详情页看过沉底策略
func DetailRecommendStrategyFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, userbehavior *behavior.UserBehavior, rankInfo *algo.RankInfo) error{
	var err error
	if userbehavior!=nil{
		behaviorCount:=behavior.MergeBehaviors(userbehavior.GetMomentListExposure(), userbehavior.GetMomentListInteract())
		if behaviorCount!=nil{
			if behaviorCount.Count>0{
				rankInfo.Level=int(-math.Min(behaviorCount.Count, 3))
			}
		}
	}
	return  err
}


// 按用户访问行为进行策略提降权
func UserBehaviorStrategyFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, userbehavior *behavior.UserBehavior, rankInfo *algo.RankInfo) error {
	var err error
	var abtest = ctx.GetAbTest()
	if abtest.GetBool("rich_strategy:behavior:moment_item_new", false) {
		if userbehavior != nil {
			// 浏览过的内容使用浏览次数及互动次数反序排列
			allBehavior := behavior.MergeBehaviors(userbehavior.GetMomentListExposure(), userbehavior.GetMomentListInteract())
			if allBehavior != nil {
				rankInfo.Level = int(-math.Min(allBehavior.Count, 5))
			}
		}
	} else {
		var currTime = float64(ctx.GetCreateTime().Unix())
		if userbehavior != nil {
			var avgExpCount float64 = 2
			listCountScore, _, listTimeScore := strategy.BehaviorCountRateTimeScore(
				userbehavior.GetMomentListExposure(), userbehavior.GetMomentListInteract(),
				avgExpCount, currTime, 36000, 18000)

			upperRate := -float32(listCountScore * listTimeScore)

			if upperRate != 0.0 {
				rankInfo.AddRecommend("UserBehavior", 1.0+upperRate)
			}
		}
	}
	return err
}
