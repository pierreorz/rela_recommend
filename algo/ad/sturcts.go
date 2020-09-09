package ad

import (
	"rela_recommend/algo"
	"rela_recommend/models/redis"
	"rela_recommend/rpc/search"
)

// 用户信息
type UserInfo struct {
	UserId    int64
	UserCache *redis.UserProfile
}

// 被推荐用户信息
type DataInfo struct {
	DataId     int64
	SearchData *search.SearchADResDataItem
	RankInfo   *algo.RankInfo
}

func (self *DataInfo) GetDataId() int64 {
	return self.DataId
}

func (self *DataInfo) GetResponseData(ctx algo.IContext) interface{} {
	os := ctx.GetRequest().GetOS()
	sData := self.SearchData
	return RecommendResponseADItemData{
		Id:        sData.Id,
		Title:     sData.Title,
		Source:    sData.AdvertSource,
		MediaType: sData.MediaType,
		MediaUrl:  sData.MediaUrl,
		DumpInfo: RecommendResponseADJump{
			DumpType:  sData.DumpType,
			Path:      sData.Path,
			ParamInfo: sData.ParamInfo,
		},
		AdwordsInfo: sData.GetPlatformAdwordsInfo(os),
		ShowTag:     sData.ShowTag,
		StartTime:   sData.StartTime,
		EndTime:     sData.EndTime,
	}
}

func (self *DataInfo) SetRankInfo(rankInfo *algo.RankInfo) {
	self.RankInfo = rankInfo
}

func (self *DataInfo) GetRankInfo() *algo.RankInfo {
	return self.RankInfo
}

type RecommendResponseADItemData struct {
	Id          int64                   `json:"id" form:"id"`
	Title       string                  `json:"title" form:"title"`
	Source      string                  `json:"source" form:"source"`       // 广告来源：own/chuanshanjia/qq/houyan/douniu/partner
	MediaType   string                  `json:"mediaType" form:"mediaType"` // 媒体类型 : image
	MediaUrl    string                  `json:"imageUrl" form:"imageUrl"`
	DumpInfo    RecommendResponseADJump `json:"dumpInfo" form:"dumpInfo"`
	AdwordsInfo string                  `json:"adwordsInfo" form:"adwordsInfo"` //广告商配置
	ShowTag     int                     `json:"showTag" form:"showTag"`         // 是否展示广告标签 0 不展示，1展示
	StartTime   int64                   `json:"startTime" form:"startTime"`     // 时间戳秒
	EndTime     int64                   `json:"endTime" form:"endTime"`         // 时间戳秒
}

type RecommendResponseADJump struct {
	DumpType  string `json:"dumpType" form:"dumpType"`
	Path      string `json:"path" form:"path"`
	ParamInfo string `json:"paramInfo" form:"paramInfo"`
}
