package mate

import (
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
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

	if params.Limit == 0 {
		params.Limit = abtest.GetInt64("default_limit", 50)
	}

	var	 role_dict=map[string]string{"0":"不想透露","1":"T","2":"P","3":"H","4":"BI","5":"其他","6":"直女","7":"腐女"}
	var	 want_dict=map[string]string{"0":"不想透露","1":"T","2":"P","3":"H", "4":"BI","5":"其他","6":"直女","7":"腐女"}
	//var	 affection_dict=map[string]string{"-1":"未设置","0":"不想透露","1":"单身","2":"约会中","3":"稳定关系","4":"已婚","5":"开放关系","6":"交往中","7":"等一个人"}
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
	horoscope_name:=horoscope_dict[user.Horoscope]
	want_name:=want_dict[user.WantRole]
	role_name:=role_dict[user.RoleName]

	//用户基础信息生成文案
	//base文案
	var roleMap=map[string]string{"T":"1","P":"1","H":"1"}
	var affection_list=map[string]string{"1":"1","7":"1"}
	var ageText string
	var roleText string
	var textList []string
	searchBase := search.SearchType{}
	searchCateg:= search.SearchType{}
	var baseVeiwList  []search.MateTextResDataItem
	userAge:=user.Age
	if userAge>=18 && userAge<=40 {
		ageText = strconv.Itoa(userAge)+"岁"
		textList=append(textList,ageText)
	}
	//log.Infof("ageText==============",ageText)
	textList=append(textList,horoscope_name)
	//自我认同
	if _, ok :=  roleMap[role_name];ok{
		log.Infof("我是"+role_name+"，你呢？")
		roleText="我是"+role_name+"，你呢？"
		textList=append(textList,role_name)
		beasSentence:=search.MateTextResDataItem{
			Id: 10002,
			Text:roleText,
			Cities:nil,
			Weight:100,
			//TextType:"10",
			//TagType:nil,
		}
		baseVeiwList=append(baseVeiwList,beasSentence)
	}
	//职业
	if user.Occupation!="" && len(user.Occupation)<=6{
		textList=append(textList,roleText)
	}
	//用户基本文案
	if len(textList)>0{
		baseText:=strings.Join(textList, "/")
		log.Infof("baseText",baseText)
		beasSentence:=search.MateTextResDataItem{
			Id: 10000,
			Text:baseText,
			Cities:nil,
			Weight:100,
			//TextType:"10",
			//TagType:nil,
		}
		baseVeiwList=append(baseVeiwList,beasSentence)

	}
	//我想找的
	if _, ok :=  roleMap[want_name];ok{
		wantText:="有"+want_name+"吗？"
		//log.Infof( "有"+want_name+"吗？")
		beasSentence:=search.MateTextResDataItem{
			Id: 10001,
			Text:wantText,
			Cities:nil,
			Weight:100,
			//TextType:"10",
			//TagType:nil,
		}
		baseVeiwList=append(baseVeiwList,beasSentence)
	}
	if user.Intro!=""{
		//log.Infof( "========Intro",user.Intro)
		beasSentence:=search.MateTextResDataItem{
			Id: 10003,
			Text:user.Intro,
			Cities:nil,
			Weight:100,
			//TextType:"10",
			//TagType:nil,
		}
		baseVeiwList=append(baseVeiwList,beasSentence)
	}
	//基础数据需要搜索
	if _,ok:=affection_list[string(user.Affection)];ok{
		//log.Infof( "========Intro",user.Affection)
		searchBase.TextType="10"
		searchBase.TagType=append(searchBase.TagType, "4")
	}
	//log.Infof("baseVeiwText=============%+v", baseVeiwList)
	//log.Infof("categSearch=============%+v", searchBase)
	//情感搜索
	//获取用户话题偏好
	userProfileMap:= map[int64]float64{}
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
	if len(themeProfileMap)>0 {
		themeProfile := themeProfileMap[params.UserId]
		themeTagLongMap := themeProfile.AiTag.UserLongTag
		themeTagShortMap := themeProfile.AiTag.UserShortTag
		if len(themeTagLongMap) > 0 {
			for k, _ := range themeTagLongMap {
				if _, ok := userProfileMap[k]; ok {
					userProfileMap[k] += 1.0
				} else {
					userProfileMap[k] = 1.0
				}
			}
		}
		if len(themeTagShortMap) > 0 {
			for k, _ := range themeTagShortMap {
				if _, ok := userProfileMap[k]; ok {
					userProfileMap[k] += 1
				} else {
					userProfileMap[k] = 1
				}
			}
		}
		//log.Infof("ThemeShortProfile=============%+v", userProfileMap)
		resultList:=rutils.SortMapByValue(userProfileMap)
		for i,v := range resultList{
			if i < 2 {
				categid:= strconv.Itoa(int(v))
				searchCateg.TagType=append(searchCateg.TagType, categid)
			}
		}
		searchCateg.TextType="20"
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
