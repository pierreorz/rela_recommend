package controllers
import (
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/log"
)

type MatchRecommendReqParams struct {
	UserId int64 `json:"userId"`
	UserIds string `json:"userIds"`
}

func MatchRecommendListHTTP(c *routers.Context) {
	var params MatchRecommendReqParams
	if err := bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(formatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}
}