package search

import (
	"fmt"
	"rela_recommend/factory"
)

const internalSearchNearMomentListUrl = "/internal/moments/recommend"

type momentListResIds struct {
	Ids []int64 `json:"ids"`
}

type momentListRes struct {
	Data momentListResIds `json:"data"`
}

// 获取附近日志列表
func CallNearMomentList(userId int64, lat, lng float32, offset, limit int, momentTypes string, insertTimestamp float32, distance string) ([]int64, error) {
	params := fmt.Sprintf("from=%d&limit=%d&lat=%f&lng=%f&user_id=%d&momentsType=%s&insertTimestamp=%f&distance=%s",
		offset, limit, lat, lng, userId, momentTypes, insertTimestamp, distance)
	res := &momentListRes{}
	err := factory.SearchRpcClient.SendGETForm(internalSearchNearMomentListUrl, params, res)
	if err != nil {
		return nil, err
	}
	return res.Data.Ids, nil
}
