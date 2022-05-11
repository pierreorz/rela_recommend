package mate

import (
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/redis"
	"rela_recommend/utils"
	"strconv"
	"strings"
)

// 内容较短，包含关键词的内容沉底
func BaseScoreStrategyItem(ctx algo.IContext, iDataInfo algo.IDataInfo, rankInfo *algo.RankInfo) error {
	abtest := ctx.GetAbTest()
	dataInfo := iDataInfo.(*DataInfo)
	sd := dataInfo.SearchData

	abSwitch := abtest.GetBool("mate_text_switch", false)
	if abSwitch {
		randomScore := float32(rand.Intn(100)) / 100.0
		rankInfo.Score = float32(sd.Weight) + randomScore
	}

	return nil
}
//多种类型的分发策略
func SortScoreItem(ctx algo.IContext) error {
	//var itemWeightMap= make(map[int64]int)
	abtest := ctx.GetAbTest()
	//后台配置曝光权重
	admin_weight := abtest.GetStrings("sentence_type_weight", "10:1,20:1,30:1,40:1,50:1")
	adminMap := make(map[int64]float64)
	for _, backtag := range admin_weight {
		type_nums :=utils.GetInt64(strings.Split(backtag, ":")[0])
		admin_weight_num :=utils.GetFloat64(strings.Split(backtag, ":")[0])
		adminMap[type_nums] = admin_weight_num
	}
	params := ctx.GetRequest()
	userCache := redis.NewUserCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	var user *redis.UserProfile
	var userCacheErr error
	var	 role_dict=map[string]string{"0":"不想透露","1":"T","2":"P","3":"H","4":"BI","5":"其他","6":"直女","7":"腐女"}
	var	 want_dict=map[string]string{"0":"不想透露","1":"T","2":"P","3":"H", "4":"BI","5":"其他","6":"直女","7":"腐女"}
	//var	 affection_dict=map[string]string{"-1":"未设置","0":"不想透露","1":"单身","2":"约会中","3":"稳定关系","4":"已婚","5":"开放关系","6":"交往中","7":"等一个人"}
	var	horoscope_dict=map[string]string{"0":"摩羯座","1":"水瓶座","2":"双鱼座","3":"白羊座","4":"金牛座","5":"双子座","6":"巨蟹座","7":"狮子座","8":"处女座","9":"天平座","10":"天蝎座","11":"射手座"}
	var roleMap=map[string]string{"T":"1","P":"1","H":"1"}
	var ageText string
	var roleText string
	var textList []string
	//var baseMap map[string]string
	if user, _, userCacheErr = userCache.QueryByUserAndUsersMap(params.UserId, []int64{}); userCacheErr != nil {
		horoscope_name:=horoscope_dict[user.Horoscope]
		want_name:=want_dict[user.WantRole]
		role_name:=role_dict[user.RoleName]

		//用户基础信息生成文案
		//base文案
		userAge:=user.Age
		if userAge>=18 && userAge<=40 {
			ageText = strconv.Itoa(userAge)
			textList=append(textList,ageText)
		}
		log.Infof("策略mmmmmm=========================")
		log.Infof("策略，ageText==============",ageText)
		textList=append(textList,horoscope_name)
		//自我认同
		if _, ok :=  roleMap[role_name];ok{
			log.Infof("我是"+role_name+"，你呢？")
			roleText=role_name
			textList=append(textList,roleText)
		}
		//职业
		if user.Occupation!="" && len(user.Occupation)<=6{
			textList=append(textList,roleText)
		}
		//用户基本文案
		if len(textList)>0{
			baseText:=strings.Join(textList, "/")
			log.Infof("baseText",baseText)
		}

		//我想找的
		if _, ok :=  roleMap[want_name];ok{
			log.Infof( "有"+want_name+"吗？")
		}
		if user.Intro!=""{
			log.Infof( "========Intro",user.Intro)
		}
	}



	//曝光逻辑
	for index := 0; index < ctx.GetDataLength(); index++ {
		randomScore := float32(rand.Intn(100)) / 100.0
		dataInfo := ctx.GetDataByIndex(index).(*DataInfo)
		sd := dataInfo.SearchData//SearchData缺少类型信息,
		//rankInfo := dataInfo.GetRankInfo()

		itemScore:=randomScore*float32(sd.Weight)
		log.Infof("itemScore===============%+v", itemScore)
		//rankInfo.AddRecommend("sortScoreItem", itemScore)
	}
	return nil
}

