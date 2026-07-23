package engine

import (
	"github.com/adimiuprix/spot-engine/order"
	"github.com/adimiuprix/spot-engine/queue"
)

type Engine struct {
	config Config

	orderQueue *queue.RingBuffer[order.Order]

	running bool
}

func New(config Config) *Engine {

	return &Engine{

		config: config,

		orderQueue: queue.New[order.Order](
			config.RingBufferSize,
		),
	}

}
