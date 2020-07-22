package moment

import (
	"errors"
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/api"
	"rela_recommend/rpc/search"
	"rela_recommend/utils"
	"time"
	"rela_recommend/models/behavior"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	var startTime = time.Now()
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	app := ctx.GetAppInfo()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	behaviorCache:=behavior.NewBehaviorCacheModule(ctx, &factory.CacheBehaviorRds)
	// search list
	dataIdList := params.DataIds
	recIdList := make([]int64, 0)
	newIdList := make([]int64, 0)
	hotIdList := make([]int64, 0)

	if dataIdList == nil || len(dataIdList) == 0 {
		// 获取推荐日志
		recListKeyFormatter := abtest.GetString("recommend_list_key", "") // moment_recommend_list:%d
		if len(recListKeyFormatter) > 5 {
			recIdList, err = momentCache.GetInt64ListOrDefault(params.UserId, -999999999, recListKeyFormatter)
		}

		// 获取最新日志
		newMomentLen := abtest.GetInt("new_moment_len", 1000)
		// if len(recIdList) == 0 {
		// 	newMomentLen = 1000
		// 	log.Warnf("recommend list is none, using new, pls check!\n")
		// }
		if newMomentLen > 0 {

			momentTypes := abtest.GetString("moment_types", "text_image,video,text,image,theme,themereply")
			radiusArray := abtest.GetStrings("radius_range", "50km")
			newMomentOffsetSecond := abtest.GetFloat("new_moment_offset_second", 60*60*24*30*3)
			newMomentStartTime := float32(ctx.GetCreateTime().Unix()) - newMomentOffsetSecond
			//当附近50km无日志，扩大范围200km,2000km,20000km直至找到日志
			for _, radius := range radiusArray {
				if abtest.GetBool("use_ai_search", true) {
					newIdList, err=search.CallNearMomentListV1(params.UserId, params.Lat, params.Lng, 0, int64(newMomentLen),
						momentTypes, newMomentStartTime, radius)
				} else {
					newIdList, err = search.CallNearMomentList(params.UserId, params.Lat, params.Lng, 0, newMomentLen,
						momentTypes, newMomentStartTime, radius)
				}
				//附近日志数量大于10即停止寻找
				if len(newIdList) > 10 {
					break
				}
			}

			if err != nil {
				return err
			}
		}
	}
	//获取热门日志
	if abtest.GetBool("real_recommend_switched", false) {
		top, _ := behaviorCache.QueryDataBehaviorTop()
		hotIdList = top.GetTopIds(100)
	}
	hotIdMap := utils.NewSetInt64FromArray(hotIdList)
	// backend recommend list
	var startBackEndTime = time.Now()
	var recIds, topMap, recMap = []int64{}, map[int64]int{}, map[int64]int{}
	if abtest.GetBool("backend_recommend_switched", false) { // 是否开启后台推荐日志
		recIds, topMap, recMap, err = api.CallBackendRecommendMomentList(2)
		if err != nil {
			log.Warnf("backend recommend list is err, %s\n", err)
		}
	}
	// 获取日志内容
	var startMomentTime = time.Now()
	var dataIds = utils.NewSetInt64FromArrays(dataIdList, recIdList, newIdList, recIds,hotIdList).ToList()
	behaviorModuleName := abtest.GetString("behavior_module_name", app.Module)  // 特征对应的module名称
	itemBehaviorMap, itemBehaviorErr := behaviorCache.QueryItemBehaviorMap(behaviorModuleName, dataIds)
	if itemBehaviorErr != nil {
		log.Warnf("user realtime cache item list is err, %s\n", itemBehaviorErr)
	}
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
	//获取日志离线画像
	var startMomentOfflineProfileTime = time.Now()
	momOfflineProfileMap, momOfflineProfileErr := momentCache.QueryMomentOfflineProfileByIdsMap(dataIds)
	if momOfflineProfileErr != nil {
		log.Warnf("moment embedding is err,%s\n", momOfflineProfileErr)
	}
	// 获取用户信息
	var startUserTime = time.Now()
	user, usersMap, err := userCache.QueryByUserAndUsersMap(params.UserId, userIds)
	if err != nil {
		log.Warnf("users list is err, %s\n", err)
	}

	//// 获取画像信息
	//var startProfileTime = time.Now()
	//emptyIds := make([]int64, 0)
	//momentUser, _, matchCacheErr := userCache.QueryMatchProfileByUserAndUsersMap(params.UserId,emptyIds)
	//if matchCacheErr != nil {
	//	log.Warnf("match profile cache list is err, %s\n", matchCacheErr)
	//}

	//获取user embedding

	var startEmbeddingTime = time.Now()

	momentUserEmbedding, momentUserEmbeddingMap, embeddingCacheErr := userCache.QueryMomentUserProfileByUserAndUsersMap(params.UserId, userIds)
	if embeddingCacheErr != nil {
		log.Warnf("moment user Embedding cache list is err, %s\n", embeddingCacheErr)
	}

	var startBuildTime = time.Now()
	userInfo := &UserInfo{
		UserId:    params.UserId,
		UserCache: user,
		//MomentProfile: momentUser,
		MomentUserProfile: momentUserEmbedding,
	}
	backendRecommendScore := abtest.GetFloat("backend_recommend_score", 1.2)
	realRecommendScore := abtest.GetFloat("real_recommend_score", 1.2)
	dataList := make([]algo.IDataInfo, 0)
	for _, mom := range moms {
		// 后期搜索完善此条件去除
		if mom.Moments.ShareTo != "all" {
			continue
		}
		if mom.Moments != nil && mom.Moments.Id > 0 {
			momUser, _ := usersMap[mom.Moments.UserId]
			//status=0 禁用用户，status=5 注销用户
			if momUser != nil {
				if momUser.Status == 0 || momUser.Status == 5 {
					continue
				}
			}
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
						recommends = append(recommends, algo.RecommendItem{Reason: "RECOMMEND", Score: backendRecommendScore, NeedReturn: true})
					}
				}
			if hotIdMap != nil{
					if isRecommend := hotIdMap.Contains(mom.Moments.Id); isRecommend {
						recommends = append(recommends, algo.RecommendItem{Reason: "REALHOT", Score: realRecommendScore, NeedReturn: true})
					}
				}
			info := &DataInfo{
				DataId:               mom.Moments.Id,
				UserCache:            momUser,
				MomentCache:          mom.Moments,
				MomentExtendCache:    mom.MomentsExtend,
				MomentProfile:        mom.MomentsProfile,
				MomentOfflineProfile: momOfflineProfileMap[mom.Moments.Id],
				RankInfo:             &algo.RankInfo{IsTop: isTop, Recommends: recommends},
				MomentUserProfile:    momentUserEmbeddingMap[mom.Moments.UserId],
				ItemBehavior: itemBehaviorMap[mom.Moments.Id],
			}
			dataList = append(dataList, info)
		}
	}
	ctx.SetUserInfo(userInfo)
	ctx.SetDataIds(dataIds)
	ctx.SetDataList(dataList)
	log.Infof("final read list dataidlist:%s",dataList)
	var endTime = time.Now()
	log.Infof("rankid %s,totallen:%d,paramlen:%d,reclen:%d,searchlen:%d;backendlen:%d;toplen:%d;total:%.3f,search:%.3f,backend:%.3f,moment:%.3f,user:%.3f,moment_offline_profile:%.3f,embedding_cache:%.3f,build:%.3f\n",
		ctx.GetRankId(), len(dataIds), len(dataIdList), len(recIdList), len(newIdList), len(recIds),len(hotIdList),
		endTime.Sub(startTime).Seconds(), startBackEndTime.Sub(startTime).Seconds(),
		startMomentTime.Sub(startBackEndTime).Seconds(),
		startUserTime.Sub(startMomentTime).Seconds(), startBuildTime.Sub(startUserTime).Seconds(), startBuildTime.Sub(startMomentOfflineProfileTime).Seconds(), startBuildTime.Sub(startEmbeddingTime).Seconds(),
		endTime.Sub(startBuildTime).Seconds())
	return nil
}

// 附近日志详情页推荐
func DoBuildMomentAroundDetailSimData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	recListKeyFormatter := abtest.GetString("around_detail_sim_list_key", "moment.around_sim_momentList:%s")
	dataIdList, err := momentCache.GetInt64ListFromGeohash(params.Lat, params.Lng, 4, recListKeyFormatter)
	if dataIdList == nil || err != nil {
		ctx.SetUserInfo(nil)
		ctx.SetDataIds(dataIdList)
		ctx.SetDataList(make([]algo.IDataInfo, 0))
		return nil
	}
	momOfflineProfileMap, momOfflineProfileErr := momentCache.QueryMomentOfflineProfileByIdsMap(dataIdList)
	if momOfflineProfileErr != nil {
		log.Warnf("moment embedding is err,%s\n", momOfflineProfileErr)
	}
	moms, err := momentCache.QueryMomentsByIds(dataIdList)
	userIds := make([]int64, 0)
	for _, mom := range moms {
		if mom.Moments != nil {
			userIds = append(userIds, mom.Moments.UserId)
		}
	}
	userIds = utils.NewSetInt64FromArray(userIds).ToList()
	user, usersMap, err := userCache.QueryByUserAndUsersMap(params.UserId, userIds)
	if err != nil {
		log.Warnf("users list is err, %s\n", err)
	}
	momentUserEmbedding, momentUserEmbeddingMap, embeddingCacheErr := userCache.QueryMomentUserProfileByUserAndUsersMap(params.UserId, userIds)
	if embeddingCacheErr != nil {
		log.Warnf("moment user Embedding cache list is err, %s\n", embeddingCacheErr)
	}
	userInfo := &UserInfo{UserId: params.UserId,
		UserCache: user,
		//MomentProfile: momentUser,
		MomentUserProfile: momentUserEmbedding}
	dataList := make([]algo.IDataInfo, 0)
	for _, mom := range moms {
		if mom.Moments.ShareTo != "all" {
			continue
		}
		if mom.Moments != nil && mom.Moments.Id > 0 {
			momUser, _ := usersMap[mom.Moments.UserId]
			//status=0 禁用用户，status=5 注销用户
			if momUser != nil {
				if momUser.Status == 0 || momUser.Status == 5 {
					continue
				}
			}

			info := &DataInfo{
				DataId:               mom.Moments.Id,
				UserCache:            momUser,
				MomentCache:          mom.Moments,
				MomentExtendCache:    mom.MomentsExtend,
				MomentProfile:        mom.MomentsProfile,
				MomentOfflineProfile: momOfflineProfileMap[mom.Moments.Id],
				RankInfo:             &algo.RankInfo{},
				MomentUserProfile:    momentUserEmbeddingMap[mom.Moments.UserId],
			}
			dataList = append(dataList, info)
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(dataIdList)
		ctx.SetDataList(dataList)
	}
	return err
}

// 关注页日志详情页推荐
func DoBuildMomentFriendDetailSimData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	if len(params.DataIds) == 0 {
		return errors.New("dataIds length must 1")
	}

	recListKeyFormatter := abtest.GetString("friend_detail_before_list_key", "moment.friend_before_moment:%d")
	momIds, err := momentCache.QueryMomentsByIds(params.DataIds)
	if err != nil {
		return errors.New("follow detail moms data not exists")
	}
	dataIdList := make([]int64, 0)
	if len(momIds) > 0 {
		dataIdList, _ := momentCache.GetInt64List(momIds[0].Moments.UserId, recListKeyFormatter)
		SetData(dataIdList, ctx)
	} else {
		SetData(dataIdList, ctx)
	}
	return err
}

// 推荐页日志详情页推荐
func DoBuildMomentRecommendDetailSimData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	if len(params.DataIds) == 0 {
		return errors.New("dataIds length must 1")
	}

	recListKeyFormatter := abtest.GetString("recommend_detail_sim_list_key", "moment.recommend_sim_momentList:%d")
	dataIdList, err := momentCache.GetInt64ListOrDefault(params.DataIds[0], -999999999, recListKeyFormatter)
	if err == nil {
		SetData(dataIdList, ctx)
	}
	return err
}

func SetData(dataIdList []int64, ctx algo.IContext) error {
	var err error
	params := ctx.GetRequest()
	userInfo := &UserInfo{UserId: params.UserId}
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	moms, err := momentCache.QueryMomentsByIds(dataIdList)
	if err != nil {
		return errors.New("query mom err")
	}
	userIds := make([]int64, 0)
	for _, mom := range moms {
		if mom.Moments != nil {
			userIds = append(userIds, mom.Moments.UserId)
		}
	}
	userIds = utils.NewSetInt64FromArray(userIds).ToList()
	usersMap, err := userCache.QueryUsersMap(userIds)
	dataList := make([]algo.IDataInfo, 0)
	for i, mom := range moms {
		if mom.Moments.ShareTo != "all" {
			continue
		}
		if mom.Moments != nil && mom.Moments.Id > 0 {
			momUser, _ := usersMap[mom.Moments.UserId]
			//status=0 禁用用户，status=5 注销用户
			if momUser != nil {
				if momUser.Status == 0 || momUser.Status == 5 {
					continue
				}
			}

			info := &DataInfo{
				DataId:               mom.Moments.Id,
				UserCache:            nil,
				MomentCache:          nil,
				MomentExtendCache:    nil,
				MomentProfile:        nil,
				MomentOfflineProfile: nil,
				RankInfo:             &algo.RankInfo{Level: i},
				MomentUserProfile:    nil,
			}
			dataList = append(dataList, info)
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(dataIdList)
		ctx.SetDataList(dataList)
	}
	return err
}
