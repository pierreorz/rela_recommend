package behavior


func (self *UserBehavior) GetLiveExposure() *Behavior {
	return self.Gets("live:exposure")
}

// 获取列表互动行为
func (self *UserBehavior) GetLiveClick() *Behavior {
	return self.Gets("live:click")
}