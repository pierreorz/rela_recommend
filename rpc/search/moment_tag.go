package search

const internalSearchMomentTagListUrl = "/search/friend_topic"

// 获取搜索日志标签列表
func CallSearchMomentTagIdList(userId int64, lat, lng float32, offset, limit int64, query string) ([]int64, error) {
	return CallSearchIdList(internalSearchMomentTagListUrl, userId, lat, lng, offset, limit, []string{}, query)
}
