package search

import (
	"encoding/json"
	"fmt"
	"rela_recommend/factory"
	"strings"
)

const internalSearchNewTheme = "/search/new_theme"

type SearchThemeResDataItem struct {
	Id int64 `json:"id"`
}

type searchThemeRes struct {
	Data      []SearchThemeResDataItem `json:"result_data"`
	TotalSize int                       `json:"total_size"`
	ErrCode   string                    `json:"errcode"`
	ErrEsc    string                    `json:"erresc"`
}


type searchThemeRequest struct {
	UserID   int64   `json:"userId" form:"userId"`
	Limit int64   `json:"limit" form:"limit"`
	Filter   string  `json:"filter" form:"filter" `
}
func CallNewThemeuserId(userId int64, limit int64,momentTypes string,recommend bool)([]int64, error){
	idlist := make([]int64, 0)
	filters := []string{
		fmt.Sprintf("{moments_type:%s}", momentTypes),
	}
	if recommend{
		filters=append(filters,fmt.Sprintf("recommended:true"))
	}
	params := searchThemeRequest{
		UserID:   userId,
		Limit:    limit,
		Filter:   strings.Join(filters, "*"),
	}
	if paramsData, err := json.Marshal(params); err == nil {
		searchRes := &searchThemeRes{}
		if err = factory.AiSearchRpcClient.SendPOSTJson(internalSearchNewTheme, paramsData, searchRes); err == nil {
			for _, element := range searchRes.Data {
				idlist = append(idlist, element.Id)
			}
			return idlist, err
		} else {
			return idlist, err
		}
	} else {
		return idlist, err
	}
}





