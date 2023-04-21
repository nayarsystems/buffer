package frame

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/jaracil/ei"
	"github.com/nayarsystems/buffer/buffer"
	"github.com/nayarsystems/buffer/vars"
)

type FieldDesc struct {
	Name         string      `json:"name"`
	Size         int         `json:"size"`
	DefaultValue interface{} `json:"defaultValue"`
}

type Frame struct {
	vars      *vars.VarsBank
	fields    []*field
	fieldsMap map[string]*field
	bitSize   int
}

func CreateFrame() *Frame {
	f := &Frame{
		vars:      vars.CreateVarsBank(),
		fields:    []*field{},
		fieldsMap: map[string]*field{},
	}
	return f
}

func (f *Frame) GetCopy() *Frame {
	fcopy := CreateFrame()
	for fieldName, field := range f.fieldsMap {
		fieldCopy := *field
		fcopy.fieldsMap[fieldName] = &fieldCopy
		fcopy.fields = append(fcopy.fields, &fieldCopy)
	}
	fcopy.vars = f.vars.GetCopy()
	fcopy.bitSize = f.bitSize
	return fcopy
}

func (f *Frame) Same(fieldName string, newValue interface{}) (same bool, err error) {
	return f.vars.Same(fieldName, newValue)
}

func (f *Frame) Set(fieldName string, newValue interface{}) (err error) {
	switch v := newValue.(type) {
	case string:
		fieldDesc, ok := f.fieldsMap[fieldName]
		if !ok {
			return fmt.Errorf("field \"%s\" does not exist", fieldName)
		}
		fieldByteSize := fieldDesc.size / 8
		if fieldDesc.size%8 != 0 {
			fieldByteSize += 1
		}
		strBuf := []byte(v)
		var newBuf []byte
		if fieldByteSize > len(v) {
			newBuf = make([]byte, fieldByteSize)
			copy(newBuf, strBuf)
			for i := len(strBuf); i < fieldByteSize; i++ {
				newBuf[i] = 0
			}
		} else {
			newBuf = strBuf
		}
		err = f.vars.Set(fieldName, newBuf)
	default:
		err = f.vars.Set(fieldName, v)
	}
	return
}

func (f *Frame) Get(fieldName string) (value interface{}, err error) {
	return f.vars.Get(fieldName)
}

func (f *Frame) GetTo(fieldName string, out interface{}) (err error) {
	return f.vars.GetTo(fieldName, out)
}

func (f *Frame) GetFieldsDesc() []*FieldDesc {
	fields := []*FieldDesc{}
	for _, ff := range f.fields {
		fields = append(fields, &FieldDesc{Name: ff.name, Size: ff.size, DefaultValue: ff.defaultValue})
	}
	return fields
}

func (f *Frame) GetBitSize() int {
	return f.bitSize
}

func (f *Frame) GetByteSize() int {
	byteSize := f.bitSize / 8
	if f.bitSize%8 != 0 {
		byteSize += 1
	}
	return byteSize
}

func (f *Frame) AddFields(newFields []*FieldDesc) error {
	for _, desc := range newFields {
		field := &field{name: desc.Name, size: desc.Size, defaultValue: desc.DefaultValue}
		f.fieldsMap[desc.Name] = field
		f.fields = append(f.fields, field)
	}
	for _, field := range f.fields {
		f.vars.InitVar(field.name, field.defaultValue, nil)
		field.offset = f.bitSize
		switch field.defaultValue.(type) {
		case bool:
			// force size
			field.size = 1
		case float32:
			// force size
			field.size = 32
		case float64:
			// force size
			field.size = 64
		default:
			if field.size <= 0 {
				// Size not specified
				switch v := field.defaultValue.(type) {
				case []byte:
					field.size = len(v) * 8
				default:
					// set default size for the field type
					field.size = int(reflect.TypeOf(field.defaultValue).Size()) * 8
				}
				if field.size <= 0 {
					return fmt.Errorf("unable to infer the size of field '%s'", field.name)
				}
			} else {
				// Size specified
				if field.size <= 0 {
					return fmt.Errorf("invalid size value (%d) for field '%s' (must be >= 0)", field.size, field.name)
				}
				switch field.defaultValue.(type) {
				case []byte:
				default:
					var maxSize int
					// get the max. size for the field type
					maxSize = int(reflect.TypeOf(field.defaultValue).Size()) * 8
					if field.size > maxSize {
						return fmt.Errorf("size set (%d) for field '%s' is out of bounds (%d)", field.size, field.name, maxSize)
					}
				}
			}
		}
		f.bitSize += field.size
	}
	return nil
}

func (f *Frame) encode(buffer *buffer.Buffer) error {
	for _, field := range f.fields {
		currentValue, _ := f.vars.Get(field.name)
		var err error
		switch actualValue := currentValue.(type) {
		case bool:
			var v bool
			if v, err = ei.N(currentValue).Bool(); err == nil {
				err = buffer.SetBit(field.offset, v)
			}
		case uint8, uint16, uint, uint32, uint64:
			var v uint64
			if v, err = ei.N(currentValue).Uint64(); err == nil {
				err = buffer.SetBitsFromUint64(field.offset, v, field.size)
			}
		case int8, int16, int, int32, int64:
			var v int64
			if v, err = ei.N(currentValue).Int64(); err == nil {
				err = buffer.SetBitsFromInt64(field.offset, v, field.size)
			}
		case []byte:
			minArraySize := field.size / 8
			if field.size%8 != 0 {
				minArraySize += 1
			}
			actualValueByteSize := len(actualValue)
			if actualValueByteSize < minArraySize {
				actualValue = append(actualValue, make([]byte, minArraySize-actualValueByteSize)...)
			}
			buf := new(bytes.Buffer)
			if err = binary.Write(buf, binary.BigEndian, actualValue); err == nil {
				err = buffer.SetBitsFromRawBuffer(field.offset, buf.Bytes(), field.size)
			}
		default:
			buf := new(bytes.Buffer)
			if err = binary.Write(buf, binary.BigEndian, currentValue); err == nil {
				err = buffer.SetBitsFromRawBuffer(field.offset, buf.Bytes(), field.size)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Frame) EncodeTo(out []byte) error {
	buffer := &buffer.Buffer{}
	buffer.InitFromRawBuffer(out)
	err := f.encode(buffer)
	if err != nil {
		return err
	}
	return nil
}

func (f *Frame) Encode() ([]byte, error) {
	buffer := &buffer.Buffer{}
	buffer.Init(f.bitSize)
	err := f.encode(buffer)
	if err != nil {
		return nil, err
	}
	return buffer.GetRawBuffer(), nil
}

func (f *Frame) Decode(rawInput []byte) error {
	input := &buffer.Buffer{}
	err := input.InitFromRawBufferN(rawInput, f.bitSize)
	if err != nil {
		return err
	}
	for _, field := range f.fields {
		currentValue, _ := f.vars.Get(field.name)
		var newValue interface{}
		switch currentValue.(type) {
		case bool:
			var err error
			if newValue, err = input.GetBit(field.offset); err != nil {
				return err
			}
		case uint8, uint16, uint, uint32, uint64:
			newRawValue, err := input.GetBitsToUint64(field.offset, field.size)
			if err != nil {
				return err
			}
			switch currentValue.(type) {
			case uint8:
				newValue = uint8(newRawValue)
			case uint16:
				newValue = uint16(newRawValue)
			case uint:
				newValue = uint(newRawValue)
			case uint32:
				newValue = uint32(newRawValue)
			case uint64:
				newValue = uint64(newRawValue)
			default:
				return fmt.Errorf("unknown type of field '%s'", field.name)
			}
		case int8, int16, int, int32, int64:
			newRawValue, err := input.GetBitsToInt64(field.offset, field.size)
			if err != nil {
				return err
			}
			switch currentValue.(type) {
			case int8:
				newValue = int8(newRawValue)
			case int16:
				newValue = int16(newRawValue)
			case int:
				newValue = int(newRawValue)
			case int32:
				newValue = int32(newRawValue)
			case int64:
				newValue = int64(newRawValue)
			default:
				return fmt.Errorf("unknown type of field '%s'", field.name)
			}
		default:
			data, err := input.GetBitsToRawBuffer(field.offset, field.size)
			if err != nil {
				return err
			}
			switch currentValue.(type) {
			case float32:
				breader := bytes.NewReader(data)
				var floatValue float32
				if err := binary.Read(breader, binary.BigEndian, &floatValue); err != nil {
					return err
				}
				newValue = floatValue
			case float64:
				var floatValue float64
				breader := bytes.NewReader(data)
				if err := binary.Read(breader, binary.BigEndian, &floatValue); err != nil {
					return err
				}
				newValue = floatValue
			case []byte:
				newValue = data
			}
		}
		if err := f.vars.Set(field.name, newValue); err != nil {
			return err
		}
	}
	return nil
}

type field struct {
	name         string
	size         int
	offset       int
	defaultValue interface{}
}
