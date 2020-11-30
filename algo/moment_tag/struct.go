package moment_tag

import (
	"rela_recommend/algo"
	// rutils "rela_recommend/utils"
	"rela_recommend/models/behavior"
	// "rela_recommend/algo/utils"
)

// 用户信息
type UserInfo struct {
	UserId int64
}

func (self *UserInfo) GetBehavior() *behavior.UserBehavior {
	return nil
}

// 标签信息
type DataInfo struct {
	DataId int64

	RankInfo *algo.RankInfo
}

func (self *DataInfo) GetDataId() int64 {
	return self.DataId
}

func (self *DataInfo) GetResponseData(ctx algo.IContext) interface{} {
	return nil
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
