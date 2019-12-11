package user

import (
	"time"
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/factory"
	// "rela_recommend/algo/utils"
	"rela_recommend/models/redis"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	var startTime = time.Now()
	// abtest := ctx.GetAbTest()
	pf := ctx.GetPerforms()
	params := ctx.GetRequest()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	
	// 确定候选用户
	dataIds := params.DataIds
	if (dataIds != nil || len(dataIds) == 0) {
		log.Warnf("user list is nil or empty!")
	}

	// 获取用户信息
	var startUserTime = time.Now()
	pf.Begin("user")
	user, usersMap, userCacheErr := userCache.QueryByUserAndUsersMap(params.UserId, dataIds)
	if userCacheErr != nil {
		log.Warnf("users cache list is err, %s\n", userCacheErr)
	}
	pf.End("user")

	// 获取画像信息
	var startProfileTime = time.Now()
	pf.Begin("profile")
	userProfile, userProfileMap, profileCacheErr := userCache.QueryNearbyProfileByUserAndUsersMap(params.UserId, dataIds)
	if profileCacheErr != nil {
		log.Warnf("match profile cache list is err, %s\n", profileCacheErr)
	}
	pf.End("profile")

	var startBuildTime = time.Now()

	pf.Begin("build")
	userInfo := &UserInfo{
		UserId: params.UserId,
		UserCache: user,
		UserProfile: userProfile}

	dataList := make([]algo.IDataInfo, 0)
	for dataId, data := range usersMap {
		info := &DataInfo{
			DataId: dataId,
			UserCache: data,
			UserProfile: userProfileMap[dataId],
			RankInfo: &algo.RankInfo{},
		}
		dataList = append(dataList, info)
	}
	ctx.SetUserInfo(userInfo)
	ctx.SetDataIds(dataIds)
	ctx.SetDataList(dataList)
	pf.End("build")
	var endTime = time.Now()

	log.Infof("userId:%d,rankid:%s,totallen:%d,cache:%d;total:%.3f,dataids:%.3f,user_cache:%.3f,profile_cache:%.3f,build:%.3f\n",
			params.UserId, ctx.GetRankId(), len(dataIds), len(dataList),
			endTime.Sub(startTime).Seconds(), startUserTime.Sub(startTime).Seconds(), 
			startProfileTime.Sub(startUserTime).Seconds(),
			startBuildTime.Sub(startProfileTime).Seconds(), endTime.Sub(startBuildTime).Seconds())
	return err
}
