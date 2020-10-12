package algo

import "encoding/json"

//********************************* 服务端日志
type RecommendLog struct {
	Module          string
	RankId          string
	Index           int64
	UserId          int64
	DataId          int64
	Algo            string
	AlgoScore       float32
	Score           float32
	RecommendScores string
	Features        string
	AbMap           string
}

// 记录request日志
type RecommendRequestLog struct {
	Module  string            `json:"module"`
	Type    string            `json:"type"` // 是推荐/热门/
	Limit   int64             `json:"limit"`
	Offset  int64             `json:"offset"`
	Ua      string            `json:"ua"`
	Os      string            `json:"os"`
	Version int               `json:"version"`
	Lat     float32           `json:"lat"`
	Lng     float32           `json:"lng"`
	UserId  int64             `json:"user_id"`
	DataIds []int64           `json:"data_ids"`
	AbMap   map[string]string `json:"ab_map"`
	Params  map[string]string `json:"params"`
	// 返回
	RankId   string `json:"rank_id"`
	Returns  int    `json:"returns"`
	performs string `json:"performs"`
}

func (self *RecommendRequestLog) ToJson() string {
	data, _ := json.Marshal(self)
	return string(data)
}
