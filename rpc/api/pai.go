package api

import (
	"encoding/json"
	"rela_recommend/factory"
	"rela_recommend/log"
	"rela_recommend/utils"
	"strconv"
	"strings"
)

const externalPaiHomeFeedUrl = "/api/rec/home_feed"

type paiHomeFeedRes struct {
	Code int  	`json:"code"`
	Message  string  `json:"msg"`
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



func GetPredictResult(lat float32,lng float32,userId int64,addr string,dataIds []int64) (map[int64]float64 ,string,string,error){
	result := make(map[int64]float64,0)
	recall_list :=make([]string,0)
	var expId= ""
	var requestErr error
	var requestId= ""
	for _,id :=range dataIds{
		recall_list=append(recall_list,strconv.FormatInt(id,10))
	}
	log.Warnf("dataid length %s",len(dataIds))
	features := Features{
		Lat:        float64(lat),
		Lng:        float64(lng),
		Os:         "android",
		Os_type:    "android 11",
		Brand:      "xiaomi",
		Model_type: "m2012k11ac",
		Net:        "wifi",
		Language:   "zh_cn_#hans",
		Ip:         addr,
	}
	params :=paiHomeFeedRequest{
		Uid:         strconv.FormatInt(userId,10),
		Size:        len(dataIds),
		Scene_id:    "home_feed",
		Recall_list: strings.Join(recall_list,","),
		Features:    features,
		Debug:       false,
	}
	if paramsData, err := json.Marshal(params); err == nil {
		paiRes := &paiHomeFeedRes{}
		if requestErr = factory.PaiRpcClient.PaiSendPOSTJson(externalPaiHomeFeedUrl, paramsData, paiRes); requestErr == nil {
			expId=paiRes.Experiment_id
			requestId=paiRes.Request_id
			if expId==""{
				expId="ER2_L2#EG2#E5"
				requestId=utils.UniqueId()
			}
			if paiRes.Code == 200 {
				for _, element := range paiRes.Items {
					user_id, err := strconv.ParseInt(element.ItemId, 10, 64)
					if err != nil {
						result[user_id] = element.Score
					}
				}
			}else{
				 for _,id :=range dataIds{
				 	result[id] = -1
				 }
			}
		}
	}
	log.Warnf("result pai%s",result)
	return result,expId,requestId,requestErr
}