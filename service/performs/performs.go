package performs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"rela_recommend/log"
	"rela_recommend/utils"
	"strings"
	"sync"
	"time"
)

type Performs struct {
	Name      string               `json:"name"`
	BeginTime *time.Time           `json:"-"`
	EndTime   *time.Time           `json:"-"`
	IsEnd     bool                 `json:"-"`
	Interval  float64              `json:"interval"`
	Result    interface{}          `json:"result,omitempty"`
	ItemsName []string             `json:"-"`
	ItemsMap  map[string]*Performs `json:"items,omitempty"`
}

func (self *Performs) check() {
	if self.ItemsMap == nil {
		self.ItemsMap = map[string]*Performs{}
	}
	now := time.Now()
	if self.BeginTime == nil {
		self.BeginTime = &now
	}
	self.EndTime = &now
	self.Interval = self.EndTime.Sub(*self.BeginTime).Seconds()
}

func (self *Performs) Length() int {
	return len(self.ItemsName)
}

func (self *Performs) addChild(name string) *Performs {
	now := time.Now()
	if val, ok := self.ItemsMap[name]; !ok {
		newItem := &Performs{Name: name, BeginTime: &now, ItemsMap: map[string]*Performs{}}
		self.ItemsName = append(self.ItemsName, name)
		self.ItemsMap[name] = newItem
		return newItem
	} else {
		log.Warnf("item is already begin:%s", name)
		return val
	}
}

// 当前节点最后一个且未关闭的对象
func (self *Performs) findNext() *Performs {
	if length := self.Length(); length > 0 {
		currName := self.ItemsName[length-1]
		if val, ok := self.ItemsMap[currName]; ok && !val.IsEnd {
			return val
		}
	}
	return nil
}

// 递归获取最后一个未关闭对象
func (self *Performs) FindCurrent() *Performs {
	if val := self.findNext(); val != nil {
		return val.FindCurrent()
	} else {
		return self
	}
}

func (self *Performs) Begin(name string) *Performs {
	self.check()
	// 递归查找当前活跃级别，如果有就执行递归开始，如果没有就创建
	if val := self.findNext(); val != nil {
		return val.Begin(name)
	} else {
		return self.addChild(name)
	}
}

func (self *Performs) End(name string) *Performs {
	return self.EndWithResult(name, nil)
}

func (self *Performs) EndWithResult(name string, result interface{}) *Performs {
	self.check()
	// 递归查找当前活跃级别，如果下级有就执行递归结束，如果没有就创建
	if val := self.findNext(); val != nil {
		return val.EndWithResult(name, result)
	} else {
		if self.Name == name {
			self.IsEnd = true
			self.Result = result
		}
		return self
	}
}

func (self *Performs) EndAndBegin(endName string, beginName string) *Performs {
	self.End(endName)
	return self.Begin(beginName)
}

func (self *Performs) Run(name string, runFunc func() interface{}) *Performs {
	self.Begin(name)
	result := runFunc()
	return self.EndWithResult(name, result)
}

func (self *Performs) RunsGo(runMap map[string]func() interface{}) *Performs {
	if len(runMap) == 0 {
		return self
	}

	runMapKeys := []string{"go"}
	for name := range runMap {
		runMapKeys = append(runMapKeys, name)
	}
	groupName := strings.Join(runMapKeys, "_")

	var group sync.WaitGroup
	groupPf := self.FindCurrent().addChild(groupName)
	for name, runFunc := range runMap {
		group.Add(1)
		childPf := groupPf.addChild(name)
		go func(name string, runFunc func() interface{}) {
			result := runFunc()
			childPf.EndWithResult(name, result)

			group.Done()
		}(name, runFunc)
	}
	group.Wait()
	return groupPf.End(groupName)
}

func (self *Performs) toString(buffer *bytes.Buffer, pre string) {
	fullName := pre + "." + self.Name
	if pre == "" {
		if self.Name == "" {
			fullName = "root"
		} else {
			fullName = self.Name
		}
	}
	var slog string
	switch result := self.Result.(type) {
	case nil:
		slog = fmt.Sprintf("%s:%.3f,", fullName, self.Interval)
	case error:
		errMsg := strings.Join(utils.Splits(result.Error(), " ,:"), "_")
		slog = fmt.Sprintf("%s:%.3f::%v,", fullName, self.Interval, errMsg)
	default:
		slog = fmt.Sprintf("%s:%.3f:%v,", fullName, self.Interval, self.Result)
	}

	buffer.WriteString(slog)
	for _, name := range self.ItemsName {
		if val, ok := self.ItemsMap[name]; ok {
			val.toString(buffer, fullName)
		}
	}
}

func (self *Performs) ToString() string {
	self.check()
	var buffer = &bytes.Buffer{}
	self.toString(buffer, "")
	return buffer.String()
}

func (self *Performs) ToJson() string {
	self.check()
	jss, _ := json.Marshal(self)
	return string(jss)
}
