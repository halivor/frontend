package connection

import (
	"unsafe"

	bp "github.com/halivor/sunshine/bufferpool"
)

type buffer interface {
	Buffer() []byte
	Release()
}

type packet struct {
	buffer
	buf []byte
	pos int
}

func newPacket() *packet {
	ptr := bp.AllocPointer(int(unsafe.Sizeof(packet{})))
	return (*packet)(ptr)
}

func resPacket(p *packet) {
}
