package label

import (
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/search"
	"rela_recommend/service/performs"
	"time"
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
		for i, nameId := range nameList {
			info := &DataInfo{
				DataId:    nameId,
				RankInfo:             &algo.RankInfo{Level: -i},
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
		for i, nameId := range nameList {
			info := &DataInfo{
				DataId:    nameId,
				RankInfo:             &algo.RankInfo{Level: -i},
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


func DoBuildLabelRec(ctx algo.IContext) error {
	var err error
	params := ctx.GetRequest()
	query :=params.Params["query"]
	idList := make([]int64, 0)
	pf := ctx.GetPerforms()
	change :=0
	// 获取用户信息
	var user *redis.UserProfile
	var usersMap = map[int64]*redis.UserProfile{}
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)

	pf.RunsGo("caches", map[string]func(*performs.Performs) interface{}{
		"user": func(*performs.Performs) interface{} {
			var userCacheErr error
			user, usersMap, userCacheErr = userCache.QueryByUserAndUsersMap(params.UserId, idList)
			if userCacheErr != nil {
				return userCacheErr
			}
			return len(usersMap)
		},
	})
	rdsPikaCache := redis.NewLiveCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	if len(query)<=0{//非文本类请求
		change=1
		if len(params.Params["image_url"])>0{//图片类日志
			if idList, err = rdsPikaCache.GetInt64ListFromString("Image", "mom_label_data:%s");err!=nil{
				return err
			}
		}
		if len(params.Params["video_url"])>0{//视频类日志
			if idList, err = rdsPikaCache.GetInt64ListFromString("Video", "mom_label_data:%s");err!=nil{
				return err
			}
		}
	}else{//请求接口数据

	}
	if change==1{//对指定数据进行打散
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(idList), func(i, j int) { idList[i], idList[j] = idList[j], idList[i] })
	}
	pf.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId:    params.UserId,
			UserCache: user,
		}

		// 组装被曝光者信息
		dataList := make([]algo.IDataInfo, 0)
		for i, nameId := range idList {
			info := &DataInfo{
				DataId:    nameId,
				RankInfo:             &algo.RankInfo{Level: -i},

			}
			dataList = append(dataList, info)
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(idList)
		ctx.SetDataList(dataList)
		return len(dataList)
	})
	return err
}


