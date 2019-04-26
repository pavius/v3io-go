package command

type DeallocationRequest struct {
	command
	JobHandle uint64
}

func (a *DeallocationRequest) Decode(encoded uint64) {

	// call inherited
	a.command.decode(encoded)

	// decode offset
	a.JobHandle = a.data & 0x2FFFFFFFFFF
}

func (a *DeallocationRequest) Encode() uint64 {
	a.data = a.JobHandle & 0x2FFFFFFFFFF

	return a.command.encode(TYPE_DEALLOCATION_REQUEST)
}
