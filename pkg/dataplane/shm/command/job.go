package command

type JobRequest struct {
	command
	JobHandle uint64
}

func (a *JobRequest) Decode(encoded uint64) {

	// call inherited
	a.command.decode(encoded)

	// decode offset
	a.JobHandle = a.data & 0x2FFFFFFFFFF
}

func (a *JobRequest) Encode() uint64 {
	a.data = a.JobHandle & 0x2FFFFFFFFFF

	return a.command.encode(TYPE_JOB_REQUEST)
}

type JobResponse struct {
	command
	JobHandle uint64
}

func (a *JobResponse) Decode(encoded uint64) {

	// call inherited
	a.command.decode(encoded)

	// decode offset
	a.JobHandle = a.data & 0x2FFFFFFFFFF
}

func (a *JobResponse) Encode() uint64 {
	a.data = a.JobHandle & 0x2FFFFFFFFFF

	return a.command.encode(TYPE_JOB_RESPONSE)
}
