/*
Copyright 2018 The v3io Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v3io

import (
	"github.com/v3io/v3io-go/pkg/dataplane/shm/job"
	"github.com/valyala/fasthttp"
)

type RequestWriter func(interface{}, *job.JobBlock)

type Request struct {
	ID uint64

	// holds the input (e.g. ListBucketInput, GetItemInput)
	Input interface{}

	// a user supplied cookie
	Cookie interface{}

	// the channel to which the response must be posted
	ResponseChan chan *Response

	// pointer to container
	RequestResponse *RequestResponse

	// Request time
	SendTimeNanoseconds int64

	// Shared memory specific
	JobType          job.Type
	PayloadSizeWords uint64
	JobBlock         *job.JobBlock

	// for requests that need to write the payload while they are popualating the input (zero copy)
	Writer RequestWriter
}

type Response struct {

	// hold a decoded output, if any
	Output interface{}

	// Equal to the ID of request
	ID uint64

	// holds the error for async responses
	Error error

	// a user supplied cookie
	Cookie interface{}

	// pointer to container
	RequestResponse *RequestResponse

	// HTTP
	HTTPResponse *fasthttp.Response

	// Shared memory specific
	JobType  job.Type
	JobBlock *job.JobBlock
	Releaser func(*Response)
	Err      error
}

func (r *Response) Release() {
	if r.HTTPResponse != nil {
		fasthttp.ReleaseResponse(r.HTTPResponse)
	}

	if r.Releaser != nil {
		r.Releaser(r)
	}
}

func (r *Response) Body() []byte {
	return r.HTTPResponse.Body()
}

func (r *Response) Request() *Request {
	return &r.RequestResponse.Request
}

// holds both a request and response
type RequestResponse struct {
	Request  Request
	Response Response
}
