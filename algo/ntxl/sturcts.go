package ntxl

import (
	"rela_recommend/algo"
	"rela_recommend/models/behavior"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/search"
)

// UserInfo 用户信息
type UserInfo struct {
	UserId    int64
	UserCache *redis.UserProfile
}

func (ui *UserInfo) GetBehavior() *behavior.UserBehavior {
	return nil
}

// DataInfo 被推荐用户信息
type DataInfo struct {
	DataId     int64
	SearchData *search.MateTextResDataItem
	RankInfo   *algo.RankInfo
}

func (data *DataInfo) GetDataId() int64 {
	return data.DataId
}

func (data *DataInfo) GetResponseData(ctx algo.IContext) interface{} {
	return nil
}

func (data *DataInfo) SetRankInfo(rankInfo *algo.RankInfo) {
	data.RankInfo = rankInfo
}

func (data *DataInfo) GetRankInfo() *algo.RankInfo {
	return data.RankInfo
}

func (data *DataInfo) GetBehavior() *behavior.UserBehavior {
	return nil
}

func (data *DataInfo) GetUserBehavior() *behavior.UserBehavior {
	return nil
}
