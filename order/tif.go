package order

type TimeInForce uint8

const (
	GTC TimeInForce = iota
	IOC
	FOK
)
