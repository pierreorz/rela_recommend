package algo

import (
	"rela_recommend/log"
)

var appMap = map[string]*AppInfo{}

func SetAppInfo(appName string, app *AppInfo) {
	appMap[appName] = app
}

func GetAppInfo(appName string) *AppInfo {
	if val, ok := appMap[appName]; ok {
		return val
	} else {
		log.Errorf("con't get app %s\n", appName)
	}
	return nil
}
