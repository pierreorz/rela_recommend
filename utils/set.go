package utils

// Set 结构
type SetInt64 struct {
	intMap map[int64]int
}

func (self *SetInt64) checkMap(setEmpty bool) {
	if self.intMap == nil || setEmpty {
		self.intMap = make(map[int64]int, 0)
	}
}

func (self *SetInt64) FromArray(vals []int64) {
	self.checkMap(true)
	for _, val := range vals {
		self.intMap[val] = 1
	}
}

func (self *SetInt64) Append(val int64) *SetInt64 {
	self.checkMap(false)
	self.intMap[val] = 1
	return self
}

func (self *SetInt64) AppendArray(vals []int64) *SetInt64 {
	self.checkMap(false)
	for _, val := range vals {
		self.intMap[val] = 1
	}
	return self
}

func (self *SetInt64) Contains(val int64) bool {
	_, ok := self.intMap[val]
	return ok
}

func (self *SetInt64) Remove(val int64) *SetInt64 {
	delete(self.intMap, val)
	return self
}

func (self *SetInt64) RemoveArray(vals []int64) *SetInt64 {
	for _, val := range vals {
		delete(self.intMap, val)
	}
	return self
}

func (self *SetInt64) ToList() []int64 {
	res := make([]int64, 0)
	for k, _ := range self.intMap {
		res = append(res, k)
	}
	return res
}

func (self *SetInt64) Len() int {
	return len(self.intMap)
}

func NewSetInt64FromArray(vals []int64) *SetInt64 {
	set := SetInt64{}
	set.FromArray(vals)
	return &set
}

func NewSetInt64FromArrays(vals ...[]int64) *SetInt64 {
	set := SetInt64{}
	for i, val := range vals {
		if i == 0 {
			set.FromArray(val)
		} else {
			set.AppendArray(val)
		}
	}
	return &set
}

// Set 结构
type SetString struct {
	intMap map[string]int
}

func (self *SetString) checkMap() {
	if self.intMap == nil {
		self.intMap = make(map[string]int, 0)
	}
}

func (self *SetString) Append(val string) *SetString {
	self.checkMap()
	self.intMap[val] = 1
	return self
}

func (self *SetString) AppendArray(vals []string) *SetString {
	self.checkMap()
	for _, val := range vals {
		self.intMap[val] = 1
	}
	return self
}

func (self *SetString) Contains(val string) bool {
	_, ok := self.intMap[val]
	return ok
}

func (self *SetString) Remove(val string) *SetString {
	delete(self.intMap, val)
	return self
}

func (self *SetString) RemoveArray(vals []string) *SetString {
	for _, val := range vals {
		delete(self.intMap, val)
	}
	return self
}

func (self *SetString) ToList() []string {
	res := make([]string, 0)
	for k, _ := range self.intMap {
		res = append(res, k)
	}
	return res
}

func (self *SetString) Len() int {
	return len(self.intMap)
}

func NewSetStringFromArray(vals []string) *SetString {
	set := SetString{}
	set.AppendArray(vals)
	return &set
}
