package shuffling

import (
	"testing"

	"github.com/nayarsystems/buffer/buffer"
	"github.com/stretchr/testify/require"
)

func Test_TransposeBits_8x8_Eq(t *testing.T) {
	b := &buffer.Buffer{}
	b.InitFromRawBuffer([]byte{
		0b_1010_1010,
		0b_0101_0101,
		0b_1010_1010,
		0b_0101_0101,
		0b_1010_1010,
		0b_0101_0101,
		0b_1010_1010,
		0b_0101_0101,
	})
	tb, err := TransposeBits(b, 8)
	require.Nil(t, err)
	require.Equal(t, b.GetRawBuffer(), tb.GetRawBuffer())
}

func Test_TransposeBits_3x8(t *testing.T) {
	b := &buffer.Buffer{}
	b.InitFromRawBuffer([]byte{
		0b_1001_0001,
		0b_1000_0001,
		0b_1001_0001,
	})

	et := &buffer.Buffer{}
	et.InitFromRawBuffer([]byte{
		0b_1110_0000,
		0b_0101_0000,
		0b_0000_0111,
	})
	tb, err := TransposeBits(b, 8)
	require.Nil(t, err)
	require.Equal(t, et.GetRawBuffer(), tb.GetRawBuffer())

	b2, err := TransposeBits(tb, 3)
	require.Nil(t, err)
	require.Equal(t, b.GetRawBuffer(), b2.GetRawBuffer())
}

func Test_TransposeBits_4x4(t *testing.T) {
	b := &buffer.Buffer{}
	b.InitFromRawBuffer([]byte{
		0b_1001_1001,
		0b_1001_1001,
	})

	et := &buffer.Buffer{}
	et.InitFromRawBuffer([]byte{
		0b_1111_0000,
		0b_0000_1111,
	})
	tb, err := TransposeBits(b, 4)
	require.Nil(t, err)
	require.Equal(t, et.GetRawBuffer(), tb.GetRawBuffer())

	b2, err := TransposeBits(tb, 4)
	require.Nil(t, err)
	require.Equal(t, b.GetRawBuffer(), b2.GetRawBuffer())
}
