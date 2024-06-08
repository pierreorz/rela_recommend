package moment

import (
	"rela_recommend/algo"
	"rela_recommend/algo/base"
	"rela_recommend/log"
	"rela_recommend/utils/routers"
	"rela_recommend/service"
	"rela_recommend/utils/response"
	"rela_recommend/utils/request"
)

func CoarseRecommendListHTTP(c *routers.Context) {
	var params = &algo.RecommendRequest{}
	if err := request.Bind(c, params); err != nil {
		log.Error(err.Error())
		c.JSON(response.FormatResponse(nil, service.WarpError(service.ErrInvaPara, "", "")))
		return
	}

	ctx := &base.ContextBase{}
	err := ctx.Do(algo.GetAppInfo("moment_coarse"), params)
	c.JSON(response.FormatResponse(ctx.GetResponse(), service.WarpError(err, "", "")))
}
