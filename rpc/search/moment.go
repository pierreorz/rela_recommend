package search

import (
	"encoding/json"
	"fmt"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/utils"
	"strings"
	"time"
)

const internalSearchNearMomentListUrlV1 = "/search/friend_moments"
const internalSearchLiveMomentListUrl = "/search/on_live"

type SearchMomentResDataItem struct {
	Id int64 `json:"id"`
}

type searchMomentRes struct {
	Data      []SearchMomentResDataItem `json:"result_data"`
	TotalSize int                       `json:"total_size"`
	ErrCode   string                    `json:"errcode"`
	ErrEsc    string                    `json:"erresc"`
}

type searchMomentRequest struct {
	UserID   int64   `json:"userId" form:"userId"`
	Offset   int64   `json:"offset" form:"offset"`
	Limit    int64   `json:"limit" form:"limit"`
	Distance string  `json:"distance" form:"distance"`
	Lng      float32 `json:"lng" form:"lng" `
	Lat      float32 `json:"lat" form:"lat" `
	Filter   string  `json:"filter" form:"filter" `
}

type searchLiveMomentRequest struct {
	Filter string `json:"filter" form:"filter" `
}

//通过useridList获取正在直播的日志列表
func CallLiveMomentList(userIdList []int64) ([]int64, error) {
	idlist := make([]int64, 0)
	userIdListstr := make([]string, 0)
	for _, userid := range userIdList {
		userIdListstr = append(userIdListstr, fmt.Sprintf("%d", userid))
	}
	log.Warnf("useridlist %s\n", userIdListstr)
	filters := []string{
		fmt.Sprintf("user_id:%s", strings.Join(userIdListstr, ",")),
	}
	params := searchLiveMomentRequest{
		Filter: strings.Join(filters, "*"),
	}

	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchMomentRes{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchLiveMomentListUrl, paramsData, searchRes); err == nil {
			for _, element := range searchRes.Data {
				idlist = append(idlist, element.Id)
			}
			return idlist, err
		} else {
			return idlist, err
		}
	} else {
		return idlist, err
	}
}

// 获取附近日志列表
func CallNearMomentListV1(userId int64, lat, lng float32, offset, limit int64, momentTypes string, insertTimestamp float32, distance string,recommend bool) ([]int64, error) {
	idlist := make([]int64, 0)
	filters := []string{
		fmt.Sprintf("{moments_type:%s}", momentTypes),     //  moments Type
		fmt.Sprintf("insert_time:[%f,)", insertTimestamp), // time
	}
	if recommend{
		filters=append(filters,fmt.Sprintf("recommended:true"))
	}
	params := searchMomentRequest{
		UserID:   userId,
		Offset:   offset,
		Limit:    limit,
		Distance: distance,
		Lng:      lng,
		Lat:      lat,
		Filter:   strings.Join(filters, "*"),
	}

	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchMomentRes{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchNearMomentListUrlV1, paramsData, searchRes); err == nil {
			for _, element := range searchRes.Data {
				idlist = append(idlist, element.Id)
			}
			return idlist, err
		} else {
			return idlist, err
		}
	} else {
		return idlist, err
	}
}

func CallNearMomentListV2(userId int64, lat, lng float32, offset, limit int64, momentTypes string, insertTimestamp float32, distance string,recommend bool) ([]int64, error) {
	idlist := make([]int64, 0)
	filters := []string{
		fmt.Sprintf("{moments_type:%s}", momentTypes),     //  moments Type
		fmt.Sprintf("insert_time:[%f,)", insertTimestamp), // time
	}
	if recommend{
		filters=append(filters,fmt.Sprintf("positive_recommend:true"))
	}
	params := searchMomentRequest{
		UserID:   userId,
		Offset:   offset,
		Limit:    limit,
		Distance: distance,
		Lng:      lng,
		Lat:      lat,
		Filter:   strings.Join(filters, "*"),
	}

	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchMomentRes{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchNearMomentListUrlV1, paramsData, searchRes); err == nil {
			for _, element := range searchRes.Data {
				idlist = append(idlist, element.Id)
			}
			return idlist, err
		} else {
			return idlist, err
		}
	} else {
		return idlist, err
	}
}
///////////////////////////////////////////  话题过滤+审核+推荐置顶

const internalSearchAuditUrlFormatter = "/search/audit_%s"

type searchMomentAuditResDataItemTopInfo struct {
	Scenery   string `json:"scenery"` // 场景
	TopType   string `json:"top_type"`
	Weight    int64  `json:"weight"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}

type SearchMomentAuditResDataItem struct {
	Id          int64                                 `json:"id"`
	ParentId    int64                                 `json:"parent_id"`
	AuditStatus int                                   `json:"audit_status"`
	TopInfo     []searchMomentAuditResDataItemTopInfo `json:"top_info"`
}

// 获取当前场景是否置顶
func (self *SearchMomentAuditResDataItem) GetCurrentTopType(scenery string) string {
	currentTime := time.Now().Unix()
	for _, top := range self.TopInfo {
		if top.Scenery == scenery {
			if top.StartTime < currentTime && currentTime < top.EndTime {
				return strings.ToUpper(top.TopType) // 返回 TOP, RECOMMEND，SOFT
			}
		}
	}
	return ""
}

type searchMomentAuditRes struct {
	Data      []SearchMomentAuditResDataItem `json:"result_data"`
	TotalSize int                            `json:"total_size"`
	ErrCode   string                         `json:"errcode"`
	ErrEsc    string                         `json:"erresc"`
}

type searchMomentAuditRequest struct {
	UserID       int64  `json:"userId" form:"userId"`
	Filter       string `json:"filter" form:"filter" `
	ReturnFields string `json:"return_fields" form:"return_fields"`
}

// 获取附近日志列表, filtedAudit 是否筛选推荐合规
func CallMomentAuditMap(userId int64, moments []int64, scenery string, momentTypes string,
	returnedRecommend bool, filtedAudit bool) (map[int64]SearchMomentAuditResDataItem, map[int64]SearchMomentAuditResDataItem, error) {

	filters := []string{
		fmt.Sprintf("moments_type:%s", momentTypes),
	}

	ids := utils.JoinInt64s(moments, ",")

	idsFilter := fmt.Sprintf("id:%s", ids)
	if filtedAudit { // 是否要求人审
		idsFilter = fmt.Sprintf("id:%s*audit_status:1", ids)
	}

	if returnedRecommend { // 返回运营推荐数据，未审或过审的都可以通过
		recommendFilter := fmt.Sprintf("{top_info.scenery:%s*top_info.top_type:top,recommend*top_info.start_time:(,now/m]*top_info.end_time:[now/m,)*!audit_status:2}", scenery)
		filters = append(filters, fmt.Sprintf("{%s|%s}", idsFilter, recommendFilter))
	} else {
		filters = append(filters, idsFilter)
	}

	params := searchMomentAuditRequest{
		UserID:       userId,
		Filter:       strings.Join(filters, "*"),
		ReturnFields: "parent_id,audit_status,top_info",
	}

	resMap := map[int64]SearchMomentAuditResDataItem{}
	themeMap := map[int64]SearchMomentAuditResDataItem{}
	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchMomentAuditRes{}
		internalSearchAuditUrl := fmt.Sprintf(internalSearchAuditUrlFormatter, scenery)
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchAuditUrl, paramsData, searchRes); err == nil {
			for i, element := range searchRes.Data {
				if element.Id > 0 {
					resMap[element.Id] = searchRes.Data[i]
				}
				if element.ParentId > 0 {
					_, themeInfoOK := themeMap[element.ParentId] // 如果话题在返回结果出现两条，优先使用有日志推荐标志的
					if (!themeInfoOK) || (themeInfoOK && element.GetCurrentTopType(scenery) != "") {
						themeMap[element.ParentId] = searchRes.Data[i]
					}
				}
			}
			return resMap, themeMap, err
		} else {
			return resMap, themeMap, err
		}
	} else {
		return resMap, themeMap, err
	}
}

/////////////////////// 日志置顶以及推荐接口

const internalSearchTopUrlFormatter = "/search/top_moment"

func CallMomentTopMap(userId int64, scenery string) (map[int64]SearchMomentAuditResDataItem, error) {

	//filters := []string{
	//	fmt.Sprintf("moments_type:%s", momentTypes),
	//}

	filters := []string{
		fmt.Sprintf("{top_info.scenery:%s*top_info.top_type:top,recommend*top_info.start_time:(,now/m]*top_info.end_time:[now/m,)}", scenery),
	}
	// 返回运营推荐数据，未审或过审的都可以通过
	//recommendFilter := fmt.Sprintf("{top_info.scenery:%s*top_info.top_type:top,recommend*top_info.start_time:(,now/m]*top_info.end_time:[now/m,)}", scenery)
	//filters = append(filters, fmt.Sprintf("%s", recommendFilter))

	params := searchMomentAuditRequest{
		UserID:       userId,
		Filter:       strings.Join(filters, "*"),
		ReturnFields: "parent_id,audit_status,top_info",
	}

	resMap := map[int64]SearchMomentAuditResDataItem{}
	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchMomentAuditRes{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchTopUrlFormatter, paramsData, searchRes); err == nil {
			for i, element := range searchRes.Data {
				resMap[element.Id] = searchRes.Data[i]
			}
			return resMap, err
		} else {
			return resMap, err
		}
	} else {
		return resMap, err
	}
}
