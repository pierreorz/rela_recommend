package match

import (
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
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

	// 确定候选用户
	dataIds := params.DataIds
	if dataIds == nil || len(dataIds) == 0 {
		log.Warnf("user list is nil or empty!")
	}
	recListKeyFormatter := abtest.GetString("match_recommend_keyformatter", "") // match_recommend_list_v1:%d
	var topMap, recMap = utils.SetInt64{}, utils.SetInt64{}
	if len(recListKeyFormatter) > 5 {
		recIdlist, errRedis := userCache.GetInt64List(params.UserId, recListKeyFormatter)
		if errRedis == nil {
			// 推荐集高分用户置顶ab开关
			topMapKeyFormatter := abtest.GetFloat("match_recommend_top", 0)
			if topMapKeyFormatter != 0 {
				// 判断推荐集长度
				if len(recIdlist) > 1 {
					topMap.Append(recIdlist[0])
					recMap.AppendArray(recIdlist[1:])
				} else {
					topMap.AppendArray(recIdlist)
				}
			} else {
				recMap.AppendArray(recIdlist)
			}
			dataIds = utils.NewSetInt64FromArrays(dataIds, recIdlist).ToList()
		} else {
			log.Warnf("user recommend list is nil or empty!")
		}
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
		UserId:       params.UserId,
		UserCache:    user,
		MatchProfile: matchUser}

	backendRecommendScore := abtest.GetFloat("backend_recommend_score", 1.5)
	dataList := make([]algo.IDataInfo, 0)

	for dataId, data := range usersMap {

		// 推荐集最高分数用户置顶
		var isTop = 0
		if topMap.Contains(data.UserId) {
			isTop = 1
		}

		// 推荐集加权
		var recommends = []algo.RecommendItem{}
		if recMap.Contains(data.UserId) {
			recommends = append(recommends, algo.RecommendItem{Reason: "RECOMMEND", Score: backendRecommendScore, NeedReturn: true})
		}

		info := &DataInfo{
			DataId:       dataId,
			UserCache:    data,
			MatchProfile: matchUserMap[dataId],
			RankInfo:     &algo.RankInfo{IsTop: isTop, Recommends: recommends},
		}
		dataList = append(dataList, info)
	}
	ctx.SetUserInfo(userInfo)
	ctx.SetDataIds(dataIds)
	ctx.SetDataList(dataList)
	var endTime = time.Now()

	log.Infof("rankid:%s,totallen:%d,cache:%d;recommend:%d,total:%.3f,dataids:%.3f,user_cache:%.3f,profile_cache:%.3f,build:%.3f\n",
		ctx.GetRankId(), len(dataIds), len(dataList), recMap.Len(),
		endTime.Sub(startTime).Seconds(), startUserTime.Sub(startTime).Seconds(),
		startProfileTime.Sub(startUserTime).Seconds(),
		startBuildTime.Sub(startProfileTime).Seconds(), endTime.Sub(startBuildTime).Seconds())
	return err
}
