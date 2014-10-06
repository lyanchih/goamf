package goamf

import (
  "io"
  "reflect"
)

const (
  AMF0 = uint16(0)
  AMF3 = uint16(3)
)

const (
  AMF0_NUMBER_MARKER         = 0x00
  AMF0_BOOLEAN_MARKER        = 0x01
  AMF0_STRING_MARKER         = 0x02
  AMF0_OBJECT_MARKER         = 0x03
  AMF0_MOVIECLIP_MARKER      = 0x04
  AMF0_NULL_MARKER           = 0x05
  AMF0_UNDEFINED_MARKER      = 0x06
  AMF0_REFERENCE_MARKER      = 0x07
  AMF0_ECMA_ARRAY_MARKER     = 0x08
  AMF0_OBJECT_END_MARKER     = 0x09
  AMF0_STRICT_ARRAY_MARKER   = 0x0a
  AMF0_DATE_MARKER           = 0x0b
  AMF0_LONG_STRING_MARKER    = 0x0c
  AMF0_UNSUPPORTED_MARKER    = 0x0d
  AMF0_RECORDSET_MARKER      = 0x0e
  AMF0_XML_DOCUMENT_MARKER   = 0x0f
  AMF0_TYPED_OBJECT_MARKER   = 0x10
  AMF0_ACMPLUS_OBJECT_MARKER = 0x11
)

const (
  AMF0_EMPTY_UTF8 = 0x0000
  AMF0_MAX_STRING_LEN = 65535
)

const (
  AMF3_UNDEFINED_MARKER = 0x00
  AMF3_NULL_MARKER      = 0x01
  AMF3_FALSE_MARKER     = 0x02
  AMF3_TRUE_MARKER      = 0x03
  AMF3_INTEGER_MARKER   = 0x04
  AMF3_DOUBLE_MARKER    = 0x05
  AMF3_STRING_MARKER    = 0x06
  AMF3_XMLDOC_MARKER    = 0x07
  AMF3_DATE_MARKER      = 0x08
  AMF3_ARRAY_MARKER     = 0x09
  AMF3_OBJECT_MARKER    = 0x0a
  AMF3_XML_MARKER       = 0x0b
  AMF3_BYTEARRAY_MARKER = 0x0c
)

const (
  AMF3_UTF8_EMPTY = 0x01
)

const (
  UndefinedKind = reflect.UnsafePointer << 1
  
)

type Reader interface {
  io.Reader
  io.ByteReader
}

type Writer interface {
  io.Writer
  io.ByteWriter
}

type Packet struct {
  Version uint16
  Headers []PacketHeader
  Messages []PacketMessage
}

type PacketHeader struct {
  HeaderName string
  MustUnderstand uint8
  Value interface{}
}

type PacketMessage struct {
  TargetUri, ResponseUri string
  Value interface{}
}

type Undefined struct{}

type AMF0Object map[string]interface{}

type AMF0TypedObject struct {
  className string
  values map[string]interface{}
}

func NewAMF0TypedObject(className string) (*AMF0TypedObject) {
  return &AMF0TypedObject{className, make(map[string]interface{})}
}

func (obj *AMF0TypedObject) AddValue(k string, v interface{}) {
  obj.values[k] = v
}

type AMF3Object struct {
  ClassName string
  Dyn bool
  keys []string
  Values map[string]interface{}
  DynValues map[string]interface{}
  isRefObj bool
  ref uint32
}

func NewAMF3Object(className string, dyn bool) (*AMF3Object) {
  return &AMF3Object{className, dyn, make([]string, 0, 1), make(map[string]interface{}), make(map[string]interface{}), false, uint32(0)}
}

func (obj *AMF3Object) AddValue(k string, v interface{}) {
  obj.Values[k] = v
}

func (obj *AMF3Object) AddDynValue(k string, v interface{}) {
  obj.DynValues[k] = v
}

type AMF3Array struct {
  DenseValues []interface{}
  AssocValues map[string]interface{}
}

func NewAMF3Array(denseCount uint) *AMF3Array {
  if denseCount == 0 {
    denseCount = 1
  }
  
  return &AMF3Array{make([]interface{}, 0, denseCount), make(map[string]interface{})}
}

func (arr *AMF3Array) AddDenseValue(v interface{}) {
  arr.DenseValues = append(arr.DenseValues, v)
}

func (arr *AMF3Array) AddAssocValue(k string, v interface{}) {
  arr.AssocValues[k] = v
}
