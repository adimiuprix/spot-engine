package protocol

// CommandType represents the type of command being executed
type CommandType uint8

const (
	CmdUnknown       CommandType = 0
	CmdPlaceOrder    CommandType = 1
	CmdCancelOrder   CommandType = 2
	CmdAmendOrder    CommandType = 3
	CmdCreateMarket  CommandType = 11
	CmdSuspendMarket CommandType = 12
	CmdResumeMarket  CommandType = 13
	CmdUpdateConfig  CommandType = 14
	CmdUserEvent     CommandType = 21
)

func (c CommandType) String() string {
	switch c {
	case CmdPlaceOrder:
		return "PlaceOrder"
	case CmdCancelOrder:
		return "CancelOrder"
	case CmdAmendOrder:
		return "AmendOrder"
	case CmdCreateMarket:
		return "CreateMarket"
	case CmdSuspendMarket:
		return "SuspendMarket"
	case CmdResumeMarket:
		return "ResumeMarket"
	case CmdUpdateConfig:
		return "UpdateConfig"
	case CmdUserEvent:
		return "UserEvent"
	default:
		return "Unknown"
	}
}
