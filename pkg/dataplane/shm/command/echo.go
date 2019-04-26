package command

type EchoRequest struct {
	command
	Payload uint64
}

func (e *EchoRequest) Decode(encoded uint64) {
	e.command.decode(encoded)

	e.Payload = e.data & 0x2FFFFFFFFFF
}

func (e *EchoRequest) Encode() uint64 {
	e.data = e.Payload & 0x2FFFFFFFFFF

	return e.command.encode(TYPE_ECHO_REQUEST)
}

type EchoResponse struct {
	command
	Payload uint64
}

func (e *EchoResponse) Decode(encoded uint64) {
	e.command.decode(encoded)

	e.Payload = e.data & 0x2FFFFFFFFFF
}

func (e *EchoResponse) Encode() uint64 {
	e.data = e.Payload & 0x2FFFFFFFFFF

	return e.command.encode(TYPE_ECHO_RESPONSE)
}
