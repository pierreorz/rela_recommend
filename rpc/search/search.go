package search

import (
	"fmt"
	"rela_recommend/factory"
)


type momentListResIds struct {
	Ids 	[]int64		`json:"ids"`
}

type momentListRes struct {
	Data	momentListResIds	`json:"data"`
}

// 获取附近日志列表
func CallNearMomentList(userId int64, lat, lng float32, offset, limit int) ([]int64, error) {
	url := "/internal/moments/recommend"
	params := fmt.Sprintf("from=%d&limit=%d&lat=%f&lng=%f&user_id=%d", offset, limit, lat, lng, userId)
	res := momentListRes{}
	err := factory.SearchRpcClient.SendGETForm(url, params, res)
	if err != nil {
		return nil, err
	}
	return res.Data.Ids, nil
}