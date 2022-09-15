package label

import (
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/search"
	"rela_recommend/service/performs"
)

func DoBuildLabelSuggest(ctx algo.IContext) error {
	var err error
	pf := ctx.GetPerforms()

	params := ctx.GetRequest()
	query :=params.Params["query"]
	nameList := make([]int64, 0)
	nameList,_ =search.CallLabelSuggestList(query)
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)

	// 获取用户信息
	var user *redis.UserProfile
	var usersMap = map[int64]*redis.UserProfile{}
	pf.RunsGo("caches", map[string]func(*performs.Performs) interface{}{
		"user": func(*performs.Performs) interface{} {
			var userCacheErr error
			user, usersMap, userCacheErr = userCache.QueryByUserAndUsersMap(params.UserId, nameList)
			if userCacheErr != nil {
				return userCacheErr
			}
			return len(usersMap)
		},
	})
	pf.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId:    params.UserId,
			UserCache: user,
		}

		// 组装被曝光者信息
		dataList := make([]algo.IDataInfo, 0)
		for _, nameId := range nameList {
			info := &DataInfo{
				DataId:    nameId,
			}
			dataList = append(dataList, info)
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(nameList)
		ctx.SetDataList(dataList)
		return len(dataList)
	})
	return err
}


func DoBuildLabelSearch(ctx algo.IContext) error {
	var err error
	pf := ctx.GetPerforms()

	params := ctx.GetRequest()
	query :=params.Params["query"]
	nameList := make([]int64, 0)
	nameList,_ =search.CallLabelSearchList(query)
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)

	// 获取用户信息
	var user *redis.UserProfile
	var usersMap = map[int64]*redis.UserProfile{}
	pf.RunsGo("caches", map[string]func(*performs.Performs) interface{}{
		"user": func(*performs.Performs) interface{} {
			var userCacheErr error
			user, usersMap, userCacheErr = userCache.QueryByUserAndUsersMap(params.UserId, nameList)
			if userCacheErr != nil {
				return userCacheErr
			}
			return len(usersMap)
		},
	})
	pf.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId:    params.UserId,
			UserCache: user,
		}

		// 组装被曝光者信息
		dataList := make([]algo.IDataInfo, 0)
		for _, nameId := range nameList {
			info := &DataInfo{
				DataId:    nameId,
			}
			dataList = append(dataList, info)
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(nameList)
		ctx.SetDataList(dataList)
		return len(dataList)
	})
	return err
}



