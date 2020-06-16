package theme

import (
	"rela_recommend/algo"
	rutils "rela_recommend/utils"

	// "rela_recommend/models/pika"
	"rela_recommend/algo/utils"
	"rela_recommend/models/redis"
)

// 用户信息
type UserInfo struct {
	UserId       int64
	UserCache    *redis.UserProfile
	UserConcerns *rutils.SetInt64
	ThemeUser    *redis.ThemeUserProfile
}

// 话题信息
type DataInfo struct {
	DataId            int64
	UserCache         *redis.UserProfile
	MomentCache       *redis.Moments
	MomentExtendCache *redis.MomentsExtend
	MomentProfile     *redis.MomentsProfile
	RankInfo          *algo.RankInfo
	Features          *utils.Features
	ThemeProfile      *redis.ThemeProfile
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
