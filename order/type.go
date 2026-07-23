package order

type Type uint8

const (
	Limit Type = iota
	Market
)
