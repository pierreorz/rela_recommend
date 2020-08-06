package behavior

//速配页面
// MatchExposure  *Behavior `json:"match:exposure"`
// MatchClick  *Behavior `json:"match:click"`
// MatchSuperLike  *Behavior `json:"match:slike"`
// MatchLike  *Behavior `json:"match:like"`
// MatchRight *Behavior `json:"match:right"`
// MatchUnlike *Behavior `json:"match:unlike"`
// MatchLeft *Behavior `json:"match:left"`

// MatchInfoExposure  *Behavior `json:"match.info:exposure"`
// MatchWhoExposure  *Behavior `json:"match.who:exposure"`

// 获取列表曝光行为
func (self *UserBehavior) GetMatchListExposure() *Behavior {
	return self.Gets("match:exposure", "match.who:exposure", "match.info:exposure")
}

// 获取列表互动行为
func (self *UserBehavior) GetMatchListInteract() *Behavior {
	return self.Gets(
		"match:click", "match:right", "match:left", "match:slike", "match:unlike", "match:like",
		"match.info:click", "match.info:slike", "match.info:like", "match.info:unlike",
		"match.who:click")
}

// 获取列表互动率
func (self *UserBehavior) GetMatchListRate() float64 {
	exposure := self.GetMatchListExposure()
	interact := self.GetMatchListInteract()
	if exposure.Count > 0 {
		return interact.Count / exposure.Count
	}
	return 0.0
}
