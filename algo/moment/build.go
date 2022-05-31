package moment

import (
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/algo/live"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/behavior"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/api"
	"rela_recommend/rpc/search"

	"rela_recommend/service/performs"
	"rela_recommend/utils"
	"time"
)


func DoBuildData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	preforms := ctx.GetPerforms()
	app := ctx.GetAppInfo()

	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	behaviorCache := behavior.NewBehaviorCacheModule(ctx)
	// search list
	custom := abtest.GetString("custom_sort_type", "ai")
	dataIdList := params.DataIds
	recIdList := make([]int64, 0)
	hourRecList :=make([]int64, 0)
	autoRecList := make([]int64, 0)
	newIdList := make([]int64, 0)
	hotIdList := make([]int64, 0)
	bussinessIdList := make([]int64, 0)
	tagRecommendIdList := make([]int64, 0)
	adList := make([]int64, 0)
	adLocationList := make([]int64, 0)
	liveMomentIds := make([]int64, 0)
	paiResult :=make(map[int64]float64,0)
	var recIds, topMap, recMap, bussinessMap = []int64{}, map[int64]int{}, map[int64]int{}, map[int64]int{}
	var liveMap = map[int64]int{}
	var expId = ""
	var requestId = ""
	var recall_expId = ""
	momentTypes := abtest.GetString("moment_types", "text_image,video,text,image,theme,themereply")
	topN :=abtest.GetInt("topn",5)
	recall_length :=abtest.GetInt("recall_length",1500)
	topScore :=abtest.GetFloat64("top_score",0.02)
	if abtest.GetBool("rec_liveMoments_switch", false) && custom != "hot" {
		liveMap = live.GetCachedLiveMomentListByTypeClassify(-1, -1)
		liveMomentIds = getMapKey(liveMap)
	}
	adInfo := abtest.GetInt64("ad_moment_id", 0)

	if adInfo != 0 { //广告类型日志
		adList = append(adList, adInfo)
	}
	//icp白名单及新注册用户
	var userTest *redis.UserProfile
	var userTestErr error
	userTest, userTestErr = userCache.QueryUserById(params.UserId)
	var userBehavior *behavior.UserBehavior // 用户实时行为
	if userTestErr != nil {
		log.Warnf("query user icp white err, %s\n", userTestErr)
		//return userTestErr
	}
	icpSwitch := abtest.GetBool("icp_switch", false)
	mayBeIcpUser := (userTest != nil) && userTest.MaybeICPUser(params.Lat, params.Lng)
	icpWhite := abtest.GetBool("icp_white", false)

	if icpSwitch && (mayBeIcpUser || icpWhite) {
		recListKeyFormatter := abtest.GetString("icp_recommend_list_key", "icp_recommend_list:%d") // moment_recommend_list:%d
		//白名单以及杭州新用户默认数据
		if len(recListKeyFormatter) > 5 {
			recIdList, err = momentCache.GetInt64ListOrDefault(-10000000, -999999999, recListKeyFormatter) //icp_recommend_list:-10000000
		}
		if abtest.GetBool("icp_around_rec", false) {
			var errSearch error
			newMomentOffsetSecond := abtest.GetFloat("new_moment_offset_second", 60*60*24*2)
			newMomentStartTime := float32(ctx.GetCreateTime().Unix()) - newMomentOffsetSecond
			newIdList, errSearch = search.CallNearMomentListV2(params.UserId, params.Lat, params.Lng, 0, 1000,
				momentTypes, newMomentStartTime, "50km", true)
			if errSearch != nil {
				log.Warnf("query user around icp err %s\n", errSearch)
				return errSearch
			}
		}
	} else {
		//原代码
		preforms.RunsGo("data", map[string]func(*performs.Performs) interface{}{
			"recommend": func(*performs.Performs) interface{} { // 获取推荐日志
				if dataIdList == nil || len(dataIdList) == 0 {
					recListKeyFormatter := abtest.GetString("recommend_list_key", "") // moment_recommend_list:%d
					if len(recListKeyFormatter) > 5 {
						var userId = params.UserId
						if custom == "hot" {
							userId = -999999999
						}
						recIdList, err = momentCache.GetInt64ListOrDefault(userId, -999999999, recListKeyFormatter)
						recall_expId=utils.RecallOwn
					}else{
						recIdList,recall_expId,_,err= api.GetRecallResult(params.UserId,recall_length)
					}
					return len(recIdList)
				}
				return nil
			},"hour": func(*performs.Performs) interface{}{
				if abtest.GetBool("hour_rec_moment",false){
					hourRecList,err =momentCache.GetInt64ListOrDefault(params.UserId,-999999999,"hour_recommend_list:%d")
					return len(hourRecList)
				}
				return nil
			},"new": func(*performs.Performs) interface{} { // 新日志 或 附近日志
				newMomentLen := abtest.GetInt("new_moment_len", 1000) //不为0即推荐添加实时日志
				if newMomentLen > 0 {
					radiusArray := abtest.GetStrings("radius_range", "50km")
					newMomentOffsetSecond := abtest.GetFloat("new_moment_offset_second", 60*60*24*30*3)
					newMomentStartTime := float32(ctx.GetCreateTime().Unix()) - newMomentOffsetSecond
					recommended := abtest.GetBool("realtime_mom_switch", false) // 是否过滤推荐审核
					if abtest.GetBool("near_liveMoments_switch", false) && abtest.GetBool("search_switched_around", true) {
						var lives []pika.LiveCache
						lives = live.GetCachedLiveListByTypeClassify(-1, -1)
						distance := abtest.GetFloat64("live_distance", 50.0)
						liveMomentIds = ReturnAroundLiveMom(lives, params.Lng, params.Lat, distance)
					}
					//当附近50km无日志，扩大范围200km,2000km,20000km直至找到日志
					var errSearch error
					if abtest.GetBool("search_switched_around", true) { //附近日志搜索开关，关闭则走兜底数据
						for _, radius := range radiusArray {
							//if abtest.GetBool("use_ai_search", false) {
							//
							//} else {
							//	newIdList, errSearch = search.CallNearMomentList(params.UserId, params.Lat, params.Lng, 0, newMomentLen,
							//		momentTypes, newMomentStartTime, radius)
							//}
							newIdList, errSearch = search.CallNearMomentListV1(params.UserId, params.Lat, params.Lng, 0, int64(newMomentLen),
								momentTypes, newMomentStartTime, radius, recommended)
							//附近日志数量大于10即停止寻找
							if len(newIdList) > 10 {
								break
							}
						}
					} else {
						recListKeyFormatter := abtest.GetString("around_list_key", "moment.around_list_data:%s")
						newIdList, errSearch = momentCache.GetInt64ListFromGeohash(params.Lat, params.Lng, 4, recListKeyFormatter)
					}
					if errSearch != nil {
						return err
					}
					return len(newIdList)
				}
				return nil
			}, "hot": func(*performs.Performs) interface{} { // 热门列表
				if abtest.GetBool("real_recommend_switched", false) {
					if top, topErr := behaviorCache.QueryDataBehaviorTop(app.Module); topErr == nil {
						if abtest.GetInt("hot_method",1)==1{
							hotIdList = top.GetTopIds(topN)
						}else{
							hotIdList = top.GetTopIdsV2(topScore)
						}
						return len(hotIdList)
					} else {
						return topErr
					}
				}
				return nil
			}, "backend": func(*performs.Performs) interface{} { // 管理后台配置推荐列表
				var errBackend error
				if abtest.GetBool("backend_recommend_switched", false) { // 是否开启后台推荐日志
					recIds, topMap, recMap, errBackend = api.CallBackendRecommendMomentList(2)
					if errBackend == nil {
						return len(recIds)
					}
				}
				return errBackend
			}, "bussiness": func(*performs.Performs) interface{} { // 业务推荐id列表
				var errBussiness error
				if abtest.GetBool("bussiness_recommend_switched", false) { // 是否开启业务推荐
					bussinessIdList, errBussiness = momentCache.GetInt64ListOrDefault(params.UserId, -9999999, "bussiness_rec_moment_data:%d")
					if len(bussinessIdList) > 0 {
						for _, id := range bussinessIdList {
							bussinessMap[id] = 1
						}
					}
					if errBussiness == nil {
						return len(bussinessIdList)
					}
				}
				return errBussiness
			}, "adLocation": func(*performs.Performs) interface{} {
				var adLocationSearchErr error
				if abtest.GetBool("adLocation_ad", true) {
					adLocationList, adLocationSearchErr = search.CallAdMomentListV1(params.UserId)
				}
				return adLocationSearchErr
			}, "better_user": func(*performs.Performs) interface{} {
				var errBetterUser error
				autoKeyFormatter := "better_user_mom_yesterday:%d"
				if abtest.GetBool("auto_recommend_switch", false) {
					autoRecList, errBetterUser = momentCache.GetInt64ListOrDefault(-999999999, -999999999, autoKeyFormatter)
					if num := abtest.GetInt("auto_recommend_random_num", 0); num > 0 && errBetterUser == nil && len(autoRecList) > 0 { //随机挑选num个优质用户日志
						rand.Seed(time.Now().UnixNano())
						rand.Shuffle(len(autoRecList), func(i, j int) { autoRecList[i], autoRecList[j] = autoRecList[j], autoRecList[i] })
						if len(autoRecList) < num {
							num = len(autoRecList)
						}
						autoRecList = autoRecList[:num-1]
					}
					if errBetterUser == nil {
						return len(autoRecList)
					}
				}
				return errBetterUser
			}, "user_behavior": func(*performs.Performs) interface{} { // 获取实时操作的内容
				realtimes, realtimeErr := behaviorCache.QueryUserBehaviorMap(app.Module, []int64{params.UserId})
				if realtimeErr == nil && abtest.GetInt("rich_strategy:user_behavior_interact:weight", 0) == 1 {
					userBehavior = realtimes[params.UserId]
					if userBehavior != nil {
						userInteract := userBehavior.GetMomentListInteract()
						if userInteract.Count > 0 {
							//获取用户实时互动日志的各个标签的实时热门数据
							tagMap := userInteract.GetTopCountTagsMap("item_tag", 5)
							//pictureTagMap :=userInteract.GetTopCountPictureTagsMap(5)
							tagList := make([]int64, 0)
							//pictureTagList :=make([]string,0)
							//for tag,_ :=range pictureTagMap{
							//	if behavior.LabelConvert(tag)!=""{
							//		pictureTagList=append(pictureTagList,tag)
							//	}
							//}
							for key, _ := range tagMap {
								//去掉情感恋爱
								if key != 23 {
									tagList = append(tagList, key)
								}
							}
							tagRecommends, _ := momentCache.QueryTagRecommendsByIds(tagList, "friends_moments_moment_tag:%d")

							tagRecommendSet := utils.SetInt64{}
							for _, tagRecommend := range tagRecommends {
								tagRecommendSet.AppendArray(tagRecommend.GetMomentIds())
							}
							tagRecommendIdList = tagRecommendSet.ToList()
						}
					}
					return len(tagRecommendIdList)
				}
				return realtimeErr
			},
		})
	}

	hotIdMap := utils.NewSetInt64FromArray(hotIdList)
	if len(hourRecList)>500&&len(recIdList)>200{
		recIdList=recIdList[:100]
	}
	var dataIds = utils.NewSetInt64FromArrays(dataIdList, recIdList,hourRecList, newIdList, recIds, hotIdList, liveMomentIds, tagRecommendIdList, autoRecList, adList, bussinessIdList, adLocationList).ToList()
	// 过滤审核
	var paiErr error
	var offTime =0
	var os = utils.GetPlatformName(params.Ua)
	if abtest.GetBool("pai_algo_switch",false){
		paiResult,expId,requestId,paiErr = api.GetPredictResult(params.Lat,params.Lng,os,params.UserId,params.Addr,dataIds,params.Ua)
		if paiErr!=nil{
			offTime=1
			expId=utils.OffTime
			requestId=utils.UniqueId()

		}
	}
	searchMomentMap := map[int64]search.SearchMomentAuditResDataItem{} // 日志推荐，置顶
	filteredAudit := abtest.GetBool("search_filted_audit", false)
	searchScenery := "moment"
	if abtest.GetBool("search_returned_recommend", false) {
		preforms.Run("search", func(*performs.Performs) interface{} {
			var searchMomentMapErr error
			searchMomentMap, searchMomentMapErr = search.CallMomentTopMap(params.UserId,
				searchScenery)
			if searchMomentMapErr == nil {
				momentIdSet := utils.SetInt64{}
				for _, searchRes := range searchMomentMap {
					momentIdSet.Append(searchRes.Id)
				}
				dataIds = momentIdSet.AppendArray(dataIds).ToList()
				return len(searchMomentMap)
			}
			return searchMomentMapErr
		})
	}

	var itemBehaviorMap = map[int64]*behavior.UserBehavior{}     // 获取日志行为
	var userItemBehaviorMap = map[int64]*behavior.UserBehavior{} //获取用户日志行为
	var moms = []redis.MomentsAndExtend{}                        // 获取日志缓存
	var userIds = make([]int64, 0)
	var momOfflineProfileMap = map[int64]*redis.MomentOfflineProfile{} // 获取日志离线画像
	var momContentProfileMap = map[int64]*redis.MomentContentProfile{}
	behaviorModuleName := abtest.GetString("behavior_module_name", app.Module) // 特征对应的module名称
	preforms.RunsGo("moment", map[string]func(*performs.Performs) interface{}{
		"item_behavior": func(*performs.Performs) interface{} { // 获取日志行为
			var itemBehaviorErr error
			itemBehaviorMap, itemBehaviorErr = behaviorCache.QueryItemBehaviorMap(behaviorModuleName, dataIds)
			if itemBehaviorErr == nil {
				return len(itemBehaviorMap)
			}
			return itemBehaviorErr
		}, "useritem_behavior": func(*performs.Performs) interface{} {
			var userItemBehaviorErr error
			userItemBehaviorMap, userItemBehaviorErr = behaviorCache.QueryUserItemBehaviorMap(behaviorModuleName, params.UserId, dataIds)
			if userItemBehaviorErr == nil {
				return len(userItemBehaviorMap)
			}
			return userItemBehaviorErr
		},
		"moment": func(*performs.Performs) interface{} { // 获取日志缓存
			var momsErr error
			if moms, momsErr = momentCache.QueryMomentsByIds(dataIds); momsErr == nil {
				for _, mom := range moms {
					if mom.Moments != nil {
						userIds = append(userIds, mom.Moments.UserId)
					}
				}
				userIds = utils.NewSetInt64FromArray(userIds).ToList()
				return len(moms)
			}
			return momsErr
		}, "profile": func(*performs.Performs) interface{} { // 获取日志离线画像
			var momOfflineProfileErr error
			momOfflineProfileMap, momOfflineProfileErr = momentCache.QueryMomentOfflineProfileByIdsMap(dataIds)
			if momOfflineProfileErr == nil {
				return len(momOfflineProfileMap)
			}
			return momOfflineProfileErr
		}, "picture_profile": func(*performs.Performs) interface{} { // 获取日志离线画像
			var momContentProfileErr error
			momContentProfileMap, momContentProfileErr = momentCache.QueryMomentContentProfileByIdsMap(dataIds)
			if momContentProfileErr == nil {
				return len(momOfflineProfileMap)
			}
			return momContentProfileErr
		},
	})
	var user *redis.UserProfile
	var usersMap = map[int64]*redis.UserProfile{}
	var momentUserEmbedding *redis.MomentUserProfile
	var userLiveProfielMap map[int64]*redis.UserLiveProfile
	var userContentProfileMap map[int64]*redis.UserContentProfile
	var momentUserEmbeddingMap = map[int64]*redis.MomentUserProfile{}
	var isVip = 0
	preforms.RunsGo("user", map[string]func(*performs.Performs) interface{}{
		"user": func(*performs.Performs) interface{} { // 获取用户信息
			var userErr error
			user, usersMap, userErr = userCache.QueryByUserAndUsersMap(params.UserId, userIds)
			if userErr == nil {
				return len(usersMap)
			}
			return userErr
		},
		"profile": func(*performs.Performs) interface{} { // 获取用户信息
			var embeddingCacheErr error
			momentUserEmbedding, momentUserEmbeddingMap, embeddingCacheErr = userCache.QueryMomentUserProfileByUserAndUsersMap(params.UserId, userIds)
			if embeddingCacheErr == nil {
				return len(momentUserEmbeddingMap)
			}
			return embeddingCacheErr
		},
		"user_live_profile": func(*performs.Performs) interface{} {
			var userLiveProfileErr error
			userLiveProfielMap, userLiveProfileErr = userCache.QueryUserLiveProfileByIdsMap([]int64{params.UserId})
			return userLiveProfileErr
		},
		"user_content_profile": func(*performs.Performs) interface{} {
			var userContentProfileErr error
			userContentProfileMap, userContentProfileErr = userCache.QueryUserContentProfileByIdsMap([]int64{params.UserId})
			return userContentProfileErr
		},
	})

	preforms.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId:             params.UserId,
			UserCache:          user,
			MomentUserProfile:  momentUserEmbedding,
			UserLiveProfile:    userLiveProfielMap[params.UserId],
			UserContentProfile: userContentProfileMap[params.UserId],
			UserBehavior:       userBehavior,
		}
		isVip = user.IsVip
		backendRecommendScore := abtest.GetFloat("backend_recommend_score", 1.2)
		realRecommendScore := abtest.GetFloat("real_recommend_score", 1.1)
		statusSwitch := abtest.GetBool("mom_status_filter", false)
		filterLive := abtest.GetBool("fileter_live", true)
		dataList := make([]algo.IDataInfo, 0)
		for _, mom := range moms {
			// 后期搜索完善此条件去除
			if icpSwitch && (mayBeIcpUser || icpWhite) { //icp白名单以及杭州新注册用户
				if !mom.CanRecommend() { //非推荐审核通过
					if !(mom.MomentsProfile != nil && mom.MomentsProfile.IsActivity) { //非活动
						continue
					}
				}
			}
			if mom.Moments == nil || mom.MomentsExtend == nil {
				continue
			}
			if mom.Moments != nil && mom.Moments.Secret == 1 && abtest.GetBool("close_secret", false) { //匿名日志且后台开关开启即关闭
				continue
			}

			//搜索过滤开关(运营推荐不管审核状态)
			if _, ok := searchMomentMap[mom.Moments.Id]; !ok {
				if filteredAudit {
					//if (mom.MomentsProfile != nil && mom.MomentsProfile.AuditStatus == 1) || (mom.MomentsProfile == nil) {
					//	continue
					//}
					if !mom.CanRecommend() {
						if mayBeIcpUser && app.Name == "moment.near" { //附近日志仅对icp用户开通推荐审核
							continue
						}
						if app.Name == "moment" { //推荐日志默认开通推荐审核
							continue
						}
					}
				}
			}
			if mom.Moments.ShareTo != "all" {
				continue
			}

			if statusSwitch && mom.Moments.Status != 1 { //状态不为1的过滤
				continue
			}

			if mom.Moments.Id > 0 {
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
				var isSoftTop = 0
				// 处理推荐
				var recommends = []algo.RecommendItem{}
				if topType, topTypeOK := searchMomentMap[mom.Moments.Id]; topTypeOK {
					topTypeRes := topType.GetCurrentTopType(searchScenery)
					isTop = utils.GetInt(topTypeRes == "TOP")
					isSoftTop = utils.GetInt(topTypeRes == "SOFT")
					if topTypeRes == "RECOMMEND" {
						recommends = append(recommends, algo.RecommendItem{Reason: "RECOMMEND", Score: backendRecommendScore, NeedReturn: true})
					}

				}
				//if isSoftTop==1{
				//	log.Warnf("soft top moment%s",mom.Moments.Id)
				//}
				var liveIndex = 0
				var isTopLiveMom = -1
				if liveMap != nil {
					if rank, isOk := liveMap[mom.Moments.Id]; isOk {
						liveIndex = rank
						momUser, _ := usersMap[mom.Moments.UserId]
						if momUser != nil {
							if isTopLive(ctx, momUser) {
								isTopLiveMom = 1
							} else {
								if isTop != 1 && filterLive { //非头部主播且非置顶直播日志进行过滤
									continue
								}
							}
						}
					}
				}
				var isBussiness = 0
				if bussinessMap != nil {
					if _, isOk := bussinessMap[mom.Moments.Id]; isOk {
						isBussiness = 1
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
				var score=0.0
				if len(paiResult) >0{
					if paiScore,isOk :=paiResult[mom.Moments.Id]; isOk{
						score=paiScore
					}
				}
				info := &DataInfo{
					DataId:               mom.Moments.Id,
					UserCache:            momUser,
					MomentCache:          mom.Moments,
					MomentExtendCache:    mom.MomentsExtend,
					MomentProfile:        mom.MomentsProfile,
					MomentOfflineProfile: momOfflineProfileMap[mom.Moments.Id],
					MomentContentProfile: momContentProfileMap[mom.Moments.Id],
					RankInfo:             &algo.RankInfo{IsTop: isTop, Recommends: recommends, LiveIndex: liveIndex, TopLive: isTopLiveMom, IsBussiness: isBussiness, IsSoftTop: isSoftTop,PaiScore:score,ExpId:utils.ConvertExpId(expId,recall_expId),RequestId:requestId,OffTime:offTime},
					MomentUserProfile:    momentUserEmbeddingMap[mom.Moments.UserId],
					ItemBehavior:         itemBehaviorMap[mom.Moments.Id],
					UserItemBehavior:     userItemBehaviorMap[mom.Moments.Id],
				}
				dataList = append(dataList, info)
			}
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(dataIds)
		ctx.SetDataList(dataList)
		return len(dataList)
	})
	return nil
}

func getMapKey(scoreMap map[int64]int) []int64 {
	res := make([]int64, 0)
	if scoreMap != nil && len(scoreMap) > 0 {
		for key, _ := range scoreMap {
			res = append(res, key)
		}
	}
	return res
}

func isTopLive(ctx algo.IContext, user *redis.UserProfile) bool {
	if user.LiveInfo != nil && user.LiveInfo.Status == 1 && (user.LiveInfo.ExpireDate > ctx.GetCreateTime().Unix()) {
		return true
	}
	return false
}
