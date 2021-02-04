package theme

import (
	"errors"
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/rpc/search"
	"rela_recommend/service/performs"

	// "rela_recommend/models/pika"
	"rela_recommend/models/behavior"
	"rela_recommend/models/redis"
	"rela_recommend/utils"
)

func DoBuildReplyData(ctx algo.IContext) error {
	var err error
	app := ctx.GetAppInfo()
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	preforms := ctx.GetPerforms()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	themeUserCache := redis.NewThemeCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	behaviorCache := behavior.NewBehaviorCacheModule(ctx, &factory.CacheBehaviorRds)

	replyIdList := []int64{}                // 话题参与 ids
	themeIdList := []int64{}                // 主话题Ids
	themeReplyMap := map[int64]int64{}      // 话题与参与话题对应关系
	var userBehavior *behavior.UserBehavior // 用户实时行为
	var tagList []int64                     //用户操作行为tag集合
	newThemeIdList :=[]int64{}
	preforms.RunsGo("recommend", map[string]func(*performs.Performs) interface{}{
		"list": func(*performs.Performs) interface{} { // 获取推荐列表
			recListKeyFormatter := abtest.GetString("recommend_list_key", "theme_reply_recommend_list:%d")
			recommendList, listErr := momentCache.GetThemeRelpyListOrDefault(params.UserId, -999999999, recListKeyFormatter)
			if listErr == nil {
				for _, recommend := range recommendList {
					replyIdList = append(replyIdList, recommend.ThemeReplyID)
					themeIdList = append(themeIdList, recommend.ThemeID)
					themeReplyMap[recommend.ThemeID] = recommend.ThemeReplyID
				}
				log.Infof("themeList======",themeIdList)
				log.Infof("replyList======",replyIdList)
				log.Infof("themeMap=======",themeReplyMap)
				return len(recommendList)
			}
			return listErr
		}, "new":func(*performs.Performs) interface{} {
			newThemeLen :=abtest.GetInt("search_theme_line",100)
			recommended :=abtest.GetBool("realtime_mom_switch",false)// 是否过滤推荐审核
			if newThemeLen >0 {
				momentTypes := abtest.GetString("new_moment_types", "theme")
				newThemeIdList, err = search.CallNewThemeuserId(params.UserId, int64(newThemeLen),momentTypes, recommended)
				themeIdList=append(themeIdList,newThemeIdList...)
				return len(newThemeIdList)
			}
			return nil
		},
		"user_behavior": func(*performs.Performs) interface{} { // 获取实时操作的内容
			realtimes, realtimeErr := behaviorCache.QueryUserBehaviorMap(app.Module, []int64{params.UserId})
			if realtimeErr == nil {
				userBehavior = realtimes[params.UserId]
				//根据实时行为获取用户操作偏好
				if userBehavior != nil {
					userInteract := userBehavior.GetThemeDetailInteract()
					if userInteract.Count > 0 {
						tagMap := userInteract.GetTopCountTagsMap("item_tag", 5)
						for key, _ := range tagMap {
							if key != 23 {
								tagList = append(tagList, key)
							}
						}
					}
				}
				return len(realtimes)
			}
			return realtimeErr
		},
	})
	preforms.Run("tag_recommend", func(*performs.Performs) interface{} {
		//根据实时行为数据召回池数据
		if userBehavior != nil {
			tagRecommends, tagErr := momentCache.QueryTagRecommendsByIds(tagList, "friends_moments_theme_tag:%d")
			if tagErr == nil {
				for _, tagRecommend := range tagRecommends {
					momentList := tagRecommend.Moments
					if len(momentList) > 0 {
						for _, themeDict := range momentList {
							replyIdList = append(replyIdList, themeDict.ReplyId)
							themeIdList = append(themeIdList, themeDict.MomentId)
							themeReplyMap[themeDict.MomentId] = themeDict.ReplyId
						}
					}
				}
				return len(tagRecommends)
			}
			return tagErr
		}
		return nil
	})
	// log.Debugf("reply_ids:%+v, theme_reply_map:%+v\n", replyIdList, themeReplyMap)
	searchScenery := "theme"
	searchReplyMap := map[int64]search.SearchMomentAuditResDataItem{} // 话题参与对应的审核与置顶结果
	searchThemeMap := map[int64]search.SearchMomentAuditResDataItem{} // 话题参与对应的审核与置顶结果

	var searchReplyThemeIds = []int64{}
	var searchThemeNoReturnIds = []int64{}
	filtedAudit := abtest.GetBool("search_filted_audit", false)
	preforms.RunsGo("search", map[string]func(*performs.Performs) interface{}{
		"reply": func(*performs.Performs) interface{} { // 搜索过状态 和 返回置顶推荐内容
			if abtest.GetBool("used_ai_search_reply", true) {
				var searchReplyMapErr error
				returnedRecommend := abtest.GetBool("search_returned_recommend", true)
				searchReplyMap, searchThemeMap, searchReplyMapErr = search.CallMomentAuditMap(params.UserId, replyIdList,
					searchScenery, "theme,themereply", returnedRecommend, filtedAudit)
				if searchReplyMapErr == nil {
					replyIdSet := utils.SetInt64{}
					for _, searchRes := range searchReplyMap {
						replyIdSet.Append(searchRes.Id)
					}
					for themeId, _ := range searchThemeMap {
						searchReplyThemeIds = append(searchReplyThemeIds, themeId)
					}
					replyIdList = replyIdSet.ToList()
					themeReplyMap = themeReplayReplaction(searchReplyMap, themeReplyMap, searchScenery) // 运营配置和算法推荐去重复，以运营配置优先
					return len(searchReplyMap)
				}
				return searchReplyMapErr
			}
			return nil
		},
		"theme": func(*performs.Performs) interface{} { // 计算不符合条件的theme
			if abtest.GetBool("used_ai_search_theme", true) {
				searchThemeResMap, _, searchThemeResMapErr := search.CallMomentAuditMap(params.UserId, themeIdList,
					searchScenery, "theme", false, filtedAudit)
				if searchThemeResMapErr == nil {
					for _, themeId := range themeIdList {
						if _, ok := searchThemeResMap[themeId]; !ok {
							searchThemeNoReturnIds = append(searchThemeNoReturnIds, themeId)
						}
					}
					return len(searchThemeNoReturnIds)
				}
				return searchThemeResMapErr
			}
			return nil
		},
	})
	// log.Debugf("reply_map:%+v, theme_reply_map:%+v\n", searchReplyMap, themeReplyMap)
	var themeIds = utils.NewSetInt64FromArray(themeIdList).AppendArray(searchReplyThemeIds).RemoveArray(searchThemeNoReturnIds).ToList()
	var replyIds = utils.NewSetInt64FromArray(replyIdList).ToList()

	var replysMap = map[int64]redis.MomentsAndExtend{}
	var replysUserIds = []int64{}
	var themes = []redis.MomentsAndExtend{}
	var themesUserIds = []int64{}
	preforms.RunsGo("moment", map[string]func(*performs.Performs) interface{}{
		"reply": func(*performs.Performs) interface{} { // 获取内容缓存
			var replyErr error
			replysMap, replyErr = momentCache.QueryMomentsMapByIds(replyIds)
			if replyErr == nil {
				for _, mom := range replysMap {
					if mom.Moments != nil {
						replysUserIds = append(replysUserIds, mom.Moments.UserId)
					}
				}
				replysUserIds = utils.NewSetInt64FromArray(replysUserIds).ToList()
				log.Infof("replysUserIds========",replysUserIds)
				return len(replysMap)
			}
			return replyErr
		},
		"theme": func(*performs.Performs) interface{} { // 获取内容缓存
			var themesMapErr error
			themes, themesMapErr = momentCache.QueryMomentsByIds(themeIds)
			if themesMapErr == nil {
				for _, mom := range themes {
					if mom.Moments != nil {
						themesUserIds = append(themesUserIds, mom.Moments.UserId)
					}
				}
				themesUserIds = utils.NewSetInt64FromArray(themesUserIds).ToList()
				log.Infof("themesUserIds==========",themesUserIds)
				return len(themes)
			}
			return themesMapErr
		},
	})

	var user *redis.UserProfile
	var usersMap = map[int64]*redis.UserProfile{}
	var usersProfileMap = map[int64]*redis.ThemeUserProfile{}
	var themeProfileMap = map[int64]*redis.ThemeProfile{}
	preforms.RunsGo("cache", map[string]func(*performs.Performs) interface{}{ // 获取用户信息
		"user": func(*performs.Performs) interface{} {
			var userInfoError error
			user, usersMap, userInfoError = userCache.QueryByUserAndUsersMap(params.UserId, themesUserIds)
			if userInfoError == nil {
				return len(usersMap)
			}
			return userInfoError
		},
		"user_profile": func(*performs.Performs) interface{} {
			var themeUserCacheErr error
			userProfileUserIds := []int64{params.UserId}
			usersProfileMap, themeUserCacheErr = themeUserCache.QueryThemeUserProfileMap(userProfileUserIds)
			if themeUserCacheErr == nil {
				return len(usersProfileMap)
			}
			return themeUserCacheErr
		},
		"theme_profile": func(*performs.Performs) interface{} {
			var themeProfileCacheErr error
			themeProfileMap, themeProfileCacheErr = themeUserCache.QueryThemeProfileMap(themeIds)
			if themeProfileCacheErr == nil {
				return len(themeProfileMap)
			}
			return themeProfileCacheErr
		},
	})

	preforms.Run("build", func(*performs.Performs) interface{} {
		dataIds := make([]int64, 0)
		userInfo := &UserInfo{
			UserId:       params.UserId,
			UserCache:    user,
			ThemeUser:    usersProfileMap[params.UserId],
			UserBehavior: userBehavior}

		backendRecommendScore := abtest.GetFloat("backend_recommend_score", 1.2)
		dataList := make([]algo.IDataInfo, 0)
		for _, theme := range themes {
			if theme.Moments != nil && theme.Moments.Id > 0 {
				themeId := theme.Moments.Id
				replyId, replyIdOk := themeReplyMap[themeId]
				reply, replyInfoOK := replysMap[replyId]
				// 计算推荐类型
				var isTop int = 0
				var recommends = []algo.RecommendItem{}
				if topType, topTypeOK := searchThemeMap[themeId]; topTypeOK {
					topTypeRes := topType.GetCurrentTopType(searchScenery)
					isTop = utils.GetInt(topTypeRes == "TOP")
					if topTypeRes == "RECOMMEND" {
						recommends = append(recommends, algo.RecommendItem{
							Reason:     "RECOMMEND",
							Score:      backendRecommendScore,
							NeedReturn: true})
					}
				}
				info := &DataInfo{
					DataId:            themeId,
					UserCache:         usersMap[theme.Moments.UserId],
					MomentCache:       theme.Moments,
					MomentExtendCache: theme.MomentsExtend,
					MomentProfile:     theme.MomentsProfile,
					ThemeProfile:      themeProfileMap[themeId],

					RankInfo: &algo.RankInfo{IsTop: isTop, Recommends: recommends},
				}
				if replyId > 0 && replyIdOk && replyInfoOK {
					info.ThemeReplyCache = reply.Moments
					info.ThemeReplyExtendCache = reply.MomentsExtend
				}
				dataIds = append(dataIds, themeId)
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

// 话题详情页的猜你喜欢
func DoBuildDetailReplyData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	pf := ctx.GetPerforms()
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	if len(params.DataIds) == 0 {
		return errors.New("dataIds length must 1")
	}
	themeId := params.DataIds[0]

	themeReplyMap := map[int64]int64{}
	pf.Run("recommend_list", func(*performs.Performs) interface{} { // 获取推荐列表
		recListKeyFormatter := abtest.GetString("recommend_list_key", "theme_reply_recommend_list:%d")
		recommendList, listErr := momentCache.GetThemeRelpyListOrDefault(params.UserId, -999999999, recListKeyFormatter)
		if listErr == nil {
			for _, recommend := range recommendList {
				themeReplyMap[recommend.ThemeID] = recommend.ThemeReplyID
			}
		}
		return len(recommendList)
	})
	searchScenery := "theme"
	pf.Run("search", func(*performs.Performs) interface{} { // 搜索过状态 和 返回置顶推荐内容
		returnedRecommend := abtest.GetBool("search_returned_recommend", true)
		filtedAudit := abtest.GetBool("search_filted_audit", false)
		searchReplyMap, _, searchReplyMapErr := search.CallMomentAuditMap(params.UserId, []int64{},
			searchScenery, "theme,themereply", returnedRecommend, filtedAudit)
		if searchReplyMapErr == nil {
			themeReplyMap = themeReplayReplaction(searchReplyMap, themeReplyMap, searchScenery)
			return len(searchReplyMap)
		}
		return searchReplyMapErr
	})

	pf.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{UserId: params.UserId}
		dataList := make([]algo.IDataInfo, 0)

		replyIds := []int64{}
		if replyId, replyIdOK := themeReplyMap[themeId]; replyIdOK {
			replyIds = append(replyIds, replyId)
		}

		for i, replyId := range replyIds {
			info := &DataInfo{
				DataId:            replyId,
				UserCache:         nil,
				MomentCache:       nil,
				MomentExtendCache: nil,
				MomentProfile:     nil,
				ThemeProfile:      nil,
				RankInfo:          &algo.RankInfo{Level: -i},
			}
			dataList = append(dataList, info)
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(replyIds)
		ctx.SetDataList(dataList)

		return len(dataList)
	})
	return err
}

// 搜索返回的推荐置顶与算法返回的进行去重，以运营配置优先
func themeReplayReplaction(searchReplyMap map[int64]search.SearchMomentAuditResDataItem, themeReplyMap map[int64]int64, scenery string) map[int64]int64 {
	for _, searchRes := range searchReplyMap {
		// 运营配置和算法推荐去重复，以运营配置优先
		if _, theThemeOK := themeReplyMap[searchRes.ParentId]; theThemeOK {
			if len(searchRes.GetCurrentTopType(scenery)) > 0 {
				themeReplyMap[searchRes.ParentId] = searchRes.Id
			}
		} else {
			themeReplyMap[searchRes.ParentId] = searchRes.Id
		}
	}
	return themeReplyMap
}
