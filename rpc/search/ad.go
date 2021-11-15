package search

import (
	"encoding/json"
	"fmt"
	"rela_recommend/algo"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/models/redis"
	"rela_recommend/utils"
	"strings"
	"time"
)

const internalSearchAdListUrl = "/search/ads"

type SearchADResDataItem struct {
	Id               int64   `json:"id"`
	Title            string  `json:"title"`
	Location         string  `json:"location"`
	Version          int     `json:"version"`
	DisplayType      string  `json:"display_type"`
	TestUsers        []int64 `json:"test_users"`
	AppSource        string  `json:"app_source"`
	AdvertSource     string  `json:"advert_source"`
	Weight           int     `json:"weight"`
	Exposure         int     `json:"exposure"`
	Cpm              int     `json:"cpm"`
	ShowTag          int     `json:"show_tag"`
	MediaType        string  `json:"media_type"`
	MediaUrl         string  `json:"media_url"`
	DumpType         string  `json:"dump_type"`
	Path             string  `json:"path"`
	ParamInfo        string  `json:"param_info"`
	AdwordsInfo      string  `json:"adwords_info" form:"adwords_info"`
	Status           int     `json:"status"`
	StartTime        int64   `json:"start_time"`
	EndTime          int64   `json:"end_time"`
	CreateTime       int64   `json:"create_time"`
	UpdateTime       int64   `json:"update_time"`
	HistoryExposures int     `json:"history_exposures"`
	HistoryClicks    int     `json:"history_clicks"`
	HistoryFails     int     `json:"history_fails"`
}

// 返回给客户端类型
type SearchADResDataItemAdwordsInfo struct {
	LocationId  string `json:"location_id,omitempty"`
	AppId       string `json:"app_id,omitempty"`
	WidthHeight string `json:"width_height,omitempty"` // "400:50"
}

// 配置的平台设定
type searchADResDataItemAdwordsInfoPlatform struct {
	IOS     *SearchADResDataItemAdwordsInfo `json:"ios,omitempty"`
	Android *SearchADResDataItemAdwordsInfo `json:"android,omitempty"`
	Other   *SearchADResDataItemAdwordsInfo `json:"other,omitempty"`
}

// 获取分平台的配置
func (self *SearchADResDataItem) GetPlatformAdwordsInfo(os string) *SearchADResDataItemAdwordsInfo {
	var res = &SearchADResDataItemAdwordsInfo{}
	if err := json.Unmarshal([]byte(self.AdwordsInfo), &res); err == nil {
		log.Debugf("adwordsInfo %d outer error:%v\n", self.Id, err)
	}

	var platforms = &searchADResDataItemAdwordsInfoPlatform{}
	if err := json.Unmarshal([]byte(self.AdwordsInfo), &platforms); err == nil {
		var platformInfo *SearchADResDataItemAdwordsInfo
		switch os {
		case "ios":
			platformInfo = platforms.IOS
		case "android":
			platformInfo = platforms.Android
		case "other":
			platformInfo = platforms.Other
		}
		if platformInfo != nil { // 重写字段
			res.LocationId = utils.CoalesceString(platformInfo.LocationId, res.LocationId)
			res.AppId = utils.CoalesceString(platformInfo.AppId, res.AppId)
			res.WidthHeight = utils.CoalesceString(platformInfo.WidthHeight, res.WidthHeight)
		}
	} else {
		log.Debugf("adwordsInfo %d platform error:%v\n", self.Id, err)
	}
	return res
}

type searchADRes struct {
	Data      []SearchADResDataItem `json:"result_data"`
	TotalSize int                   `json:"total_size"`
	ErrCode   string                `json:"errcode"`
	ErrEsc    string                `json:"erresc"`
}

type searchRequest struct {
	UserID        int64   `json:"userId" form:"userId"`
	Offset        int64   `json:"offset" form:"offset"`
	Limit         int64   `json:"limit" form:"limit"`
	Lng           float32 `json:"lng" form:"lng" `
	Lat           float32 `json:"lat" form:"lat" `
	MobileOS      string  `json:"mobileOS" form:"mobileOS"`
	ClientVersion int     `json:"clientVersion" form:"clientVersion"`
	Query         string  `json:"query" form:"query" `
	Filter        string  `json:"filter" form:"filter" `
	ReturnFields  string  `json:"return_fields" form:"return_fields" `
}

// ** 获取广告列表， 过滤条件：
// app_source = rela     //视app而定
// location = init      //视场景而定
// （ status = 2 //上线 ）or  （status = 1 and current_user_id in TestUsers）
// start_time < current_time < end_time
// (version = 0 ) or ( version > 0 and version <= current_version)
// (display_type =1) or (display_type=2 and current_user_isVip) or (display_type=3 and ！current_user_isVip)
// (exposure = 0) or (exposure < history_exposures)
// (client_os = '') or (client_os = current_os)
func CallAdList(app string, request *algo.RecommendRequest, user *redis.UserProfile) ([]SearchADResDataItem, error) {
	now := time.Now().Unix()

	displayTypes := "1"
	if user.IsVip == 1 {
		displayTypes = "1,2" // 不限制，会员可见
	} else {
		displayTypes = "1,3" // 不限制，会员不可见
	}
	//dumpType为3外部跳转，15内部跳转
	dumpType :="15"
	filters := []string{
		fmt.Sprintf("app_source:%s*location:%s", app, request.Type),        // base
		fmt.Sprintf("{status:2|{status:1*test_users:%d}}", request.UserId), // user
		fmt.Sprintf("start_time:[,%d]*end_time:[%d,]", now, now),           // time
		fmt.Sprintf("{version:0|{version:[,%d]}}", request.ClientVersion),  // version
		fmt.Sprintf("{display_type:%s}", displayTypes),                     // display vip
		fmt.Sprintf("can_exposure:true"),                                   // exposure cnt: search 端处理
		fmt.Sprintf("{client_os:|client_os:%s}", request.GetOS()),          // exposure cnt
		fmt.Sprintf("{dump_type:%s}",dumpType),          				   // fiter dump_type 不为3的数据
	}

	params := searchRequest{
		UserID:        request.UserId,
		Offset:        request.Offset,
		Limit:         request.Limit,
		Lng:           request.Lng,
		Lat:           request.Lat,
		MobileOS:      request.MobileOS,
		ClientVersion: request.ClientVersion,
		Query:         "",
		Filter:        strings.Join(filters, "*"),
		ReturnFields:  "*",
	}
	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchADRes{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchAdListUrl, paramsData, searchRes); err == nil {
			return searchRes.Data, err
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
func CallFeedAdList(app string, request *algo.RecommendRequest, user *redis.UserProfile) ([]SearchADResDataItem, error) {
	now := time.Now().Unix()
	displayTypes := "1"
	if user.IsVip == 1 {
		displayTypes = "1,2" // 不限制，会员可见
	} else {
		displayTypes = "1,3" // 不限制，会员不可见
	}
	dumpType:="3"//外部跳转类型
	//AdvertSource:="taobaoxiaoyouxi"//目前外部涞源
	filters := []string{
		fmt.Sprintf("app_source:%s*location:%s", app, request.Type),        // base
		fmt.Sprintf("{status:2|{status:1*test_users:%d}}", request.UserId), // user
		fmt.Sprintf("start_time:[,%d]*end_time:[%d,]", now, now),           // time
		fmt.Sprintf("{version:0|{version:[,%d]}}", request.ClientVersion),  // version
		fmt.Sprintf("{display_type:%s}", displayTypes),                     // display vip
		fmt.Sprintf("can_exposure:true"),                                   // exposure cnt: search 端处理
		fmt.Sprintf("{client_os:|client_os:%s}", request.GetOS()),          // exposure cnt
		fmt.Sprintf("{dump_type:%s}",dumpType),          				   // fiter dump_type 不为3的数据
		//fmt.Sprintf("{advert_source:%s}",AdvertSource),          	       // 广告涞源，淘宝小游戏
	}
	params := searchRequest{
		UserID:        request.UserId,
		Offset:        request.Offset,
		Limit:         request.Limit,
		Lng:           request.Lng,
		Lat:           request.Lat,
		MobileOS:      request.MobileOS,
		ClientVersion: request.ClientVersion,
		Query:         "",
		Filter:        strings.Join(filters, "*"),
		ReturnFields:  "*",
	}
	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchADRes{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchAdListUrl, paramsData, searchRes); err == nil {
			return searchRes.Data, err
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

