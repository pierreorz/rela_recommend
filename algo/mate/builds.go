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
	"strings"
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
	pretendList, err := awsCache.QueryPretendLoveList()
	if err == nil {
		log.Infof("pretendList=====================%+v", pretendList)
	}
	if params.Limit == 0 {
		params.Limit = abtest.GetInt64("default_limit", 50)
	}

	//获取用户信息
	var user *redis.UserProfile
	pf.Run("user", func(*performs.Performs) interface{} {
		var userCacheErr error
		if user, _, userCacheErr = userCache.QueryByUserAndUsersMap(params.UserId, []int64{}); userCacheErr != nil {
			return rutils.GetInt(user != nil)
		}
		return userCacheErr
	})
	horoscopeName := HoroscopeDict[user.Horoscope]
	wantName := WantDict[user.WantRole]
	roleName := RoleDict[user.RoleName]

	log.Infof("user.Horoscope=============%+v", user.Horoscope)
	log.Infof("user.WantRole=============%+v", user.WantRole)
	log.Infof("user.RoleName=============%+v", user.RoleName)

	log.Infof("horoscope_name=============%+v", horoscopeName)
	log.Infof("want_name=============%+v", wantName)
	log.Infof("role_name=============%+v", roleName)
	//用户基础信息生成文案
	//base文案
	var roleMap = map[string]string{"T": "1", "P": "1", "H": "1"}
	var affection_list = map[string]string{"1": "1", "7": "1"}
	var ageText string
	var roleText string
	var textList []string
	searchBase := search.SearchType{}
	searchCateg := search.SearchType{}
	var baseVeiwList []search.MateTextResDataItem
	userAge := user.Age
	if userAge >= 18 && userAge <= 40 {
		ageText = strconv.Itoa(userAge) + "岁"
		textList = append(textList, ageText)
	}
	textList = append(textList, horoscopeName)
	//自我认同
	if _, ok := roleMap[roleName]; ok {
		log.Infof("我是" + roleName + "，你呢？")
		roleText = "我是" + roleName + "，你呢？"
		textList = append(textList, roleName)
		beasSentence := GetSentenceData(10002,roleText,nil,100)

		baseVeiwList = append(baseVeiwList, beasSentence)
	}
	//职业
	if user.Occupation != "" && len(user.Occupation) <= 6 {
		textList = append(textList, roleText)
	}
	//用户基本文案
	if len(textList) > 1 {
		baseText := strings.Join(textList, "/")
		log.Infof("baseText", baseText)
		beasSentence := GetSentenceData(10000,baseText,nil,100)
		baseVeiwList = append(baseVeiwList, beasSentence)
	}
	//我想找的
	if _, ok := roleMap[wantName]; ok {
		wantText := "有" + wantName + "吗？"
		beasSentence := GetSentenceData(10001,wantText,nil,100)

		baseVeiwList = append(baseVeiwList, beasSentence)
	}
	if user.Intro != "" {
		//log.Infof( "========Intro",user.Intro)
		beasSentence := GetSentenceData(10003,user.Intro,nil,100)
		baseVeiwList = append(baseVeiwList, beasSentence)
	}
	//基础数据需要搜索
	if _, ok := affection_list[string(user.Affection)]; ok {
		//log.Infof( "========Intro",user.Affection)
		searchBase.TextType = "10"
		searchBase.TagType = append(searchBase.TagType, "4")
	}
	log.Infof("baseVeiwText=============%+v", baseVeiwList)
	log.Infof("categSearch=============%+v", searchBase)
	//情感搜索
	//获取用户话题偏好
	userThemeMap := map[int64]float64{}
	var themeProfileMap= map[int64]*redis.ThemeUserProfile{}
	pf.Run("Theme_profile", func(*performs.Performs) interface{} {
		var themeUserCacheErr error
		userProfileUserIds := []int64{params.UserId}
		themeProfileMap, themeUserCacheErr = themeUserCache.QueryThemeUserProfileMap(userProfileUserIds)
		if themeUserCacheErr == nil {
			return len(themeProfileMap)
		}
		return themeUserCacheErr
	})
	if len(themeProfileMap) > 0 {
		themeProfile := themeProfileMap[params.UserId]
		themeTagLongMap := themeProfile.AiTag.UserLongTag
		themeTagShortMap := themeProfile.AiTag.UserShortTag
		if len(themeTagLongMap) > 0 {
			for k, _ := range themeTagLongMap {
				if _, ok := userThemeMap[k]; ok {
					userThemeMap[k] += 1.0
				} else {
					userThemeMap[k] = 1.0
				}
			}
		}
		if len(themeTagShortMap) > 0 {
			for k, _ := range themeTagShortMap {
				if _, ok := userThemeMap[k]; ok {
					userThemeMap[k] += 1
				} else {
					userThemeMap[k] = 1
				}
			}
		}
	}
	//获取moment偏好
	var userBehavior *behavior.UserBehavior
	userMomMap := map[int64]float64{}
	realtimes, realtimeErr := behaviorCache.QueryUserBehaviorMap("moment", []int64{params.UserId})
	if realtimeErr == nil { // 获取flink数据
		userBehavior = realtimes[params.UserId]
		log.Infof("userBehavior=============%+v", userBehavior)
		if userBehavior != nil { //
			countMap := userBehavior.BehaviorMap["moment.recommend:exposure"]
			log.Infof("countMap=============%+v", countMap)
			if countMap != nil {
				tagMap := countMap.CountMap
				log.Infof("momentTagMap=============%+v", tagMap)
				if tagMap != nil {
					for _, v := range tagMap {
						userMomMap[v.Id] = 1.0
					}
				}
			}
		}
	}
	//合并用户偏好
	userProfile:=MergeMap(userThemeMap,userMomMap)
	log.Infof("themeMap=============%+v", userThemeMap)
	log.Infof("momMap=============%+v", userMomMap)
	log.Infof("userProfile=======%+v",userProfile)
	if len(userProfile)>0{
		resultList := rutils.SortMapByValue(userProfile)
		for i, v := range resultList {
			if i < 2 {
				categid := strconv.Itoa(int(v))
				searchCateg.TagType = append(searchCateg.TagType, categid)
			}
		}
		searchCateg.TextType = "20"
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
	//log.Infof("searchList=============%+v", searchResList)
	//合并文案数据
	//for _, searchRes := range searchResList {
	//	baseVeiwList=append(baseVeiwList,searchRes)
	//}
	pf.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId: params.UserId,
		}

		// 组装被曝光者信息
		dataIds := make([]int64, 0)
		dataList := make([]algo.IDataInfo, 0)
		for i, baseRes := range searchResList {
			info := &DataInfo{
				DataId:     baseRes.Id,
				SearchData: &searchResList[i],
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
