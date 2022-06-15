package moment

import (
	"rela_recommend/algo"
	"rela_recommend/algo/utils"
	"rela_recommend/models/behavior"
	"rela_recommend/models/redis"
	rutils "rela_recommend/utils"
)

// 日志使用者
type UserInfo struct {
	UserId       int64
	UserCache    *redis.UserProfile
	UserConcerns *rutils.SetInt64
	//MomentOfflineProfile *redis.MomentOfflineProfile
	MomentUserProfile  *redis.MomentUserProfile
	UserBehavior       *behavior.UserBehavior
	UserLiveProfile    *redis.UserLiveProfile
	UserContentProfile *redis.UserContentProfile
}

func (self *UserInfo) GetBehavior() *behavior.UserBehavior {
	return self.UserBehavior
}

// 日志发布者
type DataInfo struct {
	DataId               int64
	UserCache            *redis.UserProfile
	MomentUserProfile    *redis.MomentUserProfile
	MomentOfflineProfile *redis.MomentOfflineProfile
	MomentContentProfile *redis.MomentContentProfile
	MomentCache          *redis.Moments
	MomentExtendCache    *redis.MomentsExtend
	MomentProfile        *redis.MomentsProfile
	RankInfo             *algo.RankInfo
	Features             *utils.Features
	ItemBehavior         *behavior.UserBehavior
	ItemOfflineBehavior  map[string]int
	UserItemBehavior     *behavior.UserBehavior //用户对该发布日志的行为数据
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
	return self.ItemBehavior
}

func (self *DataInfo) GetUserBehavior() *behavior.UserBehavior {
	return nil
}

func (self *DataInfo) GetUserItemBehavior() *behavior.UserBehavior {
	return self.UserItemBehavior
}

func AdCanExposure(ctx algo.IContext, loc *redis.AdLoc, exposureRecords float64,jumpType int64) bool {
	nowTime := ctx.GetCreateTime().Unix()
	requestVersion := ctx.GetRequest().ClientVersion
	// 低版本不支持一些跳转类型
	adSkipDumpType := ctx.GetAbTest().GetInt64s("ad_skip_dump_type", "3,6")
	adMinVersion := ctx.GetAbTest().GetInt("ad_min_version", 50900)

	if loc.StartTime > nowTime {
		return false
	}

	if loc.EndTime < nowTime {
		return false
	}

	for _, _ty := range adSkipDumpType {
		if (requestVersion < adMinVersion) && (jumpType == _ty) {
			return false
		}
	}

	if exposureRecords >= loc.ExposureThreshold {
		return false
	}

	return true
}
