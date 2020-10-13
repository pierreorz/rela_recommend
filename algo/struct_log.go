package algo

import (
	"encoding/json"
	"time"
)

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
	Module  string            `json:"module"` // 对应abtest中的app
	Limit   int64             `json:"limit"`
	Offset  int64             `json:"offset"`
	Ua      string            `json:"ua,omitempty"`
	Os      string            `json:"os,omitempty"`
	Version int               `json:"version,omitempty"`
	Lat     float32           `json:"lat,omitempty"`
	Lng     float32           `json:"lng,omitempty"`
	UserId  int64             `json:"user_id,omitempty"`
	DataIds []int64           `json:"data_ids,omitempty"`
	AbMap   map[string]string `json:"ab_map,omitempty"`
	Params  map[string]string `json:"params,omitempty"`
	// 返回
	CreateTime time.Time `json:"create_time"`
	RankId     string    `json:"rank_id"`
	Returns    int       `json:"returns"`
	Performs   string    `json:"performs"`
	Error      string    `json:"error,omitempty"`
}

func (self *RecommendRequestLog) ToJson() string {
	data, _ := json.Marshal(self)
	return string(data)
}
