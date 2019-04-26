package command

type Type int

const (
	TYPE_ALLOCATION_REQUEST    = 0
	TYPE_ALLOCATION_RESPONSE   = 1
	TYPE_JOB_REQUEST           = 2
	TYPE_JOB_RESPONSE          = 3
	TYPE_DEALLOCATION_REQUEST  = 4
	TYPE_DEALLOCATION_RESPONSE = 5
	TYPE_JOB_BATCH_START       = 6
	TYPE_JOB_BATCH_SUBMIT      = 7
	TYPE_ECHO_REQUEST          = 8
	TYPE_ECHO_RESPONSE         = 9
)

type Codec interface {
	Decode(encoded uint64)
	Encode() uint64
	GetID() uint64
	SetID(ID uint64)
}
