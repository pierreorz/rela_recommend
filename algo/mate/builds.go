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
		log.Infof("pretendList=====================%+v", pretendList)
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
	//用户基础信息生成文案
	//base文案
	var affection_list= map[string]string{"1": "1", "7": "1"}
	//searchBase := search.SearchType{}
	//reqSearchCateg := search.SearchType{}
	//onlineSearchCateg := search.SearchType{}
	//请求用户基础文案
	reqUserBaseSentence := GetBaseSentenceDataById(user)
	log.Infof("reqUserBaseSentence=======================================%+v", reqUserBaseSentence)
	//在线用户基础文案
	onlineUserBaseSentenceList := GetBaseSentenceDataMap(onlineUserMap)

	//基础数据需要搜索
	var baseCategText []search.MateTextResDataItem
	if _, ok := affection_list[string(user.Affection)]; ok {
		//log.Infof( "========Intro",user.Affection)
		//searchBase.TextType = 10
		//searchBase.TagType = append(searchBase.TagType, 4)
		//情感搜索
		textType:=10
		categList:=[]int64{4}
		var baseCateg redis.TextTypeCategText
		baseCateg,err=mateCategCache.QueryMateUserCategTextList(textType,categList)
		if err==nil{
			baseCategText=GetCategSentenceData(baseCateg.TextLine,int64(textType),4)
		}
	}
	//获取假装情侣池话题偏好
	var reqUserThemeMap  map[int64]float64 //请求者
	var onlineUserThemeMap map[int64]float64 //假装情侣池
	userProfileUserIds := []int64{params.UserId}
	reqUserThemeMap=themeUserCache.QueryMatThemeProfileData(userProfileUserIds)
	if reqUserThemeMap!=nil{
		log.Infof("userThemeMap=======================================%+v", reqUserThemeMap)
	}
	onlineUserThemeMap =themeUserCache.QueryMatThemeProfileData(onlineUserList)
	if onlineUserThemeMap!=nil{
		log.Infof("onlineUserThemeMap=======================================%+v", onlineUserThemeMap)
	}
	//获取假装情侣用户moment偏好

	var reqUserMomMap  map[int64]float64 //请求者
	var onlineUserMomMap map[int64]float64 //假装情侣池
	reqUserMomMap=behaviorCache.QueryMateMomUserData(userProfileUserIds)
	if reqUserMomMap!=nil{
		log.Infof("reqUserMomMap=======================================%+v", reqUserMomMap)
	}
	onlineUserMomMap=behaviorCache.QueryMateMomUserData(onlineUserList)
	if onlineUserMomMap!=nil{
		log.Infof("onlineUserMomMap=======================================%+v", onlineUserMomMap)
	}
	//合并用户偏好(请求)
	reqUserProfile := MergeMap(reqUserThemeMap, reqUserMomMap)
	var reqCategText []search.MateTextResDataItem
	if len(reqUserProfile) > 0 {
		resultList := rutils.SortMapByValue(reqUserProfile)
		textType:=20
		var reqCateg redis.TextTypeCategText
		if len(resultList)>=2{
			for i, v := range resultList {
				if i<=2 {
					categList := []int64{int64(v)}
					reqCateg, err = mateCategCache.QueryMateUserCategTextList(textType, categList)
					if err==nil{
						reqCategText=GetCategSentenceData(reqCateg.TextLine,int64(textType),4)
					}
				}
			}
		}else{
			for _, v := range resultList {
				categList := []int64{int64(v)}
				reqCateg, err = mateCategCache.QueryMateUserCategTextList(textType, categList)
				if err==nil{
					reqCategText=GetCategSentenceData(reqCateg.TextLine,int64(textType),4)
				}
			}
		}

	}
	//合并用户偏好(假装情侣池)
	onlineUserProfile := MergeMap(onlineUserThemeMap,onlineUserMomMap)
	var onlineCategText []search.MateTextResDataItem
	if len(onlineUserProfile) > 0 {
		resultList := rutils.SortMapByValue(onlineUserProfile)
		textType:=20
		var olineCateg redis.TextTypeCategText
		if len(resultList)>=2{
			for i, v := range resultList {
				if i<=2 {
					categList := []int64{int64(v)}
					olineCateg, err = mateCategCache.QueryMateUserCategTextList(textType, categList)
					if err==nil{
						onlineCategText=GetCategSentenceData(olineCateg.TextLine,int64(textType),4)
					}
				}
			}
		}else{
			for _, v := range resultList {
				categList := []int64{int64(v)}
				olineCateg, err = mateCategCache.QueryMateUserCategTextList(textType, categList)
				if err==nil{
					onlineCategText=GetCategSentenceData(olineCateg.TextLine,int64(textType),4)
				}
			}
		}

	}
	log.Infof("reqCategText=============%+v", reqCategText)
	log.Infof("onlineCategText=============%+v", onlineCategText)
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
	//log.Infof("searchList=============%+v", searchResList)
	//合并文案数据
	var allSentenceList []search.MateTextResDataItem
	if onlineUserBaseSentenceList != nil {
		allSentenceList = append(onlineUserBaseSentenceList,searchResList...)
		allSentenceList = append(allSentenceList,baseCategText...)
		allSentenceList = append(allSentenceList,reqCategText...)
	}else{
		allSentenceList = append(reqUserBaseSentence,searchResList...)
		allSentenceList = append(allSentenceList,baseCategText...)
		allSentenceList = append(allSentenceList,onlineCategText...)
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
			info := &DataInfo{
				DataId:     baseRes.Id,
				SearchData: &allSentenceList[i],
				RankInfo:   &algo.RankInfo{},
			}
			dataIds = append(dataIds, baseRes.Id)
			dataList = append(dataList, info)
		}
		ctx.SetUserInfo(userInfo)
		ctx.SetDataIds(dataIds)
		ctx.SetDataList(dataList)

		return len(dataList)
	})

	return err
}
