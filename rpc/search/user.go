package search

import (
	"encoding/json"
	"fmt"
)

const internalSearchNearUserListUrlV1 = "/search/nearby_user"

type apiSearchStruct struct {
	MinAge         string `json:"minAge"`
	MaxAge         string `json:"maxAge"`
	RoleName       string `json:"roleName"`
	Affection      string `json:"affection"`
	IsVip          string `json:"isVip"`
	ActiveDuration string `json:"activeDuration"`
	Horoscope      string `json:"horoscope"`
	HasImage       string `json:"hasImage"`
}

// 获取附近用户列表
func CallNearUserIdList(userId int64, lat, lng float32, offset, limit int64, filterJson string) ([]int64, error) {
	filters := []string{}
	if len(filterJson) > 0 {
		apiFilter := &apiSearchStruct{}
		if apiFilterErr := json.Unmarshal([]byte(filterJson), apiFilter); apiFilterErr != nil {
			return []int64{}, apiFilterErr
		} else {
			//普通
			if apiFilter.RoleName != "" { // 自我认同
				filters = append(filters, fmt.Sprintf("role_name:%s", apiFilter.RoleName))
			}
			if apiFilter.Affection != "" { // 感情状态
				filters = append(filters, fmt.Sprintf("affection:%s", apiFilter.Affection))
			}
			if apiFilter.MinAge != "" || apiFilter.MaxAge != "" { // 年龄范围
				filters = append(filters, fmt.Sprintf("age:[%s,%s]", apiFilter.MaxAge, apiFilter.MaxAge))
			}
			//会员特权
			if apiFilter.ActiveDuration != "" { // 是否在线
				filters = append(filters, fmt.Sprintf("activity_time:(now-%sm,)", apiFilter.ActiveDuration))
			}
			if apiFilter.IsVip == "1" { //是否vip
				filters = append(filters, fmt.Sprintf("vip_end_time:(now/m,)"))
			}
			if apiFilter.HasImage == "1" { //是否有图片
				filters = append(filters, fmt.Sprintf("user_image_count:[1,)"))
			}
			if apiFilter.Horoscope != "" { // 星座多选，逗号分割
				filters = append(filters, fmt.Sprintf("horoscope:%s", apiFilter.Horoscope))
			}
		}
	}
	return CallSearchIdList(internalSearchNearUserListUrlV1, userId, lat, lng, offset, limit, filters)
}
