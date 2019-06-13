package algo

import (
	"rela_recommend/log"
)

var appMap = map[string]*AppInfo{}

func AddAppInfo(app *AppInfo) *AppInfo {
	if _, ok := appMap[app.Name]; ok {
		log.Errorf("app is exists: %s\n", app.Name)
	} else {
		appMap[app.Name] = app
	}
	return app
}

func GetAppInfo(appName string) *AppInfo {
	if val, ok := appMap[appName]; ok {
		return val
	} else {
		log.Errorf("con't get app: %s\n", appName)
	}
	return nil
}
