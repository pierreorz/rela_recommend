package live

import (
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	"rela_recommend/utils"
	"time"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	var startTime = time.Now()
	params := ctx.GetRequest()

	userCache := pika.NewUserProfileModule(&factory.CacheCluster, &factory.PikaCluster)
	rdsPikaCache := redis.NewLiveCacheModule(nil, &factory.CacheCluster, &factory.PikaCluster)

	var lives []pika.LiveCache
	// 获取主播列表
	liveType := utils.GetInt(params.Params["type"])
	classify := utils.GetInt(params.Params["classify"])
	lives = GetCachedLiveListByTypeClassify(liveType, classify)

	liveIds := make([]int64, len(lives))
	for i, _ := range lives {
		liveIds[i] = lives[i].Live.UserId
	}
	// 获取基础用户画像
	startUserTime := time.Now()
	user, users, err := userCache.QueryByUserAndUsers(params.UserId, liveIds)
	if err != nil {
		log.Warnf("QueryByUserAndUsers err: %s\n", err)
	}
	usersMap := make(map[int64]pika.UserProfile)
	for i, _ := range users {
		usersMap[users[i].UserId] = users[i]
	}
	// 获取刷新用户画像
	startLiveProfileTime := time.Now()
	user2, users2, err2 := rdsPikaCache.QueryLiveProfileByUserAndUsers(params.UserId, liveIds)
	if err2 != nil {
		log.Warnf("redis QueryLiveProfileByUserAndUsers err: %s\n", err2)
	}
	usersMap2 := make(map[int64]redis.LiveProfile)
	for i, _ := range users2 {
		usersMap2[users2[i].UserId] = users2[i]
	}

	// 获取关注信息
	startConcernsTime := time.Now()
	// concerns := make([]int64, 0)
	concerns, err := userCache.QueryConcernsByUser(params.UserId)
	if err != nil {
		log.Warnf("QueryConcernsByUser err: %s\n", err)
	}

	startBuildTime := time.Now()
	livesInfo := make([]algo.IDataInfo, 0)
	for i, _ := range lives {
		liveInfo := LiveInfo{
			UserId:    lives[i].Live.UserId,
			LiveCache: &lives[i],
			UserCache: nil, LiveProfile: nil,
			RankInfo: &algo.RankInfo{}}
		if liveUser, ok := usersMap[lives[i].Live.UserId]; ok {
			liveInfo.UserCache = &liveUser
		}
		if liveUser2, ok := usersMap2[lives[i].Live.UserId]; ok {
			liveInfo.LiveProfile = &liveUser2
		}
		livesInfo = append(livesInfo, &liveInfo)
	}

	userInfo := &UserInfo{
		UserId: user.UserId, UserCache: &user,
		LiveProfile:  &user2,
		UserConcerns: utils.NewSetInt64FromArray(concerns)}

	ctx.SetUserInfo(userInfo)
	ctx.SetDataList(livesInfo)

	var endTime = time.Now()
	log.Infof("rankid %s,type:%s,totallen:%d,backendlen:%d;total:%.3f,live:%.3f,user:%.3f,profile:%.3f,concerns:%.3f,build:%.3f\n",
		ctx.GetRankId(), params.Type, len(lives), len(users),
		endTime.Sub(startTime).Seconds(), startUserTime.Sub(startTime).Seconds(),
		startLiveProfileTime.Sub(startUserTime).Seconds(),
		startConcernsTime.Sub(startLiveProfileTime).Seconds(),
		startBuildTime.Sub(startConcernsTime).Seconds(), endTime.Sub(startBuildTime).Seconds())
	return nil
}
