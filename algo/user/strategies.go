package user

import (
	"math"
	"rela_recommend/algo"
	"rela_recommend/algo/base/strategy"
	rutils "rela_recommend/utils"
)

// 使用威尔逊算法估算内容情况：分值大概在0-0.2之间
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

// 点击过的内容降权。一小时降50%， 4小时降20%， 12小时降低7%，24小时降低4%
func UserBehaviorClickedDownItemFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	dataInfo := iDataInfo.(*DataInfo)

	if userBehavior := dataInfo.UserBehavior; userBehavior != nil {
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

func SortWithDistanceItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	request := ctx.GetRequest()
	abtest := ctx.GetAbTest()
	dataInfo := iDataInfo.(*DataInfo)
	dataLocation := dataInfo.UserCache.Location

	distance := rutils.EarthDistance(float64(request.Lng), float64(request.Lat), dataLocation.Lon, dataLocation.Lat)
	if abtest.GetString("custom_sort_type", "distance") == "distance" { // 是否按照距离排序
		rankInfo.Level = -int(distance)
	} else { // 安装距离分段排序
		sortWeightType := abtest.GetString("distance_sort_weight_type", "level")
		if sortWeightType == "weight" { // weight:按照权重，10公里为基准
			weight := float32(0.5 * math.Exp(-distance/10000.0))
			rankInfo.AddRecommend("DistanceWeight", 1.0+weight)
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

// 简单的提权策略
func SimpleUpperItemFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	abtest := ctx.GetAbTest()
	dataInfo := iDataInfo.(*DataInfo)

	// 直播用户提权
	liveUpper := abtest.GetFloat("live_upper_score", 1.0)
	if dataInfo.LiveInfo != nil && liveUpper != 1.0 {
		rankInfo.AddRecommend("LiveUpper", liveUpper)
	}
	// 曝光不足提权
	threshold := abtest.GetFloat64("exposure_upper_threshold", 0.0)
	if threshold > 0 && dataInfo.ItemBehavior != nil {
		if dataInfo.ItemBehavior.Count < threshold {
			score := float32((threshold - dataInfo.ItemBehavior.Count) / threshold)
			rankInfo.AddRecommend("ExposureUpper", 1+score*0.2)
		}
	}
	// 其他提权

	return nil
}
