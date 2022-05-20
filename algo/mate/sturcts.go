package mate

import (
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/models/behavior"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/search"
	"strconv"
	"strings"
)

// 用户信息
type UserInfo struct {
	UserId    int64
	UserCache *redis.UserProfile
}

func (self *UserInfo) GetBehavior() *behavior.UserBehavior {
	return nil
}

// 被推荐用户信息
type DataInfo struct {
	DataId     int64
	SearchData *search.MateTextResDataItem
	RankInfo   *algo.RankInfo
}

func (self *DataInfo) GetDataId() int64 {
	return self.DataId
}

func (self *DataInfo) GetResponseData(ctx algo.IContext) interface{} {
	sData := self.SearchData
	return RecommendResponseMateTextData{
		Text: sData.Text,
	}
}

func (self *DataInfo) SetRankInfo(rankInfo *algo.RankInfo) {
	self.RankInfo = rankInfo
}

func (self *DataInfo) GetRankInfo() *algo.RankInfo {
	return self.RankInfo
}

func (self *DataInfo) GetBehavior() *behavior.UserBehavior {
	return nil
}

func (self *DataInfo) GetUserBehavior() *behavior.UserBehavior {
	return nil
}

type RecommendResponseMateTextData struct {
	Text string `json:"text" form:"text"`
}

var RoleDict = map[string]string{"0": "不想透露", "1": "T", "2": "P", "3": "H", "4": "BI", "5": "其他", "6": "直女", "7": "腐女"}
var WantDict = map[string]string{"0": "不想透露", "1": "T", "2": "P", "3": "H", "4": "BI", "5": "其他", "6": "直女", "7": "腐女"}
//var	 affection_dict=map[string]string{"-1":"未设置","0":"不想透露","1":"单身","2":"约会中","3":"稳定关系","4":"已婚","5":"开放关系","6":"交往中","7":"等一个人"}
var HoroscopeDict = map[string]string{"0": "摩羯座", "1": "水瓶座", "2": "双鱼座", "3": "白羊座", "4": "金牛座", "5": "双子座", "6": "巨蟹座", "7": "狮子座", "8": "处女座", "9": "天平座", "10": "天蝎座", "11": "射手座"}

func GetSentenceData(id int64, text string, city []interface{},weight int) search.MateTextResDataItem {
	return search.MateTextResDataItem{
		Id:     id,
		Text:   text,
		Cities: city,
		Weight: weight,
		//TextType:"10",
		//TagType:nil,
	}
}
func MergeMap(mObj ...map[int64]float64) map[int64]float64 {
	newObj := map[int64]float64{}
	for _, m := range mObj {
		for k, v := range m {
			newObj[k] = v
		}
	}
	return newObj
}
//基础文案生成
var roleMap = map[string]string{"T": "1", "P": "1", "H": "1"}
var affection_list = map[string]string{"1": "1", "7": "1"}
func GetSentence(age int,horoscopeName string ,roleName string,occupation string,wantName string,intro string) []search.MateTextResDataItem{
	var baseVeiwList []search.MateTextResDataItem
	var textList []string
	if age >= 18 && age <= 40 {
		ageText := strconv.Itoa(age) + "岁"
		textList = append(textList, ageText)
	}
	textList = append(textList, horoscopeName)
	//自我认同
	if _, ok := roleMap[roleName]; ok {
		roleText := "我是" + roleName + "，你呢？"
		beasSentence := GetSentenceData(10002,roleText,nil,100)
		textList = append(textList, roleName)
		baseVeiwList = append(baseVeiwList, beasSentence)
	}
	//职业
	if occupation != "" && len(occupation) <= 6 {
		textList = append(textList, occupation)
	}
	//我想找的
	if _, ok := roleMap[wantName]; ok {
		wantText := "有" + wantName + "吗？"
		beasSentence := GetSentenceData(10001,wantText,nil,100)
		baseVeiwList = append(baseVeiwList, beasSentence)
	}
	//标签
	if intro != "" {
		beasSentence := GetSentenceData(10003,intro,nil,100)
		baseVeiwList = append(baseVeiwList, beasSentence)
	}
	log.Infof("baseSentence============+++++++========%+v",age)
	log.Infof("baseSentence============+++++++========%+v",horoscopeName)
	log.Infof("baseSentence============+++++++========%+v",roleName)
	log.Infof("baseSentence============+++++++========%+v",occupation)
	//用户基本文案
	log.Infof("baseSentence============+++++++========%+v",textList)
	log.Infof("len(textList)============+++++++========%+v",len(textList))
	if len(textList) > 1 {
		baseText := strings.Join(textList, "/")
		beasSentence := GetSentenceData(10000,baseText,nil,100)
		baseVeiwList = append(baseVeiwList, beasSentence)
	}
	return baseVeiwList
}
func GetBaseSentenceDatabyId(user *redis.UserProfile) []search.MateTextResDataItem {
	age:=user.Age
	horoscopeName:=HoroscopeDict[user.Horoscope]
	wantName := WantDict[user.WantRole]
	roleName := RoleDict[user.RoleName]
	occupation :=user.Occupation
	intro:=user.Intro
	baseSenten:=GetSentence(age,horoscopeName,wantName,roleName,occupation,intro)
	if baseSenten!=nil{
		return baseSenten
	}
	return nil
}

func GetBaseSentenceDataMap(userMap map[int64]*redis.UserProfile) []search.MateTextResDataItem {
	var onlineUserBaseMap []search.MateTextResDataItem
	var sentenceMap=make(map[string]int64)
	if len(userMap)>0 {
		for _, user := range userMap {
			age := user.Age
			horoscopeName := HoroscopeDict[user.Horoscope]
			wantName := WantDict[user.WantRole]
			roleName := RoleDict[user.RoleName]
			occupation := user.Occupation
			intro := user.Intro
			baseSenten := GetSentence(age, horoscopeName, wantName, roleName, occupation, intro)
			if len(baseSenten) > 0 {
				//文案去重
				for _, v := range baseSenten {
					id := strconv.FormatInt(v.Id, 10)
					text := v.Text
					weight := strconv.Itoa(v.Weight)
					cities := ""
					sentence := id + "," + text + "," + weight + "," + cities
					sentenceMap[sentence] = 1
				}
			}
		}
		//重新组装
		if len(sentenceMap) > 0 {
			for k, _ := range sentenceMap {
				id := strings.Split(k, ",")[0]
				text := strings.Split(k, ",")[1]
				weight := strings.Split(k, ",")[2]
				int_id, err := strconv.ParseInt(id, 10, 64)
				int_weight, err := strconv.Atoi(weight)
				if err == nil {
					resultSenten:=GetSentenceData(int_id,text,nil,int_weight)
					onlineUserBaseMap=append(onlineUserBaseMap,resultSenten)
				}
			}
		}
		return onlineUserBaseMap
	}
	return nil
}
