package controllers

import (
	"rela_recommend/utils/routers"
	"rela_recommend/algo"
	"rela_recommend/utils/response"

	_ "rela_recommend/algo/moment"
	_ "rela_recommend/algo/moment/coarse"
	_ "rela_recommend/algo/theme"
	_ "rela_recommend/algo/live"
)

func IndexHTTP(c *routers.Context) {
	rsp, err := algo.DoWithRoutersContext(c, "", "")
	c.JSON(response.FormatResponseV3(rsp, err))
}
