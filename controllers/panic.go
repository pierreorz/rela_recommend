package controllers

import (
	"errors"
	"runtime/debug"

	"rela_recommend/log"
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/utils/response"
)

func PanicHandler(c *routers.Context, i interface{}) {
	log.Errorf("panic---path:%s, err:%+v, statck:%s---", c.Request.URL.Path, i, string(debug.Stack()))
	c.JSON(response.FormatResponse(nil, service.WarpError(errors.New("painc"), "", "")))
}
