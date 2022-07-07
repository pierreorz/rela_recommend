package mate

import (
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/behavior"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/search"
	"rela_recommend/service/performs"
	rutils "rela_recommend/utils"
	"strconv"
)

func DoBuildData(ctx algo.IContext) error {
	var err error
	abtest := ctx.GetAbTest()
	pf := ctx.GetPerforms()
	params := ctx.GetRequest()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	//momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	themeUserCache := redis.NewThemeCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	behaviorCache := behavior.NewBehaviorCacheModule(ctx)
	awsCache := redis.NewMateCacheModule(&factory.CacheCluster, &factory.AwsCluster)
	mateCategCache := redis.NewMateCaegtCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	pretendList, err := awsCache.QueryPretendLoveList()
	//获取假装情侣在线用户id
	var onlineUserList []int64
	if err == nil {
		for _, v := range pretendList {
			userId, err := strconv.ParseInt(v.Userid, 10, 64)
			if err == nil {
				onlineUserList = append(onlineUserList, userId)
			}
		}
	}
	if params.Limit == 0 {
		params.Limit = abtest.GetInt64("default_limit", 50)
	}
	//获取用户实时行为
	var userBehavior redis.BehaviorMate // 用户实时行为
	berhaviorMap := map[int64]int64{} //用户近1小时曝光情况
	userBehavior,err = mateCategCache.QueryMatebehaviorMap(params.UserId)
	log.Infof("userBehavior============================%+v",userBehavior)
	if err==nil{
		for _,v:= range userBehavior.Data{
			mateID:=v.ID
			berhaviorMap[mateID]=1
		}
	}
	log.Infof("berhaviorMap============================%+v",berhaviorMap)
	//获取用户信息，在线用户信息
	var user *redis.UserProfile
	var onlineUserMap map[int64]*redis.UserProfile
	pf.Run("user", func(*performs.Performs) interface{} {
		var userCacheErr error
		if user, onlineUserMap, userCacheErr = userCache.QueryByUserAndUsersMap(params.UserId, onlineUserList); userCacheErr != nil {
			return rutils.GetInt(user != nil)
		}
		return userCacheErr
	})
	//获取距离文案(增加了用户头像)
	var distanceMap = map[int64]float64{}
	var userIamgeMap =  map[int64]string{}
	var distanceText []search.MateTextResDataItem
	reqUser := user.Location
	if len(onlineUserMap) > 0{
		for _, v := range onlineUserMap {
			onlineLocation := v.Location
			imageUrl:=v.Avatar
			distance := rutils.EarthDistance(float64(reqUser.Lon), float64(reqUser.Lat), onlineLocation.Lon, onlineLocation.Lat)
			if distance < 50000 {
				distanceMap[v.UserId] = distance
				userIamgeMap[v.UserId] = imageUrl
			}
		}
	}
	log.Infof("distanceMap=================", distanceMap)
	if len(distanceMap) > 0 {
		distanceText=GetDistanceSenten(distanceMap,distanceTextType,userIamgeMap)
		log.Infof("distanceText=================", distanceText)

	}
	//获取假装情侣池（有多少人喜欢你文案）
	//var likeText []search.MateTextResDataItem
	//if len(onlineUserList)>0{
	//	textType := 70
	//	likeText=GetLikeSenten(len(onlineUserList),int64(textType))
	//	log.Infof("likeText=================", likeText)
	//}

	//用户基础信息生成文案
	//base文案
	var affection_map = map[string]string{"1": "1", "7": "1"}

	//请求者用户基础文案（不展示）
	//reqUserBaseSentence := GetBaseSentenceDataById(user, baseTextType)
	//在线用户基础文案
	//请求者情感状态(不展示)
	//var baseCategText []search.MateTextResDataItem
	//stringAffection := strconv.Itoa(user.Affection)
	//if _, ok := affection_map[stringAffection]; ok {
	//	//情感搜索
	//	categType := int64(4)
	//	var baseCateg redis.TextTypeCategText
	//	baseCateg, err = mateCategCache.QueryMateUserCategTextList(baseTextType, categType)
	//	if err==nil{
	//		baseCategText = GetCategSentenceData(baseCateg.TextLine, baseTextType, 4,adminUserid,user.Avatar)
	//	}
	//}

	//在线用户基础文案
	onlineUserBaseSentenceList := GetBaseSentenceDataMap(onlineUserMap, baseTextType)
	//获取在线用户情感状态
	affectionNums := 0
	var onlineBaseCategText []search.MateTextResDataItem
	var userImageMap map[int64]string
	for _, userView := range onlineUserMap {
		stringAffection := strconv.Itoa(userView.Affection)
		if _, ok := affection_map[stringAffection]; ok {
			affectionNums += 1
			userImageMap[userView.UserId]=userView.Avatar
		}
	}
	if affectionNums > 0 {
		categType := int64(4)
		var onlineBaseCateg redis.TextTypeCategText
		onlineBaseCateg, err = mateCategCache.QueryMateUserCategTextList(baseTextType, categType)
		if err == nil {
			for k,v:=range userImageMap {
				if k==0 {
					//只选择一个
					onlineBaseCategText = GetCategSentenceData(onlineBaseCateg.TextLine, baseTextType, 4, k, v)
				}
			}
		}
	}
	//获取假装情侣池话题偏好
	//var reqUserThemeMap map[int64]float64    //请求者（不展示）
	//userProfileUserIds := []int64{params.UserId}
	//reqUserThemeMap = themeUserCache.QueryMatThemeProfileData(userProfileUserIds)
	////获取假装情侣用户moment偏好
	//var reqUserMomMap map[int64]float64    //请求者（不展示）
	//reqUserMomMap = behaviorCache.QueryMateMomUserData(userProfileUserIds)
	//合并用户偏好(请求者)(不展示)
	//reqUserProfile := MergeMap(reqUserThemeMap, reqUserMomMap)
	//var reqCategText []search.MateTextResDataItem
	//if len(reqUserProfile) > 0 {
	//	resultList := rutils.SortMapByValue(reqUserProfile)
	//	var reqCateg redis.TextTypeCategText
	//	randomList := GetRandomData(len(resultList), resultList)
	//	if len(randomList) > 0 {
	//		for _, v := range randomList {
	//			reqCateg, err = mateCategCache.QueryMateUserCategTextList(categTextType, v)
	//			if err == nil {
	//				reqCategText = GetCategSentenceData(reqCateg.TextLine, categTextType, 4,v,user.Avatar)
	//			}
	//		}
	//	}
	//
	//}
	////假装情侣池
	var onlineUserThemeMap map[int64]float64 //假装情侣池
	onlineUserThemeMap = themeUserCache.QueryMatThemeProfileData(onlineUserList)
	//获取假装情侣用户moment偏好
	var onlineUserMomMap map[int64]float64 //假装情侣池
	onlineUserMomMap = behaviorCache.QueryMateMomUserData(onlineUserList)
	//合并用户偏好(假装情侣池)
	onlineUserProfile := MergeMap(onlineUserThemeMap, onlineUserMomMap)
	var onlineCategText []search.MateTextResDataItem
	if len(onlineUserProfile) > 0 {
		resultList := rutils.SortMapByValue(onlineUserProfile)
		var olineCateg redis.TextTypeCategText
		randomList := GetRandomData(len(resultList), resultList)
		if len(randomList) > 0 {
			for _, v := range randomList {
				olineCateg, err = mateCategCache.QueryMateUserCategTextList(categTextType, v)
				if err == nil {
					userProfile:=onlineUserMap[v]
					onlineCategText = GetCategSentenceData(olineCateg.TextLine, categTextType, 4,v,userProfile.Avatar)
				}
			}
		}
	}
	//旧版搜索结果
	var searchResList []search.MateTextResDataItem
	pf.Run("search", func(*performs.Performs) interface{} {
		searchLimit := abtest.GetInt64("search_limit", 50)
		var searchErr error
		params.Lng = abtest.GetFloat("mate_fake_lng", params.Lng)
		params.Lat = abtest.GetFloat("mate_fake_lat", params.Lat)
		if searchResList, searchErr = search.CallMateTextList(params, searchLimit); searchErr == nil {
			return len(searchResList)
		} else {
			return searchErr
		}
	})
	//组装搜索结果（增加默认头像）
	var searchImageResult []search.MateTextResDataItem
	searchImageResult=GetSearchIamge(searchResList)

	log.Infof("searchImageResult=============%+v", searchImageResult)
	log.Infof("onlineBaseCategText=============%+v", onlineBaseCategText)
	log.Infof("onlineUserBaseSentenceList=============%+v", onlineUserBaseSentenceList)
	log.Infof("onlineCategText=============%+v", onlineCategText)
	log.Infof("distanceText=============%+v", distanceText)
	//合并文案数据
	var allSentenceList []search.MateTextResDataItem
	if onlineUserBaseSentenceList != nil {
		allSentenceList = append(allSentenceList,searchImageResult...)//搜索结果
		allSentenceList = append(allSentenceList,onlineBaseCategText...)//情感结果
		allSentenceList = append(allSentenceList,onlineUserBaseSentenceList...)//基础文案
		allSentenceList = append(allSentenceList,onlineCategText...)//用户偏好
		allSentenceList = append(allSentenceList,distanceText...)//用户距离
	}else{
		//兜底只有时间段
		allSentenceList = append(allSentenceList,searchImageResult...)
	}
	log.Infof("allSentenceList=============%+v", allSentenceList)
	pf.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId: params.UserId,
		}
		// 组装被曝光者信息
		dataIds := make([]int64, 0)
		dataList := make([]algo.IDataInfo, 0)
		for i, baseRes := range allSentenceList {
			if _, ok := berhaviorMap[baseRes.Id]; !ok {
				info := &DataInfo{
					DataId:     baseRes.Id,
					SearchData: &allSentenceList[i],
					RankInfo:   &algo.RankInfo{},
				}
				dataIds = append(dataIds, baseRes.Id)
				dataList = append(dataList, info)
			}
			log.Infof("dataIds=============%+v", dataIds)
			log.Infof("dataList=============%+v", dataList)
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(dataIds)
		ctx.SetDataList(dataList)

		return len(dataList)
	})

	return err
}
