package abtest

import (
	// "sync"
	"encoding/json"
	"errors"
	"rela_recommend/log"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

type appInfo struct {
	Name string `json:"name"`

	configPrefix string
	testPrefix   string
	whitePrefix  string

	configPlan *watch.Plan
	testPlan   *watch.Plan
	whitePlan  *watch.Plan
}

func watch_func(host string, app string, prefix string, handler func(string, string, api.KVPairs) error) *watch.Plan {
	var params = map[string]interface{}{
		"type":   "keyprefix",
		"prefix": prefix,
	}
	plan, err := watch.Parse(params)
	if err != nil {
		panic(err)
	}
	plan.Handler = func(index uint64, result interface{}) {
		var err error
		if entries, ok := result.(api.KVPairs); ok {
			err = handler(app, prefix, entries)
		} else {
			err = errors.New("type error")
		}
		if err != nil {
			log.Errorf("%s error: %s, %+v\n", prefix, err, result)
		}
		// 更新涉及的key
		if len(app) > 0 {
			// formulaKeys := GetFormulaKeys(defaultFactorMap[app], testingMap[app], whiteListMap[app])
			formulaKeys := GetFormulaKeys(getDefaultFactorMap(app), getTestingMap(app), getWhiteListMap(app))
			log.Infof("%s app:%s keys:%+v\n", prefix, app, formulaKeys)
			setFormulaListMap(app, formulaKeys)
		}
		// log.Infof("%s changed: %+v\n", prefix, result)
	}
	go func() {
		if err = plan.Run(host); err != nil {
			panic(err)
		}
	}()
	return plan
}

func updateConfig(app string, prefix string, kvs api.KVPairs) error {
	var configMap = map[string]Factor{}
	var keyPrefixLen = len(prefix)
	for _, kv := range kvs {
		configKey := kv.Key[keyPrefixLen:]
		configMap[configKey] = NewFactor(kv.Value)
	}
	// defaultFactorMap[app] = configMap
	setDefaultFactorMap(app, configMap)
	log.Infof("%s changed: %+v\n", prefix, getDefaultFactorMap(app))
	return nil
}

func updateTest(app string, prefix string, kvs api.KVPairs) error {
	var testList = []Testing{}
	for _, kv := range kvs {
		var test = Testing{}
		if err := json.Unmarshal(kv.Value, &test); err != nil {
			log.Errorf("%s test error: %s, %s\n", prefix, err, string(kv.Value))
		} else {
			testList = append(testList, test)
		}
	}
	// testingMap[app] = testList
	setTestingMap(app, testList)
	log.Infof("%s changed: %+v\n", prefix, getTestingMap(app))
	return nil
}

func updateWhite(app string, prefix string, kvs api.KVPairs) error {
	var whiteList = []WhiteName{}
	for _, kv := range kvs {
		var white = WhiteName{}
		if err := json.Unmarshal(kv.Value, &white); err != nil {
			log.Errorf("%s white error: %s, %s\n", prefix, err, string(kv.Value))
		} else {
			whiteList = append(whiteList, white)
		}
	}
	// whiteListMap[app] = whiteList
	setWhiteListMap(app, whiteList)
	log.Infof("%s changed: %+v\n", prefix, getWhiteListMap(app))
	return nil
}

func (self *appInfo) Watch(hosts string) *appInfo {
	self.configPrefix = "ai/abtest/config/" + self.Name + "/"
	self.testPrefix = "ai/abtest/test/" + self.Name + "/"
	self.whitePrefix = "ai/abtest/white/" + self.Name + "/"
	self.configPlan = watch_func(hosts, self.Name, self.configPrefix, updateConfig)
	self.testPlan = watch_func(hosts, self.Name, self.testPrefix, updateTest)
	self.whitePlan = watch_func(hosts, self.Name, self.whitePrefix, updateWhite)
	return self
}

var watchAppMap = map[string]*appInfo{}

func BeginWatching(hosts string) *watch.Plan {
	// hosts := "127.0.0.1:8500"
	return watch_func(hosts, "", "ai/abtest/app/", func(appName string, prefix string, kvs api.KVPairs) error {
		for _, kv := range kvs {
			var app = &appInfo{}
			if err := json.Unmarshal(kv.Value, app); err != nil {
				log.Errorf("%s app error: %s, %s\n", prefix, err, string(kv.Value))
			} else {
				if _, ok := watchAppMap[app.Name]; len(app.Name) > 0 && !ok {
					watchAppMap[app.Name] = app.Watch(hosts)
				}
			}
		}
		return nil
	})
}
