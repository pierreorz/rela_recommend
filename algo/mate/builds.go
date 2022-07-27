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
	mateCategCache := redis.NewMateCaegtCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)

	//获取假装情侣在线缓存数据
	var pretendList []redis.PretendLoveUser
	pf.Run("pretendUser", func(*performs.Performs) interface{} {
		awsCache := redis.NewMateCacheModule(&factory.CacheCluster, &factory.AwsCluster)
		if pretend, pretendCacheErr := awsCache.QueryPretendLoveList(); pretendCacheErr == nil {
			pretendList=append(pretendList, pretend...)
			return len(pretendList)
		}else {
			return pretendCacheErr
		}
	})
	//获取假装情侣在线用户id
	var onlineUserList []int64
	for _, v := range pretendList {
		userId, err := strconv.ParseInt(v.Userid, 10, 64)
		if err == nil {
			onlineUserList = append(onlineUserList, userId)
		}
	}
	if params.Limit == 0 {
		params.Limit = abtest.GetInt64("default_limit", 50)
	}
	//获取用户实时行为
	var userBehavior redis.BehaviorMate // 用户实时行为
	berhaviorMap := map[int64]int64{} //用户近1小时曝光情况
	//userBehavior,err = mateCategCache.QueryMatebehaviorMap(params.UserId)
	pf.Run("requser", func(*performs.Performs) interface{} {
		var requserCacheErr error
		if userBehavior, requserCacheErr = mateCategCache.QueryMatebehaviorMap(params.UserId); requserCacheErr == nil {
			return len(userBehavior.Data)
		}else{
			return requserCacheErr
		}
	})
	//kafka实时数据
	var messageBehavior *behavior.UserBehavior // 用户实时行为
	userMessageBehavior, realtimeErr := behaviorCache.QueryUserBehaviorMap("message", []int64{params.UserId})
	if realtimeErr == nil {
		messageBehavior = userMessageBehavior[params.UserId]
		if messageBehavior !=nil{
			//获取用户message页面数
			if messageBehavior.GetMateListExposure().Count > 0 {
				messageList := messageBehavior.GetMateListExposure().LastList
				if len(messageList) > 1 {
					for index := 0; index < len(messageList); index++{
						messageId := messageList[index].DataId
						berhaviorMap[messageId]=1

						}
					}
				}
			}
		}
	//解析获取实施行为(近一小时数据)
	for _,v:= range userBehavior.Data{
		mateID:=v.ID
		berhaviorMap[mateID]=1
	}
	//获取用户信息，在线用户信息
	var user *redis.UserProfile
	var onlineUserMap map[int64]*redis.UserProfile
	pf.Run("user", func(*performs.Performs) interface{} {
		var userCacheErr error
		if user, onlineUserMap, userCacheErr = userCache.QueryByUserAndUsersMap(params.UserId, onlineUserList); userCacheErr == nil {
			return rutils.GetInt(user != nil)
		}else{
			return userCacheErr
		}
	})
	//获取距离文案(增加了用户头像)
	var distanceMap = map[int64]float64{}
	var userImageMap =  map[int64]string{}
	var distanceText []search.MateTextResDataItem
	reqUser := user.Location
	if len(onlineUserMap) > 0{
		for _, v := range onlineUserMap {
			onlineLocation := v.Location
			imageUrl:=v.Avatar
			distance := rutils.EarthDistance(float64(reqUser.Lon), float64(reqUser.Lat), onlineLocation.Lon, onlineLocation.Lat)
			if distance < 50000 {
				distanceMap[v.UserId] = distance
				userImageMap[v.UserId] = imageUrl
			}
		}
	}
	log.Infof("distanceMap=================", distanceMap)
	if len(distanceMap) > 0 {
		distanceText=GetDistanceSenten(distanceMap,distanceTextType,userImageMap)
		log.Infof("distanceText=================", distanceText)

	}


	//用户基础信息生成文案
	//base文案
	var affection_map = map[string]string{"1": "1", "7": "1"}

	//在线用户基础文案
	onlineUserBaseSentenceList := GetBaseSentenceDataMap(onlineUserMap, baseTextType)
	//获取在线用户情感状态
	affectionNums := 0
	var onlineBaseCategText []search.MateTextResDataItem
	userAffectImageMap:=map[int64]string{}
	for _, userView := range onlineUserMap {
		stringAffection := strconv.Itoa(userView.Affection)
		if _, ok := affection_map[stringAffection]; ok {
			affectionNums += 1
			userAffectImageMap[userView.UserId]=userView.Avatar
		}
	}
	if affectionNums > 0 {
		categType := int64(4)
		var onlineBaseCateg redis.TextTypeCategText
		//获取情感文案
		pf.Run("affection", func(*performs.Performs) interface{} {
			var affectCacheErr error
			if onlineBaseCateg, affectCacheErr = mateCategCache.QueryMateUserCategTextList(baseTextType, categType); affectCacheErr == nil {
				return len(onlineBaseCateg.TextLine)
			}else{
				return affectCacheErr
			}
		})
		for k,_:=range userAffectImageMap {
			if k==0 {
				//只选择一个
				onlineBaseCategText = GetCategSentenceData(onlineBaseCateg.TextLine, baseTextType, 4, k)
			}
		}
	}


	//假装情侣池
	var onlineUserThemeMap map[int64][]int64 //假装情侣池
	onlineUserThemeMap = themeUserCache.QueryMatThemeProfileData(onlineUserList)
	log.Infof("ThemeMap=======%+v",onlineUserThemeMap)
	//获取假装情侣用户moment偏好
	var onlineUserMomMap map[int64][]int64 //假装情侣池
	onlineUserMomMap = behaviorCache.QueryMateMomUserData(onlineUserList)
	log.Infof("MomMap=======%+v",onlineUserThemeMap)
	//合并用户偏好(假装情侣池)
	onlineUserProfile :=GetMergeMap(onlineUserThemeMap, onlineUserMomMap)
	var onlineCategText []search.MateTextResDataItem
	if len(onlineUserProfile) > 0 {

		categMap:=GetCategoryRandomData(onlineUserProfile)
		var olineCateg redis.TextTypeCategText
		if len(categMap) > 0 {
			pf.Run("categ", func(*performs.Performs) interface{} {
				var categCacheErr error
				for categId, userId := range categMap {
					if olineCateg, categCacheErr = mateCategCache.QueryMateUserCategTextList(categTextType, categId); categCacheErr == nil {
						onlineCategText = GetCategSentenceData(olineCateg.TextLine, categTextType, categId, userId)
						return len(onlineCategText)
					} else {
						return categCacheErr
					}
				}
				return len(onlineCategText)
			})
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

	//log.Infof("searchImageResult=============%+v", searchImageResult)
	//log.Infof("onlineBaseCategText=============%+v", onlineBaseCategText)
	//log.Infof("onlineUserBaseSentenceList=============%+v", onlineUserBaseSentenceList)
	//log.Infof("onlineCategText=============%+v", onlineCategText)
	//log.Infof("distanceText=============%+v", distanceText)
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
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(dataIds)
		ctx.SetDataList(dataList)
		return len(dataList)
	})
	return err
}
