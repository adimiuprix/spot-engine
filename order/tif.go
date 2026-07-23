package order

type TimeInForce uint8

const (
	GTC      TimeInForce = iota // Good-Til-Cancel: Match what you can, rest stays in book
	IOC                         // Immediate-Or-Cancel: Match immediately, cancel remaining
	FOK                         // Fill-Or-Kill: Full fill or reject entire order
	PostOnly                    // Post-Only: Must be maker, reject if would match immediately
)

func (t TimeInForce) String() string {
	switch t {
	case GTC:
		return "GTC"
	case IOC:
		return "IOC"
	case FOK:
		return "FOK"
	case PostOnly:
		return "PostOnly"
	default:
		return "Unknown"
	}
}
