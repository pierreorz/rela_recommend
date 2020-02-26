package user

import (
	"math"
	rutils "rela_recommend/utils"
	"rela_recommend/algo"
	"rela_recommend/algo/base/strategy"
)

// 使用威尔逊算法估算内容情况：分值大概在0-0.2之间
func ItemBehaviorWilsonItemFunc(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	abtest := ctx.GetAbTest()
	dataInfo := iDataInfo.(*DataInfo)

	itemBehavior := dataInfo.ItemBehavior
	if itemBehavior != nil {
		wilsonScale := abtest.GetFloat64("rich_strategy:wilson_behavior:scale", 2.0)
		upperRate := strategy.WilsonScore(itemBehavior.GetNearbyListExposure(), itemBehavior.GetNearbyListInteract(), wilsonScale)
		rankInfo.AddRecommend("WilsonBehavior", 1.0 + float32(upperRate))
	}
	return nil
}

func SortWithDistanceItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	request := ctx.GetRequest()
	abtest := ctx.GetAbTest()
	dataInfo := iDataInfo.(*DataInfo)
	dataLocation := dataInfo.UserCache.Location

	distance := rutils.EarthDistance(float64(request.Lng), float64(request.Lat), dataLocation.Lon, dataLocation.Lat)
	if abtest.GetString("custom_sort_type", "distance") == "distance" {  // 是否按照距离排序
		rankInfo.Score = -float32(distance)
	} else {	// 安装距离分段排序
		sortWeightType := abtest.GetString("distance_sort_weight_type", "level") 
		if sortWeightType == "weight" {  // weight:按照权重，10公里为基准
			weight := float32(0.5 * math.Exp(- distance / 10000.0))
			rankInfo.AddRecommend("DistanceWeight", 1.0 + weight)
		} else {  // 按照阶段
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
