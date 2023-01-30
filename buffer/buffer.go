package buffer

import (
	"fmt"
)

type Buffer struct {
	bitSize int
	buffer  []byte
}

func (f *Buffer) Init(bitSize int) {
	f.bitSize = bitSize
	byteSize := getByteSize(bitSize)
	f.buffer = make([]byte, byteSize)
}

func (f *Buffer) InitFromRawBuffer(buff []byte) {
	f.bitSize = len(buff) * 8
	f.buffer = buff
}

func (f *Buffer) InitFromRawBufferN(buff []byte, numBits int) error {
	byteSize := getByteSize(numBits)
	if byteSize > len(buff) {
		return fmt.Errorf("not enough bits in init buffer")
	}
	f.bitSize = numBits
	f.buffer = buff
	return nil
}

func (f *Buffer) UnsetAll() {
	for i := range f.buffer {
		f.buffer[i] = 0
	}
}

func (f *Buffer) SetBit(reqidx int, v bool) (err error) {
	var idx int
	if idx, err = f.parseParams(reqidx, 1, -1); err != nil {
		return err
	}
	bytePos := int(idx / 8)
	bitPos := idx % 8
	if v {
		f.buffer[bytePos] |= (0x80 >> bitPos)
	} else {
		f.buffer[bytePos] &= ^(0x80 >> bitPos)
	}
	return nil
}

func (f *Buffer) GetBit(reqidx int) (re bool, err error) {
	var idx int
	if idx, err = f.parseParams(reqidx, 1, -1); err != nil {
		return false, err
	}
	bytePos := int(idx / 8)
	bitPos := idx % 8
	return (f.buffer[bytePos] & (0x80 >> bitPos)) != 0, nil
}

func (f *Buffer) SetBitsFromUint64(reqidx int, v uint64, size int) (err error) {
	var idx int
	if idx, err = f.parseParams(reqidx, size, 64); err != nil {
		return err
	}
	bit := 0
	for i := idx + size - 1; i >= idx; i-- {
		f.SetBit(i, (v&(uint64(1)<<bit)) != 0)
		bit++
	}
	return nil
}

func (f *Buffer) GetBitsToUint64(reqidx int, size int) (res uint64, err error) {
	var idx int
	if idx, err = f.parseParams(reqidx, size, 64); err != nil {
		return 0, err
	}
	var v uint64 = 0
	bit := 0
	for i := idx + size - 1; i >= idx; i-- {
		bitValue, _ := f.GetBit(i)
		if bitValue {
			v |= (uint64(1) << bit)
		} else {
			v &= ^(uint64(1) << bit)
		}
		bit++
	}
	return v, nil
}

func (f *Buffer) SetBitsFromInt64(idx int, v int64, size int) error {
	return f.SetBitsFromUint64(idx, uint64(v), size)
}

func (f *Buffer) GetBitsToInt64(reqidx int, size int) (res int64, err error) {
	var idx int
	if idx, err = f.parseParams(reqidx, size, 64); err != nil {
		return 0, err
	}
	var v uint64 = 0
	bit := 0
	var bitValue bool
	for i := idx + size - 1; i >= idx; i-- {
		bitValue, _ = f.GetBit(i)
		if bitValue {
			v |= (uint64(1) << bit)
		} else {
			v &= ^(uint64(1) << bit)
		}
		bit++
	}
	if bitValue {
		v |= (0xffffffffffffffff << size)
	}
	return int64(v), nil
}

func (f *Buffer) SetBitsFromRawBuffer(reqidx int, b []byte, size int) (err error) {
	var idx int
	if idx, err = f.parseParams(reqidx, size, len(b)*8); err != nil {
		return err
	}
	end := idx + size
	for i := idx; i < end; i++ {
		dstIdx := (i - idx)
		dstBytePos := dstIdx / 8
		dstBitPos := dstIdx % 8
		f.SetBit(i, ((b[dstBytePos]<<dstBitPos)&0x80) != 0)
	}
	return nil
}

func (f *Buffer) GetBitsToRawBuffer(reqidx int, size int) (res []byte, err error) {
	var idx int
	if idx, err = f.parseParams(reqidx, size, -1); err != nil {
		return nil, err
	}
	resBufSize := getByteSize(size)
	resBuf := make([]byte, resBufSize)
	end := idx + size
	for i := idx; i < end; i++ {
		dstIdx := (i - idx)
		dstBytePos := dstIdx / 8
		dstBitPos := dstIdx % 8
		bitVal, _ := f.GetBit(i)
		if bitVal {
			resBuf[dstBytePos] |= 0x80 >> dstBitPos
		}
	}
	return resBuf, nil
}

func (f *Buffer) Write(input []byte, numBits int) (err error) {
	inputBufferSize := len(input)
	inputBufferBitSize := inputBufferSize * 8
	if inputBufferBitSize < numBits {
		return fmt.Errorf("input buffer not enough size")
	}
	freeBits := len(f.buffer)*8 - f.bitSize
	if freeBits < numBits {
		extraBits := numBits - freeBits
		extraBufSize := getByteSize(extraBits)
		f.buffer = append(f.buffer, make([]byte, extraBufSize)...)
	}
	originalBitSize := f.bitSize
	f.bitSize += numBits
	return f.SetBitsFromRawBuffer(originalBitSize, input, numBits)
}

func (f *Buffer) Read(numBits int) (out *Buffer, err error) {
	if f.bitSize < numBits {
		numBits = f.bitSize
	}
	outRaw, err := f.GetBitsToRawBuffer(0, numBits)
	if err != nil {
		return nil, err
	}
	out = &Buffer{}
	out.InitFromRawBufferN(outRaw, numBits)

	byteSizeRead := getByteSize(numBits)
	newByteSize := f.GetByteSize() - byteSizeRead
	padding := numBits % 8
	offset := byteSizeRead
	if padding != 0 {
		newByteSize += 1
		offset -= 1
	}
	newBuffer := make([]byte, newByteSize)
	for newIdx := 0; newIdx < len(newBuffer); newIdx++ {
		lastIdx := newIdx + offset
		newBuffer[newIdx] |= f.buffer[lastIdx] << padding
	}
	f.buffer = newBuffer
	f.bitSize -= numBits
	return
}

func (f *Buffer) ReadEnd(numBits int) (out *Buffer, err error) {
	if f.bitSize < numBits {
		numBits = f.bitSize
	}
	outRaw, err := f.GetBitsToRawBuffer(f.bitSize-numBits, numBits)
	if err != nil {
		return nil, err
	}
	out = &Buffer{}
	out.InitFromRawBufferN(outRaw, numBits)
	f.bitSize -= numBits
	return
}

func (f *Buffer) GetBitSize() int {
	return f.bitSize
}

func (f *Buffer) GetByteSize() int {
	return getByteSize(f.bitSize)
}

func (f *Buffer) GetCopy() *Buffer {
	frame := &Buffer{}
	frame.Init(f.bitSize)
	copy(frame.buffer, f.buffer)
	return frame
}

func (f *Buffer) GetRawCopy() []byte {
	bcopy := make([]byte, len(f.buffer))
	copy(bcopy, f.buffer)
	return bcopy
}

func (f *Buffer) GetRawBuffer() []byte {
	return f.buffer
}

func (f *Buffer) parseParams(idx int, reqSize int, fromToSize int) (newIndex int, err error) {
	actualIdx := idx
	if actualIdx < 0 {
		actualIdx = f.bitSize + idx - reqSize + 1
		if actualIdx < 0 {
			return 0, fmt.Errorf("invalid index")
		}
	}
	if fromToSize >= 0 && fromToSize < reqSize {
		return 0, fmt.Errorf("not enough bits (available: %d  requested: %d)", fromToSize, reqSize)
	}
	if (f.bitSize - actualIdx) < reqSize {
		return 0, fmt.Errorf("not enough bits from index")
	}
	return actualIdx, nil
}

func getByteSize(numBits int) int {
	numBytes := numBits / 8
	if numBits%8 != 0 {
		numBytes += 1
	}
	return numBytes
}
