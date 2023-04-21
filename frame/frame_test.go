package frame

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var testFields []*FieldDesc = []*FieldDesc{
	{
		Name:         "4_BIT_INT64_DEF",
		Size:         4,
		DefaultValue: int64(-1),
	},
	{
		Name:         "4_BIT_INT64",
		Size:         4,
		DefaultValue: int64(0),
	},
	{
		Name:         "4_BIT_INT8",
		Size:         4,
		DefaultValue: int8(0),
	},
	{
		Name:         "BOOL",
		DefaultValue: false,
	},
	{
		Name:         "FLOAT32",
		DefaultValue: float32(0),
	},
	{
		Name:         "FLOAT64",
		DefaultValue: float64(0),
	},
	{
		Name:         "28_BIT_BUFFER",
		Size:         28,
		DefaultValue: []byte{},
	},
	{
		Name:         "16_BIT_BUFFER",
		DefaultValue: []byte{0x00, 0x00},
	},
}

func Test_FieldParamErrors(t *testing.T) {
	fields := getBufferFieldInfoCopy(testFields)
	frame := CreateFrame()
	fields[6].Size = 0
	err := frame.AddFields(fields)
	require.NotNil(t, err)
}

func Test_InferFieldsSize(t *testing.T) {
	fields := getBufferFieldInfoCopy(testFields)
	frame := CreateFrame()
	err := frame.AddFields(fields)
	require.Nil(t, err)

	require.Equal(t, frame.fieldsMap["4_BIT_INT64_DEF"].size, 4)
	require.Equal(t, frame.fieldsMap["4_BIT_INT64"].size, 4)
	require.Equal(t, frame.fieldsMap["4_BIT_INT8"].size, 4)
	require.Equal(t, frame.fieldsMap["BOOL"].size, 1)
	require.Equal(t, frame.fieldsMap["FLOAT32"].size, 32)
	require.Equal(t, frame.fieldsMap["FLOAT64"].size, 64)
	require.Equal(t, frame.fieldsMap["28_BIT_BUFFER"].size, 28)
	require.Equal(t, frame.fieldsMap["16_BIT_BUFFER"].size, 16)
}

func Test_EncodeDecode(t *testing.T) {
	srcFrame := CreateFrame()
	err := srcFrame.AddFields(getBufferFieldInfoCopy(testFields))
	require.Nil(t, err)
	err = srcFrame.Set("4_BIT_INT64", -3)
	require.Nil(t, err)
	err = srcFrame.Set("4_BIT_INT8", -3)
	require.Nil(t, err)
	err = srcFrame.Set("BOOL", true)
	require.Nil(t, err)
	err = srcFrame.Set("FLOAT32", 123.123)
	require.Nil(t, err)
	err = srcFrame.Set("FLOAT64", 321.321)
	require.Nil(t, err)
	err = srcFrame.Set("28_BIT_BUFFER", []byte{0x12, 0x34, 0x56, 0x7f})
	require.Nil(t, err)
	err = srcFrame.Set("16_BIT_BUFFER", []byte{0x89, 0xAB})
	require.Nil(t, err)
	srcData, err := srcFrame.Encode()
	require.Nil(t, err)

	dstFrame := CreateFrame()
	err = dstFrame.AddFields(getBufferFieldInfoCopy(testFields))
	require.Nil(t, err)
	dstFrame.Decode(srcData)
	dstData, err := dstFrame.Encode()
	require.Nil(t, err)
	require.Equal(t, srcData, dstData)

	int64_4bit_def, err := dstFrame.Get("4_BIT_INT64_DEF")
	require.Nil(t, err)
	require.Equal(t, int64(-1), int64_4bit_def)

	int64_4bit, err := dstFrame.Get("4_BIT_INT64")
	require.Nil(t, err)
	require.Equal(t, int64(-3), int64_4bit)

	int8_4bit, err := dstFrame.Get("4_BIT_INT8")
	require.Nil(t, err)
	require.Equal(t, int8(-3), int8_4bit)

	vbool, err := dstFrame.Get("BOOL")
	require.Nil(t, err)
	require.Equal(t, true, vbool)

	vfloat32, err := dstFrame.Get("FLOAT32")
	require.Nil(t, err)
	require.Equal(t, float32(123.123), vfloat32)

	vfloat64, err := dstFrame.Get("FLOAT64")
	require.Nil(t, err)
	require.Equal(t, float64(321.321), vfloat64)

	buffer16bit, err := dstFrame.Get("16_BIT_BUFFER")
	require.Nil(t, err)
	require.Equal(t, []byte{0x89, 0xAB}, buffer16bit)

	buffer28bit, err := dstFrame.Get("28_BIT_BUFFER")
	require.Nil(t, err)
	require.Equal(t, []byte{0x12, 0x34, 0x56, 0x70}, buffer28bit)

	// Test EncodeTo func
	dstFrameEncoded, err := dstFrame.Encode()
	require.NoError(t, err)
	newRaw := make([]byte, len(dstFrameEncoded))
	err = dstFrame.EncodeTo(newRaw)
	require.NoError(t, err)
	require.Equal(t, dstFrameEncoded, newRaw)
}

func Test_SetString(t *testing.T) {
	frame := CreateFrame()
	fieldName := "STRING_BUFFER"
	err := frame.AddFields([]*FieldDesc{
		{
			Name:         fieldName,
			DefaultValue: []byte{},
			Size:         (len("Hello") + 4) * 8,
		},
	})
	require.Nil(t, err)

	err = frame.Set(fieldName, "Hello")
	require.Nil(t, err)
	v, err := frame.Get(fieldName)
	require.Nil(t, err)
	ebuf := []byte{}
	ebuf = append(ebuf, "Hello"...)
	ebuf = append(ebuf, make([]byte, 4)...)
	require.Equal(t, ebuf, v)

	err = frame.Set(fieldName, "Hello123456789")
	require.Nil(t, err)
	v, err = frame.Get(fieldName)
	require.Nil(t, err)
	ebuf = []byte("Hello123456789")
	require.Equal(t, ebuf, v)

	eData, err := frame.Encode()
	require.Nil(t, err)

	err = frame.Decode(eData)
	require.Nil(t, err)
	v, err = frame.Get(fieldName)
	require.Nil(t, err)
	ebuf = []byte("Hello1234")
	require.Equal(t, ebuf, v)
}

func getBufferFieldInfoCopy(fields []*FieldDesc) []*FieldDesc {
	fieldsCopy := []*FieldDesc{}
	for _, field := range fields {
		fieldCopy := *field
		fieldsCopy = append(fieldsCopy, &fieldCopy)
	}
	return fieldsCopy
}
