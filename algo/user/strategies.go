package user

import (
	rutils "rela_recommend/utils"
	"rela_recommend/algo"
)


func SortWithDistanceItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	request := ctx.GetRequest()
	dataInfo := iDataInfo.(*DataInfo)
	dataLocation := dataInfo.UserCache.Location

	distance := rutils.EarthDistance(float64(request.Lng), float64(request.Lat), dataLocation.Lon, dataLocation.Lat)
	rankInfo.Score = -float32(distance)
	return nil
}
