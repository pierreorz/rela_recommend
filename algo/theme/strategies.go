package theme

import(
	"rela_recommend/algo"
)

var StrategyMap = map[string]algo.IStrategy{}
var SorterMap = map[string]algo.ISorter{
	"base": &algo.SorterBase{}}
var PagerMap = map[string]algo.IPager{
	"base": &algo.PagerBase{}}
var LoggerMap = map[string]algo.ILogger{
	"features": &algo.LoggerBase{},
	"performs": &algo.LoggerPerforms{}}
