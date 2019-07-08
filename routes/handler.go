package routes

import (
	"rela_recommend/utils/routers"
	logger "rela_recommend/log"
	"log"
	"rela_recommend/controllers"
	"net/http"
	"time"
	"context"
	"fmt"
)

type ApiServer struct {
	http.Server
}

func RegisterHandler(router *routers.Routers) {
	router.PanicHandler = PanicHandler
	router.Use(Logger())
}

func PanicHandler(c *routers.Context, i interface{}) {
	controllers.PanicHandler(c, i)
}

func Logger() routers.Handle {
	return func(c *routers.Context) {
		logger.Infof("%s %s", c.Request.Method, c.Request.URL.Path)
	}
}

func NewApiServer(port int, handler http.Handler) *ApiServer {
	var server ApiServer
	server.Addr = fmt.Sprintf(":%d", port)
	server.Handler = handler
	return &server
}

func (this *ApiServer) Run() {
	if err := this.ListenAndServe(); err != nil {
		log.Panicln(err)
	}
}

func (this *ApiServer) Close() {
	// shut down gracefully, but wait no longer than 10 seconds before halting
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	this.Shutdown(ctx)
}