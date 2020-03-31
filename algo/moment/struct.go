package moment

import(
	"rela_recommend/algo"
	rutils "rela_recommend/utils"
	"rela_recommend/models/redis"
	"rela_recommend/algo/utils"
)

// 日志使用者
type UserInfo struct {
	UserId int64
	UserCache *redis.UserProfile
	UserConcerns *rutils.SetInt64
	MomentOfflineProfile *redis.MomentOfflineProfile
	MomentUserProfile *redis.MomentUserProfile
}

// 日志发布者
type DataInfo struct {
	DataId 				int64
	UserCache 			*redis.UserProfile
	MomentUserProfile       *redis.MomentUserProfile
	MomentOfflineProfile *redis.MomentOfflineProfile
	MomentCache 		*redis.Moments
	MomentExtendCache 	*redis.MomentsExtend
	MomentProfile		*redis.MomentsProfile
	RankInfo			*algo.RankInfo
	Features 			*utils.Features
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
