package controllers

import (
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/utils/response"
)

func IndexHTTP(c *routers.Context) {
	var data = make(map[string]interface{})
	data["status"] = "OK"
	c.JSON(response.FormatResponse(data, service.WarpError(nil, "", "")))
}

