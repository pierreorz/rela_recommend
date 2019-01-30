package live

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

// 以此按照：打分，最后登陆时间，
func (self LiveInfoListSorter) Less(i, j int) bool {
	if self.List[i].Score == self.List[j].Score {
		return self.List[i].UserId < self.List[j].UserId
	}
	return self.List[i].Score > self.List[j].Score
}
