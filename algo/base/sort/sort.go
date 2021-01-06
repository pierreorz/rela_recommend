package sort

import (
	"rela_recommend/algo"
	"rela_recommend/log"
	"sort"
)

type SorterBase struct {
	Context algo.IContext
}

func (self SorterBase) Swap(i, j int) {
	list := self.Context.GetDataList()
	list[i], list[j] = list[j], list[i]
}
func (self SorterBase) Len() int {
	return self.Context.GetDataLength()
}

// 以此按照：打分，最后登陆时间
func (self SorterBase) Less(i, j int) bool {
	listi, listj := self.Context.GetDataByIndex(i), self.Context.GetDataByIndex(j)
	ranki, rankj := listi.GetRankInfo(), listj.GetRankInfo()

	if ranki.IsTop != rankj.IsTop {
		return ranki.IsTop > rankj.IsTop // IsTop ： 倒序， 是否置顶
	} else {
		if ranki.PagedIndex != rankj.PagedIndex { // PagedIndex: 已经被分页展示过的index, 升序排列
			return ranki.PagedIndex < rankj.PagedIndex
		} else {
			if ranki.Level != rankj.Level {
				return ranki.Level > rankj.Level // Level : 倒序， 推荐星数
			} else {
				if ranki.Score != rankj.Score {
					return ranki.Score > rankj.Score // Score : 倒序， 推荐分数
				} else {
					return listi.GetDataId() < listj.GetDataId() // UserId : 正序
				}
			}
		}
	}
}

func (self *SorterBase) Do(ctx algo.IContext) error {
	sorter := &SorterBase{Context: ctx}
	sort.Sort(sorter)
	return nil
}

//************************************************* 返回，不做排序
type SorterOrigin struct {
	Context algo.IContext
}

func (self *SorterOrigin) Do(ctx algo.IContext) error {
	return nil
}

/************************************************* 按照指定的index排序: 将指定位置的内容插到该位置，之后的向后移动
如：1，2，3，4，5
	希望将第4->2，则1,2,5,4,3
	希望将第1->4，则1,3,4,5,2
*/
type SorterHope struct {
	Context algo.IContext
}

func (self *SorterHope) Do(ctx algo.IContext) error {
	sorter := &SorterHope{Context: ctx}
	sorter.sortByIndexWithHope()
	return nil
}

// 交换位置，先取出当前值，其他值依次前移或后移，然后插入相应位置
func (self *SorterHope) swapByIndex(arr []int, currIndex, hopeIndex int) {
	if currIndex < hopeIndex { // 向后移动
		currValue := arr[currIndex]
		for i := currIndex; i < hopeIndex; i++ {
			arr[i] = arr[i+1]
		}
		arr[hopeIndex] = currValue
	} else if currIndex > hopeIndex { // 向前移动
		currValue := arr[currIndex]
		for i := currIndex; i > hopeIndex; i-- {
			arr[i] = arr[i-1]
		}
		arr[hopeIndex] = currValue
	} else {

	}
}

func (self *SorterHope) sortByIndexWithHope() error {
	// 最终排序
	var listLen = self.Context.GetDataLength()
	var list = self.Context.GetDataList()
	var indexs = make([]int, listLen)
	var hopeList = [][]int{} // 期望index []{当前index, 期望index}
	for i, data := range list {
		indexs[i] = i
		if rank := data.GetRankInfo(); 0 < rank.HopeIndex && rank.HopeIndex < listLen {
			hopeList = append(hopeList, []int{i, rank.HopeIndex})
		}
	}
	if len(hopeList) > 0 {
		sort.SliceStable(hopeList, func(i, j int) bool { // 从小到大排序
			return hopeList[i][1] < hopeList[j][1]
		})
		log.Debugf("hope list: %+v \n", hopeList)
		for _, hope := range hopeList {
			currI, hopeI := hope[0], hope[1]
			self.swapByIndex(indexs, currI, hopeI)
		}

		// 最终调整数据
		var itemMap = map[int]algo.IDataInfo{}
		for i, _ := range list {
			itemMap[i] = list[i]
		}
		for i, ni := range indexs {
			list[i] = itemMap[ni]
		}
	}
	return nil
}

/************************************************* 在每个分组中，按照推荐理由进行排序间隔打散
复杂度：快排 + 分区 + 混排： n*log2(n) + n + n * group
每个分组指：top, paged, level 等绝对隔离组
推荐理由指：recommendItem中的reason
排序间隔打散： 保持大致位置不变。同样的推荐理由尽可能间隔interval个内容，中间填充的内容为 无相关推荐理由 和 小于i + floatRange 中的内容
如： a,a,b,b,c 在间隔为2的时候返回 a, b, a, b, c
*/
type SorterWithInterval struct {
	*SorterBase
}

func (self *SorterWithInterval) Do(ctx algo.IContext) error {
	sorter := &SorterWithInterval{&SorterBase{Context: ctx}}
	sort.Sort(sorter)
	abtest := ctx.GetAbTest()
	var interval = abtest.GetInt("sort_with_interval_interval", 3)                   // 优先间隔，如果没有会逐步缩小
	var floatRange = abtest.GetInt("sort_with_interval_float", interval)             // 允许将当前向下的多少范围来填充
	var recommends = abtest.GetStrings("sort_with_interval_recommends", "RECOMMEND") // 根据哪些推荐理由进行打散
	if len(recommends) > 0 && interval > 1 {
		allIndexs := make([]int, sorter.Len())
		partitions, _ := sorter.partitions() // 分组隔离排序
		partitionStartIndex := 0
		for _, partition := range partitions {
			indexs, _ := sorter.sortWithInterval(partition, interval, floatRange, recommends) // 每组内进行间隔打散
			for i, index := range indexs {
				allIndexs[i+partitionStartIndex] = index + partitionStartIndex
			}
			partitionStartIndex += len(partition)
		}
		sorter.sortByIndex(allIndexs) // 按照最终index排序
	}

	// 期望位置处理，将HopeIndex指定大于1的内容进行移动
	if abtest.GetBool("sort_with_interval_hope_switch", true) { // 打开开关
		hopeSorter := &SorterHope{Context: ctx}
		hopeSorter.sortByIndexWithHope()
	}
	return nil
}

// 按照istop, paged, level分区隔离
func (self *SorterWithInterval) partitions() ([][]algo.IDataInfo, error) {
	res := [][]algo.IDataInfo{}
	list := self.Context.GetDataList()
	var currIsTop, currPagedIndex, currLevel, currList = 0, 0, 0, []algo.IDataInfo{}
	for i, item := range list {
		rank := item.GetRankInfo()
		if rank.IsTop == currIsTop && rank.PagedIndex == currPagedIndex && rank.Level == currLevel {
			currList = append(currList, list[i])
		} else {
			res = append(res, currList)
			currIsTop, currPagedIndex, currLevel, currList = rank.IsTop, rank.PagedIndex, rank.Level, []algo.IDataInfo{list[i]}
		}
	}
	if len(currList) > 0 {
		res = append(res, currList)
	}
	return res, nil
}

type groupItem struct {
	Name         string // 当前组名称
	Indexs       []int  // 当前组在原数组的索引,
	CurrentIndex int    // 当前该组计算到的索引, indexs 中的index
	LastIndex    int    // 已计算的新数组中当前组最后的index, 结果列表中的index
}

// 获取每个内容的排序group
func (self *SorterWithInterval) getDataInfoGroupName(data *algo.RankInfo, recommends []string) (string, error) {
	if len(data.Recommends) > 0 && len(recommends) > 0 {
		for _, recommend := range recommends {
			for _, rItem := range data.Recommends {
				if rItem.Reason == recommend {
					return rItem.Reason, nil
				}
			}
		}
	}
	return "", nil
}

// 对列表进行排序间隔打散
// list : 需要重排的列表
// interval: 两个推荐理由间的初始间隔，如果不满足会缩小
// floatRange: 从大于当前索引的偏移范围进行计算间隔填充
// recommends: 根据哪些推荐理由进行打散，优先级即顺序
func (self *SorterWithInterval) sortWithInterval(list []algo.IDataInfo, interval int, floatRange int, recommends []string) ([]int, error) {
	groups := append(recommends, "")
	var groupMap = map[string]*groupItem{}
	for _, group := range groups {
		groupMap[group] = &groupItem{
			Name:         group,
			Indexs:       []int{},
			CurrentIndex: 0,
			LastIndex:    -1}
	}
	// 分组，每组按原排序顺序排序，如找不到计入默认分组,并记录原index
	for i, item := range list {
		groupName, _ := self.getDataInfoGroupName(item.GetRankInfo(), recommends)
		currGroup, currGroupOk := groupMap[groupName]
		if !currGroupOk {
			currGroup, currGroupOk = groupMap[""]
		}
		currGroup.Indexs = append(currGroup.Indexs, i)
	}

	// 计算新的排序索引
	var newIndex = make([]int, len(list))
	for i := 0; i < len(list); i++ {
		for curInterval := interval; curInterval > 0; curInterval-- {
			var cGroup = groupMap[""] // 默认补充分组
			for _, group := range groups {
				if gItem, _ := groupMap[group]; gItem.CurrentIndex < len(gItem.Indexs) {
					intervalOk := gItem.LastIndex < 0 || i-gItem.LastIndex >= curInterval // 检测距离间隔
					if gItem.Indexs[gItem.CurrentIndex] <= i+floatRange && intervalOk {
						cGroup = gItem
						break
					}
				}
			}

			if cGroup.CurrentIndex < len(cGroup.Indexs) {
				newIndex[i] = cGroup.Indexs[cGroup.CurrentIndex] // 列表增加
				cGroup.CurrentIndex++                            // 当前计算index向后偏移
				cGroup.LastIndex = i                             // 记录每组最后位置
				// fmt.Printf("\t index %d -> %d ;interval %d; map %+v \n", i, newIndex[i], curInterval, groupMap)
				break
			}
		}
	}
	return newIndex, nil
}

// 按照给定的indexs进行重排序
func (self *SorterWithInterval) sortByIndex(indexs []int) error {
	// 最终排序
	list := self.Context.GetDataList()
	var itemMap = map[int]algo.IDataInfo{}
	for i, _ := range list {
		itemMap[i] = list[i]
	}
	for i, ni := range indexs {
		list[i] = itemMap[ni]
	}
	return nil
}
