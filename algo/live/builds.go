package live

import (
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/behavior"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/api"
	"rela_recommend/service/performs"
	"rela_recommend/utils"
	"strconv"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	params := ctx.GetRequest()
	pfms := ctx.GetPerforms()
	abtest := ctx.GetAbTest()

	sort :=params.Params["sort"]
	log.Warnf("live interface is %s",sort)
	// userCache := pika.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	rdsPikaCache := redis.NewLiveCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	redisTheCache := redis.NewUserCacheModule(ctx, &factory.CacheRds, &factory.CacheRds)
	behaviorCache := behavior.NewBehaviorCacheModule(ctx)

	var lives []pika.LiveCache
	var liveIds = []int64{}
	var liveQueryIds = []int64{}
	pfms.Run("live", func(*performs.Performs) interface{} { // 获取主播列表
		liveType := utils.GetInt(params.Params["type"])
		classify := utils.GetInt(params.Params["classify"])
		lives = GetCachedLiveListByTypeClassify(liveType, classify)
		for i, _ := range lives {
			liveIds = append(liveIds, lives[i].Live.UserId)
			id, _ := strconv.ParseInt("88888"+strconv.FormatInt(lives[i].Live.UserId, 10), 10, 64)
			liveQueryIds = append(liveQueryIds, id)
		}
		return len(lives)
	})
	var user *redis.UserProfile
	var usersMap = map[int64]*redis.UserProfile{}
	var user2 *redis.LiveProfile
	var usersMap2 = map[int64]*redis.LiveProfile{}
	var concernsSet = &utils.SetInt64{}
	var consumerSet = &utils.SetInt64{}
	var interestSet = &utils.SetInt64{}
	var hourRankMap = map[int64]api.AnchorHourRankInfo{}
	var userBehaviorMap = map[int64]*behavior.UserBehavior{}
	var userLiveContentProfileMap map[int64]*redis.UserLiveContentProfile
	var liveContentProfileMap map[int64]*redis.LiveContentProfile
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
			if userBehaviorErr != nil {
				return userBehaviorErr
			}
			return len(userBehaviorMap)
		},
		"user_interest": func(*performs.Performs) interface{} {
			if interests, interestErr := rdsPikaCache.GetInt64List(params.UserId, "user_interest_offline_%d"); interestErr == nil {
				interestSet = utils.NewSetInt64FromArray(interests)
				return interestSet.Len()
			} else {
				return interestErr
			}
		},
		"concerns": func(*performs.Performs) interface{} { // 获取关注信息
			if abtest.GetBool("live_user_concerns", true) {
				if concerns, conErr := redisTheCache.QueryConcernsByUserV1(params.UserId); conErr == nil {
					concernsSet = utils.NewSetInt64FromArray(concerns)
					return concernsSet.Len()
				} else {
					return conErr
				}
			}
			return nil
		},
		"hour_rank": func(*performs.Performs) interface{} { // 获取小时榜排名
			rankMap, hourRankErr := api.GetHourRankList(params.UserId)
			if hourRankErr == nil {
				hourRankMap = rankMap
				return len(hourRankMap)
			}
			return hourRankErr
		},
		"consume_user": func(*performs.Performs) interface{} {
			if consumer, consumerErr := rdsPikaCache.GetInt64List(params.UserId, "user_consumer:%d"); consumerErr == nil {
				consumerSet = utils.NewSetInt64FromArray(consumer)
				return consumerSet.Len()
			} else {
				return consumerErr
			}
		},
		"user_live_content_profile": func(*performs.Performs) interface{} { //用户关于直播的画像
			var userLiveContentProfileErr error
			userLiveContentProfileMap, userLiveContentProfileErr = userCache.QueryUserLiveContentProfileByIdsMap([]int64{params.UserId})
			return userLiveContentProfileErr
		},
		"live_content_profile": func(*performs.Performs) interface{} {
			var liveContentProfileErr error
			liveContentProfileMap, liveContentProfileErr = userCache.QueryLiveContentProfileByIdsMap(liveIds)
			return liveContentProfileErr
		},
	})
	pfms.Run("build", func(*performs.Performs) interface{} {
		livesInfo := make([]algo.IDataInfo, 0)
		consumer := 0
		if concernsSet.Contains(1) { //高消费用户
			consumer = 1
		}
		for i, _ := range lives {
			liveId := lives[i].Live.UserId
			id, _ := strconv.ParseInt("88888"+strconv.FormatInt(lives[i].Live.UserId, 10), 10, 64)
			liveInfo := LiveInfo{
				UserId:           liveId,
				LiveCache:        &lives[i],
				UserCache:        usersMap[liveId],
				UserItemBehavior: userBehaviorMap[id],
				LiveContentProfile:   liveContentProfileMap[liveId],
				LiveProfile:      usersMap2[liveId],
				LiveData: &LiveData{
					PreHourIndex: hourRankMap[liveId].Index,
					PreHourRank:  hourRankMap[liveId].Rank,
				},
				RankInfo: &algo.RankInfo{}}
			if lives[i].Live.IsWeekStar {
				liveInfo.GetRankInfo().AddRecommendNeedReturn("WEEK_STAR", 1.0)
				liveInfo.LiveData.AddLabel(&labelItem{
					Style: WeekStarLabel,
					NewStyle: newStyle{
						Font:       "",
						Background: "https://static.rela.me/Go5pifQDN4LnBuZzE2NjE0NzkzNjk4NzY=.png",
						Color:      "ffffff",
					},
					Title: multiLanguage{
						Chs: "闪耀周星",
						Cht: "閃耀周星",
						En:  "Weekly Star",
					},
					weight: WeekStarLabelWeight,
					level:  level1,
				})
			}

			if lives[i].Live.IsHoroscopeStar {
				liveInfo.GetRankInfo().AddRecommendNeedReturn("HOROSCOPE_STAR", 1.0)
				liveInfo.LiveData.AddLabel(&labelItem{
					Style: WeekStarLabel,
					NewStyle: newStyle{
						Font:       "",
						Background: "https://static.rela.me/Wz56WeQDN4LnBuZzE2NjE0NzkzNjk4ODU=.png",
						Color:      "ffffff",
					},
					Title: multiLanguage{
						Chs: "星座女神",
						Cht: "星座女神",
						En:  "Goddess",
					},
					weight: HoroscopeLabelWeight,
					level:  level1,
				})
			}

			if lives[i].Live.IsMonthStar {
				liveInfo.GetRankInfo().AddRecommendNeedReturn("MONTH_STAR", 1.0)
				liveInfo.LiveData.AddLabel(&labelItem{
					Style: WeekStarLabel,
					NewStyle: newStyle{
						Font:       "",
						Background: "https://static.rela.me/i75pKtQDN4LnBuZzE2NjE0NzkzNjk4ODE=.png",
						Color:      "ffffff",
					},
					Title: multiLanguage{
						Chs: "王牌主播",
						Cht: "王牌主播",
						En:  "Ace",
					},
					weight: MonthStarLabelWeight,
					level:  level1,
				})
			}

			if lives[i].Live.IsModelStudent {
				liveInfo.GetRankInfo().AddRecommendNeedReturn("MODEL_STUDENT", 1.0)
				liveInfo.LiveData.AddLabel(&labelItem{
					Style: WeekStarLabel,
					Title: multiLanguage{
						Chs: "模范生",
						Cht: "模範生",
						En:  "Model Girl",
					},
					weight: ModalStudentLabelWeight,
					level:  level1,
				})
			}

			livesInfo = append(livesInfo, &liveInfo)
		}

		userInfo := &UserInfo{
			UserId:        user.UserId,
			UserCache:     user,
			UserLiveContentProfile: userLiveContentProfileMap[params.UserId],
			LiveProfile:   user2,
			UserConcerns:  concernsSet,
			UserInterests: interestSet,
			ConsumeUser:   consumer}

		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(liveIds)
		ctx.SetDataList(livesInfo)

		return len(livesInfo)
	})
	return err

}
