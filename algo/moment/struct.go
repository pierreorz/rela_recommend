package moment

import (
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/models/redis"
	rutils "rela_recommend/utils"
	"rela_recommend/models/behavior"
)

// 日志使用者
type UserInfo struct {
	UserId       int64
	UserCache    *redis.UserProfile
	UserConcerns *rutils.SetInt64
	//MomentOfflineProfile *redis.MomentOfflineProfile
	MomentUserProfile *redis.MomentUserProfile
}

// 日志发布者
type DataInfo struct {
	DataId               int64
	UserCache            *redis.UserProfile
	MomentUserProfile    *redis.MomentUserProfile
	MomentOfflineProfile *redis.MomentOfflineProfile
	MomentCache          *redis.Moments
	MomentExtendCache    *redis.MomentsExtend
	MomentProfile        *redis.MomentsProfile
	RankInfo             *algo.RankInfo
	Features             *utils.Features
	ItemBehavior *behavior.UserBehavior


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
