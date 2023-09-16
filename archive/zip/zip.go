package zip

import (
	"fmt"
	"io"
)

type zipStruct struct {
	firstDirOffset int64
	dirEndOffset   int64
	gapOffset      int64
}

func parseZipStruct(r io.ReaderAt, size int64) (zs zipStruct, err error) {
	var buf []byte
	var dirEndOffset int64
	for i, bLen := range []int64{1024, 65 * 1024} {
		if bLen > size {
			bLen = size
		}
		buf = make([]byte, int(bLen))
		if _, err := r.ReadAt(buf, size-bLen); err != nil && err != io.EOF {
			return zipStruct{}, err
		}
		if p := findDirEndSign(buf); p >= 0 {
			buf = buf[p:]
			dirEndOffset = size - bLen + int64(p)
			break
		}
		if i == 1 || bLen == size {
			return zipStruct{}, ErrBadFile
		}
	}

	b := readBuf(buf[4:])
	b.uint16()
	b.uint16()
	b.uint16()
	b.uint16()
	b.uint32()
	firstDirOffset := uint64(b.uint32())

	const (
		fileHeadLen  = 30
		fileHeadSign = 0x04034b50
		dirLen       = 46
		dirHeadSign  = 0x02014b50
	)

	var lastFileHeadOffset int64

	dirOffset := firstDirOffset
	for int64(dirOffset) < dirEndOffset {
		rs := io.NewSectionReader(r, int64(dirOffset), size-int64(dirOffset))

		var buf [dirLen]byte
		_, err := io.ReadFull(rs, buf[:])
		if err != nil {
			return zipStruct{}, err
		}

		b := readBuf(buf[:])
		if sig := b.uint32(); sig != dirHeadSign {
			return zipStruct{}, fmt.Errorf("%w: bad directory header signature (%#x)", ErrBadFile, sig)
		}
		b.uint16()
		b.uint16()
		b.uint16()
		b.uint16()
		b.uint16()
		b.uint16()
		b.uint32()
		b.uint32()
		b.uint32()
		filenameLen := int(b.uint16())
		extraLen := int(b.uint16())
		commentLen := int(b.uint16())
		b = b[4:]
		b.uint32()
		fileHeadOffset := int64(b.uint32())
		if fileHeadOffset > lastFileHeadOffset {
			lastFileHeadOffset = fileHeadOffset
		}
		d := make([]byte, filenameLen+extraLen+commentLen)
		_, err = io.ReadFull(rs, d)
		if err != nil {
			return zipStruct{}, err
		}

		dirOffset += 46 + uint64(filenameLen+extraLen+commentLen)
	}

	rs := io.NewSectionReader(r, int64(lastFileHeadOffset), size-int64(lastFileHeadOffset))

	var fhBuf [fileHeadLen]byte
	_, err = io.ReadFull(rs, fhBuf[:])
	if err != nil {
		return zipStruct{}, err
	}

	b = readBuf(fhBuf[:])
	if sig := b.uint32(); sig != fileHeadSign {
		return zipStruct{}, fmt.Errorf("%w: bad local file header signature (%#x)", ErrBadFile, sig)
	}
	b.uint16()
	b.uint16()
	b.uint16()
	b.uint16()
	b.uint16()
	b.uint32()
	compSize := int64(b.uint32())
	b.uint32()
	filenameLen := int64(b.uint16())
	extraLen := int64(b.uint16())
	d := make([]byte, filenameLen+extraLen)
	_, err = io.ReadFull(rs, d)
	if err != nil {
		return zipStruct{}, err
	}

	gapOffset := lastFileHeadOffset + fileHeadLen + filenameLen + extraLen + compSize

	return zipStruct{
		firstDirOffset: int64(firstDirOffset),
		dirEndOffset:   dirEndOffset,
		gapOffset:      int64(gapOffset),
	}, nil
}

func findDirEndSign(b []byte) int {
	const dirEndLen = 22
	for i := len(b) - dirEndLen; i >= 0; i-- {
		if b[i] == 'P' && b[i+1] == 'K' && b[i+2] == 0x05 && b[i+3] == 0x06 {
			n := int(b[i+dirEndLen-2]) | int(b[i+dirEndLen-1])<<8
			if n+dirEndLen+i <= len(b) {
				return i
			}
		}
	}
	return -1
}
