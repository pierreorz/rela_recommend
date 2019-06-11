package performs

import (
	"time"
	"bytes"
	"fmt"
	"rela_recommend/log"
)

type PerformsItem struct {
	BeginTime	*time.Time
	EndTime		*time.Time
	Interval 	float64
	Count		int
}

type Performs struct {
	BeginTime 	*time.Time
	EndTime		*time.Time
	Interval 	float64
	ItemsName	[]string
	ItemsMap	map[string]*PerformsItem
}

func (self *Performs) Begin(name string) {
	now := time.Now()
	if self.BeginTime == nil {
		self.BeginTime = &now
	}
	if _, ok := self.ItemsMap[name]; !ok {
		self.ItemsName = append(self.ItemsName, name)
		self.ItemsMap[name] = &PerformsItem{BeginTime: &now, Count: 1}
	} else {
		log.Warnf("item is already begin:%s", name)
	}
}

func (self *Performs) End(name string) {
	if val, ok := self.ItemsMap[name]; ok {
		now := time.Now()
		self.EndTime = &now
		self.Interval = self.EndTime.Sub(*self.BeginTime).Seconds()

		val.EndTime = &now
		val.Interval = val.EndTime.Sub(*val.BeginTime).Seconds()
	} else {
		log.Warnf("item is not begin:%s", name)
	}
}

func (self *Performs) Incr(name string) {
	if val, ok := self.ItemsMap[name]; ok {
		val.Count++
	} else {
		self.Begin(name)
	}
}

func(self *Performs) EndAndBegin(endName string, beginName string) {
	self.End(endName)
	self.Begin(beginName)
}

func(self *Performs) ToString() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("total:%.3f", self.Interval))
	for _, name := range self.ItemsName {
		if val, ok := self.ItemsMap[name]; ok {
			str := fmt.Sprintf(",%s:%.3f", name, val.Interval)
			buffer.WriteString(str)
		}
	}
	return buffer.String()
}
