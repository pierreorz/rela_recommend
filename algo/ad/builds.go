package ad

import (
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/search"
	"rela_recommend/service/performs"
	rutils "rela_recommend/utils"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	pf := ctx.GetPerforms()
	params := ctx.GetRequest()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	//userData := ctx.GetUserInfo().(*UserInfo)
	// behaviorCache := behavior.NewBehaviorCacheModule(ctx, &factory.CacheBehaviorRds)

	if params.Limit == 0 {
		params.Limit = abtest.GetInt64("default_limit", 50)
	}

	// 获取用户信息
	var user *redis.UserProfile
	pf.Run("user", func(*performs.Performs) interface{} {
		var userCacheErr error
		if user, _, userCacheErr = userCache.QueryByUserAndUsersMap(params.UserId, []int64{}); userCacheErr != nil {
			return rutils.GetInt(user != nil)
		} else {
			return userCacheErr
		}
	})
	// 获取search的广告列表
	var searchResList = []search.SearchADResDataItem{}
	if abtest.GetBool("icp_switch", false) && (abtest.GetBool("is_icp_user", false) || user.MaybeICPUser(params.Lat, params.Lng)) {
		log.Infof("ad user<%s> is_icp_user", params.UserId)
	} else {
		pf.Run("search", func(*performs.Performs) interface{} {
			clientName := abtest.GetString("backend_app_name", "1") // 1: rela 2: 饭角
			var searchErr error
			//针对新老版本的请求过滤
			if params.ClientVersion >= 50802 {
				if searchResList, searchErr = search.CallFeedAdList(clientName, params, user); searchErr == nil {
					return len(searchResList)
				} else {
					return searchErr
					}
				}else{
					if searchResList, searchErr = search.CallAdList(clientName, params, user); searchErr == nil {
						return len(searchResList)
					} else{
					return searchErr
				}
			}
		})
	}
	//根据用户偏好，投放广告
	//pf.Run("user_profile", func(*performs.Performs) interface{} {
	//	//user:=userData.UserId
	//	userThemeProfile := userData.ThemeUser
	//	tagMapLine := userThemeProfile.AiTag.UserShortTag
	//	for _, tagLine := range tagMapLine {
	//		if tagLine.TagId == 19 { //19 游戏，用户偏好
	//
	//
	//		} else {
	//			return err
	//		}
	//	}
	//	return err
	//})


	pf.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId:    params.UserId,
			UserCache: user,
		}

		// 组装被曝光者信息
		dataIds := make([]int64, 0)
		dataList := make([]algo.IDataInfo, 0)
		for i, searchRes := range searchResList {
			info := &DataInfo{
				DataId:     searchRes.Id,
				SearchData: &searchResList[i],
				RankInfo:   &algo.RankInfo{},
			}
			dataIds = append(dataIds, searchRes.Id)
			dataList = append(dataList, info)
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(dataIds)
		ctx.SetDataList(dataList)

		return len(dataList)
	})

	return err
}
