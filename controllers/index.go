package controllers

import (
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/algo"
	"rela_recommend/utils/response"
)

func IndexHTTP(c *routers.Context) {
	rsp, err := algo.DoWithRoutersContext(c, "")
	c.JSON(response.FormatResponse(rsp, service.WarpError(err, "", "")))
}

