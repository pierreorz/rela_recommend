package routes

import (
	"rela_recommend/utils/routers"
	"rela_recommend/controllers"
	"rela_recommend/controllers/match"
	"rela_recommend/controllers/live"
	"rela_recommend/controllers/theme"
	"rela_recommend/controllers/moment"
	"rela_recommend/controllers/config"
	"rela_recommend/controllers/segment"
)

func RegisterRouters(router *routers.Routers) {
	router.POST("/config/abtest", config.AbTestHTTP)
	router.POST("/segment", segment.SegmentHTTP)


	// router.POST("/recommend", controllers.IndexHTTP)
	router.GET("/recommend/test", controllers.TestHTTP)
	router.POST("/recommend/matchList", match.MatchRecommendListHTTP)
	router.POST("/recommend/liveList", live.LiveRecommendListHTTP)
	router.POST("/recommend/themeList", theme.RecommendListHTTP)
	router.POST("/recommend/momentList", moment.RecommendListHTTP)
	router.POST("/recommend/coarse/momentList", moment.CoarseRecommendListHTTP)

	router.POST("/rank/:app/*type", controllers.IndexHTTP)
	router.NotFound = NotFound
}

func NotFound(c *routers.Context) {
	var notFound = make(map[string]interface{})
	notFound["errcode"] = "not_found"
	notFound["errdesc"] = "not_found"
	c.JSON(404, notFound)
}
