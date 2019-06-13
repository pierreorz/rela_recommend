package theme

import(
	"rela_recommend/algo"
	rutils "rela_recommend/utils"
	"rela_recommend/models/pika"
)

// 用户信息
type UserInfo struct {
	UserId int64
	UserCache *pika.UserProfile
	UserConcerns *rutils.SetInt64
}

// 话题信息
type DataInfo struct {
	DataId 		int64
	RankInfo	*algo.RankInfo
}

func (self *DataInfo) GetDataId() int64 {
	return self.DataId
}

func(self *DataInfo) SetRankInfo(rankInfo *algo.RankInfo) {
	self.RankInfo = rankInfo
}

func(self *DataInfo) GetRankInfo() *algo.RankInfo {
	return self.RankInfo
}
