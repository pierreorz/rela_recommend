package theme

import (
	"errors"
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/rpc/search"
	"rela_recommend/service/performs"

	// "rela_recommend/models/pika"
	"rela_recommend/models/redis"
	"rela_recommend/utils"
)

func DoBuildReplyData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	preforms := ctx.GetPerforms()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	themeUserCache := redis.NewThemeCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)

	replyIdList := []int64{}           // 话题参与 ids
	themeIdList := []int64{}           // 主话题Ids
	themeReplyMap := map[int64]int64{} // 话题与参与话题对应关系

	preforms.Run("recommend_list", func(*performs.Performs) interface{} { // 获取推荐列表
		recListKeyFormatter := abtest.GetString("recommend_list_key", "theme_reply_recommend_list:%d")
		recommendList, listErr := momentCache.GetThemeRelpyListOrDefault(params.UserId, -999999999, recListKeyFormatter)
		if listErr == nil {
			for _, recommend := range recommendList {
				replyIdList = append(replyIdList, recommend.ThemeReplyID)
				themeIdList = append(themeIdList, recommend.ThemeID)
				themeReplyMap[recommend.ThemeID] = recommend.ThemeReplyID
			}
			return len(recommendList)
		}
		return listErr
	})
	searchScenery := "theme"
	searchReplyMap := map[int64]search.SearchMomentAuditResDataItem{} // 话题参与对应的审核与置顶结果
	preforms.Run("search", func(*performs.Performs) interface{} {     // 搜索过状态 和 返回置顶推荐内容
		returnedRecommend := abtest.GetBool("search_returned_recommend", false)
		filtedAudit := abtest.GetBool("search_filted_audit", false)
		var searchReplyMapErr error
		searchReplyMap, searchReplyMapErr = search.CallMomentAuditMap(params.UserId, replyIdList, searchScenery, returnedRecommend, filtedAudit)
		if searchReplyMapErr == nil {
			replyIdSet := utils.SetInt64{}
			themeIdSet := utils.NewSetInt64FromArray(themeIdList)
			for _, searchRes := range searchReplyMap {
				replyIdSet.Append(searchRes.Id)
				themeIdSet.Append(searchRes.ParentId)
				// 运营配置和算法推荐去重复，以运营配置优先
				if _, theThemeOK := themeReplyMap[searchRes.ParentId]; theThemeOK {
					if len(searchRes.GetCurrentTopType("theme")) > 0 {
						themeReplyMap[searchRes.ParentId] = searchRes.Id
					}
				} else {
					themeReplyMap[searchRes.ParentId] = searchRes.Id
				}
			}
			replyIdList = replyIdSet.ToList()
			themeIdList = themeIdSet.ToList()
			return len(searchReplyMap)
		}
		return searchReplyMapErr
	})

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
				return len(replysMap)
			}
			return replyErr
		},
		"theme": func(*performs.Performs) interface{} { // 获取内容缓存
			var themesMapErr error
			themes, themesMapErr = momentCache.QueryMomentsByIds(themeIdList)
			if themesMapErr == nil {
				for _, mom := range themes {
					if mom.Moments != nil {
						themesUserIds = append(themesUserIds, mom.Moments.UserId)
					}
				}
				themesUserIds = utils.NewSetInt64FromArray(themesUserIds).ToList()
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
			themeProfileMap, themeProfileCacheErr = themeUserCache.QueryThemeProfileMap(themeIdList)
			if themeProfileCacheErr == nil {
				return len(themeProfileMap)
			}
			return themeProfileCacheErr
		},
	})

	preforms.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId:    params.UserId,
			UserCache: user,
			ThemeUser: usersProfileMap[params.UserId]}

		backendRecommendScore := abtest.GetFloat("backend_recommend_score", 1.2)
		dataList := make([]algo.IDataInfo, 0)
		for _, theme := range themes {
			if theme.Moments != nil && theme.Moments.Id > 0 {
				themeId := theme.Moments.Id
				replyId, replyIdOk := themeReplyMap[themeId]
				reply, replyOK := replysMap[replyId]
				if replyIdOk && replyOK {
					// 计算推荐类型
					var isTop int = 0
					var recommends = []algo.RecommendItem{}
					if topType, topTypeOK := searchReplyMap[reply.Moments.Id]; topTypeOK {
						topTypeRes := topType.GetCurrentTopType("theme")
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

						ThemeReplyCache:       reply.Moments,
						ThemeReplyExtendCache: reply.MomentsExtend,
						RankInfo:              &algo.RankInfo{IsTop: isTop, Recommends: recommends},
					}
					dataList = append(dataList, info)
				}
			}
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(replyIds)
		ctx.SetDataList(dataList)
		return nil
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

		return nil
	})
	return err
}
