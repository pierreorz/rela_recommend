package mate

import (
	"rela_recommend/algo"
	"rela_recommend/models/behavior"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/search"
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
