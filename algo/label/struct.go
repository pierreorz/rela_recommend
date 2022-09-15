package label

import (
	"rela_recommend/algo"
	"rela_recommend/models/behavior"
	"rela_recommend/models/redis"
)



type UserInfo struct {
	UserId       int64
	UserCache    *redis.UserProfile
}
func (self *UserInfo) GetBehavior() *behavior.UserBehavior {
	return nil
}

type DataInfo struct {
	DataId       int64
	RankInfo     *algo.RankInfo
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

func (self *DataInfo) GetUserItemBehavior() *behavior.UserBehavior {
	return nil
}
