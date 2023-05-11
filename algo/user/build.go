package user

import (
	"rela_recommend/algo"
	"rela_recommend/algo/live"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/pika"
	"rela_recommend/rpc/search"
	"rela_recommend/service/performs"
	"rela_recommend/tasks"
	"rela_recommend/utils"

	// "rela_recommend/algo/utils"
	"rela_recommend/models/behavior"
	"rela_recommend/models/redis"
)

func DoBuildSearchData(ctx algo.IContext) error {
	var err error
	pf := ctx.GetPerforms()
	params := ctx.GetRequest()
	abtest := ctx.GetAbTest()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)

	// 确定候选用户
	dataIds := params.DataIds
	if abtest.GetBool("always_use_search", false) { // 是否一直使用search
		pf.Run("search", func(*performs.Performs) interface{} {
			var searchErr error
			if dataIds, searchErr = search.CallSearchUserIdList(params.UserId, params.Lat, params.Lng,
				params.Offset, params.Limit, params.Params["query"]); searchErr == nil {
				return len(dataIds)
			} else {
				return searchErr
			}
		})
	}

	// 获取用户信息
	var user *redis.UserProfile
	var usersMap = map[int64]*redis.UserProfile{}
	pf.RunsGo("caches", map[string]func(*performs.Performs) interface{}{
		"user": func(*performs.Performs) interface{} {
			var userCacheErr error
			user, usersMap, userCacheErr = userCache.QueryByUserAndUsersMap(params.UserId, dataIds)
			if userCacheErr != nil {
				return userCacheErr
			}
			return len(usersMap)
		},
	})
	// 组装用户信息
	pf.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId:    params.UserId,
			UserCache: user,
		}

		// 组装被曝光者信息
		dataList := make([]algo.IDataInfo, 0)
		for _, userId := range dataIds {
			info := &DataInfo{
				DataId:    userId,
				UserCache: usersMap[userId],
				RankInfo:  &algo.RankInfo{},
			}
			dataList = append(dataList, info)
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(dataIds)
		ctx.SetDataList(dataList)
		return len(dataList)
	})
	return err
}

func DoBuildDataV1(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	app := ctx.GetAppInfo()
	pf := ctx.GetPerforms()
	params := ctx.GetRequest()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	behaviorCache := behavior.NewBehaviorCacheModule(ctx)
	liveMap := live.GetCachedLiveMap() // 当前的直播列表

	var userCurrent *redis.UserProfile
	var userCurrentErr error
	userCurrent, userCurrentErr = userCache.QueryUserById(params.UserId)
	if userCurrentErr != nil {
		log.Errorf("failed to get current user cache: %d, %s", params.UserId, userCurrentErr)
	}

	// 确定候选用户
	dataIds := params.DataIds
	userSearchMap := make(map[int64]*search.UserResDataItem, 0)
	if dataIds == nil || len(dataIds) == 0 {
		if abtest.GetBool("icp_switch", false) &&
			(abtest.GetBool("is_icp_user", false) || ((userCurrent != nil) && userCurrent.MaybeICPUser(params.Lat, params.Lng))) {

			pf.Run("get_fix_icp", func(*performs.Performs) interface{} {
				var searchErr error
				if dataIds, userSearchMap, searchErr = search.CallNearUserICPIdList(params.UserId, params.Lat, params.Lng,
					0, 2000, params.Params["search"]); searchErr == nil {
					return len(dataIds)
				} else {
					return searchErr
				}
			})
		} else if abtest.GetBool("is_icp_auditor", false) {
			pf.Run("icp_auditor", func(*performs.Performs) interface{} {
				var searchErr error
				lat := abtest.GetFloat("icp_center_lat", 30.284882)
				lon := abtest.GetFloat("icp_center_lon", 120.028722)
				if dataIds, userSearchMap, searchErr = search.CallNearUserAuditList(params.UserId, lat, lon,
					0, 2000, params.Params["search"]); searchErr == nil {
					return len(dataIds)
				} else {
					return searchErr
				}
			})
		} else {
			pf.Run("search", func(*performs.Performs) interface{} {
				var searchErr error
				if dataIds, userSearchMap, searchErr = search.CallNearUserIdList(params.UserId, params.Lat, params.Lng,
					0, 2000, params.Params["search"]); searchErr == nil {
					return len(dataIds)
				} else {
					return searchErr
				}
			})
		}
	}

	// 获取用户信息
	var user *redis.UserProfile
	var usersMap = map[int64]*redis.UserProfile{}
	// 获取用户画像
	var userProfile *redis.NearbyProfile
	var userProfileMap = map[int64]*redis.NearbyProfile{}
	// 用户实时行为
	behaviorModuleName := abtest.GetString("behavior_module_name", app.Module)
	var userBehaviorMap = map[int64]*behavior.UserBehavior{}
	var itemBehaviorMap = map[int64]*behavior.UserBehavior{}
	pf.RunsGo("caches", map[string]func(*performs.Performs) interface{}{
		"user": func(*performs.Performs) interface{} {
			var userCacheErr error
			user, usersMap, userCacheErr = userCache.QueryByUserAndUsersMap(params.UserId, dataIds)
			if userCacheErr != nil {
				return userCacheErr
			}
			return len(usersMap)
		},
		"profile": func(*performs.Performs) interface{} {
			var profileCacheErr error
			profileKeyFormatter := abtest.GetString("profile_key_formatter", "nearby_user_profile:%d")
			userProfile, userProfileMap, profileCacheErr = userCache.QueryNearbyProfileByUserAndUsersMap(params.UserId, dataIds, profileKeyFormatter)
			if profileCacheErr != nil {
				return profileCacheErr
			}
			return len(usersMap)
		},
		"realtime_useritem": func(*performs.Performs) interface{} {
			var userBehaviorErr error
			userBehaviorMap, userBehaviorErr = behaviorCache.QueryUserItemBehaviorMap(behaviorModuleName, params.UserId, dataIds)
			if userBehaviorErr != nil {
				return userBehaviorErr
			}
			return len(userBehaviorMap)
		},
		"realtime_item": func(*performs.Performs) interface{} {
			var itemBehaviorErr error
			itemBehaviorMap, itemBehaviorErr = behaviorCache.QueryItemBehaviorMap(behaviorModuleName, dataIds)
			if itemBehaviorErr != nil {
				return itemBehaviorErr
			}
			return len(itemBehaviorMap)
		},
	})

	// 组装用户信息
	pf.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId:      params.UserId,
			UserCache:   user,
			UserProfile: userProfile}

		// 组装被曝光者信息
		dataList := make([]algo.IDataInfo, 0)
		for dataId, data := range usersMap {
			if data.IsVipHiding() {
				continue
			}
			if data.DataUserCandidateCanRecommend() {
				info := &DataInfo{
					DataId:       dataId,
					UserCache:    data,
					UserProfile:  userProfileMap[dataId],
					LiveInfo:     liveMap[dataId],
					RankInfo:     &algo.RankInfo{},
					SearchFields: userSearchMap[dataId],

					UserItemBehavior: userBehaviorMap[dataId],
					ItemBehavior:     itemBehaviorMap[dataId],
				}
				dataList = append(dataList, info)
			}
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(dataIds)
		ctx.SetDataList(dataList)

		return len(dataList)
	})

	return err
}

func DoBuildNtxlData(ctx algo.IContext) error {
	var err error
	params := ctx.GetRequest()
	pf := ctx.GetPerforms()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	behaviorCache := behavior.NewBehaviorCacheModule(ctx)

	nearbyUsers := make([]int64, 0)
	userSearchMap := make(map[int64]*search.UserResDataItem, 0)
	liveUsers := make([]int64, 0)
	liveMap := make(map[int64]*pika.LiveCache, 0)
	var momentInteractBehaviors *behavior.Behavior         // 日志页面发生的互动行为
	var userInfoMomentInteractBehaviors *behavior.Behavior // 个人主页发生的日志行为
	var userInfoOtherInteractBehaviors *behavior.Behavior  // 个人主页发生的其他行为
	pf.RunsGo("recall_users", map[string]func(*performs.Performs) interface{}{
		"nearby_users": func(*performs.Performs) interface{} {
			if ctx.GetAbTest().GetBool("nearby_search_es", true) {
				var searchErr error
				nearbyUsers, userSearchMap, searchErr = search.CallNearUserList(params.UserId, params.Lat, params.Lng,
					0, 200, "", []string{})
				if searchErr != nil {
					return searchErr
				} else {
					return len(nearbyUsers)
				}
			}
			return 0
		},
		"live_users": func(*performs.Performs) interface{} {
			liveListSize := ctx.GetAbTest().GetInt("live_list_size", 5)
			liveUsers, liveMap = tasks.GetAllCachedLiveUsersAndMapByRandom(liveListSize)
			return len(liveMap)
		},
		"user_moment_behavior": func(*performs.Performs) interface{} {
			if ctx.GetAbTest().GetBool("ntxl_recall_interact", false) {
				behaviors, userCacheErr := behaviorCache.QueryUserBehaviorMap("moment", []int64{params.UserId})
				if userCacheErr != nil {
					return userCacheErr
				}
				if currentUserBehavior, ok := behaviors[params.UserId]; ok {
					momentInteractBehaviors = currentUserBehavior.GetMomentListInteract()
					if momentInteractBehaviors != nil {
						return momentInteractBehaviors.Count
					}
				}
			}
			return 0
		},
		"user_userinfo_behavior": func(*performs.Performs) interface{} {
			if ctx.GetAbTest().GetBool("ntxl_recall_interact", false) {
				behaviors, userCacheErr := behaviorCache.QueryUserBehaviorMap("userinfo", []int64{params.UserId})
				if userCacheErr == nil {
					if currentUserBehavior, ok := behaviors[params.UserId]; ok {
						userInfoMomentInteractBehaviors = currentUserBehavior.GetMomentListInteract()
						userInfoOtherInteractBehaviors = currentUserBehavior.GetUserInteracts()
					}
				}

				behaviors, userCacheErr = behaviorCache.QueryItemBehaviorMap("userinfo", []int64{params.UserId})
				if userCacheErr == nil {
					if currentUserBehavior, ok := behaviors[params.UserId]; ok {
						userInfoMomentInteractBehaviors = userInfoMomentInteractBehaviors.Merge(currentUserBehavior.GetMomentListInteract())
						userInfoOtherInteractBehaviors = userInfoOtherInteractBehaviors.Merge(currentUserBehavior.GetUserInteracts())

						var cnt float64
						if userInfoMomentInteractBehaviors != nil {
							cnt += userInfoMomentInteractBehaviors.Count
						}
						if userInfoOtherInteractBehaviors != nil {
							cnt += userInfoOtherInteractBehaviors.Count
						}
						return cnt
					}
				}
			}
			return 0
		},
	})

	momentInteractUsers, momentDataIDs, momentUserDataMap := momentInteractBehaviors.Merge(userInfoMomentInteractBehaviors).GetUserIDList()
	userinfoInteractUsers, userinfoDataIDs, _ := userInfoOtherInteractBehaviors.GetUserIDList()

	log.Debugf("ntxl mu: %+v, md: %+v", momentInteractUsers, momentDataIDs)
	log.Debugf("ntxl uu: %+v, ud: %+v", userinfoInteractUsers, userinfoDataIDs)
	var dataIds = utils.NewSetInt64FromArrays(nearbyUsers, momentInteractUsers, userinfoInteractUsers).ToList()
	var currentUserProfile *redis.UserProfile
	var usersMap = map[int64]*redis.UserProfile{}
	var userinfoUserBehavior, userinfoItemBehavior, momentUserBehavior, momentItemBehavior, messageUserBehavior,
		messageItemBehavior map[int64]*behavior.UserBehavior
	pf.RunsGo("caches", map[string]func(*performs.Performs) interface{}{
		"user": func(*performs.Performs) interface{} {
			var userCacheErr error
			currentUserProfile, usersMap, userCacheErr = userCache.QueryByUserAndUsersMap(params.UserId, dataIds)
			if userCacheErr != nil {
				return userCacheErr
			}
			return len(usersMap)
		},
		"userinfo_user_behavior": func(*performs.Performs) interface{} {
			var userCacheErr error
			userinfoUserBehavior, userCacheErr = behaviorCache.QueryUserItemBehaviorMap("userinfo", params.UserId, userinfoDataIDs)
			if userCacheErr != nil {
				return userCacheErr
			}
			return len(userinfoUserBehavior)
		},
		"userinfo_item_behavior": func(*performs.Performs) interface{} {
			var userCacheErr error
			userinfoItemBehavior, userCacheErr = behaviorCache.QueryBeenUserItemBehaviorMap("userinfo", params.UserId, userinfoDataIDs)
			if userCacheErr != nil {
				return userCacheErr
			}
			return len(userinfoItemBehavior)
		},
		"moment_user_behavior": func(*performs.Performs) interface{} {
			var userCacheErr error
			momentUserBehavior, userCacheErr = behaviorCache.QueryUserItemBehaviorMap("moment", params.UserId, momentDataIDs)
			if userCacheErr != nil {
				return userCacheErr
			}
			return len(momentUserBehavior)
		},
		"moment_item_behavior": func(*performs.Performs) interface{} {
			var userCacheErr error
			momentItemBehavior, userCacheErr = behaviorCache.QueryBeenUserItemBehaviorMap("moment", params.UserId, momentDataIDs)
			if userCacheErr != nil {
				return userCacheErr
			}
			return len(momentItemBehavior)
		},
		"message_user_behavior": func(*performs.Performs) interface{} {
			var userCacheErr error
			messageUserBehavior, userCacheErr = behaviorCache.QueryUserItemBehaviorMap("message", params.UserId, dataIds)
			if userCacheErr != nil {
				return userCacheErr
			}
			return len(messageUserBehavior)
		},
		"message_item_behavior": func(*performs.Performs) interface{} {
			var userCacheErr error
			messageItemBehavior, userCacheErr = behaviorCache.QueryBeenUserItemBehaviorMap("message", params.UserId, dataIds)
			if userCacheErr != nil {
				return userCacheErr
			}
			return len(messageItemBehavior)
		},
	})
	pf.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId:    params.UserId,
			UserCache: currentUserProfile,
		}

		// 组装被曝光者信息
		dataList := make([]algo.IDataInfo, 0)
		for dataId, data := range usersMap {
			if data.IsVipHiding() {
				continue
			}
			if data.DataUserCandidateCanRecommend() {
				info := &DataInfo{
					DataId:               dataId,
					RankInfo:             &algo.RankInfo{Score: 1.},
					UserCache:            data,
					LiveInfo:             liveMap[dataId],
					UserItemBehavior:     userinfoUserBehavior[dataId].Merge(momentUserBehavior[momentUserDataMap[dataId]]).Merge(messageUserBehavior[dataId]),
					BeenUserItemBehavior: userinfoItemBehavior[dataId].Merge(momentItemBehavior[momentUserDataMap[dataId]]).Merge(messageItemBehavior[dataId]),
				}
				dataList = append(dataList, info)
			}
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(dataIds)
		ctx.SetDataList(dataList)

		return len(dataList)
	})

	return err
}
