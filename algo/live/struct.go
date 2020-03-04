package live

import(
	"rela_recommend/algo"
	rutils "rela_recommend/utils"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	"rela_recommend/algo/utils"
)

// 用户信息
type UserInfo struct {
	UserId int64
	UserCache *pika.UserProfile
	LiveProfile *redis.LiveProfile
	UserConcerns *rutils.SetInt64
}

// 主播信息
type LiveInfo struct {
	UserId 		int64
	UserCache 	*pika.UserProfile
	LiveProfile *redis.LiveProfile
	LiveCache 	*pika.LiveCache
	RankInfo	*algo.RankInfo
	Features 	*utils.Features
}

func (self *LiveInfo) GetDataId() int64 {
	return self.UserId
}

func (self *LiveInfo) GetData() interface{} {
	return nil
}

func(self *LiveInfo) SetRankInfo(rankInfo *algo.RankInfo) {
	self.RankInfo = rankInfo
}

func(self *LiveInfo) GetRankInfo() *algo.RankInfo {
	return self.RankInfo
}

