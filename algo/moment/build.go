package moment

import (
	"errors"
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/behavior"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/api"
	"rela_recommend/rpc/search"
	"rela_recommend/service/performs"
	"rela_recommend/utils"
	"time"
	"rela_recommend/models/pika"
	"rela_recommend/algo/live"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	var startTime = time.Now()
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	preforms := ctx.GetPerforms()
	app := ctx.GetAppInfo()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	behaviorCache := behavior.NewBehaviorCacheModule(ctx, &factory.CacheBehaviorRds)
	// search list
	dataIdList := params.DataIds
	recIdList := make([]int64, 0)
	newIdList := make([]int64, 0)
	hotIdList := make([]int64, 0)
	liveIdList :=make([]int64,0)
	momentTypes := abtest.GetString("moment_types", "text_image,video,text,image,theme,themereply")
	if abtest.GetBool("rec_liveMoments_switch",false){
		var lives []pika.LiveCache

		lives = live.GetCachedLiveListByTypeClassify(-1,-1)

		liveIds := make([]int64, len(lives))
		for i, _ := range lives {
			liveIds[i] = lives[i].Live.UserId
		}
		if len(liveIds)>0{
			liveIdList,err=search.CallLiveMomentList(liveIds)
		}
	}

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
			radiusArray := abtest.GetStrings("radius_range", "50km")
			newMomentOffsetSecond := abtest.GetFloat("new_moment_offset_second", 60*60*24*30*3)
			newMomentStartTime := float32(ctx.GetCreateTime().Unix()) - newMomentOffsetSecond
			if abtest.GetBool("near_liveMoments_switch",false){
				//服务端优化附近日志接口后可以直接将momentstype 添加live以及voice_live
				aroundliveMomIdList, _ := search.CallNearMomentList(params.UserId, params.Lat, params.Lng, 0, 100,
					"live,voice_live", 60*60*24, "50km")
				liveMoms,liveMomerr:=momentCache.QueryMomentsByIds(aroundliveMomIdList)
				if err!=nil{
					log.Warnf("near live mom err,%s\n",liveMomerr)
				}
				nearLiveUserIds:= make([]int64, 0)
				for _,mom :=range liveMoms{
					nearLiveUserIds=append(nearLiveUserIds,mom.Moments.UserId)
				}
				nearLiveUserIds=utils.NewSetInt64FromArray(nearLiveUserIds).ToList()
				liveIdList=ReturnLiveList(nearLiveUserIds,aroundliveMomIdList)
			}
			//当附近50km无日志，扩大范围200km,2000km,20000km直至找到日志
			for _, radius := range radiusArray {
				if abtest.GetBool("use_ai_search", false) {
					newIdList, err = search.CallNearMomentListV1(params.UserId, params.Lat, params.Lng, 0, int64(newMomentLen),
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

	var dataIds = utils.NewSetInt64FromArrays(dataIdList, recIdList, newIdList, recIds, hotIdList,liveIdList).ToList()
	// 过滤审核
	searchMomentMap := map[int64]search.SearchMomentAuditResDataItem{} // 日志推荐，置顶

	searchScenery := "moment"
	if abtest.GetBool("search_audit_switched", false) {
		preforms.Run("search", func(*performs.Performs) interface{} {
			returnedRecommend := abtest.GetBool("search_returned_recommend", false)
			filtedAudit := abtest.GetBool("search_filted_audit", false)
			var searchMomentMapErr error
			searchMomentMap, searchMomentMapErr= search.CallMomentAuditMap(params.UserId, dataIds,
				searchScenery, momentTypes, returnedRecommend, filtedAudit)
			if searchMomentMapErr == nil {
				momentIdSet := utils.SetInt64{}
				for _, searchRes := range searchMomentMap {
					momentIdSet.Append(searchRes.Id)
				}
				dataIds = momentIdSet.ToList()
				return len(searchMomentMap)
			}
			return searchMomentMapErr
		})
	}
	log.Warnf("searchMomMap ,%s\n",searchMomentMap)
	// 获取日志内容
	var startMomentTime = time.Now()
	behaviorModuleName := abtest.GetString("behavior_module_name", app.Module) // 特征对应的module名称
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
			if topType, topTypeOK := searchMomentMap[mom.Moments.Id]; topTypeOK {
				topTypeRes := topType.GetCurrentTopType(searchScenery)
				log.Warnf("types of mom %s,%s\n",topTypeRes,mom.Moments.Id)
				isTop = utils.GetInt(topTypeRes == "TOP")
				if topTypeRes == "RECOMMEND" {
					recommends = append(recommends, algo.RecommendItem{Reason: "RECOMMEND", Score: backendRecommendScore, NeedReturn: true})
				}
			}
			if recMap != nil {
				if _, isRecommend := recMap[mom.Moments.Id]; isRecommend {
					recommends = append(recommends, algo.RecommendItem{Reason: "RECOMMEND", Score: backendRecommendScore, NeedReturn: true})
				}
			}
			if hotIdMap != nil {
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
				ItemBehavior:         itemBehaviorMap[mom.Moments.Id],
			}
			dataList = append(dataList, info)
		}
	}
	ctx.SetUserInfo(userInfo)
	ctx.SetDataIds(dataIds)
	ctx.SetDataList(dataList)
	var endTime = time.Now()
	log.Infof("rankid %s,totallen:%d,paramlen:%d,reclen:%d,searchlen:%d;backendlen:%d;toplen:%d;total:%.3f,search:%.3f,backend:%.3f,moment:%.3f,user:%.3f,moment_offline_profile:%.3f,embedding_cache:%.3f,build:%.3f\n",
		ctx.GetRankId(), len(dataIds), len(dataIdList), len(recIdList), len(newIdList), len(recIds), len(hotIdList),
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
	dataIdList := make([]int64, 0)
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	livemoms, liveMomerr := momentCache.QueryMomentsByIds(params.DataIds)
	if liveMomerr != nil {
		return errors.New("liveMomerr ")
	}
	momsType := livemoms[0].Moments.MomentsType

	if momsType == "live" || momsType == "voice_live" {
		//判断日志是否是直播日志
		var lives []pika.LiveCache
		liveLen := abtest.GetInt("live_moment_len", 3)
		lives = live.GetCachedLiveListByTypeClassify(-1, -1)

		liveIds := make([]int64, len(lives))
		liveMap :=make(map[int64]float64,0)
		for i, _ := range lives {
			if lives[i].Live.UserId != livemoms[0].Moments.UserId {
				//获取直播日志：得分的map
				liveMap[lives[i].Live.UserId] = float64(lives[i].Score)
			}
		}
		//根据score分数得到liveidlist
		liveIds=utils.SortMapByValue(liveMap)
		if len(liveIds) > 0 {
			liveIdList, err := search.CallLiveMomentList(liveIds)
			if liveLen >= len(liveIdList) {
				liveLen = len(liveIdList)
			}
			if err == nil {
				dataIdList = liveIdList[:liveLen]
			}
		}
	} else {
		recListKeyFormatter := abtest.GetString("around_detail_sim_list_key", "moment.around_sim_momentList:%s")
		dataIdList, err = momentCache.GetInt64ListFromGeohash(params.Lat, params.Lng, 4, recListKeyFormatter)
	}

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
				UserCache:            momUser,
				MomentCache:          mom.Moments,
				MomentExtendCache:    mom.MomentsExtend,
				MomentProfile:        mom.MomentsProfile,
				MomentOfflineProfile: momOfflineProfileMap[mom.Moments.Id],
				RankInfo:             &algo.RankInfo{},
				MomentUserProfile:    momentUserEmbeddingMap[mom.Moments.UserId],
			}
			if momsType == "live" || momsType == "voice_live"{
				info.RankInfo.Level=-i
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
	moms,liveMomerr:=momentCache.QueryMomentsByIds(params.DataIds)
	if liveMomerr != nil {
		return errors.New("liveMomerr ")
	}
	momsType:=moms[0].Moments.MomentsType
	if momsType == "live"|| momsType == "voice_live" {
		//判断日志类型
		var lives []pika.LiveCache
		liveLen:=abtest.GetInt("live_moment_len",3)
		lives = live.GetCachedLiveListByTypeClassify(-1,-1)

		liveIds := make([]int64, len(lives))
		liveMap :=make(map[int64]float64,0)
		for i, _ := range lives {
			if lives[i].Live.UserId != moms[0].Moments.UserId {
				//获取用户：分数map
				liveMap[lives[i].Live.UserId] = float64(lives[i].Score)
			}
		}
		liveIds=utils.SortMapByValue(liveMap)
		log.Warnf("live ids %s\n",liveIds)
		if len(liveIds) > 0 {
			liveIdList, err := search.CallLiveMomentList(liveIds)
			if liveLen>=len(liveIdList){
				liveLen=len(liveIdList)
			}
			if err == nil {
				SetData(liveIdList[:liveLen], ctx)
			}
		}
	} else {
		recListKeyFormatter := abtest.GetString("recommend_detail_sim_list_key", "moment.recommend_sim_momentList:%d")
		dataIdList, err := momentCache.GetInt64ListOrDefault(params.DataIds[0], -999999999, recListKeyFormatter)
		if err == nil {
			SetData(dataIdList, ctx)
		}
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
				RankInfo:             &algo.RankInfo{Level: -i},
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

func ReturnLiveList(userIdList,momIdList []int64) []int64{
	//根据userid和momid获取每个用户最新的直播日志
	var idMap = map[int64]int64{}
	res := make([]int64, 0)
	index:=0
	for _,userId:=range userIdList{
		if _, ok := idMap[userId]; ok {
			continue
		}else{
			idMap[userId]=momIdList[index]
			res=append(res,momIdList[index])
		}
		index+=1
	}
	return res
}