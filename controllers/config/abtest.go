package config

import (
	"rela_recommend/utils/routers"
	"rela_recommend/service"
	"rela_recommend/utils/response"
	"rela_recommend/utils/request"
	"rela_recommend/service/abtest"
	"rela_recommend/log"
)

type AbTestRequest struct {
	App  		string  			`json:"app" form:"app"`
	Ua      	string 				`json:"ua" form:"ua"`
	Lat			float32 			`json:"lat" form:"lat"`
	Lng			float32 			`json:"lng" form:"lng"`
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
	ab := abtest.GetAbTestWithUaLocSetting(params.App, params.UserId, params.Ua, params.Lat, params.Lng, params.ParamsMap)
	c.JSON(response.FormatResponseV3(ab, nil))
}

