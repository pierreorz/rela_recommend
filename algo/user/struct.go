package user

import (
	"rela_recommend/algo"
	// rutils "rela_recommend/utils"
	"rela_recommend/models/behavior"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	// "rela_recommend/algo/utils"
)

// 用户信息
type UserInfo struct {
	UserId      int64
	UserCache   *redis.UserProfile
	UserProfile *redis.NearbyProfile
}

// 被推荐用户信息
type DataInfo struct {
	DataId      int64
	UserCache   *redis.UserProfile
	UserProfile *redis.NearbyProfile
	LiveInfo    *pika.LiveCache
	RankInfo    *algo.RankInfo

	UserBehavior *behavior.UserBehavior
	ItemBehavior *behavior.UserBehavior
}

func (self *DataInfo) GetDataId() int64 {
	return self.DataId
}

// 返回给服务端的数据
type responseItem struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (self *DataInfo) GetResponseData(ctx algo.IContext) interface{} {
	return &responseItem{
		Lat: self.UserCache.Location.Lat,
		Lng: self.UserCache.Location.Lon,
	}
}

func (self *DataInfo) SetRankInfo(rankInfo *algo.RankInfo) {
	self.RankInfo = rankInfo
}

func (self *DataInfo) GetRankInfo() *algo.RankInfo {
	return self.RankInfo
}
