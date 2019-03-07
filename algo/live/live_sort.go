package live

import (
	"sort"
)

// ************************************** 用户排序
type LiveInfoListSorter struct {
	List []LiveInfo
	Context *LiveAlgoContext
} 

func (self LiveInfoListSorter) Swap(i, j int) {
	self.List[i], self.List[j] = self.List[j], self.List[i]
}
func (self LiveInfoListSorter) Len() int {
	return len(self.List)
}
// 以此按照：打分，最后登陆时间
func (self LiveInfoListSorter) Less(i, j int) bool {
	ranki, rankj := self.List[i].RankInfo, self.List[j].RankInfo

	if ranki.IsTop != rankj.IsTop {
		return ranki.IsTop > rankj.IsTop		// IsTop ： 倒序， 是否置顶
	} else {
		if ranki.Level != rankj.Level {
			return ranki.Level > rankj.Level		// Level : 倒序， 推荐星数
		} else {
			if ranki.Score != rankj.Score {
				return ranki.Score > rankj.Score		// Score : 倒序， 推荐分数
			} else {
				return self.List[i].UserId < self.List[j].UserId	// UserId : 正序
			}
		}
	}
}
func (self *LiveInfoListSorter) Sort() {
	sort.Sort(self)
}
func (self *LiveInfoListSorter) DoStrategies() *LiveInfoListSorter {
	strategies := []LiveStrategy{
		&LiveTopStrategy{},
		&LiveLevelStrategy{},
		&LiveOldStrategy{},
	}
	for i, _ := range strategies {
		strategies[i].Do(self.Context, self.List)
	}
	return self
}
