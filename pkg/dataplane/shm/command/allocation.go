package command

import "github.com/v3io/v3io-go/pkg/dataplane/shm/job"

type AllocationRequest struct {
	command
	JobType          job.Type
	PayloadSizeWords uint64
}

func (a *AllocationRequest) Decode(encoded uint64) {
	a.command.decode(encoded)

	// decode job type (10 bits)
	a.JobType = job.Type((a.data & 0x2FF00000000) >> 32)

	// decode allocation size (32 bits)
	a.PayloadSizeWords = a.data & 0xFFFFFFFF
}

func (a *AllocationRequest) Encode() uint64 {

	// encode job type (10 bits)
	a.data = a.PayloadSizeWords & 0xFFFFFFFF

	// encode job type
	a.data |= (uint64(a.JobType) << 32) & 0x2FF00000000

	// ask parent to encode
	return a.command.encode(TYPE_ALLOCATION_REQUEST)
}

type AllocationResponse struct {
	command
	JobHandle uint64
}

func (a *AllocationResponse) Decode(encoded uint64) {

	// call inherited
	a.command.decode(encoded)

	// decode offset
	a.JobHandle = a.data & 0x2FFFFFFFFFF
}

func (a *AllocationResponse) Encode() uint64 {

	// encode job type (10 bits)
	a.data = a.JobHandle & 0x2FFFFFFFFFF

	return a.command.encode(TYPE_ALLOCATION_RESPONSE)
}
