package moment

import (
	"rela_recommend/algo"
	"rela_recommend/algo/live"
	"rela_recommend/factory"
	"rela_recommend/models/behavior"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/api"
	"rela_recommend/rpc/search"
	"rela_recommend/service/performs"
	"rela_recommend/utils"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
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
	tagRecommendIdList := make([]int64, 0)
	liveMomentIds := make([]int64, 0)
	var recIds, topMap, recMap = []int64{}, map[int64]int{}, map[int64]int{}
	momentTypes := abtest.GetString("moment_types", "text_image,video,text,image,theme,themereply")

	if abtest.GetBool("rec_liveMoments_switch", false) {
		liveMomentIds = live.GetCachedLiveMomentListByTypeClassify(-1, -1)
	}

	var userBehavior *behavior.UserBehavior // 用户实时行为
	preforms.RunsGo("data", map[string]func(*performs.Performs) interface{}{
		"recommend": func(*performs.Performs) interface{} { // 获取推荐日志
			if dataIdList == nil || len(dataIdList) == 0 {
				recListKeyFormatter := abtest.GetString("recommend_list_key", "") // moment_recommend_list:%d
				if len(recListKeyFormatter) > 5 {
					recIdList, err = momentCache.GetInt64ListOrDefault(params.UserId, -999999999, recListKeyFormatter)
					return len(recIdList)
				}
			}
			return nil
		}, "new": func(*performs.Performs) interface{} { // 新日志 或 附近日志
			newMomentLen := abtest.GetInt("new_moment_len", 1000)
			if newMomentLen > 0 {
				radiusArray := abtest.GetStrings("radius_range", "50km")
				newMomentOffsetSecond := abtest.GetFloat("new_moment_offset_second", 60*60*24*30*3)
				newMomentStartTime := float32(ctx.GetCreateTime().Unix()) - newMomentOffsetSecond
				if abtest.GetBool("near_liveMoments_switch", false) {
					var lives []pika.LiveCache

					lives = live.GetCachedLiveListByTypeClassify(-1, -1)
					liveMomentIds = ReturnAroundLiveMom(lives, params.Lng, params.Lat)
				}
				//当附近50km无日志，扩大范围200km,2000km,20000km直至找到日志
				var errSearch error
				for _, radius := range radiusArray {
					if abtest.GetBool("use_ai_search", false) {
						newIdList, errSearch = search.CallNearMomentListV1(params.UserId, params.Lat, params.Lng, 0, int64(newMomentLen),
							momentTypes, newMomentStartTime, radius)
					} else {
						newIdList, errSearch = search.CallNearMomentList(params.UserId, params.Lat, params.Lng, 0, newMomentLen,
							momentTypes, newMomentStartTime, radius)
					}
					//附近日志数量大于10即停止寻找
					if len(newIdList) > 10 {
						break
					}
				}

				if errSearch != nil {
					return err
				}
				return len(newIdList)
			}
			return nil
		}, "hot": func(*performs.Performs) interface{} { // 热门列表
			if abtest.GetBool("real_recommend_switched", false) {
				if top, topErr := behaviorCache.QueryDataBehaviorTop(); topErr == nil {
					hotIdList = top.GetTopIds(100)
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
		}, "user_behavior": func(*performs.Performs) interface{} { // 获取实时操作的内容
			realtimes, realtimeErr := behaviorCache.QueryUserBehaviorMap(app.Module, []int64{params.UserId})
			if realtimeErr == nil&&abtest.GetInt("rich_strategy:user_behavior_interact:weight",1)==1 {
				userBehavior = realtimes[params.UserId]
				userInteract := userBehavior.GetMomentListInteract()
				if userInteract.Count > 0 {
					//获取用户实时互动日志的各个标签的实时热门数据
					tagMap := userInteract.GetTopCountTagsMap("item_tag", 5)
					tagList := make([]int64, 0)
					for key, _ := range tagMap {
						//去掉情感恋爱
						if key != 23 {
							tagList = append(tagList, key)
						}
					}
					tagRecommends, _ := momentCache.QueryTagRecommendsByIds(tagList, "friends_moments_moment_tag:%d")
					tagRecommendSet := utils.SetInt64{}
					for _,tagRecommend :=range tagRecommends{
						tagRecommendSet.AppendArray(tagRecommend.GetMomentIds())
					}
					tagRecommendIdList = tagRecommendSet.ToList()
				}
				return len(tagRecommendIdList)
			}
			return realtimeErr
		},
	})

	hotIdMap := utils.NewSetInt64FromArray(hotIdList)
	var dataIds = utils.NewSetInt64FromArrays(dataIdList, recIdList, newIdList, recIds, hotIdList, liveMomentIds,tagRecommendIdList).ToList()
	// 过滤审核
	searchMomentMap := map[int64]search.SearchMomentAuditResDataItem{} // 日志推荐，置顶
	filteredAudit := abtest.GetBool("search_filted_audit", false)
	searchScenery := "moment"
	if abtest.GetBool("search_returned_recommend", false){
		preforms.Run("search", func(*performs.Performs) interface{} {
			var searchMomentMapErr error
			searchMomentMap, searchMomentMapErr = search.CallMomentTopMap(params.UserId,
				searchScenery, momentTypes)
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

	var itemBehaviorMap = map[int64]*behavior.UserBehavior{} // 获取日志行为
	var moms = []redis.MomentsAndExtend{}                    // 获取日志缓存
	var userIds = make([]int64, 0)
	var momOfflineProfileMap = map[int64]*redis.MomentOfflineProfile{} // 获取日志离线画像

	behaviorModuleName := abtest.GetString("behavior_module_name", app.Module) // 特征对应的module名称
	preforms.RunsGo("moment", map[string]func(*performs.Performs) interface{}{
		"item_behavior": func(*performs.Performs) interface{} { // 获取日志行为
			var itemBehaviorErr error
			itemBehaviorMap, itemBehaviorErr = behaviorCache.QueryItemBehaviorMap(behaviorModuleName, dataIds)
			if itemBehaviorErr == nil {
				return len(itemBehaviorMap)
			}
			return itemBehaviorErr
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
		},
	})

	var user *redis.UserProfile
	var usersMap = map[int64]*redis.UserProfile{}
	var momentUserEmbedding *redis.MomentUserProfile
	var momentUserEmbeddingMap = map[int64]*redis.MomentUserProfile{}
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
	})

	preforms.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId:            params.UserId,
			UserCache:         user,
			MomentUserProfile: momentUserEmbedding,
			UserBehavior:      userBehavior,
		}

		backendRecommendScore := abtest.GetFloat("backend_recommend_score", 1.2)
		realRecommendScore := abtest.GetFloat("real_recommend_score", 1.2)
		dataList := make([]algo.IDataInfo, 0)
		for _, mom := range moms {
			// 后期搜索完善此条件去除
			if mom.Moments == nil || mom.MomentsExtend == nil {
				continue
			}
			//搜索过滤开关(运营推荐不管审核状态)
			if _,ok :=searchMomentMap[mom.Moments.Id];!ok{
				if filteredAudit {
					if (mom.MomentsProfile != nil && mom.MomentsProfile.AuditStatus == 0)||(mom.MomentsProfile==nil) {
						continue
					}
				}
			}
			if mom.Moments.ShareTo != "all" {
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
				// 处理推荐
				var recommends = []algo.RecommendItem{}
				if topType, topTypeOK := searchMomentMap[mom.Moments.Id]; topTypeOK {
					topTypeRes := topType.GetCurrentTopType(searchScenery)
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
		return len(dataList)
	})
	return nil
}
