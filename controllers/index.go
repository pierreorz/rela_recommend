package controllers

import (
	"rela_recommend/algo/base"
	"rela_recommend/utils/response"
	"rela_recommend/utils/routers"

	_ "rela_recommend/algo/ad"
	_ "rela_recommend/algo/live"
	_ "rela_recommend/algo/match"
	_ "rela_recommend/algo/mate"
	_ "rela_recommend/algo/moment"
	_ "rela_recommend/algo/moment/coarse"
	_ "rela_recommend/algo/moment_tag"
	_ "rela_recommend/algo/theme"
	_ "rela_recommend/algo/user"
	_ "rela_recommend/algo/label"
)

// curl 127.0.0.1:3200/rank/ -H "Content-Type: application/json" -d "{\"limit\":10,\"offset\":0,\"lat\":31.245714,\"lng\":121.486158,\"userId\":104708381,\"abMap\":{\"redis.json.thread.threshold\":\"100\"}}"
func IndexHTTP(c *routers.Context) {
	rsp, err := base.DoWithRoutersContext(c, "", "")
	c.JSON(response.FormatResponseV3(rsp, err))
}
