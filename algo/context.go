package algo

import (
	"rela_recommend/models/behavior"
	"rela_recommend/service/abtest"
	"rela_recommend/service/performs"
	"time"
)

// ************************************************** 上下文
type IUserInfo interface {
	GetBehavior() *behavior.UserBehavior // 获取当前用户行为
}

type IDataInfo interface {
	GetDataId() int64                        // 获取数据的ID
	GetResponseData(IContext) interface{}    // 获取数据返回的定制化数据
	GetRankInfo() *RankInfo                  // 获取推荐排名信息
	SetRankInfo(*RankInfo)                   // 设置推荐排名信息
	GetBehavior() *behavior.UserBehavior     // 获取数据的用户行为
	GetUserBehavior() *behavior.UserBehavior // 获取当前用户对此数据的行为
}

type IContext interface {
	GetRankId() string
	GetCreateTime() time.Time
	GetPlatform() int
	GetAppInfo() *AppInfo
	GetRequest() *RecommendRequest
	GetAbTest() *abtest.AbTest
	GetUserInfo() IUserInfo
	SetUserInfo(IUserInfo)
	GetDataIds() []int64
	SetDataIds([]int64)
	GetDataList() []IDataInfo
	GetDataLength() int
	SetDataList([]IDataInfo)
	GetDataByIndex(int) IDataInfo

	SetResponse(*RecommendResponse)
	GetResponse() *RecommendResponse

	GetPerforms() *performs.Performs

	DoNew(*AppInfo, *RecommendRequest) error
	DoInit() error
	DoBuildData() error
	DoFeatures() error
	DoAlgo() error
	DoStrategies() error
	DoSort() error
	DoPage() error
	DoLog() error
	Do(*AppInfo, *RecommendRequest) error
}
