package zip

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/hjr265/sneak/internal/sneak"
)

var (
	ErrBadFile = fmt.Errorf("invalid ZIP file: %w", sneak.ErrBadFile)
)

type Reader struct {
	io.Reader
	head sneak.Header
}

func NewReader(r io.ReaderAt, size int64) (*Reader, error) {
	header, dataOffset, err := readSneakHeader(r, size)
	if err != nil {
		return nil, err
	}
	return &Reader{io.NewSectionReader(r, int64(dataOffset), header.Size), header}, nil
}

func (r Reader) Header() sneak.Header {
	return r.head
}

func readSneakHeader(r io.ReaderAt, size int64) (head sneak.Header, dataOffset uint64, err error) {
	zs, err := parseZipStruct(r, size)
	if err != nil {
		return sneak.Header{}, 0, err
	}

	rs := io.NewSectionReader(r, int64(zs.gapOffset), size-int64(zs.gapOffset))

	var shBuf [20]byte
	_, err = io.ReadFull(rs, shBuf[:])
	if err != nil {
		return sneak.Header{}, 0, err
	}

	b := readBuf(shBuf[:])
	if sig := b.uint32(); sig != 0x01020304 {
		return sneak.Header{}, 0, fmt.Errorf("%w: bad sneak signature (%#x)", ErrBadFile, sig)
	}
	version := int(b.uint16())
	size = int64(b.uint64())
	filenameLenX := int(b.uint16())
	commentLen := int(b.uint32())

	d := make([]byte, filenameLenX+commentLen)
	_, err = io.ReadFull(rs, d)
	if err != nil {
		return sneak.Header{}, 0, err
	}

	return sneak.Header{
		Version: version,
		Name:    string(d[:filenameLenX]),
		Comment: string(d[filenameLenX : filenameLenX+commentLen]),
		Size:    size,
	}, uint64(zs.gapOffset) + 20 + uint64(filenameLenX+commentLen), nil
}

type readBuf []byte

func (b *readBuf) uint8() uint8 {
	v := (*b)[0]
	*b = (*b)[1:]
	return v
}

func (b *readBuf) uint16() uint16 {
	v := binary.LittleEndian.Uint16(*b)
	*b = (*b)[2:]
	return v
}

func (b *readBuf) uint32() uint32 {
	v := binary.LittleEndian.Uint32(*b)
	*b = (*b)[4:]
	return v
}

func (b *readBuf) uint64() uint64 {
	v := binary.LittleEndian.Uint64(*b)
	*b = (*b)[8:]
	return v
}

func (b *readBuf) sub(n int) readBuf {
	b2 := (*b)[:n]
	*b = (*b)[n:]
	return b2
}
