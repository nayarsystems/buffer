package vars

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/jaracil/ei"
)

type Var struct {
	Name         string
	Value        interface{}
	IsSet        bool
	Meta         map[string]interface{}
	VarUpdatedCb VarUpdatedCbFunc

	unsetValue interface{}
}

func (r *Var) GetCopy() *Var {
	// Warn: values based on reference types are not "deep copied"
	vcopy := &Var{
		Name:       r.Name,
		IsSet:      r.IsSet,
		Meta:       getMsiCopy(r.Meta),
		unsetValue: r.unsetValue,
	}
	switch v := r.Value.(type) {
	case []byte:
		dstValue := make([]byte, len(v))
		copy(dstValue, v)
		vcopy.Value = dstValue
	default:
		vcopy.Value = v
	}
	return vcopy
}

type VarUpdatedCbFunc = func(varName string, isSet bool, newValue interface{}, meta map[string]interface{})
type VarsBank struct {
	varMap map[string]*Var
	mutex  sync.Mutex
}

func CreateVarsBank() *VarsBank {
	return &VarsBank{
		varMap: map[string]*Var{},
	}
}

func (r *VarsBank) GetCopy() *VarsBank {
	vcopy := &VarsBank{
		varMap: map[string]*Var{},
	}
	for name, reg := range r.varMap {
		vcopy.varMap[name] = reg.GetCopy()
	}
	return vcopy
}

func (r *VarsBank) SetUpdatedCb(varName string, cb VarUpdatedCbFunc) error {
	reg, exists := r.varMap[varName]
	if !exists {
		return fmt.Errorf("%s does not exist", varName)
	}
	reg.VarUpdatedCb = cb
	return nil
}

func (r *VarsBank) InitVar(varName string, unsetValue interface{}, meta map[string]interface{}) {
	if meta == nil {
		meta = map[string]interface{}{}
	}
	r.varMap[varName] = &Var{
		Name:       varName,
		Value:      unsetValue,
		unsetValue: unsetValue,
		IsSet:      false,
		Meta:       meta,
	}
}

func (r *VarsBank) Set(varName string, newValue interface{}) (err error) {
	r.lock()
	defer r.unlock()
	return r.UnsafeSet(varName, newValue)
}

func (r *VarsBank) UnsafeSet(varName string, newValue interface{}) (err error) {
	reg, exists := r.varMap[varName]
	if !exists {
		return fmt.Errorf("%s does not exist", varName)
	}
	oldValue := reg.Value
	var setValue interface{}
	if setValue, err = getSetValue(oldValue, newValue); err != nil {
		return fmt.Errorf("can't update %s (%T) by a value of different type (%T): %s", varName, oldValue, newValue, err.Error())
	}
	if !reg.IsSet || !reflect.DeepEqual(oldValue, newValue) {
		reg.Value = setValue
		reg.IsSet = true
		if reg.VarUpdatedCb != nil {
			reg.VarUpdatedCb(reg.Name, reg.IsSet, reg.Value, getMsiCopy(reg.Meta))
		}
	}
	return nil
}

func (r *VarsBank) Unset(varName string) error {
	r.lock()
	defer r.unlock()
	return r.UnsafeUnset(varName)
}

func (r *VarsBank) UnsafeUnset(varName string) error {
	if reg, ok := r.varMap[varName]; ok {
		if reg.IsSet {
			reg.IsSet = false
			reg.Value = reg.unsetValue
			if reg.VarUpdatedCb != nil {
				reg.VarUpdatedCb(reg.Name, reg.IsSet, reg.Value, getMsiCopy(reg.Meta))
			}
		}
	} else {
		return fmt.Errorf("%s not found", varName)
	}
	return nil
}

func (r *VarsBank) UnsetAll() {
	r.lock()
	defer r.unlock()
	for varName := range r.varMap {
		r.UnsafeUnset(varName)
	}
}

func (r *VarsBank) SetMetaRegister(varName string, metaId string, meta interface{}) error {
	r.lock()
	defer r.unlock()
	return r.UnsafeSetMetaRegister(varName, metaId, meta)
}

func (r *VarsBank) UnsafeSetMetaRegister(varName string, metaId string, meta interface{}) error {
	reg, exists := r.varMap[varName]
	if !exists {
		return fmt.Errorf("%s does not exist", varName)
	}
	reg.Meta[metaId] = meta
	return nil
}

func (r *VarsBank) GetMetaRegister(varName string, metaId string, meta *interface{}) (interface{}, error) {
	r.lock()
	defer r.unlock()
	return r.UnsafeGetMetaRegister(varName, metaId, meta)
}

func (r *VarsBank) UnsafeGetMetaRegister(varName string, metaId string, meta *interface{}) (interface{}, error) {
	reg, exists := r.varMap[varName]
	if !exists {
		return nil, fmt.Errorf("%s does not exist", varName)
	}
	var m interface{}
	if m, exists = reg.Meta[metaId]; !exists {
		return nil, fmt.Errorf("%s does not contain a meta named %s", varName, metaId)
	}
	return m, nil
}

func (r *VarsBank) Get(varName string) (interface{}, error) {
	r.lock()
	defer r.unlock()
	return r.UnsafeGet(varName)
}

func (r *VarsBank) Same(varName string, newValue interface{}) (bool, error) {
	r.lock()
	defer r.unlock()
	curretValue, err := r.UnsafeGet(varName)
	if err != nil {
		return false, err
	}
	var setValue interface{}
	if setValue, err = getSetValue(curretValue, newValue); err != nil {
		return false, nil
	}
	return reflect.DeepEqual(curretValue, setValue), nil
}

func (r *VarsBank) UnsafeGet(varName string) (interface{}, error) {
	reg, exists := r.varMap[varName]
	if !exists {
		return nil, fmt.Errorf("%s does not exist", varName)
	}
	return reg.Value, nil
}

func (r *VarsBank) GetVar(varName string) (*Var, error) {
	r.lock()
	defer r.unlock()
	return r.UnsafeGetVar(varName)
}

func (r *VarsBank) UnsafeGetVar(varName string) (*Var, error) {
	reg, exists := r.varMap[varName]
	if !exists {
		return nil, fmt.Errorf("%s does not exist", varName)
	}
	return reg, nil
}

func (r *VarsBank) GetMeta(varName string) (map[string]interface{}, error) {
	r.lock()
	defer r.unlock()
	return r.UnsafeGetMeta(varName)
}

func (r *VarsBank) UnsafeGetMeta(varName string) (map[string]interface{}, error) {
	reg, exists := r.varMap[varName]
	if !exists {
		return nil, fmt.Errorf("%s does not exist", varName)
	}
	return getMsiCopy(reg.Meta), nil
}

func (r *VarsBank) SetMeta(varName string, meta map[string]interface{}) error {
	r.lock()
	defer r.unlock()
	return r.UnsafeSetMeta(varName, meta)
}

func (r *VarsBank) UnsafeSetMeta(varName string, meta map[string]interface{}) error {
	reg, exists := r.varMap[varName]
	if !exists {
		return fmt.Errorf("%s does not exist", varName)
	}
	reg.Meta = getMsiCopy(meta)
	return nil
}

func (r *VarsBank) SetTrue(varName string) error {
	r.lock()
	defer r.unlock()
	return r.UnsafeSetTrue(varName)
}

func (r *VarsBank) UnsafeSetTrue(varName string) error {
	return r.UnsafeSet(varName, true)
}

func (r *VarsBank) SetFalse(varName string) error {
	r.lock()
	defer r.unlock()
	return r.UnsafeSetFalse(varName)
}

func (r *VarsBank) UnsafeSetFalse(varName string) error {
	return r.UnsafeSet(varName, false)
}

func (r *VarsBank) GetVarList() (varList []*Var) {
	r.lock()
	defer r.unlock()
	return r.UnsafeGetVarList()
}

func (r *VarsBank) UnsafeGetVarList() (varList []*Var) {
	varList = []*Var{}
	for _, reg := range r.varMap {
		varCopy := reg.GetCopy()
		varList = append(varList, varCopy)
	}
	return
}

func (r *VarsBank) GetBoolN(varName string) bool {
	r.lock()
	defer r.unlock()
	return r.UnsafeGetBoolN(varName)
}

func (r *VarsBank) UnsafeGetBoolN(varName string) bool {
	var dst bool
	_ = r.UnsafeGetTo(varName, &dst)
	return dst
}

func (r *VarsBank) GetTo(varName string, dst interface{}) error {
	r.lock()
	defer r.unlock()
	return r.UnsafeGetTo(varName, dst)
}

func (r *VarsBank) UnsafeGetTo(varName string, dst interface{}) error {
	rawValue, err := r.UnsafeGet(varName)
	if err != nil {
		return err
	}
	switch v := dst.(type) {
	case *bool:
		setValue, err := ei.N(rawValue).Bool()
		if err != nil {
			return err
		}
		*v = setValue
	case *uint8:
		setValue, err := ei.N(rawValue).Uint8()
		if err != nil {
			return err
		}
		*v = setValue
	case *uint16:
		setValue, err := ei.N(rawValue).Uint16()
		if err != nil {
			return err
		}
		*v = setValue
	case *uint:
		setValue, err := ei.N(rawValue).Uint()
		if err != nil {
			return err
		}
		*v = setValue
	case *uint32:
		setValue, err := ei.N(rawValue).Uint32()
		if err != nil {
			return err
		}
		*v = setValue
	case *uint64:
		setValue, err := ei.N(rawValue).Uint64()
		if err != nil {
			return err
		}
		*v = setValue
	case *int8:
		setValue, err := ei.N(rawValue).Int8()
		if err != nil {
			return err
		}
		*v = setValue
	case *int16:
		setValue, err := ei.N(rawValue).Int16()
		if err != nil {
			return err
		}
		*v = setValue
	case *int:
		setValue, err := ei.N(rawValue).Int()
		if err != nil {
			return err
		}
		*v = setValue
	case *int32:
		setValue, err := ei.N(rawValue).Int32()
		if err != nil {
			return err
		}
		*v = setValue
	case *int64:
		setValue, err := ei.N(rawValue).Int64()
		if err != nil {
			return err
		}
		*v = setValue
	case *float32:
		setValue, err := ei.N(rawValue).Float32()
		if err != nil {
			return err
		}
		*v = setValue
	case *float64:
		setValue, err := ei.N(rawValue).Float64()
		if err != nil {
			return err
		}
		*v = setValue
	case *[]byte:
		setValue, err := ei.N(rawValue).Bytes()
		if err != nil {
			return err
		}
		*v = setValue
	default:
		return fmt.Errorf("type '%v' not supported", v)
	}
	return nil
}

type msi = map[string]interface{}

func getMsiCopy(data map[string]interface{}) map[string]interface{} {
	resp := msi{}
	for key, val := range data {
		if v, ok := val.(msi); ok {
			resp[key] = getMsiCopy(v)
		} else {
			resp[key] = val
		}
	}
	return resp
}

func sameType(a, b interface{}) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(b)
}

func getSetValue(oldValue interface{}, newValue interface{}) (setValue interface{}, err error) {
	if sameType(oldValue, newValue) {
		setValue = newValue
	} else {
		switch oldValue.(type) {
		case bool:
			setValue, err = ei.N(newValue).Bool()
		case uint8:
			setValue, err = ei.N(newValue).Uint8()
		case uint16:
			setValue, err = ei.N(newValue).Uint16()
		case uint:
			setValue, err = ei.N(newValue).Uint()
		case uint32:
			setValue, err = ei.N(newValue).Uint32()
		case uint64:
			setValue, err = ei.N(newValue).Uint64()
		case int8:
			setValue, err = ei.N(newValue).Int8()
		case int16:
			setValue, err = ei.N(newValue).Int16()
		case int:
			setValue, err = ei.N(newValue).Int()
		case int32:
			setValue, err = ei.N(newValue).Int32()
		case int64:
			setValue, err = ei.N(newValue).Int64()
		case float32:
			setValue, err = ei.N(newValue).Float32()
		case float64:
			setValue, err = ei.N(newValue).Float64()
		default:
			err = fmt.Errorf("conversion not possible")
		}
	}
	return
}

func (r *VarsBank) lock() {
	r.mutex.Lock()
}

func (r *VarsBank) unlock() {
	r.mutex.Unlock()
}
