package moment

import (
	"math"
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/algo/base/strategy"
	"rela_recommend/algo/utils"
	"rela_recommend/log"
	"rela_recommend/models/behavior"
	"rela_recommend/models/redis"
	rutils "rela_recommend/utils"
	"sort"
	"strings"
	"time"
)

const(
	Android = "is_android"
	All = "all"
	NotVip = "not_vip"
	IsVip = "is_vip"
	Ios = "is_ios"
	FeedRecPage = "moment.recommend"
)


type momLiveSorter   []momLive
type momLive struct {
	momId int64
	score float64
}

func (a momLiveSorter) Len() int      { return len(a) }
func (a momLiveSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a momLiveSorter) Less(i, j int) bool { // 按照 score , id 倒序
	if a[i].score == a[j].score {
		return a[i].momId>a[j].momId
	}else{
		return a[i].score>a[j].score
	}
}

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



// 标签下日志按照7天内日志优先策略
func MomentLabelDoTimeLevel(ctx algo.IContext, index int) error {
	abtest := ctx.GetAbTest()
	if hourStrategy := abtest.GetInt("DoTimeLevel:time_interval", 168); hourStrategy > 0 {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		hours := int(ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Hours()) / hourStrategy
		rankInfo.Level = -hours
	}
	return nil
}

// 按照时间优先策略
func SortByTimeLevel(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	min := int(ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Minutes())
	rankInfo.Level = -min
	return nil
}

//日志提权策略v2
func DoTimeWeightLevelV2(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	timeLevel := int(ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Hours()) / 2
	if timeLevel <= 12 {
		if timeLevel >= 0 && timeLevel < 2 { //近4个小时提权权重高1.2
			rankInfo.AddRecommend("momentNearTimeWeightV2", 1.2)
		}
		if timeLevel >= 2 && timeLevel < 4 { //4-8小时1.05
			rankInfo.AddRecommend("momentNearTimeWeightV2", 1.05)
		}
		if timeLevel >= 4 {
			rankInfo.AddRecommend("momentNearTimeWeightV2", 1.01)
		}
	}
	return nil
}

//优质用户推荐策略
func BetterUserMomAddWeight(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	if dataInfo.UserCache != nil && dataInfo.UserCache.Grade > 0 {
		if dataInfo.UserCache.Grade < 50 {
			rankInfo.AddRecommend("betterUserWeight", 1+float32(dataInfo.UserCache.Grade)/50*0.2)
		} else {
			rankInfo.AddRecommend("betterUserWeight", 1.2)
		}
	}
	return nil
}

//用户短期偏好提取
func ShortPrefAddWeight(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	tagList := ""
	userInfo := ctx.GetUserInfo().(*UserInfo)
	if userInfo.MomentUserProfile != nil && dataInfo.MomentOfflineProfile != nil {
		shortPrefs := userInfo.MomentUserProfile.AiTag["short"]
		tags := dataInfo.MomentOfflineProfile.AiTag
		if len(tags) > 0 && len(shortPrefs) > 0 {
			for _, tag := range tags {
				tagList += tag.Name
			}
			for _, shortPref := range shortPrefs {
				//对情感恋爱以及宠物的短期偏好不提权
				if strings.Contains(tagList, shortPref.Name) && shortPref.Name != "情感恋爱" && shortPref.Name != "宠物" {
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
		if len(tagList) > 0 {
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
	editTag := abtest.GetString("edit_tags", "")
	if dataInfo.MomentOfflineProfile != nil {
		tags := dataInfo.MomentOfflineProfile.AiTag
		if len(tags) > 0 && len(editTag) > 1 {
			for _, nameMap := range tags {
				if strings.Contains(editTag, nameMap.Name) {
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

//文本日志打上标签进行打散
func TextMomInterval(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	if dataInfo.MomentCache != nil {
		if dataInfo.MomentCache.MomentsType == "text" || dataInfo.MomentCache.MomentsType == "themereply" {
			rankInfo.AddRecommend("textmom", 1.0)
		}
	}
	return nil
}

//推荐日志偏好提权策略
func DoPrefWeightLevel(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	userInfo := ctx.GetUserInfo().(*UserInfo)
	if dataInfo.MomentCache != nil {
		tagList := dataInfo.MomentCache.MomentsExt.TagList
		if userInfo.MomentUserProfile != nil && len(userInfo.MomentUserProfile.UserPref) > 0 {
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
func DoTimeWeightLevel(ctx algo.IContext, index int) error {
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
func MomLabelAddWeight(ctx algo.IContext, index int) error {
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	labelScore := ctx.GetAbTest().GetFloat("label_score", 1.1)
	if dataInfo.MomentCache!=nil{
		if dataInfo.MomentCache.MomentsExt.TagList != "" {
			rankInfo.AddRecommend("labelMomWeight", labelScore)
		}
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
func AroundNewUserAddWeightFunc(ctx algo.IContext, index int) error {
	abtest := ctx.GetAbTest()
	dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
	rankInfo := dataInfo.GetRankInfo()
	if newUserDefine := abtest.GetInt("new_user_define", 0); newUserDefine > 0 {
		if dataInfo.UserCache != nil {
			hourInterval := int(ctx.GetCreateTime().Sub(dataInfo.UserCache.CreateTime.Time).Hours()) / 24
			if hourInterval < newUserDefine {
				rankInfo.AddRecommend("newUserWeight", 1.2)
			}
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
			rankInfo.AddRecommend("ItemBehaviorV1", 1.0+upperRate)
		}
	} else {
		var avgExpCount float64 = 50

		listCountScore, listRateScore, listTimeScore := strategy.BehaviorCountRateTimeScore(
			itembehavior.GetMomentListExposure(), itembehavior.GetMomentListInteract(), avgExpCount, 0, 0, 0)

		upperRate := float32(listCountScore * listRateScore * listTimeScore)

		if upperRate != 0.0 {
			rankInfo.AddRecommend("ItemBehavior", 1.0+upperRate)
		}
	}
	return err
}

//日志详情页看过沉底策略
func DetailRecommendStrategyFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, userbehavior *behavior.UserBehavior, rankInfo *algo.RankInfo) error {
	var err error
	if userbehavior != nil {
		behaviorCount := behavior.MergeBehaviors(userbehavior.GetMomentListExposure(), userbehavior.GetMomentListInteract())
		if behaviorCount != nil {
			if behaviorCount.Count > 0 {
				rankInfo.Level = int(-math.Min(behaviorCount.Count, 3))
			}
		}
	}
	return err
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

//根据日志的类别来进行相应的提权
func ContentAddWeight(ctx algo.IContext) error {
	var err error
	var abtest = ctx.GetAbTest()
	contentType := abtest.GetStringSet("content_type", "theme,themereply")
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		if dataInfo.MomentCache != nil {
			rankInfo := dataInfo.GetRankInfo()
			if contentType.Contains(dataInfo.MomentCache.MomentsType) {
				rankInfo.AddRecommend("contentTypeWeight", 1.1)
			}
		}
	}
	return err
}

//回流用户日志提权策略
func RecallUserAddWeight(ctx algo.IContext) error {
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if dataInfo.UserCache != nil {
			newRecall := dataInfo.UserCache.Recall
			if newRecall == 1 {
				rankInfo.AddRecommend("RecallUserWeight", 1.2)
			}
		}
	}
	return nil
}

//用户直播画像
func UserLiveWeight(ctx algo.IContext) error {
	userData := ctx.GetUserInfo().(*UserInfo)
	userLiveLongPref := make(map[int64]float32)
	userLiveShortPref := make(map[int64]float32)
	userConsumeLongPref := make(map[int64]float32)
	userConsumeShortPref := make(map[int64]float32)
	if userData.UserLiveProfile != nil {
		userLiveLongPref = userData.UserLiveProfile.LiveLongPref
		userLiveShortPref = userData.UserLiveProfile.LiveShortPref
		userConsumeLongPref = userData.UserLiveProfile.ConsumeLongPref
		userConsumeShortPref = userData.UserLiveProfile.ConsumeShortPref
	}
	for index := 0; index < ctx.GetDataLength(); index++ {
		DataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := DataInfo.GetRankInfo()
		if DataInfo.MomentCache != nil {
			userId := DataInfo.MomentCache.UserId
			if strings.Contains(DataInfo.MomentCache.MomentsType, "live") {
				var score float32 = 0.0
				if w1, ok := userLiveLongPref[userId]; ok {
					score += w1
				}
				if w2, ok := userLiveShortPref[userId]; ok {
					score += w2
				}
				if w3, ok := userConsumeLongPref[userId]; ok {
					score += w3
				}
				if w4, ok := userConsumeShortPref[userId]; ok {
					score += w4
				}
				rankInfo.AddRecommend("UserLiveProFileWeight", 1+score)
			} else {
				continue
			}
		}
	}
	return nil
}



func MomentContentStrategy(ctx algo.IContext) error{
	userInfo :=ctx.GetUserInfo().(*UserInfo)
	abtest := ctx.GetAbTest()
	var max =abtest.GetInt("offline_tag_max",3)
	var recommend1 =0
	var recommend2 =0
	if userInfo.UserContentProfile!=nil{
		momContentPref :=userInfo.UserContentProfile.PicturePref
		for index :=0 ;index< ctx.GetDataLength();index++{
			dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
			rankInfo := dataInfo.GetRankInfo()
			if dataInfo.MomentContentProfile != nil {
				tags := dataInfo.MomentContentProfile.Tags
				if len(tags) >= 1 {
					var score float32 = 0.0
					var personPref float32 = 0.0
					var facePref float32 = 0.0
					for _, tag := range strings.Split(tags, ",") {
						if pref, isOk := momContentPref[tag]; isOk {
							if tag != "人" && tag != "人脸" {
								score += pref
							} else {
								if tag == "人" {
									personPref = pref
								} else {
									facePref = pref
								}
							}
						}
					}
					if score>0{
						rankInfo.AddRecommend("Momcontent",1+0.3*utils.Expit(score))
						recommend1+=1
					}
					if personPref>0&&facePref>0{
						if facePref/personPref>0.95{
							rankInfo.AddRecommend("faceup",1.2)
							recommend2+=1
						}else if facePref/personPref<0.7{
							rankInfo.AddRecommend("facedown",0.8)
						}
					}
				}
			}
			if recommend2>=3&&recommend1>=max{
				break
			}
		}
	}
	return nil
}

// 针对指定categ提权
func MomentCategWeight(ctx algo.IContext) error {
	userData := ctx.GetUserInfo().(*UserInfo)
	abtest := ctx.GetAbTest()
	//后台配置增加曝光内容类型
	editTags := abtest.GetInt64Set("edit_tags_weight", "1,5,6,8,11,12,13,14,15,17,18,19,20,21,22,24,25")
	userTagMap := make(map[string]float64)
	//获取用户日志偏好名和话题偏好名
	if userData.MomentUserProfile != nil {
		momShortPrefs := userData.MomentUserProfile.AiTag["short"]
		for _, shortPref := range momShortPrefs {
			userTagMap[shortPref.Name] = 0.75
		}
	}
	if editTags != nil {
		for index := 0; index < ctx.GetDataLength(); index++ {
			dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
			rankInfo := dataInfo.GetRankInfo()
			if dataInfo.MomentProfile != nil {
				ThemetagList := dataInfo.MomentProfile.Tags
				if len(ThemetagList) > 0 && len(userTagMap) > 0 {
					var score float64 = 0.0
					var count float64 = 0.0
					for _, tag := range ThemetagList {
						if editTags.Contains(tag.Id) {
							if tagScore, ok := userTagMap[tag.Name]; ok {
								score += tagScore
							} else {
								score += 0.6
							}
							count += 1.0
						}
					}
					if count > 0.0 && score > 0.0 {
						avg := float32(1.0 + (score / count))
						rankInfo.AddRecommend("MomentCategWeight", avg)
					}
				}
			}
		}
	}
	return nil
}

//热门日志打压策略
func HotMomentSuppressStrategyFunc(ctx algo.IContext) error {
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if dataInfo.ItemBehavior != nil {
			itemBehavior := dataInfo.ItemBehavior
			if itemBehavior != nil {
				willSonScore := strategy.WilsonScore(itemBehavior.GetMomentListExposure(), itemBehavior.GetMomentListInteract(), 3)
				rankInfo.AddRecommend("willSonSuppressWeight", 1-float32(willSonScore))
			}
		}
	}
	return nil
}

func UserPictureInteractStrategyFunc(ctx algo.IContext) error {
	var err error
	var currTime = float64(ctx.GetCreateTime().Unix())
	var abtest = ctx.GetAbTest()
	var max =abtest.GetInt("user_picture_tag_max",2)
	var userInfo = ctx.GetUserInfo().(*UserInfo)
	if userInfo.UserBehavior != nil {
		userInteract := userInfo.UserBehavior.GetMomentListInteract()
		if userInteract.Count > 0 {
			pictureTagMap := userInteract.GetTopCountPictureTagsMap(5)
			if pictureTagMap != nil && len(pictureTagMap) > 0 {
				var recommend = 0
				for index := 0; index < ctx.GetDataLength(); index++ {
					dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
					if dataInfo.MomentProfile != nil {
						rankInfo := dataInfo.GetRankInfo()
						var score float64 = 0.0
						var count float64 = 0.0
						for _, tag := range dataInfo.MomentProfile.ShuMeiLabels {
							if userTag, ok := pictureTagMap[tag]; ok && userTag != nil {
								if behavior.LabelConvert(tag) != "" {
									rate := math.Max(math.Min(userTag.Count/userInteract.Count, 1.0), 0.0)
									hour := math.Max(currTime-userTag.LastTime, 0.0) / (60 * 60)
									score += utils.ExpLogit(rate) * math.Exp(-hour)
									count += 1.0
								}
							}
						}
						if count > 0.0 && score > 0.0 {
							var finalScore = float32(1.0 + utils.Norm(score/count,0.2))
							rankInfo.AddRecommend("UserPictureTagInteract", finalScore)
							recommend+=1
						}
					}
					if recommend>=max{
						break
					}
				}
			}
		}
	}
	return err
}

func UserMomTagInteractStrategyFunc(ctx algo.IContext) error {
	var err error
	var currTime = float64(ctx.GetCreateTime().Unix())
	var max = 3
	var userInfo = ctx.GetUserInfo().(*UserInfo)
	if userInfo.UserBehavior != nil {
		userInteract := userInfo.UserBehavior.GetMomentListInteract()
		if userInteract.Count > 0 {
			momTagMap := userInteract.GetTopCountMomTagsMap(5)
			if momTagMap != nil && len(momTagMap) > 0 {
				var recommend = 0
				for index := 0; index < ctx.GetDataLength(); index++ {
					dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
					if dataInfo.MomentCache != nil {
						rankInfo := dataInfo.GetRankInfo()
						var score float64 = 0.0
						var count float64 = 0.0
						var tagList = strings.Split(dataInfo.MomentCache.MomentsExt.TagList, ",")
						if tagList != nil && len(tagList) > 0 {
							for _, tag := range tagList {
								if userTag, ok := momTagMap[tag]; ok && userTag != nil {
									rate := math.Max(math.Min(userTag.Count/userInteract.Count, 1.0), 0.0)
									hour := math.Max(currTime-userTag.LastTime, 0.0) / (60 * 60)
									score += utils.ExpLogit(rate) * math.Exp(-hour)
									count += 1.0
								}
							}
							if count > 0.0 && score > 0.0 {
								var finalScore = float32(1.0 + utils.Norm(score/count,0.2))
								rankInfo.AddRecommend("UserMomTagInteract", finalScore)
								recommend += 1
							}
						}
					}
					if recommend >= max {
						break
					}
				}
			}
		}
	}
	return err
}

// 根据用户实时行为偏好，进行的策略
func UserBehaviorInteractStrategyFunc(ctx algo.IContext) error {
	var err error
	var currTime = float64(ctx.GetCreateTime().Unix())
	var userInfo = ctx.GetUserInfo().(*UserInfo)
	if userInfo.UserBehavior != nil {
		userInteract := userInfo.UserBehavior.GetMomentListInteract()
		if userInteract.Count > 0 {
			tagMap := userInteract.GetTopCountTagsMap("item_tag", 5)
			// todo 用户实时偏好
			for index := 0; index < ctx.GetDataLength(); index++ {
				dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
				if dataInfo.MomentProfile != nil { // todo 对每个进行提权
					rankInfo := dataInfo.GetRankInfo()
					var score float64 = 0.0
					var count float64 = 0.0
					for _, tag := range dataInfo.MomentProfile.Tags {
						if userTag, ok := tagMap[tag.Id]; ok && userTag != nil && tag.Id != 23 {
							rate := math.Max(math.Min(userTag.Count/userInteract.Count, 1.0), 0.0)
							hour := math.Max(currTime-userTag.LastTime, 0.0) / (60 * 60)
							score += utils.ExpLogit(rate) *  math.Exp(-hour)
							count += 1.0
							//log.Debugf("UserBehaviorInteractStrategyFunc:%d,rate:%f,hour:%f,score:%f,count:%f,userTag:%s", tag.Id, rate, hour, score, count,userTag)
						}
					}
					if count > 0.0 && score > 0.0 {
						var finalScore = float32(1.0 + utils.Norm(score/count,0.2))
						rankInfo.AddRecommend("UserTagInteract", finalScore)
					}
				}
			}
		}
	}
	return err
}

//附近日志未看过的提权
func NeverSeeStrategyFunc(ctx algo.IContext) error {
	abtest := ctx.GetAbTest()
	exposureNum := abtest.GetFloat64("exposure_num", 1.0)
	weight := abtest.GetFloat("never_see_weight", 0.2)
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if dataInfo.UserItemBehavior == nil || dataInfo.UserItemBehavior.Count < exposureNum {
			if dataInfo.MomentCache != nil && int(ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Hours()) < 13 {
				rankInfo.AddRecommend("NeverSeeWeight", 1+weight)
			}
		}
	}
	return nil
}

//附近日志未被互动提权
func NeverInteractStrategyFunc(ctx algo.IContext) error {
	abtest := ctx.GetAbTest()
	interactNum := abtest.GetFloat64("interact_num", 1.0)
	hour := abtest.GetInt("not_interact_hour", 3)
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if dataInfo.ItemBehavior == nil || dataInfo.ItemBehavior.GetMomentListInteract().Count < interactNum { //
			if dataInfo.MomentCache != nil && int(ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Hours()) < hour {
				rankInfo.AddRecommend("NeverInteractWeight", 1.2)
			}
		}
	}
	return nil
}

//活动提权策略仅针对白名单
func increaseEventExpose(ctx algo.IContext) error {
	abtest := ctx.GetAbTest()
	icpSwitch := abtest.GetBool("icp_switch", false)
	icpWhite := abtest.GetBool("icp_white", false)
	userInfo := ctx.GetUserInfo().(*UserInfo)
	params := ctx.GetRequest()
	if userInfo.UserCache != nil {
		if icpSwitch && (userInfo.UserCache.MaybeICPUser(params.Lat, params.Lng) || icpWhite) {
			for index := 0; index < ctx.GetDataLength(); index++ {
				dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
				rankInfo := dataInfo.GetRankInfo()
				if dataInfo.MomentProfile != nil {
					if dataInfo.MomentProfile.IsActivity {
						rankInfo.AddRecommend("icpActivityWeight", 1.2)
					}
				}
			}
		}
	}
	return nil
}

func adLocationAroundExposureThresholdItemFunc(ctx algo.IContext) error {
	var ua = ctx.GetRequest().GetUa()  //ios,android,other

	userInfo := ctx.GetUserInfo().(*UserInfo)
	var isVip=0
	if userInfo!=nil{
		isVip=userInfo.UserCache.IsVip
	}
	log.Warnf("vip type %s",ctx.GetRequest().GetUa())
	log.Warnf("version %s",ctx.GetRequest().GetVersion())
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if adLocation := dataInfo.MomentCache.MomentsExt.AdLocation; adLocation != nil {
			if aroundAd := adLocation.MomentAround; aroundAd != nil {
				var userType = dataInfo.MomentCache.MomentsExt.UserType
				var jumpType = dataInfo.MomentCache.MomentsExt.JumpType
				if (userType==IsVip&&isVip==1)||(userType==NotVip&&isVip==0)||(strings.Contains(userType,ua))||(userType==All)||(userType==""){
					userBehavior := dataInfo.UserItemBehavior
					var count = 0.0
					if userBehavior != nil {
						count = userBehavior.GetAroundExposure().Count
					}
					log.Warnf("can exposure %s",AdCanExposure(ctx, aroundAd, count,jumpType))

					if AdCanExposure(ctx, aroundAd, count,jumpType) {
						if aroundAd.Index==0{
							rankInfo.IsTop=1
						}else{
							rankInfo.HopeIndex=aroundAd.Index
						}
					}
				}

			}
		}

	}

	return nil
}



//func adLocationAroundExposureThresholdFunc(ctx algo.IContext) error {
//	var adMapIndex = make(map[int64]int, 0)
//	var indexMapAd = make(map[int]int64, 0)
//	for index := 0; index < ctx.GetDataLength(); index++ {
//		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
//		if adLocation := dataInfo.MomentCache.MomentsExt.AdLocation; adLocation != nil {
//			if val := adLocation.MomentAround; val != nil {
//				userBehavior := dataInfo.UserItemBehavior
//				if userBehavior != nil {
//					if AdCanExposure(ctx, val, userBehavior.GetAroundExposure().Count) {
//						if _, ok := indexMapAd[val.Index]; ok {
//							var adIndex = val.Index
//							for {
//								adIndex += 1
//								if _, ok := indexMapAd[adIndex]; !ok {
//									indexMapAd[adIndex] = dataInfo.MomentCache.Id
//									break
//								}
//							}
//						} else {
//							indexMapAd[val.Index] = dataInfo.MomentCache.Id
//						}
//					}
//				}
//			}
//		}
//	}
//	if len(indexMapAd) > 0 {
//		for index, momId := range indexMapAd {
//			adMapIndex[momId] = index
//		}
//		for index := 0; index < ctx.GetDataLength(); index++ {
//			dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
//			rankInfo := dataInfo.GetRankInfo()
//			if _, ok := adMapIndex[dataInfo.MomentCache.Id]; ok {
//				rankInfo.HopeIndex = adMapIndex[dataInfo.MomentCache.Id]
//			}
//		}
//	}
//	return nil
//}

func adLocationRecExposureThresholdFunc(ctx algo.IContext) error {
	var isTop = 0                     //判断是否有置顶日志
	var isSoftTop = 0                 //判断是否有软置顶非直播日志
	var softTopId int64               //最先日志id
	var ua = ctx.GetRequest().GetUa() //ios,android,other
	userInfo := ctx.GetUserInfo().(*UserInfo)
	var isVip = 0
	if userInfo != nil {
		isVip = userInfo.UserCache.IsVip
	}
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if rankInfo.IsTop == 1 {
			isTop = 1
		}
		if rankInfo.IsSoftTop == 1 {
			if dataInfo.UserItemBehavior == nil || dataInfo.UserItemBehavior.GetRecExposure().Count < 1 {
				//if !strings.Contains(dataInfo.MomentCache.MomentsType,"live"){//过滤直播日志
					softTopId = dataInfo.MomentCache.Id
					isSoftTop = 1
				//}
			}
		}
	}
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if dataInfo.MomentCache.Id == softTopId {
			if isTop == 0 {
				rankInfo.IsTop = 1
			} else {
				rankInfo.HopeIndex = 1
			}
		}
		var change = 0
		if adLocation := dataInfo.MomentCache.MomentsExt.AdLocation; adLocation != nil {
			if recAd := adLocation.MomentRecommend; recAd != nil {
				var userType = dataInfo.MomentCache.MomentsExt.UserType
				var jumpType = dataInfo.MomentCache.MomentsExt.JumpType
				if (userType==IsVip&&isVip==1)||(userType==NotVip&&isVip==0)||(strings.Contains(userType,ua))||(userType==All)||(userType==""){
					userBehavior := dataInfo.UserItemBehavior
					var count = 0.0
					if userBehavior != nil {
						count = userBehavior.GetRecExposure().Count
					}
					if AdCanExposure(ctx, recAd, count,jumpType) {
						if recAd.Index < isTop+isSoftTop {
							rankInfo.HopeIndex = isSoftTop + isTop
						} else {
							rankInfo.HopeIndex = recAd.Index
						}
						if isTop == 0 && isSoftTop == 0 && change == 0 {
							if recAd.Index == 0 {
								rankInfo.IsTop = 1
								change += 1
							}
						}
					}
				}
			}
		}

	}

	return nil
}

func softTopAndExposureFunc(ctx algo.IContext) error {
	var isTop = 0
	var softTopList = make([]int64, 0)
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if dataInfo.UserItemBehavior == nil || dataInfo.UserItemBehavior.GetRecExposure().Count < 1 {
			if rankInfo.IsSoftTop == 1 {
				softTopList = append(softTopList, dataInfo.MomentCache.Id)
			}
		}
		if rankInfo.IsTop == 0 {
			continue
		} else {
			isTop = 1
		}
	}
	if len(softTopList) > 0 {
		for index := 0; index < ctx.GetDataLength(); index++ {
			dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
			rankInfo := dataInfo.GetRankInfo()
			if dataInfo.MomentCache.Id == softTopList[0] {
				rankInfo.HopeIndex = isTop
			}
		}
	}

	return nil
}

//晚上9-12点热门直播日志前2名会被放置去指定位置，看过后沉底
func hotLiveHopeIndexStrategyFunc(ctx algo.IContext) error {
	abtest := ctx.GetAbTest()
	lower := abtest.GetInt("top_mom_lower", 21)
	interval := abtest.GetInt("live_interval", 2) //指定位置间隔
	upper := abtest.GetInt("top_mom_upper", 24)
	maxShowLive := abtest.GetInt("max_show_live", 2)     //最多显示top的直播日志条数
	maxSeeTime := abtest.GetFloat64("max_see_time", 1.0) //最大看过次数
	if ctx.GetCreateTime().Hour() >= lower && ctx.GetCreateTime().Hour() <= upper {
		for index := 0; index < ctx.GetDataLength(); index++ {
			dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
			rankInfo := dataInfo.GetRankInfo()
			if dataInfo.UserItemBehavior == nil || dataInfo.UserItemBehavior.Count <= maxSeeTime {
				if rankInfo.LiveIndex > 0 && rankInfo.LiveIndex <= maxShowLive {
					rankInfo.HopeIndex = 1 + interval*(rankInfo.LiveIndex-1) //位置从1开始，间隔interval
				}
			}
		}
	}
	return nil
}

func RecExposureAssignmentsStrategyFunc(ctx algo.IContext) error {
	abtest := ctx.GetAbTest()
	maxNum := abtest.GetFloat64("max_exposure_count", 100.0)
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if dataInfo.ItemBehavior != nil {
			exposureCount := dataInfo.ItemBehavior.GetRecExposure().Count
			if exposureCount > maxNum {
				rankInfo.AddRecommend("max_exposure_down", 0.9)
			}
		}
	}
	return nil
}

func AroundExposureAssignmentsStrategyFunc(ctx algo.IContext) error {
	abtest := ctx.GetAbTest()
	maxNum := abtest.GetFloat64("max_exposure_count", 100.0)
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if dataInfo.ItemBehavior != nil {
			exposureCount := dataInfo.ItemBehavior.GetAroundExposure().Count
			if exposureCount > maxNum {
				rankInfo.AddRecommend("max_exposure_down", 0.9)
			}
		}
	}
	return nil
}

func adHopeIndexStrategyFunc(ctx algo.IContext) error {
	abtest := ctx.GetAbTest()
	adInfo := abtest.GetInt64("ad_moment_id", 0)
	adLocation := abtest.GetInt("ad_location", 1)
	start := abtest.GetInt64("ad_starttime", 1628492400) //活动开始时间
	end := abtest.GetInt64("ad_endtime", 1628956800)     //活动结束时间
	if adInfo != 0 {
		for index := 0; index < ctx.GetDataLength(); index++ {
			dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
			rankInfo := dataInfo.GetRankInfo()
			if dataInfo.MomentCache != nil && dataInfo.MomentCache.Id == adInfo {
				if ctx.GetCreateTime().Unix() >= start && ctx.GetCreateTime().Unix() <= end {
					if rankInfo.IsTop != 1 {
						rankInfo.HopeIndex = adLocation
					}
				}
			}
		}
	}
	return nil
}

func  LiveMomAddWeightFunc(ctx algo.IContext) error{
	liveArrMap := make(map[int64]int, 0)
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		if strings.Contains(dataInfo.MomentCache.MomentsType, "live"){
			liveArrMap[dataInfo.DataId]=1
		}
	}
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if dataInfo.UserItemBehavior == nil || dataInfo.UserItemBehavior.Count < 1 {
			if !strings.Contains(dataInfo.MomentCache.MomentsType, "live"){
				if _,isOk :=liveArrMap[dataInfo.MomentCache.Id];isOk{
					rankInfo.AddRecommend("liveMomAddWeight", 1.1)
				}
			}
		}
	}
	return nil
}

func aroundLiveExposureFunc(ctx algo.IContext) error {
	liveArr := make([]int64, 0)
	liveArrMap := make(map[int64]int, 0)
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		//没有看过的日志
		if dataInfo.UserItemBehavior == nil || dataInfo.UserItemBehavior.Count < 1 {
			if strings.Contains(dataInfo.MomentCache.MomentsType, "live") && rankInfo.IsSoftTop == 0 && rankInfo.IsTop == 0 { ////直播日志且非置顶日志且非软置顶日志
				liveArr = append(liveArr, dataInfo.DataId)
			}

		}
	}
	//对每个数组打散
	if len(liveArr)>0{
		liveArrMap = Shuffle(liveArr)
		for index := 0; index < ctx.GetDataLength(); index++ {
			dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
			rankInfo := dataInfo.GetRankInfo()
			if sortIndex, ok := liveArrMap[dataInfo.DataId]; ok { //运营推荐主播每隔5位随机进行展示
				rankInfo.HopeIndex = sortIndex*6 + GenerateRangeNum(1, 7)
			}
		}
	}
	return nil
}

func liveRecommendStrategyFunc(ctx algo.IContext) error{
	userInfo := ctx.GetUserInfo().(*UserInfo)
	abtest := ctx.GetAbTest()
	haveSoft :=0
	interval :=abtest.GetInt("live_interval_index",7)
	w1 :=0.0
	w2 :=0.0
	w3 :=0.0
	w4 :=0.0
	w5 :=0.0
	w6 :=0.0
	w7 :=0.0
	label :=0
	blindWeight :=abtest.GetFloat64("blind_weight",0.2)
	var res =momLiveSorter{}
	sortIds :=make(map[int64]int,0)
	if userInfo.UserCache!=nil{
		hourInterval := int(ctx.GetCreateTime().Sub(userInfo.UserCache.CreateTime.Time).Hours()) / 24
		if hourInterval<=1||userInfo.UserCache.Age>=30||!(userInfo.UserCache.InChina()){
			label=1
		}

	}
	for index := 0; index < ctx.GetDataLength(); index++ {
		w6 = 0
		w7 = 0
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if dataInfo.UserItemBehavior == nil || dataInfo.UserItemBehavior.Count < 1 {
			if strings.Contains(dataInfo.MomentCache.MomentsType, "live")&&rankInfo.IsSoftTop ==1&&haveSoft==0{
				haveSoft=1
			}
			if strings.Contains(dataInfo.MomentCache.MomentsType, "live") && rankInfo.IsTop == 0 &&dataInfo.MomentCache!=nil&&rankInfo.IsSoftTop ==0 { //非置顶直播日志  //非软置顶直播日志
			    var mom momLive
			    mom.momId = dataInfo.MomentCache.Id

				if live := dataInfo.LiveContentProfile; live != nil {//必须有主播相关的画像
					//新用户不管
					if dataInfo.MomentCache.MomentsType=="live"{//日志类型得分，视频直播类型占0.65分
						w1 =0.6
					}else{
						w1 =0.4
					}
					w2 =live.LiveContentScore//直播内容得分
					w3 =live.LiveValueScore //直播价值得分
					if user :=userInfo.UserLiveContentProfile;user!=nil{
						if user.WantRole+live.Role==1{//角色属性相关
							w4 =1
						}
						if pref,isOk :=user.UserLivePref[dataInfo.MomentCache.UserId];isOk{
							w5 =pref
						}
					}
				}
				if rankInfo.IsBlindMom==2&&label==1{
					w6=0.2
				}else if rankInfo.IsBlindMom==1{
					w6=0
					w7=1
				}
				score :=utils.Norm(w1,1) *0.3 + utils.Norm(w2,1)*0.2 +utils.Norm(w3,1)*0.1 +utils.Norm(w4,1)*0.1+utils.Norm(w5,1)*0.3+blindWeight*float64(rankInfo.IsBlindMom)*w7+w6
				mom.score=score
				res=append(res, mom)
			}
		}
	}
	sort.Sort(res)
	for index,mom :=range res{
		sortIds[mom.momId] = index
	}
	for index:=0;index<ctx.GetDataLength();index++{
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if sortIndex,ok :=sortIds[dataInfo.DataId];ok{//运营推荐主播每隔5位随机进行展示
			rankInfo.HopeIndex=(sortIndex+haveSoft)*(interval-1)+GenerateRangeNum(1,interval)
		}
	}
	return nil
}

func tagMomStrategyFunc(ctx algo.IContext) error{
	change :=0
	for index :=0;index<ctx.GetDataLength();index++{
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if dataInfo.UserItemBehavior == nil || dataInfo.UserItemBehavior.Count < 1 {
			if rankInfo.IsTagMom==1{
				rankInfo.HopeIndex=1
				change=1
			}//小时关注标签日志召回
			if rankInfo.IsTagMom==2{
				rankInfo.HopeIndex=2+change
			}//离线关注标签日志召回
		}
	}
	return nil
}
func editRecommendStrategyFunc(ctx algo.IContext) error {
	recommendArr :=make([]int64,0)
	recommendArrMap :=make(map[int64]int,0)
	for index :=0;index<ctx.GetDataLength(); index++{
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if dataInfo.UserItemBehavior == nil || dataInfo.UserItemBehavior.Count < 1 {//没有看过的日志
			if strings.Contains(rankInfo.ReasonString(),"RECOMMEND")&&! strings.Contains(dataInfo.MomentCache.MomentsType,"live")&&rankInfo.IsSoftTop==0&&rankInfo.IsTop==0{////如果是运营推荐且不为直播日志且非置顶日志且非软置顶日志
				recommendArr=append(recommendArr,dataInfo.DataId)
			}
		}
	}
	//对每个数组打散
	recommendArrMap = Shuffle(recommendArr)
	for index :=0;index<ctx.GetDataLength();index++{
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if sortIndex,ok :=recommendArrMap[dataInfo.DataId];ok{//运营推荐主播每隔5位随机进行展示
			rankInfo.HopeIndex=sortIndex*5+GenerateRangeNum(1,6)
		}
	}
	return nil
}

//头部主播上线即指定位置曝光推荐页
func topLiveIncreaseExposureFunc(ctx algo.IContext) error {
	abtest := ctx.GetAbTest()
	var startIndex = 1
	interval := abtest.GetInt("top_live_interval", 2) //指定位置间隔
	var liveIndex = 0
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if rankInfo.IsTop == 1 {
			startIndex = 2
			break
		}
	}
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		if dataInfo.UserItemBehavior == nil || dataInfo.UserItemBehavior.Count <= 1 {
			if rankInfo.TopLive == 1 && rankInfo.IsTop != 1&&rankInfo.IsSoftTop!=1 {//不可软置顶以及硬置顶
				rankInfo.HopeIndex = startIndex + interval*liveIndex //位置从1开始，间隔interval
				liveIndex += 1
			}
		}
	}
	return nil
}

func BussinessExposureFunc(ctx algo.IContext) error {
	bussinessIdList := make([]int64, 0)
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		userItemBehavior := dataInfo.UserItemBehavior
		if rankInfo.IsBussiness > 0 {
			if moms := dataInfo.MomentCache; moms != nil {
				if userItemBehavior == nil {
					bussinessIdList = append(bussinessIdList, moms.Id)
				}
			}
		}
	}
	choice := int64(0)
	if len(bussinessIdList) > 0 {
		choice = RandChoiceOne(bussinessIdList)
	}
	//var intNum =rand.Intn(10)
	if choice != 0 {
		for index := 0; index < ctx.GetDataLength(); index++ {
			dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
			rankInfo := dataInfo.GetRankInfo()
			if moms := dataInfo.MomentCache; moms != nil {
				if moms.Id == choice {
					rankInfo.HopeIndex = 4
				}
			}
		}
	}
	return nil
}
func ThemeReplyIndexFunc(ctx algo.IContext) error {
	rand.Seed(time.Now().UnixNano())
	var choice = rand.Intn(9)+1
	var choice1 =[]int64{5,6}
	var choice2=[]int64{2,3}
	var change1 = 0
	var change2 = 0
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		userItemBehavior := dataInfo.UserItemBehavior
		if moms := dataInfo.MomentCache; moms != nil {
			if moms.MomentsType == "themereply" && userItemBehavior == nil&&change1==0{
				if choice<3{
					if strings.Contains(rankInfo.RecommendsString(),"RECOMMEND"){
						if ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Hours()<24{
							rankInfo.HopeIndex=int(RandChoiceOne(choice1))
							change1+=1
						}
					}
				}else if choice>2&&choice<9{
					if !strings.Contains(rankInfo.RecommendsString(),"RECOMMEND"){
						rankInfo.HopeIndex=int(RandChoiceOne(choice1))
						change1+=1
					}
				}
			}
			if moms.MomentsType=="theme" && userItemBehavior==nil &&change2==0{
				if choice<3{
					if strings.Contains(rankInfo.RecommendsString(),"RECOMMEND"){
						if ctx.GetCreateTime().Sub(dataInfo.MomentCache.InsertTime).Hours()<24{
							rankInfo.HopeIndex=int(RandChoiceOne(choice2))
							change2+=1
						}
					}
				}else if choice>2&&choice<9{
					if !strings.Contains(rankInfo.RecommendsString(),"RECOMMEND"){
						rankInfo.HopeIndex=int(RandChoiceOne(choice2))
						change2+=1
					}
				}
			}
		}
	}
	return nil
}

func MaybeTopLive(ctx algo.IContext, user *redis.UserProfile) bool {
	if user.LiveInfo != nil && user.LiveInfo.Status == 1 && (user.LiveInfo.ExpireDate > ctx.GetCreateTime().Unix()) {
		return true
	}
	return false
}

func TestHopIndexStrategyFunc(ctx algo.IContext) error {
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index)
		rankInfo := dataInfo.GetRankInfo()
		if index == 5 {
			rankInfo.HopeIndex = 2
		} else if index == 7 {
			rankInfo.HopeIndex = 10
		} else if index == 10 {
			rankInfo.HopeIndex = 15
		}
	}
	return nil
}

func RandChoiceOne(list []int64) int64 {
	if list != nil && len(list) > 0 {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
		return list[0]
	}
	return 0
}

func Shuffle(list []int64) map[int64]int{
	result :=make(map[int64]int,0)
	if list != nil && len(list) > 0 {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
		for index,item :=range list{
			result[item]=index
		}
	}
	return result
}

func GenerateRangeNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max - min) + min
	return randNum
}

func AddRecommendReasonFunc(ctx algo.IContext) error {
	params := ctx.GetRequest()
	for index := 0; index < ctx.GetDataLength(); index++ {
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		rankInfo := dataInfo.GetRankInfo()
		moms :=dataInfo.MomentExtendCache
		if rankInfo.IsBussiness==1{
			rankInfo.AddRecommendWithType("follow",1,algo.TypeYouFollow)
		}
		if moms!=nil{
			lat :=moms.Lat
			lng :=moms.Lng
			if rutils.EarthDistance(float64(params.Lng), float64(params.Lat), lng, lat)<=30000{//少于30km即为附近的人标签
				if params.Lng>0.0&&params.Lat>0.0&&lng>0.0&&lat>0.0{
					rankInfo.AddRecommendWithType("nearby",1,algo.TypeNearby)
				}
			}
		}
		if dataInfo.ItemOfflineBehavior!=nil{
			if GetMomentLikeNum(dataInfo.ItemOfflineBehavior.PageMap,"moment.recommend:like","moment.friend:like")>300{
				rankInfo.AddRecommendWithType("hot",1,algo.TypeHot)
			}
		}
		//if {
		//	rankInfo.AddRecommendWithType("hot",1,algo.TypeHot)
		//}//判断是否热门
	}
	return nil
}

func getFeedRecExposure(pageMap map[string]int) int{
	result :=0
	if count,ok :=pageMap[FeedRecPage];ok{
		result = count
	}
	return result
}

func GetMomentLikeNum(pageMap map[string]int,names ...string) int{
	result :=0
	for _, name := range names {
		if count,ok :=pageMap[name];ok{
			result+=count
		}
	}
	return result
}