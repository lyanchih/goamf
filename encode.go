package goamf

import (
  "io"
  "bytes"
  "errors"
  "runtime"
  "reflect"
  "encoding/binary"
)

type marshaler interface {
  marshalAmf(e *encodeState) error
}

func marshal(version uint16, v interface{}) ([]byte, error) {
  e := &encodeState{
    refStore: new(refStore),
    version: version,
  }
  err := e.marshal(v)
  if err != nil {
    return nil, err
  }
  return e.Bytes(), nil
}

var Marshal = MarshalAmf0

func MarshalAmf0(v interface{}) ([]byte, error) {
  return marshal(AMF0, v)
}

func MarshalAmf3(v interface{}) ([]byte, error) {
  return marshal(AMF3, v)
}

type encodeState struct {
  bytes.Buffer
  *refStore
  version uint16
}

func (e *encodeState) marshal(v interface{}) (err error) {
  defer func() {
    if r := recover(); r != nil {
      if _, ok := r.(runtime.Error); ok {
        panic(r)
      }
      if s, ok := r.(string); ok {
        panic(s)
      }
      err = r.(error)
    }
  }()
  
  e.reflectValue(reflect.ValueOf(v))
  return nil
}

func (e *encodeState) marshalNew(v interface{}) (e2 *encodeState, err error) {
  e2 = &encodeState{
    refStore: e.refStore,
    version: e.version,
  }
  err = e2.marshal(v)
  if err != nil {
    return nil, err
  }
  
  return
}

func (e *encodeState) reflectValue(v reflect.Value) {
  err := valueEncoder(e.version, v)(e, v)
  if err != nil {
    panic(err)
  }
}

type encoderFunc func(e *encodeState, v reflect.Value) (err error)

func valueEncoder(version uint16, v reflect.Value) encoderFunc {
  if !v.IsValid() {
    return nilValueEncoder
  }
  
  return typeEncoder(v.Type())
}

// AMF0_NULL_MARKER, AMF3_NULL_MARKER
func nilValueEncoder(e *encodeState, v reflect.Value) error {
  marker := byte(AMF0_NULL_MARKER)
  if e.version != AMF0 {
    marker = byte(AMF3_NULL_MARKER)
  }
  return e.WriteByte(marker)
}

func invalidValueEncoder(e *encodeState, v reflect.Value) error {
  return errors.New("Invalid value encoder")
}

var marshalerType = reflect.TypeOf(new(marshaler)).Elem()

func typeEncoder(t reflect.Type) encoderFunc {
  if t.Implements(marshalerType) {
    return marshalerEncoder
  }
  
  switch t.Kind() {
  case reflect.String:
    return stringEncoder
  case reflect.Bool:
    return booleanEncoder
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    return numberEncoder
  case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
    return numberEncoder
  case reflect.Float32, reflect.Float64:
    return numberEncoder
  case reflect.Array, reflect.Slice:
    return arrayEncoder
  case reflect.Map:
  }
  return invalidValueEncoder
}

func marshalerEncoder(e *encodeState, v reflect.Value) error {
  if v.Kind() == reflect.Ptr && v.IsNil() {
    return errors.New("The marshaler value shouldn't be nil pointer")
  }
  
  m := v.Interface().(marshaler)
  return m.marshalAmf(e)
}

// AMF0_NUMBER_MARKER, AMF3_INTEGER_MARKER
func numberEncoder(e *encodeState, v reflect.Value) (err error) {
  if e.version == AMF0 {
    _, err = writeDouble(e, v.Float())
  } else {
    _, err = writeInteger(e, uint32(v.Uint()))
  }
  return
}

// AMF0_BOOLEAN_MARKER, AMF3_TRUE_MARKER, AMF3_FALSE_MARKER
func booleanEncoder(e *encodeState, v reflect.Value) (err error) {
  if e.version == AMF0 {
    _, err = writeBoolean(e, v.Bool())
  } else {
    _, err = writeTrueOrFalse(e, v.Bool())
  }
  return err
}

// AMF0_STRING_MARKER, AMF3_STRING_MARKER
func stringEncoder(e *encodeState, v reflect.Value) (err error) {
  if e.version == AMF0 {
    _, err = writeAMF0String(e, v.String())
  } else {
    err = e.WriteByte(byte(AMF3_STRING_MARKER))
    if err != nil {
      return err
    }
    
    _, err = writeUTF8Vr(e, v.String())
  }
  return err
}

// AMF0_STRICT_ARRAY_MARKER
func arrayEncoder(e *encodeState, v reflect.Value) (err error) {
  if e.version == AMF0 {
    err = writeStrictArray(e, v.Interface().([]interface{}))
  }
  return err
}

func (p *Packet) marshalAmf(e *encodeState) (err error) {
  err = binary.Write(e, binary.BigEndian, &p.Version)
  if err != nil {
    return
  }
  
  err = writeU16(e, uint16(len(p.Headers)))
  if err != nil {
    return
  }
  for _, header := range p.Headers {
    _, err = writeUTF8(e, header.HeaderName)
    if err != nil {
      return
    }
    
    err = e.WriteByte(byte(header.MustUnderstand))
    if err != nil {
      return
    }
    
    var e2 *encodeState
    e2, err = e.marshalNew(header.Value)
    if err != nil {
      return
    }
    
    err = writeU32(e, uint32(e2.Len()))
    if err != nil {
      return
    }
    
    _, err = io.Copy(e, e2)
    if err != nil {
      return
    }
  }
  
  err = writeU16(e, uint16(len(p.Messages)))
  if err != nil {
    return
  }
  for _, message := range p.Messages {
    _, err = writeUTF8(e, message.TargetUri)
    if err != nil {
      return
    }
    
    _, err = writeUTF8(e, message.ResponseUri)
    if err != nil {
      return
    }
    
    var e2 *encodeState
    e2, err = e.marshalNew(message.Value)
    if err != nil {
      return
    }
    
    length := uint32(e2.Len())
    err = binary.Write(e, binary.BigEndian, &length)
    if err != nil {
      return
    }
    
    _, err = io.Copy(e, e2)
    if err != nil {
      return
    }
  }
  return
}

// AMF0_OBJECT_MARKER
func (obj AMF0Object) marshalAmf(e *encodeState) error {
  err := e.WriteByte(AMF0_OBJECT_MARKER)
  if err != nil {
    return err
  }
  
  for k, v := range obj {
    _, err = writeUTF8(e, k)
    if err != nil {
      return err
    }
    
    err = e.marshal(v)
    if err != nil {
      return err
    }
  }
  
  err = writeAMF0EmptyUTF8(e)
  if err != nil {
    return err
  }
  
  err = e.WriteByte(AMF0_OBJECT_END_MARKER)
  if err != nil {
    return err
  }
  
  e.addObjectRef(obj)
  return nil
}

// AMF0_UNDEFINED_MARKER, AMF3_UNDEFINED_MARKER
func (undef Undefined) marshalAmf(e *encodeState) error {
  marker := byte(AMF0_UNDEFINED_MARKER)
  if e.version != AMF0 {
    marker = byte(AMF3_UNDEFINED_MARKER)
  }
  
  return e.WriteByte(marker)
}

// AMF3_ARRAY_MARKER
func (arr AMF3Array) marshalAmf(e *encodeState) (err error) {
  err = e.WriteByte(byte(AMF3_ARRAY_MARKER))
  if err != nil {
    return
  }
  
  if index, hasRef := e.findObjectRef(&arr); hasRef {
    return writeU29Ref(e, index)
  }
  
  length := uint32(0)
  if arr.DenseValues != nil && len(arr.DenseValues) != 0 {
    length = uint32(len(arr.DenseValues))
  }
  
  u29 := length << 1 | 0x01
  _, err = writeU29(e, u29)
  if err != nil {
    return
  }
  
  for k, v := range arr.AssocValues {
    err = writeAssocValue(e, k, v);
    if err != nil {
      return
    }
  }
  
  err = writeAMF3EmptyUTF8(e)
  if err != nil {
    return
  }
  
  for _, v := range arr.DenseValues {
    err = e.marshal(v)
    if err != nil {
      return
    }
  }
  
  if err != nil {
    e.objectRef = append(e.objectRef, arr)
  }
  return
}

// AMF3_OBJECT_MARKER
func (obj *AMF3Object) marshalAmf(e *encodeState) (err error) {
  version := e.version
  if e.version != AMF3 {
    err = e.WriteByte(byte(AMF0_ACMPLUS_OBJECT_MARKER))
    if err != nil {
      return
    }
    e.version = AMF3
  }
  
  err = e.WriteByte(byte(AMF3_OBJECT_MARKER))
  if err != nil {
    return
  }
  
  if index, hasRef := e.findTraitsRef(obj); hasRef {
    return writeTraitsRef(e, index)
  }
  valueLen := uint32(len(obj.Values)) << 4 | 0x03
  if obj.Dyn {
    valueLen = valueLen | 0x08
  }
  _, err = writeU29(e, valueLen)
  if err != nil {
    return
  }
  
  _, err = writeUTF8Vr(e, obj.ClassName)
  if err != nil {
    return
  }
  
  tmpValues := make([]interface{}, 0, len(obj.Values))
  for k, v := range obj.Values {
    _, err = writeUTF8Vr(e, k)
    if err != nil {
      return err
    }
    tmpValues = append(tmpValues, v)
  }
  
  for _, v := range tmpValues {
    err = e.marshal(v)
    if err != nil {
      return err
    }
  }
  
  if obj.Dyn {
    for k, v := range obj.DynValues {
      _, err = writeUTF8Vr(e, k)
      if err != nil {
        return err
      }
      err = e.marshal(v)
      if err != nil {
        return err
      }
    }
    
    err = writeAMF3EmptyUTF8(e)
    if err != nil {
      return
    }
  }
  
  e.addTraitsRef(obj)
  e.version = version
  return
}
