package shm

import (
	"fmt"

	"github.com/v3io/v3io-go/pkg/dataplane"

	"github.com/nuclio/logger"
)

type ResponseHandler func(*v3io.Response) error

type Dispatcher struct {
	logger              logger.Logger
	context             *Context
	responseHandlerByID []ResponseHandler
}

func NewDispatcher(parentLogger logger.Logger, context *Context, maxInflightRequests int) (*Dispatcher, error) {
	loggerInstance := parentLogger.GetChild("dispatcher").(logger.Logger)

	loggerInstance.DebugWith("Created", "maxInflightRequests", maxInflightRequests)

	return &Dispatcher{
		logger:              loggerInstance,
		context:             context,
		responseHandlerByID: make([]ResponseHandler, maxInflightRequests*2),
	}, nil
}

func (d *Dispatcher) RegisterResponseHandler(ID uint64, handler ResponseHandler) {
	if d.responseHandlerByID[ID] != nil {
		panic("ID already taken")
	}

	d.responseHandlerByID[ID] = handler
}

func (d *Dispatcher) Dispatch() error {
	for {
		response := d.context.GetNextResponse()

		// get response handler and nullify it
		responseHandler := d.responseHandlerByID[response.ID]
		d.responseHandlerByID[response.ID] = nil

		if responseHandler == nil {
			return fmt.Errorf("No response registered for response: %d", response.ID)
		}

		// handle pending request
		err := responseHandler(response)
		if err != nil {
			return err
		}
	}
}
