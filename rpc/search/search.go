package search

import (
	"fmt"
	"rela_recommend/factory"
)
const internalSearchNearMomentListUrl = "/internal/moments/recommend"

type momentListResIds struct {
	Ids 	[]int64		`json:"ids"`
}

type momentListRes struct {
	Data	momentListResIds	`json:"data"`
}

type momentsListReq struct {
	UserId          int64    `form:"userId" json:"userId"`
	Lng             float64  `form:"lng" json:"lng"`
	Lat             float64  `form:"lat" json:"lat"`
	From            int      `form:"from" json:"from"`
	Limit           int      `form:"limit" json:"limit"`
	MomentsType     []string `form:"limit" json:"momentsType"`
	InsertTimestamp float64  `form:"insertTimestamp" json:"insertTimestamp"`
}

// 获取附近日志列表
func CallNearMomentList(userId int64, lat, lng float32, offset, limit int, momentTypes string, insertTimestamp float32) ([]int64, error) {
	params := fmt.Sprintf("from=%d&limit=%d&lat=%f&lng=%f&user_id=%d&momentsType=%s&insertTimestamp=%f", 
						  offset, limit, lat, lng, userId, momentTypes, insertTimestamp)
	res := &momentListRes{}
	err := factory.SearchRpcClient.SendGETForm(internalSearchNearMomentListUrl, params, res)
	if err != nil {
		return nil, err
	}
	return res.Data.Ids, nil
}