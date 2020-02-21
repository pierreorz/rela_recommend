package user

import (
	"time"
	"rela_recommend/algo"
	"rela_recommend/algo/live"
	"rela_recommend/log"
	"rela_recommend/factory"
	// "rela_recommend/algo/utils"
	"rela_recommend/models/redis"
	"rela_recommend/models/behavior"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	var startTime = time.Now()
	abtest := ctx.GetAbTest()
	app := ctx.GetAppInfo()
	pf := ctx.GetPerforms()
	params := ctx.GetRequest()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	behaviorCache := behavior.NewBehaviorCacheModule(ctx, &factory.CacheBehaviorRds)
	liveMap := live.GetCachedLiveMap()	// 当前的直播列表

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

	// 获取实时信息
	var startRealTime = time.Now()
	pf.Begin("realtime")
	behaviorModuleName := abtest.GetString("behavior_module_name", app.Module)  // 特征对应的module名称
	userBehaviorMap, userBehaviorErr := behaviorCache.QueryUserBehaviorMap(behaviorModuleName, params.UserId, dataIds)
	itemBehaviorMap, itemBehaviorErr := behaviorCache.QueryItemBehaviorMap(behaviorModuleName, dataIds)
	if userBehaviorErr != nil {
		log.Warnf("user realtime cache user list is err, %s\n", userBehaviorErr)
	}
	if itemBehaviorErr != nil {
		log.Warnf("user realtime cache item list is err, %s\n", itemBehaviorErr)
	}
	pf.End("realtime")

	// 组装用户信息
	var startBuildTime = time.Now()
	pf.Begin("build")
	userInfo := &UserInfo{
		UserId: params.UserId,
		UserCache: user,
		UserProfile: userProfile}

	// 组装被曝光者信息
	dataList := make([]algo.IDataInfo, 0)
	for dataId, data := range usersMap {
		info := &DataInfo{
			DataId: dataId,
			UserCache: data,
			UserProfile: userProfileMap[dataId],
			LiveInfo: liveMap[dataId],
			RankInfo: &algo.RankInfo{},

			UserBehavior: userBehaviorMap[dataId],
			ItemBehavior: itemBehaviorMap[dataId],
		}
		dataList = append(dataList, info)
	}
	ctx.SetUserInfo(userInfo)
	ctx.SetDataIds(dataIds)
	ctx.SetDataList(dataList)
	pf.End("build")
	var endTime = time.Now()

	log.Infof("userId:%d,rankid:%s,totallen:%d,cache:%d;total:%.3f,dataids:%.3f,user_cache:%.3f,profile_cache:%.3f,realtime_cache:%.3f,build:%.3f\n",
			params.UserId, ctx.GetRankId(), len(dataIds), len(dataList),
			endTime.Sub(startTime).Seconds(), startUserTime.Sub(startTime).Seconds(), 
			startProfileTime.Sub(startUserTime).Seconds(), startRealTime.Sub(startProfileTime).Seconds(),
			startBuildTime.Sub(startRealTime).Seconds(), endTime.Sub(startBuildTime).Seconds())
	return err
}
