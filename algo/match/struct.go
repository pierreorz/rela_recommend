package match

import(
	"rela_recommend/algo"
	"rela_recommend/models/pika"
)

type UserInfo struct {
	UserId int64
	UserCache *pika.UserProfile
}

type DataInfo struct {
	DataId 				int64
	UserCache 			*pika.UserProfile
	RankInfo			*algo.RankInfo
}

func (self *DataInfo) GetDataId() int64 {
	return self.DataId
}

func (self *DataInfo) GetRankInfo() *algo.RankInfo {
	return self.RankInfo
}

func (self *DataInfo) SetRankInfo(rankInfo *algo.RankInfo) {
	self.RankInfo = rankInfo
}
