package routes

import (
	"rela_recommend/routers"
	"rela_recommend/controllers"
	"rela_recommend/controllers/match"
	"rela_recommend/controllers/live"
)

func RegisterRouters(router *routers.Routers) {
	router.GET("/recommend", controllers.IndexHTTP)
	router.GET("/recommend/test", controllers.TestHTTP)
	router.POST("/recommend/matchList", match.MatchRecommendListHTTP)
	router.POST("/recommend/liveList", live.LiveRecommendListHTTP)
	router.NotFound = NotFound
}

func NotFound(c *routers.Context) {
	var notFound = make(map[string]interface{})
	notFound["errcode"] = "not_found"
	notFound["errdesc"] = "not_found"
	c.JSON(404, notFound)
}