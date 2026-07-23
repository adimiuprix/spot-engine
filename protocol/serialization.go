package protocol

import (
	"encoding/json"
	"fmt"
)

// Request is a union type for all command requests
type Request struct {
	Type CommandType
	Data interface{} // One of the typed request structs
}

// MarshalRequest serializes a typed request to JSON bytes
func MarshalRequest(req interface{}) ([]byte, error) {
	return json.Marshal(req)
}

// UnmarshalRequest deserializes JSON bytes to a typed request
func UnmarshalRequest(data []byte) (interface{}, error) {
	// First, parse just the type field
	var base BaseCommand
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, fmt.Errorf("failed to unmarshal base command: %w", err)
	}

	// Then unmarshal into the appropriate typed struct
	switch base.Type {
	case CmdPlaceOrder:
		var req PlaceOrderRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal PlaceOrderRequest: %w", err)
		}
		return &req, nil

	case CmdCancelOrder:
		var req CancelOrderRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal CancelOrderRequest: %w", err)
		}
		return &req, nil

	case CmdAmendOrder:
		var req AmendOrderRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal AmendOrderRequest: %w", err)
		}
		return &req, nil

	case CmdCreateMarket:
		var req CreateMarketRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal CreateMarketRequest: %w", err)
		}
		return &req, nil

	case CmdSuspendMarket:
		var req SuspendMarketRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal SuspendMarketRequest: %w", err)
		}
		return &req, nil

	case CmdResumeMarket:
		var req ResumeMarketRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ResumeMarketRequest: %w", err)
		}
		return &req, nil

	case CmdUpdateConfig:
		var req UpdateConfigRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal UpdateConfigRequest: %w", err)
		}
		return &req, nil

	case CmdUserEvent:
		var req UserEventRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("failed to unmarshal UserEventRequest: %w", err)
		}
		return &req, nil

	default:
		return nil, fmt.Errorf("unknown command type: %d", base.Type)
	}
}
