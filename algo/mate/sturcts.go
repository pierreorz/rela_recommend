package mate

import (
	"fmt"
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/models/behavior"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/search"
	"rela_recommend/utils"
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
		UserId: sData.UserId,
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
	UserId int64 `json:"userId" form:"userId"`
}

var RoleDict = map[string]string{"0": "不想透露", "1": "T", "2": "P", "3": "H", "4": "BI", "5": "其他", "6": "直女", "7": "腐女"}
var WantDict = map[string]string{"0": "不想透露", "1": "T", "2": "P", "3": "H", "4": "BI", "5": "其他", "6": "直女", "7": "腐女"}
//var	 affection_dict=map[string]string{"-1":"未设置","0":"不想透露","1":"单身","2":"约会中","3":"稳定关系","4":"已婚","5":"开放关系","6":"交往中","7":"等一个人"}
var HoroscopeDict = map[string]string{"0": "摩羯座", "1": "水瓶座", "2": "双鱼座", "3": "白羊座", "4": "金牛座", "5": "双子座", "6": "巨蟹座", "7": "狮子座", "8": "处女座", "9": "天平座", "10": "天蝎座", "11": "射手座"}
var CategNumsList=map[int64]int64{1:1,2:1,3:1,4:1,5:1,7:1,8:1,9:1,10:1,11:1,12:1,13:1,14:1,15:1,17:1,18:1,19:1,20:1,21:1,22:1,24:1,25:1}

func GetSentenceData(id int64, text string, city []interface{},weight int,TextType int64,TagType int64,UserId int64) search.MateTextResDataItem {
	return search.MateTextResDataItem{
		Id:     id,
		Text:   text,
		Cities: city,
		Weight: weight,
		TextType:TextType,
		TagType:TagType,
		UserId:UserId,
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

//var affection_list = map[string]string{"1": "1", "7": "1"}
func GetSentence(age int,horoscopeName string ,roleName string,occupation string,wantName string,intro string,textType int64,userId int64) []search.MateTextResDataItem{
	var baseVeiwList []search.MateTextResDataItem
	var textList []string
	if age >= 18 && age <= 40 {
		ageText := strconv.Itoa(age) + "岁"
		textList = append(textList, ageText)
	}
	textList = append(textList, horoscopeName)
	//自我认同
	if _, ok := roleMap[roleName]; ok {//10002
		roleText := "我是" + roleName + "，你呢？"
		beasSentence := GetSentenceData(10002, roleText, nil, 100, textType, 2, userId)
		baseVeiwList = append(baseVeiwList, beasSentence)
		textList = append(textList, roleName)
	}
	//职业
	if occupation != "" && len(occupation) <= 6 {
		textList = append(textList, occupation)
	}
	//我想找的
	if _, ok := roleMap[wantName]; ok { //10001
		wantText := "有" + wantName + "吗？"
		beasSentence := GetSentenceData(10001, wantText, nil, 100, textType, 1, userId)
		baseVeiwList = append(baseVeiwList, beasSentence)
	}
	//签名
	//if intro != "" {
	//	beasSentence := GetSentenceData(10003,intro,nil,100,textType,3)
	//	baseVeiwList = append(baseVeiwList, beasSentence)
	//}
	//用户基本文案
	if len(textList) > 1 { //10000
		baseText := strings.Join(textList, "/")
		beasSentence := GetSentenceData(10000, baseText, nil, 100, textType, 0, userId)
		baseVeiwList = append(baseVeiwList, beasSentence)

	}
	return baseVeiwList
}
func GetBaseSentenceDataById(user *redis.UserProfile,textType int64) []search.MateTextResDataItem {
	age:=user.Age
	horoscopeName:=HoroscopeDict[user.Horoscope]
	wantName := WantDict[user.WantRole]
	roleName := RoleDict[user.RoleName]
	occupation :=user.Occupation
	intro:=user.Intro
	userId:=user.UserId
	baseSenten:=GetSentence(age,horoscopeName,wantName,roleName,occupation,intro,textType,userId)
	if baseSenten!=nil{
		return baseSenten
	}
	return nil
}

func GetBaseSentenceDataMap(userMap map[int64]*redis.UserProfile,textType int64) []search.MateTextResDataItem {
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
			userId:=user.UserId
			baseSenten := GetSentence(age, horoscopeName, wantName, roleName, occupation, intro,textType,userId)
			if len(baseSenten) > 0 {
				//文案去重
				for _, v := range baseSenten {
					id := strconv.FormatInt(v.Id, 10)
					text := v.Text
					weight := strconv.Itoa(v.Weight)
					cities := ""
					textType :=strconv.FormatInt(v.TextType, 10)
					tagType  :=strconv.FormatInt(v.TagType, 10)
					userId  :=strconv.FormatInt(v.UserId, 10)
					sentence := id + "|$|" + text + "|$|" + weight + "|$|" + cities+"|$|"+textType+"|$|"+tagType+"|$|"+userId
					sentenceMap[sentence] = 1
				}
			}
		}
		//重新组装
		if len(sentenceMap) > 0 {
			for k, _ := range sentenceMap {
				id := strings.Split(k, "|$|")[0]
				text := strings.Split(k, "|$|")[1]
				weight := strings.Split(k, "|$|")[2]
				textType := strings.Split(k, "|$|")[4]
				tagType := strings.Split(k, "|$|")[5]
				userId := strings.Split(k, "|$|")[6]
				int_id, err := strconv.ParseInt(id, 10, 64)
				int_weight, err := strconv.Atoi(weight)
				int_textType,err:=strconv.ParseInt(textType, 10, 64)
				int_tagType,err:=strconv.ParseInt(tagType, 10, 64)
				int_userId,err:=strconv.ParseInt(userId, 10, 64)
				if err == nil {
					resultSenten:=GetSentenceData(int_id,text,nil,int_weight,int_textType,int_tagType,int_userId)
					onlineUserBaseMap=append(onlineUserBaseMap,resultSenten)
				}
			}
		}
		return onlineUserBaseMap
	}
	return nil
}

func GetCategSentenceData(text string,textType int64 ,categType int64,userId int64) []search.MateTextResDataItem {
	var categSentceList []search.MateTextResDataItem
	if len(text) > 0 {
		textList := strings.Split(text, "|$|")
		for i, v := range textList {
			id := textType*1000 + categType*100 + int64(i)
			text := v
			categSenten := GetSentenceData(id, text, nil, 100,textType,categType,userId)
			categSentceList = append(categSentceList, categSenten)
		}
		//log.Infof("categSentceList======================%+v",categSentceList)
		return categSentceList
	}
	return categSentceList
}


func GetRandomData(listLength int,categList [] int64) []int64 {
	var randomNum []int64
	if listLength > 0 {
		if listLength > 5{
			for i := 0; i <= 5; i++ {
				randomIndex := rand.Intn(listLength - 1)
				categNum:=categList[randomIndex]
				if _, ok := CategNumsList[categNum]; ok {
					randomNum=append(randomNum, categNum)
				}
			}
		}else{
			for i := 0; i < listLength; i++ {
				categNum:=categList[i]
				if _, ok := CategNumsList[categNum]; ok {
					randomNum=append(randomNum, categNum)
				}
			}
		}
		return randomNum
	}
	return randomNum
}

func min(l []float64) (min float64) {
	min = l[0]
	for _, v := range l {
		if v < min {
			min = v
		}
	}
	return
}

func GetDistanceSenten(kmMap map[int64]float64 ,textType int64 )[]search.MateTextResDataItem { //地理位置信息 textType:60
	var distanceList []search.MateTextResDataItem
	if len(kmMap) > 0{
		copyDict := make(map[int64]float64)
		for k,v:=range kmMap{
			copyDict[k]=v
		}
		minUser := utils.SortMapByValue(kmMap)
		minDistance := int(copyDict[minUser[len(minUser)-1]] / 1000.0)
		strKm := fmt.Sprintf("%d", minDistance)
		distanceText := "她距离你" + strKm + "公里"
		distanceSentence := GetSentenceData(60101, distanceText, nil, 100, textType, 1, minUser[len(minUser)-1])
		distanceList = append(distanceList, distanceSentence)
		return distanceList
	}
	return distanceList
}

func GetLikeSenten(nums int,textType int64)[]search.MateTextResDataItem {
	var likeList []search.MateTextResDataItem
	strNum:=strconv.Itoa(nums)
	likeText:="又有" + strNum +"人喜欢了你！"
	likeSentence := GetSentenceData(70101, likeText, nil, 100, textType, 1, 3568)
	likeList=append(likeList,likeSentence)
	return likeList
}

