package controllers

import (
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/log"
)

type UserCardParams struct {
	Key string `form:"key" json:"key" binding:"required"`
	Language string `form:"language" json:"language"`
}

func UserCardHTTP(c *routers.Context) {
	var params UserCardParams
	if err := bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(formatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}
	userId, err := service.DefaultKeyService().GetUserIdByKey(params.Key)
	if err != nil {
		c.JSON(formatResponse(nil, service.WarpError(service.ErrInvaKey, params.Language, "")))
		return
	}
	data, err := service.GetUserCard(userId)
	c.JSON(formatResponse(data, service.WarpError(err, params.Language, "")))
}