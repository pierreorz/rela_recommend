package user

import (
	"fmt"
	"math"
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/algo/base/strategy"
	rutils "rela_recommend/utils"
)

// ItemBehaviorWilsonItemFunc 使用威尔逊算法估算内容情况：分值大概在0-0.2之间
func ItemBehaviorWilsonItemFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	abtest := ctx.GetAbTest()
	dataInfo := iDataInfo.(*DataInfo)

	itemBehavior := dataInfo.ItemBehavior
	if itemBehavior != nil {
		wilsonScale := abtest.GetFloat64("rich_strategy:wilson_behavior:scale", 2.0)
		upperRate := strategy.WilsonScore(itemBehavior.GetNearbyListExposure(), itemBehavior.GetNearbyListInteract(), wilsonScale)
		rankInfo.AddRecommend("WilsonBehavior", 1.0+float32(upperRate))
	}
	return nil
}

// BehaviorClickedDownItemFunc 点击过的内容降权。一小时降50%， 4小时降20%， 12小时降低7%，24小时降低4%
func BehaviorClickedDownItemFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	dataInfo := iDataInfo.(*DataInfo)

	if userBehavior := dataInfo.UserItemBehavior; userBehavior != nil {
		interactItem := userBehavior.GetNearbyListInteract()
		if interactItem.Count > 0 {
			timeSec := (float64(ctx.GetCreateTime().Unix()) - interactItem.LastTime) / 60.0 / 60.0 // 离最后操作了多少小时
			if timeSec > 0 {
				rankInfo.AddRecommend("ClickedDown", 1.0-float32(1.0/(1.0+timeSec)))
			}
		}
	}
	return nil
}

// ExpoTooMuchDownItemFunc 单用户曝光过多降权
func ExpoTooMuchDownItemFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	dataInfo := iDataInfo.(*DataInfo)

	if userBehavior := dataInfo.UserItemBehavior; userBehavior != nil {
		exposuresItem := userBehavior.GetNearbyListExposure()
		expoThreshold := ctx.GetAbTest().GetFloat64("single_expo_threshold", 3.)
		if exposuresItem.Count >= expoThreshold {
			timeMinute := (float64(ctx.GetCreateTime().Unix()) - exposuresItem.LastTime) / 60
			if timeMinute > 0 {
				decay := rutils.GaussDecay(exposuresItem.Count, 0., expoThreshold, timeMinute)
				rankInfo.AddRecommend("ExpoTooMuchDown", float32(decay))
			}
		}
	}
	return nil
}

func SortWithDistanceItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	request := ctx.GetRequest()
	abtest := ctx.GetAbTest()
	dataInfo := iDataInfo.(*DataInfo)
	dataLocation := dataInfo.UserCache.Location

	distance := rutils.EarthDistance(float64(request.Lng), float64(request.Lat), dataLocation.Lon, dataLocation.Lat)
	if abtest.GetString("custom_sort_type", "distance") == "distance" { // 是否按照距离排序
		rankInfo.Level = -int(distance)
	} else { // 安装距离分段排序
		if randomArea := abtest.GetInt("random_distance_area", 0); (randomArea > 0) && (request.Offset == 0) {
			distance = distance + float64(rand.Intn(randomArea))
		}
		sortWeightType := abtest.GetString("distance_sort_weight_type", "level")
		if sortWeightType == "weight" { // weight:按照权重，10公里为基准
			weight := float32(0.5 * math.Exp(-distance/10000.0))
			rankInfo.AddRecommend("DistanceWeight", 1.0+weight)
		} else if sortWeightType == "weight01" { // weight:按照权重，10公里和100公里为基准，解决远距离权重消失问题
			weight := float32(0.25 * (math.Exp(-distance/10000.0) + math.Exp(-distance/100000.0)))
			rankInfo.AddRecommend("DistanceWeight01", 1.0+weight)
		} else { // 按照阶段
			if distance < 1000 {
				rankInfo.Level = 7
			} else if distance < 3000 {
				rankInfo.Level = 6
			} else if distance < 5000 {
				rankInfo.Level = 5
			} else if distance < 10000 {
				rankInfo.Level = 4
			} else if distance < 30000 {
				rankInfo.Level = 3
			} else if distance < 50000 {
				rankInfo.Level = 2
			} else if distance < 100000 {
				rankInfo.Level = 1
			} else {
				rankInfo.Level = 0
			}
		}

	}
	return nil
}

// SimpleUpperItemFunc 简单的提权策略
func SimpleUpperItemFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	abtest := ctx.GetAbTest()
	dataInfo := iDataInfo.(*DataInfo)

	// 直播用户提权
	liveUpper := abtest.GetFloat("live_upper_score", 1.0)
	if dataInfo.LiveInfo != nil && liveUpper != 1.0 {
		rankInfo.AddRecommend("LiveUpper", liveUpper)
	}
	// 其他提权

	return nil
}

// CoverFaceUpperItem 对有头像的用户进行提权
func CoverFaceUpperItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	matchUser := iDataInfo.(*DataInfo)
	if (matchUser.SearchFields != nil) && (matchUser.SearchFields.CoverHasFace) {
		upperRate := ctx.GetAbTest().GetFloat("cover_face_upper", 0)
		rankInfo.AddRecommend("CoverFaceUpper", 1+upperRate)
	}
	return nil
}

func WeekExposureNoInteractFunc(ctx algo.IContext) error {
	abTest := ctx.GetAbTest()
	userInfo := ctx.GetUserInfo().(*UserInfo)
	nearbyProfile := userInfo.UserProfile

	if nearbyProfile != nil && nearbyProfile.WeekExposures != nil {
		overExposureThreshold := abTest.GetInt("over_exposure_threshold", 10)
		for index := 0; index < ctx.GetDataLength(); index++ {
			dataInfo := ctx.GetDataByIndex(index)
			rankInfo := dataInfo.GetRankInfo()

			dataIDStr := fmt.Sprintf("%d", dataInfo.GetDataId())
			if exposures, ok := nearbyProfile.WeekExposures[dataIDStr]; ok {
				if exposures.Exposures >= overExposureThreshold && exposures.Clicks <= 0 {
					decreaseRatio := abTest.GetFloat("over_decrease_ratio", 0.2)
					rankInfo.AddRecommend("WeekExposureNoInteract", 1-decreaseRatio)
				}
			}
		}
	}

	return nil
}

// NtxlActiveDecayWeightFunc 女通讯录活跃加权
func NtxlActiveDecayWeightFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	dataInfo := iDataInfo.(*DataInfo)

	if userProfile := dataInfo.UserCache; userProfile != nil {
		activeThreshold := ctx.GetAbTest().GetFloat64("user_active_threshold", 15*60)
		if userProfile.IsActive(int64(activeThreshold)) {
			rankInfo.Level = 7
			ratio := rutils.GaussDecay(float64(userProfile.ActiveInSeconds()), 0, activeThreshold, 3600)
			if ratio > 0 {
				rankInfo.AddRecommendWithType("ActiveDecay", float32(ratio), algo.TypeActive)
			}
		}
	}
	return nil
}

// NtxNearbyDecayWeightFunc 女通讯录距离近加权
func NtxNearbyDecayWeightFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	currentLat := float64(ctx.GetRequest().Lat)
	currentLng := float64(ctx.GetRequest().Lng)
	if currentUser := ctx.GetUserInfo(); currentUser != nil {
		dataInfo := iDataInfo.(*DataInfo)
		if userProfile := dataInfo.UserCache; userProfile != nil {
			distance := userProfile.Distance(currentLng, currentLat)
			//log.Debugf("<%d> <%d> distance %.2f", ctx.GetRequest().UserId, dataInfo.DataId, distance)
			if distance <= 3000 {
				rankInfo.AddRecommendWithType("NearbyUser", 1, algo.TypeNearbyUser)
			}
			ratio := rutils.GaussDecay(distance, 0, 3000, 50000)
			if ratio > 0 {
				rankInfo.AddRecommend("DistantDecay", float32(ratio))
			}
		}
	}
	return nil
}

// NtxOnLiveWeightFunc 女通讯录直播加权
func NtxOnLiveWeightFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	dataInfo := iDataInfo.(*DataInfo)
	if liveProfile := dataInfo.LiveInfo; liveProfile != nil {
		rankInfo.AddRecommendWithType("OnLiveUser", 1.05, algo.TypeOnLiveUser)
	}
	return nil
}

// NtxNotSingleDecayFunc 女通讯录非单身降权
func NtxNotSingleDecayFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	dataInfo := iDataInfo.(*DataInfo)
	if userCache := dataInfo.UserCache; userCache != nil {
		if !userCache.IsSingle() {
			rankInfo.AddRecommend("NotSingleDecay", 0.7)
		}
	}
	return nil
}

// NtxlUserPageViewItemFunc 主页访问/被访问过
func NtxlUserPageViewItemFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	dataInfo := iDataInfo.(*DataInfo)

	if userItemBehavior := dataInfo.UserItemBehavior; userItemBehavior != nil {
		interactItem := userItemBehavior.GetUserPageViews()
		if interactItem.Count > 0 {
			score := 1 + math.Log(interactItem.Count/10+1)
			rankInfo.AddRecommendWithType("ViewUserPage", float32(score), algo.TypeVisit)
		}
	}
	if beenUserItemBehavior := dataInfo.BeenUserItemBehavior; beenUserItemBehavior != nil {
		interactItem := beenUserItemBehavior.GetUserPageViews()
		if interactItem.Count > 0 {
			score := 1 + math.Log(interactItem.Count/10+1)
			rankInfo.AddRecommendWithType("BeenViewUserPage", float32(score), algo.TypeVisit)
		}
	}
	return nil
}

// NtxlUserWinkItemFunc 主页挤眼/被挤眼
func NtxlUserWinkItemFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	dataInfo := iDataInfo.(*DataInfo)

	if userItemBehavior := dataInfo.UserItemBehavior; userItemBehavior != nil {
		interactItem := userItemBehavior.GetUserWinks()
		if interactItem.Count > 0 {
			score := 1 + math.Log(interactItem.Count/2+1)
			rankInfo.AddRecommendWithType("WinkUser", float32(score), algo.TypeWink)
		}
	}
	if beenUserItemBehavior := dataInfo.BeenUserItemBehavior; beenUserItemBehavior != nil {
		interactItem := beenUserItemBehavior.GetUserWinks()
		if interactItem.Count > 0 {
			score := 1 + math.Log(interactItem.Count/2+1)
			rankInfo.AddRecommendWithType("BeenWinkUser", float32(score), algo.TypeWink)
		}
	}
	return nil
}

// NtxlMomentInteractItemFunc 日志互动/被互动
func NtxlMomentInteractItemFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	dataInfo := iDataInfo.(*DataInfo)
	interactRatio := ctx.GetAbTest().GetFloat("interact_ratio", 1.0)

	if userItemBehavior := dataInfo.UserItemBehavior; userItemBehavior != nil {
		interactItem := userItemBehavior.GetMomentListInteract()
		if interactItem.Count > 0 {
			score := 1 + math.Log(interactItem.Count/3+1)
			rankInfo.AddRecommendWithType("MomentInteract", interactRatio*float32(score), algo.TypeMomentInteract)
		}
	}
	if beenUserItemBehavior := dataInfo.BeenUserItemBehavior; beenUserItemBehavior != nil {
		interactItem := beenUserItemBehavior.GetMomentListInteract()
		if interactItem.Count > 0 {
			score := 1 + math.Log(interactItem.Count/3+1)
			rankInfo.AddRecommendWithType("BeenMomentInteract", interactRatio*float32(score), algo.TypeMomentInteract)
		}
	}
	return nil
}

// NtxlSendMessageItemFunc 聊天/被聊天降权，女通讯录本身是为了促进聊天，如果双方已经频繁聊天，就没必要在这里曝光了
func NtxlSendMessageItemFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	dataInfo := iDataInfo.(*DataInfo)

	if userItemBehavior := dataInfo.UserItemBehavior; userItemBehavior != nil {
		interactItem := userItemBehavior.GetSendMessages()
		if interactItem.Count > 0 {
			decay := rutils.GaussDecay(interactItem.Count, 0., 0, 3)
			rankInfo.AddRecommend("SendMsgDecay", float32(decay))
		}
	}

	if beenUserItemBehavior := dataInfo.BeenUserItemBehavior; beenUserItemBehavior != nil {
		interactItem := beenUserItemBehavior.GetSendMessages()
		if interactItem.Count > 0 {
			decay := rutils.GaussDecay(interactItem.Count, 0., 0, 3)
			rankInfo.AddRecommend("BeenSendMsgDecay", float32(decay))
		}
	}
	return nil
}
