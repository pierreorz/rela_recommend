package routes

import (
	"rela_recommend/routers"
	"rela_recommend/controllers"
)

func RegisterRouters(router *routers.Routers) {
	router.GET("/recommend", controllers.IndexHTTP)
	router.GET("/recommend/userCard", controllers.UserCardHTTP)
	router.POST("/recommend/matchList", controllers.MatchRecommendListHTTP)
	router.NotFound = NotFound
}

func NotFound(c *routers.Context) {
	var notFound = make(map[string]interface{})
	notFound["errcode"] = "not_found"
	notFound["errdesc"] = "not_found"
	c.JSON(404, notFound)
}