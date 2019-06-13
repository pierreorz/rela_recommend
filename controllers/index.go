package controllers

import (
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/algo"
	"rela_recommend/utils/response"

	_ "rela_recommend/algo/moment"
	_ "rela_recommend/algo/moment/coarse"
	_ "rela_recommend/algo/theme"
)

func IndexHTTP(c *routers.Context) {
	rsp, err := algo.DoWithRoutersContext(c, "")
	c.JSON(response.FormatResponse(rsp, service.WarpError(err, "", "")))
}

