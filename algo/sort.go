package algo

import "sort"

/*
在每个分组中，按照推荐理由进行排序间隔打散
每个分组指：top, paged, level 等绝对隔离组
推荐理由指：recommendItem中的reason
排序间隔打散： 保持大致位置不变。同样的推荐理由尽可能间隔interval个内容，中间填充的内容为 无相关推荐理由 和 小于i + floatRange 中的内容
如： a,a,b,b,c 在间隔为2的时候返回 a, b, a, b, c
*/
type SorterWithInterval struct {
	SorterBase
}

func (self *SorterWithInterval) Do(ctx IContext) error {
	sorter := &SorterWithInterval{SorterBase{Context: ctx}}
	sort.Sort(sorter)
	var interval = 3                         // 优先间隔，如果没有会逐步缩小
	var floatRange = 3                       // 允许将当前向下的多少范围来填充
	var recommends = []string{"a", "b", "c"} // 根据哪些推荐理由进行打散
	if len(recommends) > 0 && interval > 1 {
		allIndexs := make([]int, self.Len())
		groups, _ := self.groups() // 分组隔离排序
		groupStartIndex := 0
		for _, group := range groups {
			indexs, _ := self.sortWithInterval(group, interval, floatRange, recommends) // 每组内进行间隔打散
			for i, index := range indexs {
				allIndexs[i+groupStartIndex] = index + groupStartIndex
			}
			groupStartIndex += len(group)
		}
		self.sortByIndex(allIndexs) // 按照最终index排序
	}
	// 间隔处理
	return nil
}

// 按照istop, paged, level分组隔离
func (self *SorterWithInterval) groups() ([][]IDataInfo, error) {
	res := [][]IDataInfo{}
	list := self.Context.GetDataList()
	var currIsTop, currPagedIndex, currLevel, currList = 0, 0, 0, []IDataInfo{}
	for i, item := range list {
		rank := item.GetRankInfo()
		if rank.IsTop == currIsTop && rank.PagedIndex == currPagedIndex && rank.Level == currLevel {
			currList = append(currList, list[i])
		} else {
			res = append(res, currList)
			currIsTop, currPagedIndex, currLevel, currList = rank.IsTop, rank.PagedIndex, rank.Level, []IDataInfo{list[i]}
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
func (self *SorterWithInterval) getDataInfoGroupName(data *RankInfo, recommends []string) (string, error) {
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
func (self *SorterWithInterval) sortWithInterval(list []IDataInfo, interval int, floatRange int, recommends []string) ([]int, error) {
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
	var itemMap = map[int]IDataInfo{}
	for i, _ := range list {
		itemMap[i] = list[i]
	}
	for i, ni := range indexs {
		list[i] = itemMap[ni]
	}
	return nil
}
