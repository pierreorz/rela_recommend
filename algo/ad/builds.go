package ad

import (
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/behavior"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/search"
	"rela_recommend/service/performs"
	rutils "rela_recommend/utils"
	"sort"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	pf := ctx.GetPerforms()
	params := ctx.GetRequest()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	behaviorCache := behavior.NewBehaviorCacheModule(ctx)
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
	//获取用户实时行为
	var userBehavior *behavior.UserBehavior // 用户实时行为
	userAdIdFeedMap := map[int64]int64{}        //广告曝光数据
	userAdIdInitMap := map[int64]int64{}        //广告曝光数据
	realtimes, realtimeErr := behaviorCache.QueryUserBehaviorMap("ad", []int64{params.UserId})
	if realtimeErr == nil { // 获取flink数据
		userBehavior = realtimes[params.UserId]
		if userBehavior != nil { //开屏广告和feed流广告id
			//获取用户广告的行为(feed广告)
			if userBehavior.GetAdFeedListExposure().Count > 0 {
				userFeedList := userBehavior.GetAdFeedListExposure().LastList
				if len(userFeedList) > 1 {
					userAdTimeMap := map[float64]int64{}
					usserAdTimeList := []float64{}
					for index := 0; index < len(userFeedList); index++ {
						userFeedId := userFeedList[index].DataId
						userFeedTime := userFeedList[index].LastTime
						usserAdTimeList = append(usserAdTimeList, userFeedTime)
						userAdTimeMap[userFeedTime] = userFeedId
					}
					lastTime := sort.Float64Slice(usserAdTimeList)
					sort.Sort(lastTime)
					adId := userAdTimeMap[lastTime[len(lastTime)-1]]
					userAdIdFeedMap[adId] = 1
				}
			}
			//获取用户广告的行为(init广告)
			if userBehavior.GetAdInitListExposure().Count > 0 {
				userInitList := userBehavior.GetAdInitListExposure().LastList
				if len(userInitList) > 1 {
					userInitId := userInitList[len(userInitList)-1].DataId
					userAdIdInitMap[userInitId] = 1
				}
			}
		}
	}
	log.Infof("userAdIdFeedMap=================%+v", userAdIdFeedMap)
	log.Infof("userAdIdInitMap=================%+v", userAdIdInitMap)
	// 获取search的广告列表
	var searchResList = []search.SearchADResDataItem{}
	if abtest.GetBool("icp_switch", false) && (abtest.GetBool("is_icp_user", false) || user.MaybeICPUser(params.Lat, params.Lng)) {
		log.Infof("ad user<%s> is_icp_user", params.UserId)
	} else {
		pf.Run("search", func(*performs.Performs) interface{} {
			clientName := abtest.GetString("backend_app_name", "1") // 1: rela 2: 饭角
			var searchErr error
			if user.CanRecommend() {
				//针对新老版本的请求过滤
				if params.ClientVersion >= 50802 { //params.Type == feedType 不对广告类型限制
					if searchResList, searchErr = search.CallFeedAdList(clientName, params, user); searchErr == nil {
						return len(searchResList)
					} else {
						return searchErr
					}
				} else {
					if searchResList, searchErr = search.CallAdList(clientName, params, user); searchErr == nil {
						return len(searchResList)
					} else {
						return searchErr
					}
				}
			} else {
				return rutils.GetInt(user)
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
		if len(searchResList) != 1 {
			for i, searchRes := range searchResList {
				if _, ok := userAdIdFeedMap[searchRes.Id]; !ok {
					if _, ok := userAdIdInitMap[searchRes.Id]; !ok {

						info := &DataInfo{
							DataId:     searchRes.Id,
							SearchData: &searchResList[i],
							RankInfo:   &algo.RankInfo{},
						}
						dataIds = append(dataIds, searchRes.Id)
						dataList = append(dataList, info)
					}
				}
			}
		} else { //当只有一条广告时不过滤上一次广告
			for i, searchRes := range searchResList {
				info := &DataInfo{
					DataId:     searchRes.Id,
					SearchData: &searchResList[i],
					RankInfo:   &algo.RankInfo{},
				}
				dataIds = append(dataIds, searchRes.Id)
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
