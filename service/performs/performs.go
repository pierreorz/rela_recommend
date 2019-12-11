package performs

import (
	"time"
	"bytes"
	"fmt"
	"rela_recommend/log"
)

type Performs struct {
	Name 		string
	BeginTime 	*time.Time
	EndTime		*time.Time
	IsEnd		bool
	Interval 	float64
	ItemsName	[]string
	ItemsMap	map[string]*Performs
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

func(self *Performs) findNext() *Performs {
	if length := self.Length(); length > 0 {
		currName := self.ItemsName[length - 1]
		if val, ok := self.ItemsMap[currName]; ok && !val.IsEnd {
			return val
		}
	}
	return nil
}

func (self *Performs) Begin(name string) {
	self.check()
	// 递归查找当前活跃级别，如果有就执行递归开始，如果没有就创建
	if val := self.findNext(); val != nil {
		val.Begin(name)
	} else {
		self.addChild(name)
	}
}

func (self *Performs) End(name string) {
	self.check()
	// 递归查找当前活跃级别，如果下级有就执行递归结束，如果没有就创建
	if val := self.findNext(); val != nil {
		val.End(name)
	} else {
		if self.Name == name {
			self.IsEnd = true
		}
	}
}

func(self *Performs) EndAndBegin(endName string, beginName string) {
	self.End(endName)
	self.Begin(beginName)
}

func(self *Performs) toString(buffer *bytes.Buffer, pre string) {
	fullName := pre + "." + self.Name
	if pre == "" {
		if self.Name == "" {
			fullName = "root"
		} else {
			fullName = self.Name
		}
	}
	buffer.WriteString(fmt.Sprintf("%s:%.3f,", fullName, self.Interval))
	for _, name := range self.ItemsName {
		if val, ok := self.ItemsMap[name]; ok {
			val.toString(buffer, fullName)
		}
	}
}

func(self *Performs) ToString() string {
	self.check()
	var buffer = &bytes.Buffer{}
	self.toString(buffer, "")
	return buffer.String()
}
