package segment

import (
	"time"
	"rela_recommend/routers"
	"rela_recommend/service"
	"rela_recommend/log"
	"rela_recommend/utils/response"
	"rela_recommend/utils/request"
	"rela_recommend/factory"
)

type SegmentParams struct {
	Sentence   	string     	`form:"sentence" json:"sentence"`
	Sentences   []string    `form:"sentences" json:"sentences"`
}

type SegmentResult struct {
	Words    	[]string  	`form:"words" json:"words"`
	WordsList	[][]string	`form:"wordsList" json:"wordsList"`
}

func SegmentHTTP(c *routers.Context) {
	var startTime = time.Now()
	var params SegmentParams
	if err := request.Bind(c, &params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}
	
	var dataLen int = len(params.Sentences)		// 语句数量
	var charLen int = 0							// 单词数量
	var res = &SegmentResult{}
	if dataLen > 0 {
		for _, sen := range params.Sentences {
			charLen += len(sen)
			var words = factory.Segmenter.Cut(sen)
			res.WordsList = append(res.WordsList, words)
		}
	} else {
		charLen += len(params.Sentence)
		dataLen = 1
		res.Words = factory.Segmenter.Cut(params.Sentence)
	}
	var endTime = time.Now()
	log.Infof("sentenceLen %d,charLen %d,total:%.3f\n",
			  dataLen, charLen, endTime.Sub(startTime).Seconds())
	c.JSON(response.FormatResponse(res, service.WarpError(nil, "", "")))
}
