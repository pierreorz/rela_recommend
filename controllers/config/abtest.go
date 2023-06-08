package config

import (
	"rela_recommend/log"
	"rela_recommend/service"
	"rela_recommend/service/abtest"
	"rela_recommend/utils/request"
	"rela_recommend/utils/response"
	"rela_recommend/utils/routers"
)

type AbTestRequest struct {
	App       string            `json:"app" form:"app"`
	Ua        string            `json:"ua" form:"ua"`
	From      string            `json:"from" form:"from"`
	Region    string            `json:"region" form:"region"`
	Lat       float32           `json:"lat" form:"lat"`
	Lng       float32           `json:"lng" form:"lng"`
	UserId    int64             `json:"user_id" form:"user_id"`
	ParamsMap map[string]string `json:"params_map" from:"params_map"`
}

func AbTestHTTP(c *routers.Context) {
	var params AbTestRequest
	if err := request.Bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}

	if params.ParamsMap == nil {
		params.ParamsMap = make(map[string]string)
	}
	params.ParamsMap["from"] = params.From
	params.ParamsMap["region"] = params.Region
	log.Debugf("ab request: %+v", params)

	ab := abtest.GetAbTestWithUaLocSetting(params.App, params.UserId, params.Ua, params.Lat, params.Lng, params.ParamsMap)
	c.JSON(response.FormatResponseV3(ab, nil))
}
