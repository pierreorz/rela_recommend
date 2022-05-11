package mate

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
	//momentCache := redis.NewMomentCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	themeUserCache := redis.NewThemeCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)

	if params.Limit == 0 {
		params.Limit = abtest.GetInt64("default_limit", 50)
	}

	var	 role_dict=map[string]string{"0":"不想透露","1":"T","2":"P","3":"H","4":"BI","5":"其他","6":"直女","7":"腐女"}
	var	 want_dict=map[string]string{"0":"不想透露","1":"T","2":"P","3":"H", "4":"BI","5":"其他","6":"直女","7":"腐女"}
	var	 affection_dict=map[string]string{"-1":"未设置","0":"不想透露","1":"单身","2":"约会中","3":"稳定关系","4":"已婚","5":"开放关系","6":"交往中","7":"等一个人"}
	var	horoscope_dict=map[string]string{"0":"摩羯座","1":"水瓶座","2":"双鱼座","3":"白羊座","4":"金牛座","5":"双子座","6":"巨蟹座","7":"狮子座","8":"处女座","9":"天平座","10":"天蝎座","11":"射手座"}
	//获取用户信息
	var user *redis.UserProfile
	pf.Run("user", func(*performs.Performs) interface{} {
		var userCacheErr error
		if user, _, userCacheErr = userCache.QueryByUserAndUsersMap(params.UserId, []int64{}); userCacheErr != nil {
			return rutils.GetInt(user != nil)
		}
		return userCacheErr
	})
	log.Infof("======================user profile:%+v",user)
	log.Infof("mate===============id",user.UserId)
	log.Infof("mate===============Occupation",user.Occupation)
	log.Infof("mate===============Intro",user.Intro)
	log.Infof("mate===============WantRole",user.WantRole)
	log.Infof("mate===============Affection",user.Affection)
	log.Infof("mate===============Horoscope",user.Horoscope)
	horoscope_name:=horoscope_dict[user.Horoscope]
	affection_name:=affection_dict[string(user.Affection)]
	want_name:=want_dict[user.WantRole]
	role_name:=role_dict[user.RoleName]
	log.Infof("==========================================")
	log.Infof("mate===============role_name",role_name)
	log.Infof("mate===============want_name",want_name)
	log.Infof("mate===============affection_name",affection_name)
	log.Infof("mate===============horoscope_name",horoscope_name)
	//用户基础信息生成文案

	//获取用户话题偏好
	UserProfileList:=[]int64{}
	var themeProfileMap = map[int64]*redis.ThemeUserProfile{}
	pf.Run("Theme_profile", func(*performs.Performs) interface{} {
		var themeUserCacheErr error
		userProfileUserIds := []int64{params.UserId}
		themeProfileMap, themeUserCacheErr = themeUserCache.QueryThemeUserProfileMap(userProfileUserIds)
		if themeUserCacheErr == nil {
			return len(themeProfileMap)
		}
		return themeUserCacheErr
	})
	log.Infof("ThemeProfileMap=============%+v",themeProfileMap)
	themeProfile:=themeProfileMap[params.UserId]
	themeTagLongMap:=themeProfile.AiTag.UserLongTag
	themeTagShortMap:=themeProfile.AiTag.UserShortTag
	for k,_ := range themeTagLongMap{
		UserProfileList=append(UserProfileList, k)
		log.Infof("ThemeLongProfile=============%+v",UserProfileList)
	}
	for k,_ := range themeTagShortMap{
		UserProfileList=append(UserProfileList, k)
		log.Infof("ThemeShortProfile=============%+v",UserProfileList)
	}

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

	pf.Run("build", func(*performs.Performs) interface{} {
		userInfo := &UserInfo{
			UserId: params.UserId,
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
