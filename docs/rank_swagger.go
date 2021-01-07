package docs

import (
	"rela_recommend/algo"
	"rela_recommend/utils/response"
)

// swagger:route POST /rank/{app}/{rType} tag idRecommendEndpoint
// 速配推荐接口，支持筛选.
// responses:
//   200: recommendResponse

// app.
//
// 功能
// swagger:parameters idRecommendEndpoint
type app struct {
	// app
	//
	// in: path
	// required: true
	// enum: moment,theme,user,live,match
	App string `json:"app"`
}

// rType.
//
// 类型
// swagger:parameters idRecommendEndpoint
type rType struct {
	// rType
	//
	// in: path
	// required: true
	// enum: nearby,reply,detail_reply
	RType string `json:"r_type"`
}

// 速配推荐接口返回.
// swagger:response recommendResponse
type recommendResponseWrapper struct {
	// in:body
	Body response.ResponseV3
	Data algo.RecommendResponse
}

// swagger:parameters idRecommendEndpoint
type recommendParamsWrapper struct {
	// 搜索请求参数.
	// in:body
	Body algo.RecommendRequest
}
