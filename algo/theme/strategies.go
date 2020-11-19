package theme

import (
	"math"
	"rela_recommend/algo"
	"rela_recommend/algo/base/strategy"
	autils "rela_recommend/algo/utils"
	"rela_recommend/models/behavior"
	"rela_recommend/utils"
	"unicode/utf8"
)

func ItemBehaviorStrategyFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, itembehavior *behavior.UserBehavior, rankInfo *algo.RankInfo) error {
	var err error
	var upperRate float32
	var abTest = ctx.GetAbTest()

	if abTest.GetBool("rich_strategy:behavior:item_new", false) {
		// 使用威尔逊算法估算内容情况：分值大概在0-0.2之间
		listRate := strategy.WilsonScore(itembehavior.GetThemeListExposure(), itembehavior.GetThemeListInteract(), 3)
		infoRate := strategy.WilsonScore(itembehavior.GetThemeListExposure(), itembehavior.GetThemeListInteract(), 8)
		upperRate = float32(listRate*0.6 + infoRate*0.4)

		if upperRate != 0.0 {
			rankInfo.AddRecommend("ItemBehaviorV1", 1.0+upperRate)
		}
	} else {
		var avgExpCount float64 = 20000
		var avgInfCount float64 = 500

		listCountScore, listRateScore, listTimeScore := strategy.BehaviorCountRateTimeScore(
			itembehavior.GetThemeListExposure(), itembehavior.GetThemeListInteract(), avgExpCount, 0, 0, 0)
		infoCountScore, infoRateScore, infoTimeScore := strategy.BehaviorCountRateTimeScore(
			itembehavior.GetThemeDetailExposure(), itembehavior.GetThemeDetailInteract(), avgInfCount, 0, 0, 0)

		upperRate = float32(0.4*listCountScore*listRateScore*listTimeScore +
			0.6*infoCountScore*infoRateScore*infoTimeScore)

		if upperRate != 0.0 {
			rankInfo.AddRecommend("ItemBehavior", 1.0+upperRate)
		}
	}

	return err
}

func UserBehaviorStrategyFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, userbehavior *behavior.UserBehavior, rankInfo *algo.RankInfo) error {
	var err error
	var abTest = ctx.GetAbTest()
	var currTime = float64(ctx.GetCreateTime().Unix())

	if abTest.GetBool("rich_strategy:behavior:user_new", false) {
		if userbehavior != nil {
			// 浏览过的内容使用浏览次数反序排列，3:未浏览过，2：浏览一次，1：浏览2次，0：浏览3次以上
			allBehavior := behavior.MergeBehaviors(userbehavior.GetThemeListExposure(), userbehavior.GetThemeListInteract(),
				userbehavior.GetThemeDetailExposure(), userbehavior.GetThemeDetailInteract())
			if allBehavior != nil {
				rankInfo.Level = int(-math.Min(allBehavior.Count, 5))
			}
		}
	} else {
		if userbehavior != nil {
			var upperRate float32
			var avgExpCount float64 = 2
			var avgInfCount float64 = 1

			listCountScore, _, listTimeScore := strategy.BehaviorCountRateTimeScore(
				userbehavior.GetThemeListExposure(), userbehavior.GetThemeListInteract(),
				avgExpCount, currTime, 18000, 18000)
			infoCountScore, _, infoTimeScore := strategy.BehaviorCountRateTimeScore(
				userbehavior.GetThemeDetailExposure(), userbehavior.GetThemeDetailInteract(),
				avgInfCount, currTime, 36000, 18000)

			// upperRate = - float32(0.4 * listCountScore * listTimeScore + 0.6 * infoCountScore * infoTimeScore)
			upperRate = -float32(math.Max(listCountScore*listTimeScore, infoCountScore*infoTimeScore))

			if upperRate != 0.0 {
				rankInfo.AddRecommend("UserBehavior", 1.0+upperRate)
			}
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
					activities := abTest.GetStrings("rich_strategy:behavior:self_top_actvivties", "theme.hotweek:exposure,theme.hotweek:click")
					behaviors := userbehavior.Gets(activities...)
					lastBehaviorTime = behaviors.LastTime
				}

				if currTime-lastBehaviorTime >= selfTopTime {
					rankInfo.IsTop = 1
				}
			}
		}
	}
	return err
}

//根据历史用户行为短期偏好提权
func UserShortTagWegiht(ctx algo.IContext, index int) error {
	userData := ctx.GetUserInfo().(*UserInfo)
	dataInfo:=ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	tagMapLine :=userData.ThemeUser
	if tagMapLine!=nil && dataInfo.MomentProfile!=nil{
		shortTagList := tagMapLine.AiTag.UserShortTag
		ThemetagList := dataInfo.MomentProfile.Tags
		if shortTagList!=nil && ThemetagList!=nil && len(ThemetagList) > 0 && len(shortTagList) > 0  {
			var score float64 = 0.0
			var count float64 = 0.0
			for _, tag := range ThemetagList {
				//对情感话题和宠物不提权
				if tag.Id!=23 && tag.Id!=7 {
					if tagIdDict,ok :=shortTagList[tag.Id];ok{
						rate :=tagIdDict.TagScore
						score += rate
						count += 1.0
					}

				}
			}
			if count > 0.0 && score > 0.0 {
				avg:=float32(1.0+(score/count))
				rankInfo.AddRecommend("UserShortTagProfile", avg)
			}
		}

	}
	return nil
}
// 根据标签提权
func themeCategItem(ctx algo.IContext, index int ) error {
	data:=ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := data.GetRankInfo()
	if data.MomentProfile!=nil {
		ThemetagList := data.MomentProfile.Tags
		if ThemetagList!=nil && len(ThemetagList) > 0 {
			for _, tag := range ThemetagList {
				if tag.Id!=23 && tag.Id!=7 && tag.Id!=2 && tag.Id!=8 && tag.Id!=9 && tag.Id!=22 && tag.Id!=24 {
					rankInfo.AddRecommend("themeCateg", 1.5)
				}
			}
		}
	}
	return nil
}

// 内容较短，包含关键词的内容沉底
func TextDownStrategyItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	var abTest = ctx.GetAbTest()
	dataInfo := iDataInfo.(*DataInfo)
	if dataInfo != nil && dataInfo.MomentCache != nil && ctx.GetUserInfo() != nil {
		downLen := abTest.GetInt("rich_strategy:text_down:len", 8)
		downWords := abTest.GetStrings("rich_strategy:text_down:words", "对象,加群,骗子")

		text := dataInfo.MomentCache.MomentsText
		if utf8.RuneCountInString(text) < downLen || utils.StringContains(text, downWords) {
			rankInfo.AddRecommend("TextDown", 0.01)
		}
	}
	return nil
}

// 根据用户实时行为偏好，进行的策略
func UserBehaviorInteractStrategyFunc(ctx algo.IContext) error {
	var err error
	var abtest = ctx.GetAbTest()
	var currTime = float64(ctx.GetCreateTime().Unix())
	var userInfo = ctx.GetUserInfo().(*UserInfo)
	if userInfo.UserBehavior != nil {
		userInteract := userInfo.UserBehavior.GetThemeDetailInteract()
		if userInteract.Count > 0 {
			weight := abtest.GetFloat64("user_behavior_interact_weight", 1.0)
			tagMap := userInteract.GetTopCountTagsMap("item_tag", 5)
			// todo 用户实时偏好
			for index := 0; index < ctx.GetDataLength(); index++ {
				dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
				if dataInfo.MomentProfile != nil { // todo 对每个进行提权
					rankInfo := dataInfo.GetRankInfo()
					var score float64 = 0.0
					var count float64 = 0.0
					for _, tag := range dataInfo.MomentProfile.Tags {
						if userTag, ok := tagMap[tag.Id]; ok && userTag != nil && tag.Id != 23  {
							rate := math.Max(math.Min(userTag.Count/userInteract.Count, 1.0), 0.0)
							hour := math.Max(currTime-userTag.LastTime, 0.0) / (60 * 60)
							score += autils.ExpLogit(rate) * math.Exp(-hour)
							count += 1.0
							// log.Debugf("UserBehaviorInteractStrategyFunc:%d,rate:%f,hour:%f,score:%f,count:%f", tag.Id, rate, hour, score, count)
						}
					}
					if count > 0.0 && score > 0.0 {
						var finalScore = float32(1.0 + score/count*weight)
						rankInfo.AddRecommend("UserTagIteract", finalScore)
					}
				}
			}
		}
	}
	return err
}
