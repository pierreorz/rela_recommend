package api

import (
	"errors"
	"fmt"
	"rela_recommend/factory"
	"time"
)

const internalLiveHourRankListUrl = "/internal/live/anchorHourRank"

type AnchorHourRankRes struct {
	CreatTime     time.Time                `json:"creatTime"`
	NextCreatTime time.Time                `json:"nextCreatTime"`
	List          []AnchorHourRankItem     `json:"list"`
	Detail        AnchorHourRankUserDetail `json:"detail"`
}

type AnchorHourRankItem struct {
	Id         string  `json:"id"`
	IdInt      int64   `json:"idInt"`
	Score      string  `json:"score"`
	ScoreFloat float64 `json:"scoreFloat"`
	NickName   string  `json:"nickName"`
	Avatar     string  `json:"avatar"`
	LiveStatus int     `json:"liveStatus"`
	IsFollow   bool    `json:"isFollow"`
}

type AnchorHourRankUserDetail struct {
	UserId     string  `json:"userId"`
	Gem        float64 `json:"gem"`
	DeltaToPre float64 `json:"deltaToPre"`
}

type AnchorHourRankData struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	TTL     int               `json:"ttl"`
	Data    AnchorHourRankRes `json:"data"`
}

type AnchorHourRankInfo struct {
	Index int
	Rank  int
}

// 获取主播在上个小时列表中的排名, {userId: {index, rank}}，rankId从1开始
func CallLiveHourRankMap(userId int64) (map[int64]AnchorHourRankInfo, error) {
	params := fmt.Sprintf("userId=%d&dataVersion=last", userId)
	res := &AnchorHourRankData{}
	err := factory.LiveRpcClient.SendGETForm(internalLiveHourRankListUrl, params, res)
	if err == nil {
		if res != nil && res.Code == 0 {
			resMap := map[int64]AnchorHourRankInfo{}
			if res.Data.List != nil { // 获取每个id的排名，可以并列排名
				var lastScore float64 = 0.0
				var currRank int = 1
				for i, item := range res.Data.List {
					resMap[item.IdInt] = AnchorHourRankInfo{Index: i, Rank: currRank}
					if item.ScoreFloat != lastScore {
						currRank += 1
					}
					lastScore = item.ScoreFloat
				}
			}
			return resMap, nil
		} else {
			errMsg := fmt.Sprintf("CallLiveHourRankList error, %+v", res)
			return nil, errors.New(errMsg)
		}
	} else {
		return nil, err
	}
}
