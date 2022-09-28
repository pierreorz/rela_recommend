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
const externalPaiHomeFeedRecallUrl = "/api/rec/home_feed_recall"
const externalPaiLabelRecUrl = "/label_rec_v1"

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



func GetPredictResult(lat float32,lng float32,os string,userId int64,addr string,dataIds []int64,ua string) (map[int64]float64 ,string,string,error){
	result := make(map[int64]float64,0)
	recall_list :=make([]string,0)
	var expId= ""
	var requestErr error
	var requestId= ""
	for _,id :=range dataIds{
		recall_list=append(recall_list,strconv.FormatInt(id,10))
	}
	os_type,brand,model_type,net,language :=utils.UaAnalysis(ua)
	features := Features{
		Lat:        float64(lat),
		Lng:        float64(lng),
		Os:         os,
		Os_type:    os_type,
		Brand:      brand,
		Model_type: model_type,
		Net:        net,
		Language:   language,
		Ip:         addr,
	}
	log.Warnf("moment features%s",features)
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
				expId=utils.RequestErr
				requestId=utils.UniqueId()
			}
			if paiRes.Code == 200 {
				for _, element := range paiRes.Items {
					user_id, err := strconv.ParseInt(element.ItemId, 10, 64)
					if err == nil {
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
	return result,expId,requestId,requestErr
}


type paiHomeFeedRecallRes struct {
	Code int  	`json:"code"`
	Message  string  `json:"msg"`
	Request_id  string  `json:"request_id"`
	Experiment_id  string `json:"experiment_id"`
	Size    int  `json:"size"`
	Items  []PaiResRecallDataItem  `json:"items"`
}


type paiLabelRecRes struct {
	Code    int `json:"errcode"`
	Reason  string `json:"reason"`
	Ids     []int64  `json:"ids"`
}
type PaiResRecallDataItem struct {
	ItemId   string  `json:"item_id"`
	Score    float64  `json:"score"`
}

type paiHomeFeedRecallRequest struct {
	Uid   string   `json:"uid"`
	Size   int   `json:"size"`
	Scene_id    string   `json:"scene_id"`
	Request_id string  `json:"request_id"`
	Debug   bool  `json:"debug"`
}

type paiLabelRecRequest struct {
	Query  string  `json:"query"`
	ImageUrl string `json:"image_url"`
	VideoUrl string  `json:"video_url"`
}

func GetRecallResult(userId int64,size int) ([]int64 ,string,string,error){
	result := make([]int64,0)
	expId :=""
	requestId :=""
	var requestErr error
	params :=paiHomeFeedRecallRequest{
		Uid:         strconv.FormatInt(userId,10),
		Size:        size,
		Scene_id:    "home_feed_recall",
		Debug:       false,
	}
	if paramsData, err := json.Marshal(params); err == nil {
		paiRes := &paiHomeFeedRecallRes{}
		if requestErr = factory.PaiRpcRecallClient.PaiRecallSendPOSTJson(externalPaiHomeFeedRecallUrl, paramsData, paiRes); requestErr == nil {
			if paiRes.Code == 200 {
				for _, element := range paiRes.Items {
					item, err := strconv.ParseInt(element.ItemId, 10, 64)
					if err == nil {
						result=append(result,item)
					}
				}
			}
			if expArr :=strings.Split(paiRes.Experiment_id,"_");len(expArr)==2{
				expId = expArr[1]
			}
		}else{
			expId = utils.RecallOffTime
		}
	}
	log.Warnf("request err,%s",requestErr)
	return result,expId,requestId,requestErr
}


func GetLabelRecResult(query string,video string,image string) ([]int64,string,error){
	result :=&paiLabelRecRes{}
	reason :=""
	ids :=make([]int64,0)
	var requestErr error
	params := paiLabelRecRequest{
		Query:    query,
		ImageUrl: video,
		VideoUrl: image,
	}
	if paramsData, err := json.Marshal(params); err == nil {
		if requestErr = factory.PaiRpcLabelRecClient.PaiLabelRecSendPOSTJson(externalPaiLabelRecUrl, paramsData, result); requestErr == nil {
			for _, element := range result.Ids {
				ids = append(ids, element)
			}
			reason = result.Reason
		}
		return ids,reason,requestErr
	}
	return ids,reason,requestErr
}


