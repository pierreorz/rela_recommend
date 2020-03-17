package theme

import (
	"time"
	"errors"
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

	themeUserCache := redis.NewThemeCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)

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

	//userInfo := &UserInfo{
	//	UserId: params.UserId,
	//	UserCache: user}

	// 增加新特征


	userid :=make([]int64,0)
	userList :=append(userid, params.UserId)
	userMap,themeUserCacheErr := themeUserCache.QueryThemeUserProfileMap(userList)
	if themeUserCacheErr != nil {
		log.Warnf("themeUserProfile cache list is err, %s\n", themeUserCacheErr)
	}
	userInfo := &UserInfo{
		UserId: params.UserId,
		UserCache: user,
		ThemeUser:userMap[params.UserId]}
	themeMap,themeCacheErr :=themeUserCache.QueryThemeProfileMap(dataIdList)
	if themeCacheErr != nil {
		log.Warnf("themeProfile cache list is err, %s\n", themeCacheErr)
	}


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
				ThemeProfile:themeMap[mom.Moments.Id],
				//ThemeUserCache:userScore[momUser.UserId],
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


// 话题详情页的猜你喜欢
func DoBuildMayBeLikeData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	pf := ctx.GetPerforms()
	rdsPikaCache := redis.NewLiveCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	if len(params.DataIds) == 0 {
		return errors.New("dataIds length must 1")
	} 

	pf.Begin("cache")
	recListKeyFormatter := abtest.GetString("recommend_list_key", "theme_recommend_maybelike_list:%d")
	dataIdList, _ := rdsPikaCache.GetInt64List(params.DataIds[0], recListKeyFormatter)
	pf.End("cache")


	pf.Begin("build")
	userInfo := &UserInfo{UserId: params.UserId}

	dataList := make([]algo.IDataInfo, 0)
	for i, dataId := range dataIdList {
		info := &DataInfo{
			DataId: dataId,
			UserCache: nil,
			MomentCache: nil,
			MomentExtendCache: nil,
			MomentProfile: nil,
			ThemeProfile: nil,
			RankInfo: &algo.RankInfo{Score: float32(-i)},
		}
		dataList = append(dataList, info)
	}
	ctx.SetUserInfo(userInfo)
	ctx.SetDataIds(dataIdList)
	ctx.SetDataList(dataList)

	pf.End("build")
	return err
}