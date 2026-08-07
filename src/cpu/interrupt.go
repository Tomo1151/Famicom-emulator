package cpu

const (
	TYPE_NMI InterruptType = iota
	TYPE_IRQ
)

type InterruptType uint8

type Interrupt struct {
	Type          InterruptType
	VectorAddress uint16
	BFlagMask     uint8
	Cycles        uint8
}

var NMI = Interrupt{
	Type:          TYPE_NMI,
	VectorAddress: 0xFFFA,
	BFlagMask:     0b0010_0000,
	Cycles:        7,
}

var IRQ = Interrupt{
	Type:          TYPE_IRQ,
	VectorAddress: 0xFFFE,
	BFlagMask:     0b0010_0000,
	Cycles:        7,
}
