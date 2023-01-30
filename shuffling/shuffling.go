package shuffling

import (
	"fmt"

	"github.com/nayarsystems/buffer/buffer"
)

func TransposeBits(input *buffer.Buffer, alignBits int) (out *buffer.Buffer, err error) {
	if alignBits <= 0 {
		return nil, fmt.Errorf("invalid align value (%d). Must be > 0", alignBits)
	}
	bitSize := input.GetBitSize()
	if bitSize%alignBits != 0 {
		return nil, fmt.Errorf("not aligned")
	}
	out = &buffer.Buffer{}
	out.Init(input.GetBitSize())

	numRows := bitSize / alignBits
	dstBit := 0
	for col := 0; col < alignBits; col++ {
		for row := 0; row < numRows; row++ {
			srcBit := row*alignBits + col
			v, err := input.GetBit(srcBit)
			if err != nil {
				return nil, err
			}
			err = out.SetBit(dstBit, v)
			if err != nil {
				return nil, err
			}
			dstBit++
		}
	}
	return out, nil
}
