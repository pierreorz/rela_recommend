package user

import (
	rutils "rela_recommend/utils"
	"rela_recommend/algo"
)


func SortWithDistanceItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	request := ctx.GetRequest()
	abtest := ctx.GetAbTest()
	dataInfo := iDataInfo.(*DataInfo)
	dataLocation := dataInfo.UserCache.Location

	distance := rutils.EarthDistance(float64(request.Lng), float64(request.Lat), dataLocation.Lon, dataLocation.Lat)
	if abtest.GetString("custom_sort_type", "distance") == "distance" {  // 是否按照距离排序
		rankInfo.Score = -float32(distance)
	} else {	// 安装距离分段排序
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
	return nil
}
