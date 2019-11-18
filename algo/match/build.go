package match

import (
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"time"

	// "rela_recommend/algo/utils"
	"rela_recommend/models/redis"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	var startTime = time.Now()
	// abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)

	// 确定候选用户
	dataIds := params.DataIds
	if dataIds == nil || len(dataIds) == 0 {
		log.Warnf("user list is nil or empty!")
	}

	// 获取用户信息
	var startUserTime = time.Now()
	user, usersMap, userCacheErr := userCache.QueryByUserAndUsersMap(params.UserId, dataIds)
	if userCacheErr != nil {
		log.Warnf("users cache list is err, %s\n", userCacheErr)
	}

	// 获取画像信息
	var startProfileTime = time.Now()
	matchUser, matchUserMap, matchCacheErr := userCache.QueryMatchProfileByUserAndUsersMap(params.UserId, dataIds)
	if matchCacheErr != nil {
		log.Warnf("match profile cache list is err, %s\n", matchCacheErr)
	}

	// 生成数据
	var startBuildTime = time.Now()

	userInfo := &UserInfo{
		UserId:    params.UserId,
		UserCache: user,
		MatchProfile: matchUser,}

	dataList := make([]algo.IDataInfo, 0)
	for dataId, data := range usersMap {
		info := &DataInfo{
			DataId:    dataId,
			UserCache: data,
			MatchProfile: matchUserMap[dataId],
			RankInfo:  &algo.RankInfo{},
		}
		dataList = append(dataList, info)
	}
	ctx.SetUserInfo(userInfo)
	ctx.SetDataIds(dataIds)
	ctx.SetDataList(dataList)
	var endTime = time.Now()

	log.Infof("rankid:%s,totallen:%d,cache:%d;total:%.3f,dataids:%.3f,user_cache:%.3f,profile_cache:%.3f,build:%.3f\n",
		ctx.GetRankId(), len(dataIds), len(dataList),
		endTime.Sub(startTime).Seconds(), startUserTime.Sub(startTime).Seconds(), 
		startProfileTime.Sub(startUserTime).Seconds(),
		startBuildTime.Sub(startProfileTime).Seconds(), endTime.Sub(startBuildTime).Seconds())
	return err
}
