package moment

import (
	"errors"
	"rela_recommend/algo"
	"rela_recommend/algo/live"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	"rela_recommend/utils"
)

// 附近日志详情页推荐
func DoBuildMomentAroundDetailSimData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	dataIdList := make([]int64, 0)
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	livemoms, liveMomerr := momentCache.QueryMomentsByIds(params.DataIds)
	if liveMomerr != nil {
		return errors.New("liveMomerr ")
	}
	if len(livemoms)>0{
		momsType := livemoms[0].Moments.MomentsType

		if momsType == "live" || momsType == "voice_live" {
			//判断日志是否是直播日志
			var lives []pika.LiveCache
			liveLen := abtest.GetInt("live_moment_len", 3)
			lives = live.GetCachedLiveListByTypeClassify(-1, -1)
			dataIdList = ReturnTopnScoreLiveMom(lives, liveLen, params.DataIds[0])
		} else {
			recListKeyFormatter := abtest.GetString("around_detail_sim_list_key", "moment.around_sim_momentList:%s")
			dataIdList, err = momentCache.GetInt64ListFromGeohash(params.Lat, params.Lng, 4, recListKeyFormatter)
		}

		if dataIdList == nil || err != nil {
			ctx.SetUserInfo(nil)
			ctx.SetDataIds(dataIdList)
			ctx.SetDataList(make([]algo.IDataInfo, 0))
			return nil
		}
		momOfflineProfileMap, momOfflineProfileErr := momentCache.QueryMomentOfflineProfileByIdsMap(dataIdList)
		if momOfflineProfileErr != nil {
			log.Warnf("moment embedding is err,%s\n", momOfflineProfileErr)
		}
		moms, err := momentCache.QueryMomentsByIds(dataIdList)
		userIds := make([]int64, 0)
		for _, mom := range moms {
			if mom.Moments != nil {
				userIds = append(userIds, mom.Moments.UserId)
			}
		}
		userIds = utils.NewSetInt64FromArray(userIds).ToList()
		user, usersMap, err := userCache.QueryByUserAndUsersMap(params.UserId, userIds)
		if err != nil {
			log.Warnf("users list is err, %s\n", err)
		}
		momentUserEmbedding, momentUserEmbeddingMap, embeddingCacheErr := userCache.QueryMomentUserProfileByUserAndUsersMap(params.UserId, userIds)
		if embeddingCacheErr != nil {
			log.Warnf("moment user Embedding cache list is err, %s\n", embeddingCacheErr)
		}
		userInfo := &UserInfo{UserId: params.UserId,
			UserCache: user,
			//MomentProfile: momentUser,
			MomentUserProfile: momentUserEmbedding}
		dataList := make([]algo.IDataInfo, 0)
		for i, mom := range moms {
			if mom.Moments != nil && mom.Moments.Id > 0 {
				if mom.Moments.ShareTo != "all" {
					continue
				}
				if mom.Moments.Id==params.DataIds[0]{
					continue
				}
				momUser, _ := usersMap[mom.Moments.UserId]
				//status=0 禁用用户，status=5 注销用户
				if momUser != nil {
					if momUser.Status == 0 || momUser.Status == 5 {
						continue
					}
				}

				info := &DataInfo{
					DataId:               mom.Moments.Id,
					UserCache:            momUser,
					MomentCache:          mom.Moments,
					MomentExtendCache:    mom.MomentsExtend,
					MomentProfile:        mom.MomentsProfile,
					MomentOfflineProfile: momOfflineProfileMap[mom.Moments.Id],
					RankInfo:             &algo.RankInfo{},
					MomentUserProfile:    momentUserEmbeddingMap[mom.Moments.UserId],
				}
				if momsType == "live" || momsType == "voice_live" {
					info.RankInfo.Level = -i
				}
				dataList = append(dataList, info)
			}
			ctx.SetUserInfo(userInfo)
			ctx.SetDataIds(dataIdList)
			ctx.SetDataList(dataList)
		}
	}


	return err
}

// 关注页日志详情页推荐
func DoBuildMomentFriendDetailSimData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	if len(params.DataIds) == 0 {
		return errors.New("dataIds length must 1")
	}

	recListKeyFormatter := abtest.GetString("friend_detail_before_list_key", "moment.friend_before_moment:%d")
	momIds, err := momentCache.QueryMomentsByIds(params.DataIds)
	if err != nil {
		return errors.New("follow detail moms data not exists")
	}
	dataIdList := make([]int64, 0)
	if len(momIds) > 0 && momIds[0].Moments != nil && momIds[0].Moments.Secret == 0 {
		dataIdList, err = momentCache.GetInt64List(momIds[0].Moments.UserId, recListKeyFormatter)
		if err != nil {
			return errors.New("follow detail query redis err")
		}
	}
	SetData(dataIdList, ctx,params.DataIds[0],true)

	return err
}

// 推荐页日志详情页推荐
func DoBuildMomentRecommendDetailSimData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	params := ctx.GetRequest()
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	if len(params.DataIds) == 0 {
		return errors.New("dataIds length must 1")
	}
	moms, liveMomerr := momentCache.QueryMomentsByIds(params.DataIds)
	if liveMomerr != nil {
		return errors.New("liveMomerr ")
	}
	if len(moms)>0{
		momsType := moms[0].Moments.MomentsType
		if momsType == "live" || momsType == "voice_live" {
			//判断日志类型
			var lives []pika.LiveCache
			liveLen := abtest.GetInt("live_moment_len", 3)
			lives = live.GetCachedLiveListByTypeClassify(-1, -1)
			liveIds := ReturnTopnScoreLiveMom(lives, liveLen, moms[0].Moments.UserId)
			SetData(liveIds, ctx,params.DataIds[0],false)

		} else {
			recListKeyFormatter := abtest.GetString("recommend_detail_sim_list_key", "moment.recommend_sim_momentList:%d")
			var defaultId = -999999999
			if momsType == "video" {
				defaultId = -999999998
			}
			dataIdList, err := momentCache.GetInt64ListOrDefault(params.DataIds[0], int64(defaultId), recListKeyFormatter)
			if err == nil {
				SetData(dataIdList, ctx,params.DataIds[0],false)
			}
		}
	}
	return err
}

func SetData(dataIdList []int64, ctx algo.IContext,momsId int64,filter bool) error {
	var err error
	params := ctx.GetRequest()
	userInfo := &UserInfo{UserId: params.UserId}
	momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	moms, err := momentCache.QueryMomentsByIds(dataIdList)
	if err != nil {
		return errors.New("query mom err")
	}
	userIds := make([]int64, 0)
	for _, mom := range moms {
		if mom.Moments != nil {
			userIds = append(userIds, mom.Moments.UserId)
		}
	}
	userIds = utils.NewSetInt64FromArray(userIds).ToList()
	usersMap, err := userCache.QueryUsersMap(userIds)
	dataList := make([]algo.IDataInfo, 0)
	for i, mom := range moms {
		if mom.Moments != nil && mom.Moments.Id > 0 {
			if mom.Moments.ShareTo != "all" {
				continue
			}
			if mom.Moments.Id==momsId{
				continue
			}
			//关注日志去掉匿名日志
			if filter&&mom.Moments.Secret==1{
				continue
			}
			momUser, _ := usersMap[mom.Moments.UserId]
			//status=0 禁用用户，status=5 注销用户
			if momUser != nil {
				if momUser.Status == 0 || momUser.Status == 5 {
					continue
				}
			}

			info := &DataInfo{
				DataId:               mom.Moments.Id,
				UserCache:            nil,
				MomentCache:          nil,
				MomentExtendCache:    nil,
				MomentProfile:        nil,
				MomentOfflineProfile: nil,
				RankInfo:             &algo.RankInfo{Level: -i},
				MomentUserProfile:    nil,
			}
			dataList = append(dataList, info)
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(dataIdList)
		ctx.SetDataList(dataList)
	}
	return err
}

func ReturnAroundLiveMom(lives []pika.LiveCache, userLng float32, userLat float32) []int64 {
	res := make([]int64, 0)
	if len(lives) > 0 {
		for i, _ := range lives {
			//直接获取日志id
			//liveIds[i] = lives[i].Live.MomentsID
			momLat := lives[i].Live.Lat
			momLng := lives[i].Live.Lng
			if utils.EarthDistance(float64(momLng), float64(momLat), float64(userLng), float64(userLat))/1000.0 < 50 {
				res = append(res, lives[i].Live.MomentsID)
			}
		}
	}
	return res
}

func ReturnTopnScoreLiveMom(lives []pika.LiveCache, topn int, userId int64) []int64 {
	res := make([]int64, 0)
	liveMap := make(map[int64]float64, 0)
	for i, _ := range lives {
		if lives[i].Live.UserId != userId {
			//获取用户：分数map
			liveMap[lives[i].Live.MomentsID] = float64(lives[i].Score)
		}
	}
	res = utils.SortMapByValue(liveMap)
	if topn >= len(res) {
		topn = len(res)
	}
	return res[:topn]
}
func ReturnLiveList(userIdList, momIdList []int64) []int64 {
	//根据userid和momid获取每个用户最新的直播日志
	var idMap = map[int64]int64{}
	res := make([]int64, 0)
	index := 0
	for _, userId := range userIdList {
		if _, ok := idMap[userId]; ok {
			continue
		} else {
			idMap[userId] = momIdList[index]
			res = append(res, momIdList[index])
		}
		index += 1
	}
	return res
}
