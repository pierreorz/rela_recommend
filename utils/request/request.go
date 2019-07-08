package request

import (
	"rela_recommend/factory"
	"rela_recommend/utils/routers"
)

func Bind(c *routers.Context, i interface{}) error {
	if factory.IsProduction {
		return c.BindAndSingnature(i)
	}
	return c.Bind(i)
}
