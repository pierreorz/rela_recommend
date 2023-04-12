package user

import (
	"rela_recommend/algo"
	"rela_recommend/rpc/search"

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

func (self *UserInfo) GetBehavior() *behavior.UserBehavior {
	return nil
}

// 被推荐用户信息
type DataInfo struct {
	DataId       int64
	UserCache    *redis.UserProfile
	UserProfile  *redis.NearbyProfile
	LiveInfo     *pika.LiveCache
	RankInfo     *algo.RankInfo
	SearchFields *search.UserResDataItem

	UserItemBehavior     *behavior.UserBehavior
	BeenUserItemBehavior *behavior.UserBehavior
	ItemBehavior         *behavior.UserBehavior
}

func (self *DataInfo) GetDataId() int64 {
	return self.DataId
}

// 返回给服务端的数据
type responseItem struct {
	Lat            float64 `json:"lat"`            // 纬度
	Lng            float64 `json:"lng"`            // 经度
	LastActiveTime int64   `json:"lastActiveTime"` // 最后活跃时间
	IsOnLive       bool    `json:"is_on_live"`     // 正在直播
}

func (self *DataInfo) GetResponseData(ctx algo.IContext) interface{} {
	res := &responseItem{}
	if self.UserCache != nil {
		res.Lat = self.UserCache.Location.Lat
		res.Lng = self.UserCache.Location.Lon
		res.LastActiveTime = self.UserCache.LastUpdateTime
	}
	if self.LiveInfo != nil {
		res.IsOnLive = true
	}
	return res
}

func (self *DataInfo) SetRankInfo(rankInfo *algo.RankInfo) {
	self.RankInfo = rankInfo
}

func (self *DataInfo) GetRankInfo() *algo.RankInfo {
	return self.RankInfo
}

func (self *DataInfo) GetBehavior() *behavior.UserBehavior {
	return self.ItemBehavior
}

func (self *DataInfo) GetUserBehavior() *behavior.UserBehavior {
	return self.UserItemBehavior
}
