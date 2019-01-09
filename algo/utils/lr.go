package utils

import (
	"os"
	"fmt"
	"math"
	"io/ioutil"
	"rela_recommend/log"
	"encoding/json"
	"compress/gzip"
)

type LR struct {
	FeatureCount int 		`json:"n_features"`
	ClassCount int 			`json:"n_classes"`
	CoefMap map[int]float32 `json:"coef"`
	Intercept float32 		`json:"intercept"`
}

func (self *LR) Init(path string)  {
	fr, oerr := os.Open(path)
	defer fr.Close()
	if oerr != nil {
		fmt.Println("lr:open file err", path, oerr.Error())
	}
	gzf, gerr := gzip.NewReader(fr)
	defer gzf.Close()
	if gerr != nil {
		fmt.Println("lr:read gzip file err", gerr.Error())
	}
	data, rerr := ioutil.ReadAll(gzf)
	if rerr != nil {
		fmt.Println("lr:read all err", rerr.Error())
	}
	jerr := json.Unmarshal(data, self)
	if jerr != nil {
		fmt.Println("lr:load json err")
	}
	log.Infof("lr:init ok %s", path)
}

func (self *LR) PredictSingle(features []float32) float32 {
	var score float32 = self.Intercept
	for i, feature := range features {
		if feature > 0 {
			score += self.CoefMap[i] * feature
		}
	}
	return 1.0 / (1.0 + float32(math.Exp(-float64(score))))
}
