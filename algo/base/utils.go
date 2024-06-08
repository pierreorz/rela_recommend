package base

import (
	"errors"
	"fmt"
	"rela_recommend/algo"
	"rela_recommend/log"
	"rela_recommend/utils"
	"rela_recommend/utils/request"
	"rela_recommend/utils/routers"
	"runtime/debug"
	"strings"
)

// 通过处理请求参数灵活处理所有算法
func DoWithRoutersContext(c *routers.Context, appName, typeName string) (*algo.RecommendResponse, error) {
	var ctx = &ContextBase{}
	var params = &algo.RecommendRequest{}
	var err = request.Bind(c, params)
	if err == nil {
		params.App = utils.CoalesceString(appName, c.Params.ByName("app"), params.App)

		// url中的*type会出现前后/
		appType := utils.CoalesceString(strings.Split(c.Params.ByName("type"), "/")...)
		params.Type = utils.CoalesceString(typeName, appType, params.Type)

		name := params.App
		if len(params.Type) > 0 {
			name = fmt.Sprintf("%s.%s", name, params.Type)
		}
		var app = algo.GetAppInfo(name)
		if app != nil {
			err = ctx.Do(app, params)
		} else {
			err = errors.New("invalid app: " + name)
		}
	}
	if err != nil {
		log.Errorf("utils.panic---path:%s, err:%+v, statck:%s---", c.Request.URL.Path,err, string(debug.Stack()))
		log.Error(err.Error())
	}
	return ctx.GetResponse(), err
}
