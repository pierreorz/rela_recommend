package mate

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/behavior"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/search"
	"rela_recommend/utils"
	"strconv"
	"strings"
	"time"
)

const (
	distanceTextType int64 = 60
	baseTextType int64 =10
	categTextType int64 =20
	adminUserid int64 =3568
	defaultImage string ="https://static.rela.me/ctMTEwMTEwLnBuZzE2NTcxNjI4ODY5NDc=.png"
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
		TagId: sData.TagType,
		TypeId: sData.TextType,
		DataId:sData.Id,
		MyAvatar:sData.MyAvatar,
		MatchAvatar:sData.MatchAvatar,
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
	TagId int64 `json:"tagId" form:"tagId"`
	TypeId int64 `json:"typeId" form:"typeId"`
	DataId int64 `json:"data_id" form:"data_id"`
	MyAvatar string `json:"myAvatar" form:"myAvatar"`
	MatchAvatar string `json:"matchAvatar" form:"matchAvatar"`
}

var RoleDict = map[string]string{"0": "不想透露", "1": "T", "2": "P", "3": "H", "4": "BI", "5": "其他", "6": "直女", "7": "腐女"}
var WantDict = map[string]string{"0": "不想透露", "1": "T", "2": "P", "3": "H", "4": "BI", "5": "其他", "6": "直女", "7": "腐女"}
//var	 affection_dict=map[string]string{"-1":"未设置","0":"不想透露","1":"单身","2":"约会中","3":"稳定关系","4":"已婚","5":"开放关系","6":"交往中","7":"等一个人"}
var HoroscopeDict = map[string]string{"0": "摩羯座", "1": "水瓶座", "2": "双鱼座", "3": "白羊座", "4": "金牛座", "5": "双子座", "6": "巨蟹座", "7": "狮子座", "8": "处女座", "9": "天平座", "10": "天蝎座", "11": "射手座"}
var CategNumsList=map[int64]int64{1:1,2:1,3:1,4:1,5:1,7:1,8:1,9:1,10:1,11:1,12:1,13:1,14:1,15:1,17:1,18:1,19:1,20:1,21:1,22:1,24:1,25:1}

var imageList =[]string{
	//"http://static.rela.me/game/avatar/1?imageslim",
	//"http://static.rela.me/game/avatar/2?imageslim",
	//"http://static.rela.me/game/avatar/3?imageslim",
	//"http://static.rela.me/game/avatar/4?imageslim",
	//"http://static.rela.me/game/avatar/5?imageslim",
	//"http://static.rela.me/game/avatar/6?imageslim",
	"https://static.rela.me/aw5bu657uELnBuZzE2NjE5MzI1MjczNTg=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTI5LnBuZzE2NjE5MzI1MjczODk=.png?imageView2/2/w/87",
	"https://static.rela.me/aw5bu657uELnBuZzE2NjE5MzI1MjczNzY=.png?imageView2/2/w/87",
	"https://static.rela.me/aw5bu657uELnBuZzE2NjE5MzI1MjczNzI=.png?imageView2/2/w/87",
	"https://static.rela.me/aw5bu657uELnBuZzE2NjE5MzI1MjczODE=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTY5LnBuZzE2NjE5MzI1MjczNjc=.png?imageView2/2/w/87",
	"https://static.rela.me/uO6YCJ5Yy6LnBuZzE2NjE5MzI1Mjc0MDA=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTcxLnBuZzE2NjE5MzI1MjczOTI=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTczLnBuZzE2NjE5MzI1MjczOTY=.png?imageView2/2/w/87",
	"https://static.rela.me/aw5bu657uELnBuZzE2NjE5MzI1MjczODU=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTc5LnBuZzE2NjE5MzI1Mjc0MDY=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTc3LnBuZzE2NjE5MzI1Mjc0MDM=.png?imageView2/2/w/87",
	"https://static.rela.me/uO6YCJ5Yy6LnBuZzE2NjE5MzI1Mjc0MDg=.png?imageView2/2/w/87",
	"https://static.rela.me/uO6YCJ5Yy6LnBuZzE2NjE5MzI1Mjc0MTE=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTg3LnBuZzE2NjE5MzI1Mjc0MTk=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTg1LnBuZzE2NjE5MzI1Mjc0MTc=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTgzLnBuZzE2NjE5MzI1Mjc0MTQ=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTk4LnBuZzE2NjE5MzI1Mjc0MjE=.png?imageView2/2/w/87",
	"https://static.rela.me/uO6YCJ5Yy6LnBuZzE2NjE5MzI1Mjc0MjY=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTkyLnBuZzE2NjE5MzI1Mjc0Mjg=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTk2LnBuZzE2NjE5MzI1Mjc0MjQ=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTYyLnBuZzE2NjE5MzI1Mjc0MzE=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTYyLnBuZzE2NjE5MzI1Mjc0MzM=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTYyLnBuZzE2NjE5MzI1Mjc0MzU=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTMxLnBuZzE2NjE5MzI1Mjc0NDE=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTQzLnBuZzE2NjE5MzI1Mjc0NDM=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTYyLnBuZzE2NjE5MzI1Mjc0Mzc=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTQ1LnBuZzE2NjE5MzI1Mjc0NDU=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTI5LnBuZzE2NjE5MzI1Mjc0Mzk=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTQ5LnBuZzE2NjE5MzI1Mjc0NDk=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTQ3LnBuZzE2NjE5MzI1Mjc0NDc=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTEwMy5wbmcxNjYxOTMyNTI3NDUx.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTU4LnBuZzE2NjE5MzI1Mjc0NTc=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTU1LnBuZzE2NjE5MzI1Mjc0NTU=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTUzLnBuZzE2NjE5MzI1Mjc0NTM=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTU5LnBuZzE2NjE5MzI1Mjc0NTk=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTEwMC5wbmcxNjYxOTMyNTI3NDYz.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTYyLnBuZzE2NjE5MzI1Mjc0NjE=.png?imageView2/2/w/87",
	"https://static.rela.me/u+5bGCLTEwNi5wbmcxNjYxOTMyNTI3NDY1.png?imageView2/2/w/87",
}
//根据时间戳随机到2个不同url
func GetRandomImage()[]string{
	var randomImageList []string
	timeSecond :=time.Now().Unix()
	randomNum:=timeSecond%int64(len(imageList))-1
	youSecond:=timeSecond+randomNum
	myIndex:=timeSecond%int64(len(imageList))
	matchIndex:=youSecond%int64(len(imageList))
	randomImageList=append(randomImageList,imageList[matchIndex])
	randomImageList=append(randomImageList,imageList[myIndex])
	return randomImageList
}

//获取返回数据
func GetSentenceData(id int64, text string, weight int,textType int64,tagType int64,userId int64,myAvatar string,matchAvatar string) search.MateTextResDataItem {
	return search.MateTextResDataItem{
		Id:     id,
		Text:   text,
		Weight: weight,
		TextType:textType,
		TagType:tagType,
		UserId:userId,
		MyAvatar:myAvatar,
		MatchAvatar:matchAvatar,
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
func GetSentence(age int,horoscopeName string ,roleName string,occupation string,wantName string,intro string,textType int64,userId int64,imageUrl string) []search.MateTextResDataItem{
	var baseVeiwList []search.MateTextResDataItem
	var textList []string
	//增加随机图片
	IamgeList:=GetRandomImage()
	if age >= 18 && age <= 40 {
		ageText := strconv.Itoa(age) + "岁"
		textList = append(textList, ageText)
	}
	textList = append(textList, horoscopeName)
	//自我认同
	if _, ok := roleMap[roleName]; ok {//10002
		roleText := "我是" + roleName + "，你呢？"
		beasSentence := GetSentenceData(10002, roleText, 100, 100, textType, userId,IamgeList[0],IamgeList[1])
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
		beasSentence := GetSentenceData(10001, wantText, 100, 100, textType, userId,IamgeList[0],IamgeList[1])
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
		beasSentence := GetSentenceData(10000, baseText, 100, 100, textType,  userId,IamgeList[0],IamgeList[1])
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
	imageUrl:=user.Avatar
	baseSenten:=GetSentence(age,horoscopeName,wantName,roleName,occupation,intro,textType,userId,imageUrl)
	if baseSenten!=nil{
		return baseSenten
	}
	return nil
}

func GetBaseSentenceDataMap(userMap map[int64]*redis.UserProfile,textType int64) []search.MateTextResDataItem {
	var onlineUserBaseMap []search.MateTextResDataItem
	if len(userMap)>0 {
		for _, user := range userMap {
			age := user.Age
			horoscopeName := HoroscopeDict[user.Horoscope]
			wantName := WantDict[user.WantRole]
			roleName := RoleDict[user.RoleName]
			occupation := user.Occupation
			intro := user.Intro
			userId:=user.UserId
			imageUrl:=user.Avatar
			baseSenten := GetSentence(age, horoscopeName, wantName, roleName, occupation, intro,textType,userId,imageUrl)
			if baseSenten!=nil {
				onlineUserBaseMap = append(onlineUserBaseMap, baseSenten...)
			}
		}
		return onlineUserBaseMap
	}
	return nil
}

func GetCategSentenceData(text string,textType int64 ,categType int64,userId int64) []search.MateTextResDataItem {
	var categSentceList []search.MateTextResDataItem
	if len(text) > 0 {
		//增加随机图片
		IamgeList:=GetRandomImage()
		textList := strings.Split(text, "|$|")
		for i, v := range textList {
			id := textType*1000 + categType*100 + int64(i)
			text := v
			categSenten := GetSentenceData(id, text, 100,textType,categType,userId,IamgeList[0],IamgeList[1])
			categSentceList = append(categSentceList, categSenten)
		}
		//log.Infof("categSentceList======================%+v",categSentceList)
		return categSentceList
	}
	return categSentceList
}


func GetRandomData(listLength int,categList [] int64) []int64 {
	var randomNum []int64
	if listLength > 0 {//对于偏好不做限制，原来限制最多有5个偏好出现。
		for i := 0; i < listLength; i++ {
			categNum:=categList[i]
			if _, ok := CategNumsList[categNum]; ok {
				randomNum=append(randomNum, categNum)
			}
		}
		return randomNum
	}
	return randomNum
}


func GetDistanceSenten(kmMap map[int64]float64 ,textType int64,IamgeMap map[int64]string )[]search.MateTextResDataItem { //地理位置信息 textType:60
	var distanceList []search.MateTextResDataItem
	if len(kmMap) > 0 {
		copyDict := make(map[int64]float64)
		for k, v := range kmMap {
			copyDict[k] = v
		}
		//增加随机图片
		IamgeList:=GetRandomImage()
		minUser := utils.SortMapByValue(kmMap)
		minDistance := copyDict[minUser[len(minUser)-1]] / 1000.0
		if minDistance < 1.0{
			strKm := fmt.Sprintf("%d", int(minDistance*1000))
			distanceText := "她距离你" + strKm + "米"
			distanceSentence := GetSentenceData(60101, distanceText, 100, textType, 1, minUser[len(minUser)-1],IamgeList[0],IamgeList[1])
			distanceList = append(distanceList, distanceSentence)
		}else{
			strKm := fmt.Sprintf("%d", int(minDistance))
			distanceText := "她距离你" + strKm + "公里"
			distanceSentence := GetSentenceData(60101, distanceText, 100,textType, 1, minUser[len(minUser)-1],IamgeList[0],IamgeList[1])
			distanceList = append(distanceList, distanceSentence)
		}
		return distanceList
	}
	return distanceList
}
//重新组装search结果，增加默认图片
func GetSearchIamge( searchResult []search.MateTextResDataItem) []search.MateTextResDataItem{
	var searchImageResult []search.MateTextResDataItem
	for _,v:=range searchResult {
		//增加随机图片
		IamgeList := GetRandomImage()
		imageResult := GetSentenceData(v.Id, v.Text, v.Weight, v.TextType, v.TagType, v.UserId, IamgeList[0], IamgeList[1])
		searchImageResult = append(searchImageResult, imageResult)
	}
	return searchImageResult
}

//合并偏好用户
func GetMergeMap(themeMap map[int64][]int64,momMap map[int64][]int64) map[int64][]int64{
	allMap:= make(map[int64][]int64)
	for k,v:=range themeMap{
		allMap[k]=v
	}
	for k,v:=range momMap{
		if _,ok:=allMap[k];ok{
			allMap[k]=append(allMap[k],v...)
		}else {
			allMap[k]=v
		}
	}

	return allMap
}
//根据结果抽取偏好
func GetCategoryRandomData(categMap map[int64][]int64 ) map[int64]int64{
	resultMap:= make(map[int64]int64)
	for k,v:= range categMap{
		index:=rand.Intn(len(v))
		userId:=v[index]
		resultMap[k]=userId
	}

	return resultMap
}

//建立结构pika缓存结构体
type MatePika struct {
	DataIds []int64 `json:"dataIds"`
	DataList []algo.IDataInfo `json:"dataList"`
}

func GetPikaUser(ctx algo.IContext,userid int64) ([]int64,[]algo.IDataInfo,error){
	mateCategCache := redis.NewMateCaegtCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)

	var matePikaString string
	var matePikaCacheErr error
	var matePika MatePika
	if matePikaString,matePikaCacheErr = mateCategCache.QueryMataUserPikaMap(userid); matePikaCacheErr == nil{
		log.Info("matePikaStringr=========",matePikaString)
		if err := json.Unmarshal([]byte(matePikaString), &matePika); err != nil {
			log.Info("mate pika JSON error=========",err)
		}
	}
	dataIds:=matePika.DataIds
	dataList:=matePika.DataList
	return dataIds,dataList,matePikaCacheErr
}

func SetPikaUser(ctx algo.IContext,userid int64,dataIds []int64,dataList []algo.IDataInfo) (int,error){
	reqMatePika:=MatePika{dataIds,dataList}
	reqMatePikaString, _ := json.Marshal(reqMatePika)

	mateCategCache := redis.NewMateCaegtCacheModule(ctx, &factory.CacheCluster, &factory.PikaCluster)
	var cacheTime=10*60

	dataLen,err:=mateCategCache.SetMataUserPikaMap(userid,reqMatePikaString,cacheTime)

	return dataLen,err
}





//func GetLikeSenten(nums int,textType int64)[]search.MateTextResDataItem {
//	var likeList []search.MateTextResDataItem
//	strNum:=strconv.Itoa(nums)
//	likeText:="又有" + strNum +"人喜欢了你！"
//	likeSentence := GetSentenceData(70101, likeText, nil, 100, textType, 1, 3568)
//	likeList=append(likeList,likeSentence)
//	return likeList
//}

