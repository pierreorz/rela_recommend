package match

import (
	"rela_recommend/algo"
	"rela_recommend/algo/base"
	"rela_recommend/log"
	"rela_recommend/service"
	"rela_recommend/utils"
	"rela_recommend/utils/request"
	"rela_recommend/utils/response"
	"rela_recommend/utils/routers"
	"strings"
)

type MatchRecommendReqParams struct {
	Limit   int64             `json:"limit" form:"limit"`
	Offset  int64             `json:"offset" form:"offset"`
	Ua      string            `json:"ua" form:"ua"`
	UserId  int64             `json:"userId" form:"userId"`
	UserIds string            `json:"userIds" form:"userIds"`
	AbMap   map[string]string `json:"abMap" form:"abMap"`
}

type MatchRecommendResponse struct {
	Status  string
	RankId  string
	UserIds []int64
}

type MatchRecommendLog struct {
	RankId     string
	Index      int64
	UserId     int64
	ReceiverId int64
	Algo       string
	AlgoScore  float32
	Score      float32
	Features   string
	AbMap      string
}

func MatchRecommendListHTTP(c *routers.Context) {
	var params MatchRecommendReqParams
	if err := request.Bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}
	var userIds = make([]int64, 0)
	var userIdsStrs = strings.Split(params.UserIds, ",")
	for _, uid := range userIdsStrs {
		userIds = append(userIds, utils.GetInt64(uid))
	}

	var params2 = &algo.RecommendRequest{
		Limit:   params.Limit,
		Offset:  params.Offset,
		Ua:      params.Ua,
		Lat:     0.0,
		Lng:     0.0,
		UserId:  params.UserId,
		DataIds: userIds,
		AbMap:   params.AbMap,
	}
	ctx := &base.ContextBase{}
	err := ctx.Do(algo.GetAppInfo("match"), params2)
	res2 := ctx.GetResponse()
	res := MatchRecommendResponse{
		Status:  res2.Status,
		RankId:  res2.RankId,
		UserIds: res2.DataIds,
	}
	c.JSON(response.FormatResponse(res, service.WarpError(err, "", "")))
}
