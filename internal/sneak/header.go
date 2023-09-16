package sneak

import (
	"encoding/binary"
	"io"
)

type Header struct {
	Version int
	Name    string
	Comment string
	Size    int64
}

func WriteHeader(w io.Writer, head Header) (z int, err error) {
	var buf [20]byte
	b := writeBuf(buf[:])
	b.uint32(0x01020304)
	b.uint16(uint16(head.Version))
	b.uint64(uint64(head.Size))
	b.uint16(uint16(len(head.Name)))
	b.uint32(uint32(len(head.Comment)))
	n, err := w.Write(buf[:])
	if err != nil {
		return n, err
	}
	z += n
	n, err = io.WriteString(w, head.Name)
	if err != nil {
		return z + n, err
	}
	z += n
	n, err = io.WriteString(w, head.Comment)
	z += n
	return z, err
}

type writeBuf []byte

func (b *writeBuf) uint8(v uint8) {
	(*b)[0] = v
	*b = (*b)[1:]
}

func (b *writeBuf) uint16(v uint16) {
	binary.LittleEndian.PutUint16(*b, v)
	*b = (*b)[2:]
}

func (b *writeBuf) uint32(v uint32) {
	binary.LittleEndian.PutUint32(*b, v)
	*b = (*b)[4:]
}

func (b *writeBuf) uint64(v uint64) {
	binary.LittleEndian.PutUint64(*b, v)
	*b = (*b)[8:]
}

// TODO: Remove
type HeaderRaw struct {
	Signature   []byte // 4 bytes
	Version     uint16 // 2 bytes
	Size        uint64 // 8 bytes
	FilenameLen uint16 // 2 bytes
	CommentLen  uint32 // 4 bytes
	Filename    []byte
	Comment     []byte
}
