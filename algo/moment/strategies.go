package moment

import (
	"rela_recommend/algo"
	"rela_recommend/models/behavior"
	"rela_recommend/algo/base/strategy"
	"math"
	"strings"
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

//日志提权策略v2
func DoTimeWeightLevelV2(ctx algo.IContext, index int) error{
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	timeLevel := int(ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Hours()) / 3
	if timeLevel <= 8 {
		//避免脏数据的影响
		rankInfo.AddRecommend("momentNearTimeWeightV2", 1.0+float32(0.1/(1.0+math.Max(float64(timeLevel),0))))
	}
	return nil
}


//用户短期偏好提取
func ShortPrefAddWeight(ctx algo.IContext, index int) error{
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	tagList :=""
	userInfo := ctx.GetUserInfo().(*UserInfo)
	if userInfo.MomentUserProfile!=nil&&dataInfo.MomentOfflineProfile!=nil{
		shortPrefs :=userInfo.MomentUserProfile.AiTag["short"]
		tags :=dataInfo.MomentOfflineProfile.AiTag
		if len(tags)>0{
			for _,tag :=range tags{
				tagList+=tag.Name
			}
			for _,shortPref :=range shortPrefs{
				if strings.Contains(tagList,shortPref.Name){
					rankInfo.AddRecommend("shortPrefWeight", 1+shortPref.Score)
				}
			}
		}

	}
	return nil
}

//指定标签提权
func AssignTagAddWeight(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	abtest := ctx.GetAbTest()
	if dataInfo.MomentCache != nil {
		tagList := dataInfo.MomentCache.MomentsExt.TagList
		if len(tagList)>0{
			assignTag := abtest.GetString("assign_tag", "_")
			if strings.Contains(tagList, assignTag) {
				rankInfo.AddRecommend("AssignTagWeight", 1.1)
			}
		}

	}
	return nil
}
//运营后台配置提权
func EditTagWeight(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	abtest := ctx.GetAbTest()
	editTag :=abtest.GetString("edit_tags","")
	if dataInfo.MomentOfflineProfile!=nil{
		tags :=dataInfo.MomentOfflineProfile.AiTag
		if len(tags)>0&&len(editTag)>1{
			for _,nameMap :=range tags{
				if strings.Contains(editTag,nameMap.Name){
					rankInfo.AddRecommend("EditTagWeight", 1.1)
				}
			}
		}
	}
	return nil
}
//附近日志详情页视频日志提权
func VideoMomWeight(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	if dataInfo.MomentCache != nil {
		if dataInfo.MomentCache.MomentsType == "video" {
			rankInfo.AddRecommend("VideoMomWeight", 1.2)
		}
	}
	return nil
}

//推荐日志偏好提权策略
func DoPrefWeightLevel(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	userInfo := ctx.GetUserInfo().(*UserInfo)
	if dataInfo.MomentCache!= nil {
		tagList := dataInfo.MomentCache.MomentsExt.TagList
		if userInfo.MomentUserProfile!=nil&&len(userInfo.MomentUserProfile.UserPref) > 0 {
			for _, tag := range userInfo.MomentUserProfile.UserPref {
				if strings.Contains(tagList, tag) {
					rankInfo.AddRecommend("UserTagPref", 1.1)
				}
			}
		}
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
	labelScore := ctx.GetAbTest().GetFloat("label_score", 1.2)
	if dataInfo.MomentCache.MomentsExt.TagList!=""{
		rankInfo.AddRecommend("labelMomWeight",labelScore)
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
