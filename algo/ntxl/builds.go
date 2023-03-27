package ntxl

import (
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/search"
	"rela_recommend/service/performs"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	params := ctx.GetRequest()
	pf := ctx.GetPerforms()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)

	dataIds := make([]int64, 0)
	userSearchMap := make(map[int64]*search.UserResDataItem, 0)
	pf.Run("search_nearby", func(*performs.Performs) interface{} {
		var searchErr error
		dataIds, userSearchMap, searchErr = search.CallNearUserIdList(params.UserId, params.Lat, params.Lng,
			0, 2000, "")
		if searchErr != nil {
			return searchErr
		} else {
			return len(dataIds)
		}
	})

	var user *redis.UserProfile
	var usersMap = map[int64]*redis.UserProfile{}
	pf.RunsGo("caches", map[string]func(*performs.Performs) interface{}{
		"user": func(*performs.Performs) interface{} {
			var userCacheErr error
			user, usersMap, userCacheErr = userCache.QueryByUserAndUsersMap(params.UserId, dataIds)
			if userCacheErr != nil {
				return userCacheErr
			}
			return len(usersMap)
		}})
	pf.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId:    params.UserId,
			UserCache: user,
		}

		// 组装被曝光者信息
		dataList := make([]algo.IDataInfo, 0)
		for dataId, data := range usersMap {
			if data.IsVipHiding() {
				continue
			}
			if data.DataUserCandidateCanRecommend() {
				info := &DataInfo{
					DataId:   dataId,
					RankInfo: &algo.RankInfo{},
				}
				if data.IsActive(1800) {
					info.RankInfo.AddRecommendWithType("active", 1, algo.TypeActive)
				}
				if data.Distance(user.Location.Lat, user.Location.Lon) <= 30000 {
					info.RankInfo.AddRecommendWithType("active", 1.5, algo.TypeNearbyUser)
				}
				dataList = append(dataList, info)
			}
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(dataIds)
		ctx.SetDataList(dataList)

		return len(dataList)
	})

	return err
}
