package algo

import (
	"rela_recommend/service/abtest"
	"rela_recommend/service/performs"
	"time"
)

// ************************************************** 上下文
type IUserInfo interface {
}

type IDataInfo interface {
	GetDataId() int64
	GetResponseData(IContext) interface{}
	GetRankInfo() *RankInfo
	SetRankInfo(*RankInfo)
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
