package match

import (
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/rpc/search"
	"rela_recommend/utils"
	"time"

	// "rela_recommend/algo/utils"
	"rela_recommend/models/redis"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	var startTime = time.Now()
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)

	var user *redis.UserProfile
	dataIds := []int64{}

	// 获取用户信息，修正经纬度
	if params.Lat == 0 || params.Lng == 0 {
		user, err = userCache.QueryUserById(params.UserId)
		if err == nil && user != nil {
			params.Lat = float32(user.Location.Lat)
			params.Lng = float32(user.Location.Lon)
		}
	}

	recListKeyFormatter := abtest.GetString("match_recommend_keyformatter", "") // match_recommend_list_v1:%d
	var recMap = utils.SetInt64{}
	if len(recListKeyFormatter) > 5 {
		recIdlist, errRedis := userCache.GetInt64List(params.UserId, recListKeyFormatter)
		if errRedis == nil {
			recMap.AppendArray(recIdlist)
			dataIds = utils.NewSetInt64FromArrays(dataIds, recIdlist).ToList()
		} else {
			log.Warnf("user recommend list is nil or empty!")
		}
	}

	var startSearchTime = time.Now()
	if abtest.GetBool("used_ai_search", false) {
		if !abtest.GetBool("filter_role_name", false) {
			log.Infof("If search user:%+v", user)
			user = nil
		}
		log.Infof("Build call search user:%+v", user)

		searchIds, searchErr := search.CallMatchList(params.UserId, params.Lat, params.Lng, dataIds, user)
		if searchErr == nil {
			dataIds = searchIds
			log.Infof("get searchlist len %d\n, ", len(dataIds))
		} else {
			log.Warnf("search list is err, %s\n", searchErr)
		}
	} else {
		dataIds = utils.NewSetInt64FromArrays(dataIds, params.DataIds).ToList()
	}

	// 获取用户信息
	var startUserTime = time.Now()
	user, usersMap, usersCacheErr := userCache.QueryByUserAndUsersMap(params.UserId, dataIds)
	if usersCacheErr != nil {
		log.Warnf("users cache list is err, %s\n", usersCacheErr)
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
		UserId:       params.UserId,
		UserCache:    user,
		MatchProfile: matchUser}

	backendRecommendScore := abtest.GetFloat("backend_recommend_score", 1.5)
	dataList := make([]algo.IDataInfo, 0)

	for dataId, data := range usersMap {

		// 推荐集加权
		var recommends = []algo.RecommendItem{}
		if recMap.Contains(data.UserId) {
			recommends = append(recommends, algo.RecommendItem{Reason: "RECOMMEND", Score: backendRecommendScore, NeedReturn: true})
		}

		info := &DataInfo{
			DataId:       dataId,
			UserCache:    data,
			MatchProfile: matchUserMap[dataId],
			RankInfo:     &algo.RankInfo{Recommends: recommends},
		}
		dataList = append(dataList, info)
	}
	ctx.SetUserInfo(userInfo)
	ctx.SetDataIds(dataIds)
	ctx.SetDataList(dataList)
	var endTime = time.Now()

	log.Infof("rankid:%s,totallen:%d,cache:%d;recommend:%d,total:%.3f,dataids:%.3f,search:%.3f,user_cache:%.3f,profile_cache:%.3f,build:%.3f\n",
		ctx.GetRankId(), len(dataIds), len(dataList), recMap.Len(),
		endTime.Sub(startTime).Seconds(), startSearchTime.Sub(startTime).Seconds(), startUserTime.Sub(startSearchTime).Seconds(),
		startProfileTime.Sub(startUserTime).Seconds(),
		startBuildTime.Sub(startProfileTime).Seconds(), endTime.Sub(startBuildTime).Seconds())
	return err
}
