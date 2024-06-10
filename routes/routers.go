package routes

import (
	"net/http"
	"rela_recommend/controllers"
	_ "rela_recommend/controllers/config"
	"rela_recommend/controllers/moment"
	"rela_recommend/utils/routers"
)

func RegisterRouters(router *routers.Routers) {
	//router.POST("/config/abtest", config.AbTestHTTP)

	router.POST("/recommend", controllers.IndexHTTP)
	//router.GET("/recommend/test", controllers.TestHTTP)
	router.POST("/recommend/momentList", moment.RecommendListHTTP)
	router.POST("/recommend/coarse/momentList", moment.CoarseRecommendListHTTP)

	// 动态路由
	router.POST("/rank/:app", controllers.IndexHTTP)
	router.POST("/rank/:app/*type", controllers.IndexHTTP)
	router.NotFound = NotFound

	// 代理静态文件，swagger.json 之类的
	router.ServeFiles("/relarecommend/*filepath", http.Dir("static"))
}

func NotFound(c *routers.Context) {
	var notFound = make(map[string]interface{})
	notFound["errcode"] = "not_found"
	notFound["errdesc"] = "not_found"
	c.JSON(404, notFound)
}
