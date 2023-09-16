package zip

import (
	"encoding/binary"
	"io"
	"io/fs"

	"github.com/hjr265/sneak/internal/sneak"
)

type Writer struct {
	w    io.Writer
	r    io.ReaderAt
	size int64
	zs   zipStruct
}

func NewWriter(w io.Writer, r io.ReaderAt, size int64) (*Writer, error) {
	zs, err := parseZipStruct(r, size)
	if err != nil {
		return nil, err
	}
	return &Writer{
		w:    w,
		r:    r,
		size: size,
		zs:   zs,
	}, nil
}

func (k Writer) SneakFile(f fs.File) error {
	var err error
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	_, err = io.Copy(k.w, io.NewSectionReader(k.r, 0, k.zs.gapOffset))
	if err != nil {
		return err
	}
	headSize, err := sneak.WriteHeader(k.w, sneak.Header{
		Version: 1,
		Name:    fi.Name(),
		Size:    fi.Size(),
		Comment: "",
	})
	if err != nil {
		return err
	}
	dataSize, err := io.Copy(k.w, f)
	if err != nil {
		return err
	}
	_, err = io.Copy(k.w, io.NewSectionReader(k.r, k.zs.gapOffset, k.zs.dirEndOffset-k.zs.gapOffset))
	if err != nil {
		return err
	}
	_, err = k.writeDirEnd(k.zs, int64(headSize)+dataSize)
	return err
}

func (k Writer) writeDirEnd(zs zipStruct, dirOffsetPad int64) (int, error) {
	rs := io.NewSectionReader(k.r, zs.dirEndOffset, k.size-zs.dirEndOffset)

	const dirEndLen = 22

	buf := make([]byte, int(k.size-zs.dirEndOffset))
	_, err := io.ReadFull(rs, buf[:])
	if err != nil {
		return 0, err
	}

	b := writeBuf(buf[16:])
	b.uint32(uint32(zs.firstDirOffset + dirOffsetPad))

	return k.w.Write(buf)
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
