package behavior

// 获取列表曝光行为
func (self *UserBehavior) GetNearbyListExposure() *Behavior {
	return self.Gets("around.list:exposure")
}

// 获取列表互动行为
func (self *UserBehavior) GetNearbyListInteract() *Behavior {
	return self.Gets("around.list:click")
}

// 获取列表互动率
func (self *UserBehavior) GetNearbyListRate() float64 {
	exposure := self.GetNearbyListExposure()
	interact := self.GetNearbyListInteract()
	if exposure.Count > 0 {
		return interact.Count / exposure.Count
	}
	return 0.0
}
