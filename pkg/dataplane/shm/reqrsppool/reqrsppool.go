package reqrsppool

import (
	"github.com/v3io/v3io-go/pkg/dataplane"
)

type RequestResponsePool struct {
	requestResponses    []*v3io.RequestResponse
	currentIndex        int
	numRequestResponses int
}

func NewRequestResponsePool(numRequestResponses int) *RequestResponsePool {

	// create an empty requestResponse pool
	return &RequestResponsePool{
		requestResponses:    make([]*v3io.RequestResponse, numRequestResponses, numRequestResponses),
		currentIndex:        0,
		numRequestResponses: numRequestResponses,
	}
}

func (rp *RequestResponsePool) Get() *v3io.RequestResponse {
	if rp.currentIndex == 0 {
		panic("Failed to allocate requestResponse, pool empty")
	}

	// don't use defer here
	requestResponse := rp.requestResponses[rp.currentIndex-1]
	rp.currentIndex--

	return requestResponse
}

func (rp *RequestResponsePool) Put(requestResponse *v3io.RequestResponse) {
	if rp.currentIndex >= rp.numRequestResponses {
		panic("RequestResponse pool is full, can't receive another requestResponse")
	}

	rp.requestResponses[rp.currentIndex] = requestResponse
	rp.currentIndex++
}
