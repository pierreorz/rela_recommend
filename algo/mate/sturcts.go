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
