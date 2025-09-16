package cpu


type InterruptType uint8

const (
	TYPE_NMI InterruptType = iota
	TYPE_IRQ
)

type Interrupt struct {
	Type InterruptType
	VectorAddress uint16
	BFlagMask uint8
	CPUCycles uint8
}

var NMI = Interrupt{Type: TYPE_NMI, VectorAddress: 0xFFFA, BFlagMask: 0b0010_0000, CPUCycles: 2}

var IRQ = Interrupt{Type: TYPE_IRQ, VectorAddress: 0xFFFE, BFlagMask: 0b0001_0000, CPUCycles: 2}