package moment

type Strategy interface {
	Do(ctx *AlgoContext, list []DataInfo)
}

type StrategyBase struct {}
