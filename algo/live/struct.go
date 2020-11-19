package live

import (
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/models/behavior"
	"rela_recommend/models/pika"
	"rela_recommend/models/redis"
	rutils "rela_recommend/utils"
)

// 用户信息
type UserInfo struct {
	UserId       int64
	UserCache    *redis.UserProfile
	LiveProfile  *redis.LiveProfile
	UserConcerns *rutils.SetInt64
}

func (self *UserInfo) GetBehavior() *behavior.UserBehavior {
	return nil
}

type LiveData struct {
	PreHourIndex int // 小时榜排名，1开始
	PreHourRank  int // 小时榜排名，1开始, 相同分数有并列名次
}

// 主播信息
type LiveInfo struct {
	UserId      int64
	UserCache   *redis.UserProfile
	LiveProfile *redis.LiveProfile
	LiveCache   *pika.LiveCache
	LiveData    *LiveData
	RankInfo    *algo.RankInfo
	Features    *utils.Features
}

func (self *LiveInfo) GetDataId() int64 {
	return self.UserId
}

func (self *LiveInfo) GetResponseData(ctx algo.IContext) interface{} {
	if self.LiveCache != nil {
		return self.LiveCache.Data4Api
	} else {
		return nil
	}
}

func (self *LiveInfo) SetRankInfo(rankInfo *algo.RankInfo) {
	self.RankInfo = rankInfo
}

func (self *LiveInfo) GetRankInfo() *algo.RankInfo {
	return self.RankInfo
}

func (self *LiveInfo) GetBehavior() *behavior.UserBehavior {
	return nil
}

func (self *LiveInfo) GetUserBehavior() *behavior.UserBehavior {
	return nil
}
