package buffer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_AddBitsFromBuffer(t *testing.T) {
	frame := &Buffer{}
	err := frame.Write([]byte{0x055, 0x44, 0x3f}, 20)
	require.Nil(t, err)
	require.Equal(t, []byte{0x055, 0x44, 0x30}, frame.GetRawCopy())
	require.Equal(t, 20, frame.GetBitSize())
	err = frame.Write([]byte{0x32, 0x21, 0x10, 0x0f}, 28)
	require.Nil(t, err)
	require.Equal(t,
		[]byte{0x55, 0x44, 0x33, 0x22, 0x11, 0x00},
		frame.GetRawCopy(),
	)
	require.Equal(t, 48, frame.GetBitSize())
}

func Test_SetBitOutOfBounds_0(t *testing.T) {
	frame := &Buffer{}
	frame.Init(64)
	err := frame.SetBit(64, false)
	require.NotNil(t, err)
}

func Test_SetLastBitTrue(t *testing.T) {
	frame := &Buffer{}
	frame.Init(64)
	err := frame.SetBit(63, true)
	require.Nil(t, err)
	v, err := frame.GetBit(63)
	require.Nil(t, err)
	require.True(t, v)
	v, err = frame.GetBit(-1)
	require.Nil(t, err)
	require.True(t, v)
}

func Test_UnsetAll(t *testing.T) {
	frame := &Buffer{}
	frame.Init(64)
	err := frame.SetBit(63, true)
	require.Nil(t, err)
	frame.UnsetAll()
	v, err := frame.GetBit(63)
	require.Nil(t, err)
	require.False(t, v)
}

func Test_SetLastBitFalse(t *testing.T) {
	frame := &Buffer{}
	frame.Init(64)
	err := frame.SetBit(63, true)
	require.Nil(t, err)
	err = frame.SetBit(62, true)
	require.Nil(t, err)
	err = frame.SetBit(63, false)
	require.Nil(t, err)
	v, err := frame.GetBit(63)
	require.Nil(t, err)
	require.False(t, v)
	v, err = frame.GetBit(62)
	require.Nil(t, err)
	require.True(t, v)
	v, err = frame.GetBit(-2)
	require.Nil(t, err)
	require.True(t, v)
	err = frame.SetBit(-2, false)
	require.Nil(t, err)
	v, err = frame.GetBit(-2)
	require.Nil(t, err)
	require.False(t, v)
}

func Test_GetUint8Bits(t *testing.T) {
	frame := &Buffer{}
	frame.InitFromRawBuffer([]byte{0x12, 0x34, 0x56, 0x78})
	v, err := frame.GetBitsToUint64(4, 8)
	require.Nil(t, err)
	require.Equal(t, uint64(0x23), v)
}

func Test_GetUint4Bits(t *testing.T) {
	frame := &Buffer{}
	frame.InitFromRawBuffer([]byte{0x12, 0x34, 0x56, 0x78})
	v, err := frame.GetBitsToUint64(24, 4)
	require.Nil(t, err)
	require.Equal(t, uint64(0x07), v)
}

func Test_GetUint6Bits(t *testing.T) {
	frame := &Buffer{}
	frame.InitFromRawBuffer([]byte{0x12, 0b_0000_00010, 0b_1000_0000, 0x78})
	v, err := frame.GetBitsToUint64(14, 4)
	require.Nil(t, err)
	require.Equal(t, uint64(0b_1010), v)
}

func Test_GetUint8Bits_RightBoundsError(t *testing.T) {
	frame := &Buffer{}
	frame.InitFromRawBuffer([]byte{0x12, 0x34, 0x56, 0x78})
	_, err := frame.GetBitsToUint64(25, 8)
	require.NotNil(t, err)
}

func Test_GetUint8BitsRight(t *testing.T) {
	frame := &Buffer{}
	frame.InitFromRawBuffer([]byte{0x12, 0x34, 0x56, 0x78})
	v, err := frame.GetBitsToUint64(-5, 8)
	require.Nil(t, err)
	require.Equal(t, uint64(0x67), v)
}

func Test_Set3bitSignedInt(t *testing.T) {
	frame := &Buffer{}
	frame.InitFromRawBuffer([]byte{0x00, 0x00, 0x00, 0x00})
	v0 := int8(-2)
	err := frame.SetBitsFromInt64(0, int64(v0), 3)
	require.Nil(t, err)
	b := frame.GetRawCopy()
	e := []byte{0b1100_0000, 0, 0, 0}
	require.Equal(t, e, b)
	v1, err := frame.GetBitsToInt64(0, 3)
	require.Nil(t, err)
	require.Equal(t, v0, int8(v1))
}

func Test_SetBitsFromBuffer(t *testing.T) {
	frame := &Buffer{}
	frame.InitFromRawBuffer([]byte{0x00, 0x00, 0x00, 0x00})
	input := []byte{0x55, 0x44, 0x0f}
	err := frame.SetBitsFromRawBuffer(4, input, 20)
	require.Nil(t, err)
	b := frame.GetRawCopy()
	e := []byte{0x05, 0x54, 0x40, 0x00}
	require.Equal(t, e, b)
}

func Test_GetBitsToBuffer(t *testing.T) {
	frame := &Buffer{}
	frame.InitFromRawBuffer([]byte{0x05, 0x54, 0x43, 0x32, 0x2f})
	o, err := frame.GetBitsToRawBuffer(4, 28)
	require.Nil(t, err)
	e := []byte{0x55, 0x44, 0x33, 0x20}
	require.Equal(t, e, o)
}

func Test_GetActualIndex(t *testing.T) {
	frame := &Buffer{}
	frame.Init(10)

	reqsize := 10
	_, err := frame.parseParams(-11, reqsize, reqsize)
	require.NotNil(t, err)

	reqsize = 1
	idx, err := frame.parseParams(-10, reqsize, reqsize)
	require.Nil(t, err)
	require.Equal(t, 0, idx)

	reqsize = 2
	_, err = frame.parseParams(-10, reqsize, reqsize)
	require.NotNil(t, err)

	reqsize = 11
	_, err = frame.parseParams(-10, reqsize, reqsize)
	require.NotNil(t, err)

	reqsize = 10
	_, err = frame.parseParams(-10, reqsize, reqsize-1)
	require.NotNil(t, err)

	reqsize = 10
	idx, err = frame.parseParams(0, reqsize, reqsize)
	require.Nil(t, err)
	require.Equal(t, 0, idx)

	reqsize = 10
	_, err = frame.parseParams(1, reqsize, reqsize)
	require.NotNil(t, err)

	reqsize = 2
	_, err = frame.parseParams(9, reqsize, reqsize)
	require.NotNil(t, err)

	reqsize = 1
	_, err = frame.parseParams(10, reqsize, reqsize)
	require.NotNil(t, err)
}

func Test_Read_4(t *testing.T) {
	buf := &Buffer{}
	buf.InitFromRawBufferN([]byte{0x11, 0x22, 0x33, 0x4f}, 28)
	out, err := buf.Read(4)
	require.Nil(t, err)
	require.Equal(t, []byte{0x10}, out.GetRawBuffer())
}

func Test_Read_8(t *testing.T) {
	buf := &Buffer{}
	buf.InitFromRawBufferN([]byte{0x11, 0x22, 0x33, 0x4f}, 28)
	out, err := buf.Read(8)
	require.Nil(t, err)
	require.Equal(t, []byte{0x11}, out.GetRawBuffer())
}

func Test_Read_12(t *testing.T) {
	buf := &Buffer{}
	buf.InitFromRawBufferN([]byte{0x11, 0x22, 0x33, 0x4f}, 28)
	out, err := buf.Read(12)
	require.Nil(t, err)
	require.Equal(t, []byte{0x11, 0x20}, out.GetRawBuffer())
}

func Test_ReadEnd_8(t *testing.T) {
	buf := &Buffer{}
	buf.InitFromRawBufferN([]byte{0x11, 0x22, 0x33, 0x4f}, 28)
	out, err := buf.ReadEnd(8)
	require.Nil(t, err)
	require.Equal(t, []byte{0x34}, out.GetRawBuffer())
}
