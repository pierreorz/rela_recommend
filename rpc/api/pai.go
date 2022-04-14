package api

import (
	"encoding/json"
	"rela_recommend/algo"
	"rela_recommend/factory"
	"strconv"
)

const externalPaiHomeFeedUrl = "/api/rec/home_feed"

type paiHomeFeedRes struct {
	Code int  	`json:"code"`
	Message  string  `json:"message"`
	Request_id  string  `json:"request_id"`
	Experiment_id  string `json:"experiment_id"`
	Size    int  `json:"size"`
	Items  []PaiResDataItem  `json:"items"`
}

type PaiResDataItem struct {
	ItemId   string  `json:"item_id"`
	RetrieveId string  `json:"retrieve_id"`
	Score    float64  `json:"score"`
}

type paiHomeFeedRequest struct {
	Uid   string   `json:"uid"`
	Size   int   `json:"size"`
	Scene_id    string   `json:"scene_id"`
	Request_id string  `json:"request_id"`
	Recall_list      string  `json:"recall_list" `
	Features      Features `json:"features"`
	Debug   bool  `json:"debug"`
}

type Features struct {
	Lat   float64   `json:"lat"`
	Lng   float64    `json:"lng"`
	Os    string    `json:"os"`
	Os_type  string  `json:"os_type"`
	Brand  string    `json:"brand"`
	Model_type  string `json:"model_type"`
	Net    string    `json:"net"`
	Language  string  `json:"language"`
	Ip      string    `json:"ip"`
}



func GetPredictResult(recommendRequest *algo.RecommendRequest,dataIds []int64) (map[int64]float64 ,error){
	result := make(map[int64]float64,0)
	features := Features{
		Lat:        float64(recommendRequest.Lat),
		Lng:        float64(recommendRequest.Lng),
		Os:         "android",
		Os_type:    "android 11",
		Brand:      "xiaomi",
		Model_type: "m2012k11ac",
		Net:        "wifi",
		Language:   "zh_cn_#hans",
		Ip:         recommendRequest.Addr,
	}
	params :=paiHomeFeedRequest{
		Uid:         strconv.FormatInt(recommendRequest.UserId,10),
		Size:        0,
		Scene_id:    "home_feed",
		Recall_list: recall_list,
		Features:    features,
		Debug:       false,
	}
	if paramsData, err := json.Marshal(params); err == nil {
		paiRes := &paiHomeFeedRes{}
		if err = factory.PaiRpcClient.SendPOSTJson(externalPaiHomeFeedUrl, paramsData, paiRes); err == nil {
			for _, element := range paiRes.Items {
				user_id, err := strconv.ParseInt(element.ItemId, 10, 64)
				if err !=nil{
						result[user_id]=element.Score
					}

			}
			return result, err
		} else {
			return result, err
		}
	} else {
		return result, err
	}
}