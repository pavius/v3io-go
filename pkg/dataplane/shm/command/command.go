package command

type command struct {
	ID   uint64
	data uint64
}

func DecodeType(encoded uint64) Type {
	return Type(encoded >> 60)
}

func (c *command) GetID() uint64 {
	return c.ID
}

func (c *command) SetID(ID uint64) {
	c.ID = ID
}

func (c *command) decode(encoded uint64) {

	// extract data (42 bits)
	c.data = encoded & 0x3FFFFFFFFFF
	encoded >>= 42

	// extract command id (18 bits)
	c.ID = encoded & 0x3FFFF
}

func (c *command) encode(t Type) uint64 {

	return uint64((t&0xF)<<60) | ((c.ID & 0x3FFFF) << 42) | (c.data & 0x3FFFFFFFFFF)
}
