package controllers

import (
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/factory"
)

func IndexHTTP(c *routers.Context) {
	var data = make(map[string]interface{})
	data["status"] = "OK"
	c.JSON(formatResponse(data, service.WarpError(nil, "", "")))
}


func bind(c *routers.Context, i interface{}) error {
	if factory.IsProduction {
		return c.BindAndSingnature(i)
	}
	return c.Bind(i)
}