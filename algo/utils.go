package algo
import (
	"os"
	"fmt"
	"strings"
	"errors"
	"rela_recommend/log"
	"rela_recommend/utils/routers"
	"rela_recommend/utils/request"
	"rela_recommend/utils"
)

// 获取当前工作目录
func GetWorkDir(path string) string {
	work_dir, _ := os.Getwd()
	return work_dir + path
}

// 将算法列表初始化，并且生成到指定Map
func AlgoListInitToMap(algoList []IAlgo) map[string]IAlgo {
	algoMap := map[string]IAlgo{}
	for index, _ := range algoList {
		algoList[index].Init()
		algoMap[algoList[index].Name()] = algoList[index]
	}
	return algoMap
}

// 通过处理请求参数灵活处理所有算法
func DoWithRoutersContext(c *routers.Context, appName string) (*RecommendResponse, error) {
	var ctx = &ContextBase{}
	var params = &RecommendRequest{}
	var err = request.Bind(c, params)
	if err == nil {
		params.App = utils.CoalesceString(appName, c.Params.ByName("app"), params.App)

		// url中的*type会出现前后/
		appType := utils.CoalesceString(strings.Split(c.Params.ByName("type"), "/")...)
		params.Type = utils.CoalesceString(appType, params.Type)

		name := params.App
		if len(params.Type) > 0 {
			name = fmt.Sprintf("%s.%s", name, params.Type)
		}
		var app = GetAppInfo(name)
		if app != nil {
			err = ctx.Do(app, params)
		} else {
			err = errors.New("invalid app: " + name)
		}
	}
	if err != nil {
		log.Error(err.Error())
	}
	return ctx.GetResponse(), err
}
