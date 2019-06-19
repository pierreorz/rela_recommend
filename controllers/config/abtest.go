package config

import (
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/utils/response"
	"rela_recommend/utils/request"
	"rela_recommend/service/abtest"
	"rela_recommend/log"
)

type AbTestRequest struct {
	App  		string  			`json:"app" form:"app"`
	UserId  	int64  				`json:"user_id" form:"user_id"`
	ParamsMap 	map[string]string	`json:"params_map" from:"params_map"`
}

func AbTestHTTP(c *routers.Context) {
	var params AbTestRequest
	if err := request.Bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}
	ab := abtest.GetAbTest(params.App, params.UserId)
	c.JSON(response.FormatResponseV3(ab, nil))
}

