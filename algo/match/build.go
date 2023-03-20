package match

import (
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/rpc/search"
	"rela_recommend/service/performs"
	"rela_recommend/utils"

	// "rela_recommend/algo/utils"
	"rela_recommend/models/redis"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	pfms := ctx.GetPerforms()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	dataIds := params.DataIds

	// 获取用户信息，修正经纬度
	var user *redis.UserProfile
	pfms.Run("user", func(*performs.Performs) interface{} { // 获取推荐列表
		var userCacheErr error
		if user, userCacheErr = userCache.QueryUserById(params.UserId); userCacheErr == nil {
			if params.Lat == 0 || params.Lng == 0 {
				if user != nil {
					params.Lat = float32(user.Location.Lat)
					params.Lng = float32(user.Location.Lon)
				}
			}
			return 1
		}
		return userCacheErr
	})

	var recMap = utils.SetInt64{}
	userSearchMap := make(map[int64]*search.UserResDataItem, 0)
	pfms.RunsGo("ids", map[string]func(*performs.Performs) interface{}{
		"search": func(*performs.Performs) interface{} {
			var searchErr error
			if abtest.GetBool("used_ai_search", false) {
				if dataIds, userSearchMap, searchErr = search.CallMatchList(ctx, params.UserId, params.Lat, params.Lng, dataIds, user); searchErr == nil {
					return len(dataIds)
				}
			}
			return searchErr
		},
		"recommend": func(*performs.Performs) interface{} {
			var recErr error
			recListKeyFormatter := abtest.GetString("match_recommend_keyformatter", "") // match_recommend_list_v1:%d
			if len(recListKeyFormatter) > 5 {
				var recIdlist = []int64{}
				if recIdlist, recErr = userCache.GetInt64List(params.UserId, recListKeyFormatter); recErr == nil {
					recMap.AppendArray(recIdlist)
					return len(recIdlist)
				}
			}
			return recErr
		},
	})
	dataIds = utils.NewSetInt64FromArrays(dataIds, recMap.ToList()).ToList()

	var usersMap = map[int64]*redis.UserProfile{}
	var matchUser *redis.MatchProfile
	var matchUserMap = map[int64]*redis.MatchProfile{}
	pfms.RunsGo("caches", map[string]func(*performs.Performs) interface{}{
		"user": func(*performs.Performs) interface{} { // 获取用户信息
			var usersCacheErr error
			if usersMap, usersCacheErr = userCache.QueryUsersMap(dataIds); usersCacheErr == nil {
				return len(usersMap)
			}
			return usersCacheErr
		},
		"profile": func(*performs.Performs) interface{} { // 获取画像信息
			var matchCacheErr error
			matchUser, matchUserMap, matchCacheErr = userCache.QueryMatchProfileByUserAndUsersMap(params.UserId, dataIds)
			if matchCacheErr == nil {
				return len(matchUserMap)
			}
			return matchCacheErr
		},
	})

	if ctx.GetRequest().UserId == 110758574 {
		log.Infof("userSearchMap: %+v", userSearchMap)
	}
	pfms.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId:       params.UserId,
			UserCache:    user,
			MatchProfile: matchUser}

		backendRecommendScore := abtest.GetFloat("backend_recommend_score", 1.5)
		dataList := make([]algo.IDataInfo, 0)

		for dataId, data := range usersMap {
			if data.DataUserCandidateCanRecommend() {
				// 推荐集加权
				var recommends = []algo.RecommendItem{}
				if recMap.Contains(data.UserId) {
					recommends = append(recommends, algo.RecommendItem{Reason: "RECOMMEND", Score: backendRecommendScore, NeedReturn: true})
				}

				info := &DataInfo{
					DataId:       dataId,
					UserCache:    data,
					MatchProfile: matchUserMap[dataId],
					RankInfo:     &algo.RankInfo{Recommends: recommends},
					SearchFields: userSearchMap[dataId],
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
