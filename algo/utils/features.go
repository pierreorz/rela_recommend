package utils


import (
	"fmt"
	"math"
	"bytes"
	"rela_recommend/utils"
)

//********************************* 特征列表
type Features struct {
	featuresMap map[int]float32
}

func (self *Features) checkInit() {
	if self.featuresMap == nil {
		self.featuresMap = make(map[int]float32)
	}
}

func (self *Features) ToString() string {
	self.checkInit()
	var buffer bytes.Buffer
	var i int = 0
	for key, val := range self.featuresMap {
		if i != 0 {
			buffer.WriteString(",")
		}
		str := fmt.Sprintf("%d:%f", key, val)
		buffer.WriteString(str)
		i++
	}
	return buffer.String()
}

func (self *Features) ToMap() map[int]float32 {
	self.checkInit()
	return self.featuresMap
}

func (self *Features) FromMap(featuresMap map[int]float32) {
	for key, val := range featuresMap {
		self.Add(key, val)
	}
}

func (self *Features) FromArray(features []float32) {
	for key, val := range features {
		self.Add(key, val)
	}
}

func (self *Features) Add(key int, val float32) bool {
	self.checkInit()
	if key >= 0 && math.Abs(float64(val)) >= 0.000001 {
		self.featuresMap[key] = val
		return true
	}
	return false
}

func (self *Features) AddCategory(start, length, minVal, val, def int) bool {
	new_val := val - minVal
	if new_val >= 0 && new_val < length {
		return self.Add(start + new_val, 1.0)
	} else {
		return self.Add(start + (def - minVal), 1.0)
	}
}

func (self *Features) AddCategories(start, length, minVal int, vals []int, def int) bool {
	for _, val := range vals {
		self.AddCategory(start, length, minVal, val, def)
	}
	return true
}

// 根据hash值映射为稀疏特征
func (self *Features) AddHash(start, length int, val interface{}) bool {
	bytesVal := utils.GetBytes(val)
	hashVal := utils.Md5Sum32(bytesVal)
	restVal := int(hashVal % uint32(length))
	return self.AddCategory(start, length, 0, restVal, 0)
}

func (self *Features) Get(key int) float32 {
	self.checkInit()
	if val, ok := self.featuresMap[key]; ok {
		return val
	}
	return 0.0
}
