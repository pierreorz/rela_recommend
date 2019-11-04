package theme

import (
	"time"
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/factory"
	"rela_recommend/rpc/search"
	"rela_recommend/rpc/api"
	// "rela_recommend/models/pika"
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
	rdsPikaCache := redis.NewLiveCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)

	// search list
	var startSearchTime = time.Now()
	dataIdList := params.DataIds
	newIdList := []int64{}
	if  (dataIdList == nil || len(dataIdList) == 0) {
		recListKeyFormatter := abtest.GetString("recommend_list_key", "theme_recommend_list:%d")
		if len(recListKeyFormatter) > 5 {
			dataIdList, err = rdsPikaCache.GetInt64List(params.UserId, recListKeyFormatter)
			if err == nil {
				log.Warnf("theme recommend list is nil, %s\n", err)
			}
			if len(dataIdList) == 0 {
				dataIdList, _ = rdsPikaCache.GetInt64List(-999999999, recListKeyFormatter)
			}
		}
	
		newMomentLen := abtest.GetInt("new_moment_len", 100)
		if newMomentLen > 0 {
			momentTypes := abtest.GetString("new_moment_types", "theme")
			radiusRange := abtest.GetString("new_moment_radius_range", "1000km")
			newMomentOffsetSecond := abtest.GetFloat("new_moment_offset_second", 60 * 60 * 24)
			
			startNewTime := float32(ctx.GetCreateTime().Unix()) - newMomentOffsetSecond
			newIdList, err = search.CallNearMomentList(params.UserId, params.Lat, params.Lng, 0, newMomentLen,
													   momentTypes, startNewTime , radiusRange)
			if err != nil {
				log.Warnf("theme new list error %s\n", err)
			}
		}
	}
	//backend recommend list
	var startBackEndTime = time.Now()
	var recIds, topMap, recMap = []int64{}, map[int64]int{}, map[int64]int{}
	if abtest.GetBool("backend_recommend_switched", true) {	// 等待该接口5.0上线，别忘记配置conf文件
		recIds, topMap, recMap, err = api.CallBackendRecommendMomentList(1)
		if err != nil {
			log.Warnf("backend recommend list is err, %s\n", err)
		}
	}

	// 获取日志内容
	var startMomentTime = time.Now()
	var dataIds = utils.NewSetInt64FromArray(dataIdList).AppendArray(newIdList).AppendArray(recIds).ToList()
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
		
	backendRecommendScore := abtest.GetFloat("backend_recommend_score", 1.2)
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
	log.Infof("rankid %s,totallen:%d,oldlen:%d,searchlen:%d;backendlen:%d;total:%.3f,old:%.3f,search:%.3f,backend:%.3f,moment:%.3f,user:%.3f,build:%.3f\n",
			  ctx.GetRankId(), len(dataIds), len(dataIdList), len(newIdList), len(recIds),
			  endTime.Sub(startTime).Seconds(), startSearchTime.Sub(startTime).Seconds(), 
			  startBackEndTime.Sub(startSearchTime).Seconds(), startMomentTime.Sub(startBackEndTime).Seconds(),
			  startUserTime.Sub(startMomentTime).Seconds(), startBuildTime.Sub(startUserTime).Seconds(),
			  endTime.Sub(startBuildTime).Seconds() )
	return nil
}
