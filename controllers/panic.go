package controllers

import (
	"errors"
	"runtime/debug"

	"rela_recommend/log"
	"rela_recommend/routers"
	"rela_recommend/service"
)

func PanicHandler(c *routers.Context, i interface{}) {
	log.Errorf("panic---path:%s, err:%+v, statck:%s---", c.Request.URL.Path, i, string(debug.Stack()))
	c.JSON(formatResponse(nil, service.WarpError(errors.New("painc"), "", "")))
}
