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
)

// 搜索接口
const internalSearchMatchListUrl = "/search/quick_match"

type searchMatchRequest struct {
	UserID       int64   `json:"userId" form:"userId"`
	Lng          float32 `json:"lng" form:"lng" `
	Lat          float32 `json:"lat" form:"lat" `
	PinnedIds    string  `json:"pinned_ids" form:"pinned_ids" `
	Filter       string  `json:"filter" form:"filter" `
	ReturnFields string  `json:"return_fields" form:"return_fields"`
}

// 已读接口
const internalSearchSeenListUrl = "/seen/quick_match"

type seenListRes struct {
	Data      string `json:"data"`
	Result    string `json:"result"`
	ErrCode   string `json:"errcode"`
	ErrDesc   string `json:"errdesc"`
	ErrDescEn string `json:"errdesc_en"`
	ReqeustID string `json:"request_id"`
}

type searchMatchSeenRequest struct {
	UserID     int64  `json:"userId" form:"userId"`
	Expiration int64  `json:"expiration" form:"expiration" `
	Scenery    string `json:"scenery" form:"scenery" `
	SeenIds    string `json:"seen_ids" form:"seen_ids" `
	Sink       string `json:"sink,omitempty" form:"seen_ids"`
}

// 获取用户列表, 过滤条件：
// role_name = "1,2,3"
func CallMatchList(ctx algo.IContext, userId int64, lat, lng float32, userIds []int64, user *redis.UserProfile) ([]int64, map[int64]*UserResDataItem, error) {
	abtest := ctx.GetAbTest()
	idList := make([]int64, 0)
	userSearchMap := make(map[int64]*UserResDataItem, 0)

	strIds := make([]string, len(userIds))
	for k, v := range userIds {
		strIds[k] = fmt.Sprintf("%d", v)
	}
	strsIds := strings.Join(strIds, ",")

	filters := []string{}
	if abtest.GetBool("filter_role_name", false) && user != nil {
		// 角色(0=不想透漏 1=T 2=P 3=H 4=Bi 5=其他 6=直女 7=腐女)
		wantRoles := strings.Join(strings.Split(user.WantRole, ""), ",")
		wantRole := utils.GetInt64s(wantRoles)
		wantRole0 := utils.Removes(wantRole, []int64{0, 5})
		wantRoleStrs := utils.JoinInt64s(wantRole0, ",")

		if wantRoleStrs != "" {
			filters = append(filters, fmt.Sprintf("role_name:%s", wantRoleStrs))
		}
	}

	params := searchMatchRequest{
		UserID:       userId,
		Lng:          lng,
		Lat:          lat,
		PinnedIds:    strsIds,
		Filter:       strings.Join(filters, "*"),
		ReturnFields: "id,cover_has_face,avatar_has_face,cover_beautiful,avatar_beautiful",
	}
	if paramsData, err := json.Marshal(params); err == nil {
		res := &userListRes{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchMatchListUrl, paramsData, res); err == nil {
			for _, element := range res.Data {
				newElement := element // 闭包
				idList = append(idList, newElement.Id)
				userSearchMap[newElement.Id] = &newElement
			}
			log.Infof("get paramsData:%s", string(paramsData))
			return idList, userSearchMap, err
		} else {
			return idList, userSearchMap, err
		}
	} else {
		return idList, userSearchMap, err
	}

}

// 获取已读用户列表
func CallMatchSeenList(userId, expiration int64, scenery string, userIds []int64) bool {

	strIds := make([]string, len(userIds))
	for k, v := range userIds {
		strIds[k] = fmt.Sprintf("%d", v)
	}
	strsIds := strings.Join(strIds, ",")

	params := searchMatchSeenRequest{
		UserID:     userId,
		Expiration: expiration,
		Scenery:    scenery,
		SeenIds:    strsIds,
	}
	if paramsData, err := json.Marshal(params); err == nil {
		log.Infof("paramsData%s", string(paramsData))
		res := &seenListRes{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchSeenListUrl, paramsData, res); err == nil {
			return res.Data == "ok"
		} else {
			return false
		}
	} else {
		return false
	}

}

// 获取已读用户列表
func CallNearbySeenList(userId int64, sink string, userIds []int64) bool {

	strIds := make([]string, len(userIds))
	for k, v := range userIds {
		strIds[k] = fmt.Sprintf("%d", v)
	}
	strsIds := strings.Join(strIds, ",")

	params := searchMatchSeenRequest{
		UserID:  userId,
		Sink:    sink,
		SeenIds: strsIds,
	}
	if paramsData, err := json.Marshal(params); err == nil {
		log.Infof("paramsData%s", string(paramsData))
		res := &seenListRes{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchSeenListUrl, paramsData, res); err == nil {
			return res.Data == "ok"
		} else {
			return false
		}
	} else {
		return false
	}

}
