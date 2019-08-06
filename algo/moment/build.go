package moment

import (
	"time"
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/factory"
	"rela_recommend/rpc/search"
	"rela_recommend/rpc/api"
	"rela_recommend/models/redis"
	"rela_recommend/utils"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	var startTime = time.Now()
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)

	// search list
	dataIdList := params.DataIds
	if dataIdList == nil || len(dataIdList) == 0 {
		radiusRange := abtest.GetString("radius_range", "50km")
		dataIdList, err = search.CallNearMomentList(params.UserId, params.Lat, params.Lng, 0, 1000, 
												 "text_image,video,text,image", 0.0, radiusRange)
		if err != nil {
			return err
		}
	}

	// backend recommend list
	var startBackEndTime = time.Now()
	var recIds, topMap, recMap = []int64{}, map[int64]int{}, map[int64]int{}
	if abtest.GetBool("backend_recommend_switched", false) {	// 是否开启后台推荐日志
		recIds, topMap, recMap, err = api.CallBackendRecommendMomentList(2)
		if err != nil {
			log.Warnf("backend recommend list is err, %s\n", err)
		}
	}

	// 获取日志内容
	var startMomentTime = time.Now()
	var dataIds = utils.NewSetInt64FromArray(dataIdList).AppendArray(recIds).ToList()
	moms, err := momentCache.QueryMomentsByIds(dataIds)
	userIds := make([]int64, 0)
	if err != nil {
		log.Warnf("moment list is err, %s\n", err)
	} else {
		for _, mom := range moms {
			if mom.Moments != nil {
				userIds = append(userIds, mom.Moments.UserId)
			}
		}
		userIds = utils.NewSetInt64FromArray(userIds).ToList()
	}
	// 获取用户信息
	var startUserTime = time.Now()
	user, usersMap, err := userCache.QueryByUserAndUsersMap(params.UserId, userIds)
	if err != nil {
		log.Warnf("users list is err, %s\n", err)
	}

	var startBuildTime = time.Now()
	userInfo := &UserInfo{
		UserId: params.UserId,
		UserCache: user}

	backendRecommendScore := abtest.GetFloat("backend_recommend_score", 1.1)
	dataList := make([]algo.IDataInfo, 0)
	for _, mom := range moms {
		if mom.Moments != nil && mom.Moments.Id > 0 {
			momUser, _ := usersMap[mom.Moments.UserId]

			// 处理置顶
			var isTop = 0
			if topMap != nil {
				if _, isTopOk := topMap[mom.Moments.Id]; isTopOk {
					isTop = 1
				}
			}

			// 处理推荐
			var recommends = []algo.RecommendItem{}
			if recMap != nil {
				if _, isRecommend := recMap[mom.Moments.Id]; isRecommend {
					recommends = append(recommends, algo.RecommendItem{ Reason: "RECOMMEND", Score: backendRecommendScore, NeedReturn: true })
				}
			}
				
			info := &DataInfo{
				DataId: mom.Moments.Id,
				UserCache: momUser,
				MomentCache: mom.Moments,
				MomentExtendCache: mom.MomentsExtend,
				MomentProfile: mom.MomentsProfile,
				RankInfo: &algo.RankInfo{IsTop: isTop, Recommends: recommends},
			}
			dataList = append(dataList, info)
		}
	}
	ctx.SetUserInfo(userInfo)
	ctx.SetDataIds(dataIds)
	ctx.SetDataList(dataList)
	var endTime = time.Now()
	log.Infof("rankid %s,totallen:%d,searchlen:%d;backendlen:%d;total:%.3f,search:%.3f,backend:%.3f,moment:%.3f,user:%.3f,build:%.3f\n",
			  ctx.GetRankId(), len(dataIds), len(dataIdList), len(recIds),
			  endTime.Sub(startTime).Seconds(), startBackEndTime.Sub(startTime).Seconds(), 
			  startMomentTime.Sub(startBackEndTime).Seconds(),
			  startUserTime.Sub(startMomentTime).Seconds(), startBuildTime.Sub(startUserTime).Seconds(),
			  endTime.Sub(startBuildTime).Seconds() )
	return nil
}
