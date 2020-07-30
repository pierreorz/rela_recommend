package live

import (
	"rela_recommend/algo"
	"rela_recommend/algo/base"
	"rela_recommend/log"
	"rela_recommend/service"
	"rela_recommend/utils"
	"rela_recommend/utils/request"
	"rela_recommend/utils/response"
	"rela_recommend/utils/routers"
)

type LiveRecommendRequest struct {
	Limit     int64  `json:"limit" form:"limit"`
	Offset    int64  `json:"offset" form:"offset"`
	Ua        string `json:"ua" form:"ua"`
	UserId    int64  `json:"userId" form:"userId"`
	LiveIdStr string `json:"liveIds" form:"liveIds"`
	LiveIds   []int64
}

type LiveRecommendResponse struct {
	Status  string  `json:"status" form:"status"`
	Message string  `json:"message" form:"message"`
	RankId  string  `json:"rankId" form:"rankId"`
	LiveIds []int64 `json:"liveIds" form:"liveIds"`
}

func LiveRecommendListHTTP(c *routers.Context) {
	var params LiveRecommendRequest
	if err := request.Bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}
	params.LiveIds = utils.GetInt64s(params.LiveIdStr)

	var params2 = &algo.RecommendRequest{
		Limit:   params.Limit,
		Offset:  params.Offset,
		Ua:      params.Ua,
		Lat:     0.0,
		Lng:     0.0,
		UserId:  params.UserId,
		DataIds: params.LiveIds,
	}
	ctx := &base.ContextBase{}
	err := ctx.Do(algo.GetAppInfo("live"), params2)
	res2 := ctx.GetResponse()
	res := LiveRecommendResponse{
		Status:  res2.Status,
		Message: res2.Message,
		RankId:  res2.RankId,
		LiveIds: res2.DataIds,
	}
	c.JSON(response.FormatResponse(res, service.WarpError(err, "", "")))
}
