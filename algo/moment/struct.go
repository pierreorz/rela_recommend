package moment

import(
	"rela_recommend/algo"
	rutils "rela_recommend/utils"
	"rela_recommend/models/redis"
	"rela_recommend/algo/utils"
)

// 用户信息
type UserInfo struct {
	UserId int64
	UserCache *redis.UserProfile
	UserConcerns *rutils.SetInt64
	MatchProfile *redis.MatchProfile
}

// 主播信息
type DataInfo struct {
	DataId 				int64
	UserCache 			*redis.UserProfile
	MomentCache 		*redis.Moments
	MomentExtendCache 	*redis.MomentsExtend
	MomentProfile		*redis.MomentsProfile
	RankInfo			*algo.RankInfo
	Features 			*utils.Features
	MatchProfile *redis.MatchProfile

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
