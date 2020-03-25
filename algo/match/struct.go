package match

import (
	"rela_recommend/algo"
	"rela_recommend/models/redis"
)

type UserInfo struct {
	UserId       int64
	UserCache    *redis.UserProfile
	MatchProfile *redis.MatchProfile
}

type DataInfo struct {
	DataId       int64
	UserCache    *redis.UserProfile
	MatchProfile *redis.MatchProfile
	RankInfo     *algo.RankInfo
}

func (self *DataInfo) GetDataId() int64 {
	return self.DataId
}

func (self *DataInfo) GetResponseData() interface{} {
	return nil
}

func (self *DataInfo) GetRankInfo() *algo.RankInfo {
	return self.RankInfo
}

func (self *DataInfo) SetRankInfo(rankInfo *algo.RankInfo) {
	self.RankInfo = rankInfo
}
