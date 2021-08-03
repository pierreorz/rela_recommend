package live

import (
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/api"
	"rela_recommend/service/performs"
	"rela_recommend/utils"
	"rela_recommend/models/behavior"
	"strconv"
	"rela_recommend/log"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	params := ctx.GetRequest()
	pfms := ctx.GetPerforms()
	// userCache := pika.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	rdsPikaCache := redis.NewLiveCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	redisTheCache := redis.NewUserCacheModule(ctx, &factory.CacheRds, &factory.CacheRds)
	behaviorCache := behavior.NewBehaviorCacheModule(ctx, &factory.CacheBehaviorRds)


	var lives []pika.LiveCache
	var liveIds = []int64{}
	var liveQueryIds = []int64{}
	pfms.Run("live", func(*performs.Performs) interface{} { // 获取主播列表
		liveType := utils.GetInt(params.Params["type"])
		classify := utils.GetInt(params.Params["classify"])
		lives = GetCachedLiveListByTypeClassify(liveType, classify)

		for i, _ := range lives {
			liveIds = append(liveIds, lives[i].Live.UserId)
			id,err :=strconv.ParseInt("88888"+lives[i].Live.UserIdStr,10,64)
			if err!=nil{
				liveQueryIds=append(liveQueryIds,id)
			}
		}
		return len(lives)
	})

	var user *redis.UserProfile
	var usersMap = map[int64]*redis.UserProfile{}
	var user2 *redis.LiveProfile
	var usersMap2 = map[int64]*redis.LiveProfile{}
	var concernsSet = &utils.SetInt64{}
	var hourRankMap = map[int64]api.AnchorHourRankInfo{}
	var userBehaviorMap = map[int64]*behavior.UserBehavior{}

	pfms.RunsGo("cache", map[string]func(*performs.Performs) interface{}{
		"user": func(*performs.Performs) interface{} { // 获取基础用户画像
			var userErr error
			user, usersMap, userErr = userCache.QueryByUserAndUsersMap(params.UserId, liveIds)
			if userErr == nil {
				return len(usersMap)
			}
			return userErr
		},
		"profile": func(*performs.Performs) interface{} { // 获取刷新用户画像
			var userProfileErr error
			user2, usersMap2, userProfileErr = rdsPikaCache.QueryLiveProfileByUserAndUsersMap(params.UserId, liveIds)
			if userProfileErr == nil {
				return len(usersMap2)
			}
			return userProfileErr
		},
		"realtime_useritem": func(*performs.Performs) interface{} {
			var userBehaviorErr error
			userBehaviorMap, userBehaviorErr = behaviorCache.QueryUserItemBehaviorMap("live", params.UserId, liveQueryIds)
			log.Warnf("user live ids%s",liveQueryIds)
			log.Warnf("userlive behavior%s",userBehaviorMap)
			if userBehaviorErr != nil {
				return userBehaviorErr
			}
			return len(userBehaviorMap)
		},
		"concerns": func(*performs.Performs) interface{} { // 获取关注信息
			if concerns, conErr := redisTheCache.QueryConcernsByUser(params.UserId); conErr == nil {
				concernsSet = utils.NewSetInt64FromArray(concerns)
				return concernsSet.Len()
			} else {
				return conErr
			}
		},
		"hour_rank": func(*performs.Performs) interface{} { // 获取小时榜排名
			rankMap, hourRankErr := api.CallLiveHourRankMap(params.UserId)
			if hourRankErr == nil {
				hourRankMap = rankMap
				return len(hourRankMap)
			}
			return hourRankErr
		},
	})

	pfms.Run("build", func(*performs.Performs) interface{} {
		livesInfo := make([]algo.IDataInfo, 0)
		for i, _ := range lives {
			liveId := lives[i].Live.UserId
			id,_ := strconv.ParseInt("88888"+lives[i].Live.UserIdStr,10,64)
			liveInfo := LiveInfo{
				UserId:      liveId,
				LiveCache:   &lives[i],
				UserCache:   usersMap[liveId],
				UserItemBehavior: userBehaviorMap[id],
				LiveProfile: usersMap2[liveId],
				LiveData: &LiveData{
					PreHourIndex: hourRankMap[liveId].Index,
					PreHourRank:  hourRankMap[liveId].Rank,
				},
				RankInfo: &algo.RankInfo{}}
			if lives[i].Live.IsWeekStar {
				liveInfo.GetRankInfo().AddRecommendNeedReturn("WEEK_STAR", 1.0)
			}
			livesInfo = append(livesInfo, &liveInfo)
		}

		userInfo := &UserInfo{
			UserId:       user.UserId,
			UserCache:    user,
			LiveProfile:  user2,
			UserConcerns: concernsSet}

		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(liveIds)
		ctx.SetDataList(livesInfo)

		return len(livesInfo)
	})
	return err
}
