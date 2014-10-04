package goamf

import (
  "log"
  "bytes"
  "errors"
  "runtime"
)

func unmarshal(version uint16, data []byte) (interface{}, error) {
  d := &decodeState{
    bytes.NewBuffer(data),
    new(refStore),
    version,
  }
  return unmarshalPacket(d)
}

var Unmarshal = UnmarshalAmf0

func UnmarshalAmf0(data []byte) (interface{}, error) {
  return unmarshal(AMF0, data)
}

func UnmarshalAmf3(data []byte) (interface{}, error) {
  return unmarshal(AMF3, data)
}

type decodeState struct {
  *bytes.Buffer
  *refStore
  version uint16
}

func unmarshalPacket(d *decodeState) (p *Packet, err error) {
  version, err := readU16(d)
  if err == nil && version != AMF0 && version != AMF3 {
    err = errors.New("AMF version should be 0 or 3")
  }
  if err != nil {
    return nil, err
  }
  
  p, err = NewAmfPacket(version)
  if err != nil {
    return
  }
  
  headerCount, err := readU16(d)
  for index := uint16(0); index < headerCount; index++ {
    headerName, err := readUTF8(d)
    if err != nil {
      return nil, err
    }
    
    mustUnderstand, err := readU8(d)
    if err != nil {
      return nil, err
    }
    
    length, err := readU32(d)
    if err != nil {
      return nil, err
    }
    
    var v interface{}
    if length == uint32(0xffffffff) {
      v, err = d.unmarshal()
      if err != nil {
        return nil, err
      }
    } else {
      data := make([]byte, length)
      _, err = d.Read(data)
      if err != nil {
        return nil, err
      }
      
      v, err = d.unmarshalNew(data)
      if err != nil {
        return nil, err
      }
    }
    err = p.AddHeader(headerName, mustUnderstand, v)
    if err != nil {
      return nil, err
    }
  }
  
  messageCount, err := readU16(d)
  for index := uint16(0); index < messageCount; index++ {
    targetUri, err := readUTF8(d)
    if err != nil {
      return nil, err
    }
    
    responseUri, err := readUTF8(d)
    if err != nil {
      return nil, err
    }
    
    length, err := readU32(d)
    if err != nil {
      return nil, err
    }
    
    var v interface{}
    if length == uint32(0xffffffff) {
      v, err = d.unmarshal()
      if err != nil {
        return nil, err
      }
    } else {
      data := make([]byte, length)
      _, err = d.Read(data)
      if err != nil {
        return nil, err
      }
      
      v, err = d.unmarshalNew(data)
      if err != nil {
        return nil, err
      }
    }
    
    err = p.AddMessage(targetUri, responseUri, v)
    if err != nil {
      return nil, err
    }
  }
  
  return
}

func (d *decodeState) unmarshal() (v interface{}, err error) {
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
  
  f, err := typeDecoder(d)
  if err != nil {
    log.Println(err)
    return nil, err
  }
  
  return f(d)
}

func (d *decodeState) unmarshalNew(data []byte) (interface{}, error){
  d2 := &decodeState {
    bytes.NewBuffer(data),
    d.refStore,
    d.version,
  }
  
  return d2.unmarshal()
}

type decoderFunc func(d *decodeState) (v interface{}, err error)

func typeDecoder(d *decodeState) (f decoderFunc, err error) {
  f = nil
  marker, err := d.ReadByte()
  if err != nil {
    return
  }

  if d.version == AMF0 {
    switch marker {
    case AMF0_NUMBER_MARKER: f = amf0NumberDecoder
    case AMF0_BOOLEAN_MARKER: f = amf0BooleanDecoder
    case AMF0_STRING_MARKER: f = amf0StringDecoder
    case AMF0_OBJECT_MARKER: f = amf0ObjectDecoder
    case AMF0_MOVIECLIP_MARKER: err = errors.New("Movie Clip mark is not supported yet")
    case AMF0_NULL_MARKER: f = amf0NullDecoder
    case AMF0_UNDEFINED_MARKER: f = amf0UndefindedDecoder
    case AMF0_REFERENCE_MARKER: err = errors.New("Reference mark is not supported yet")
    case AMF0_ECMA_ARRAY_MARKER: err = errors.New("ECMA array mark is not supported yet")
    case AMF0_OBJECT_END_MARKER: err = errors.New("It shouldn't only occur object end mark")
    case AMF0_STRICT_ARRAY_MARKER: f = amf0StrictArrayDecoder
    case AMF0_DATE_MARKER: err = errors.New("Date mark is not supported yet")
    case AMF0_LONG_STRING_MARKER: f = amf0LongString
    case AMF0_UNSUPPORTED_MARKER: err = errors.New("Unsupported mark is not supported yet")
    case AMF0_RECORDSET_MARKER: err = errors.New("Record Set mark is not supported yet")
    case AMF0_XML_DOCUMENT_MARKER: err = errors.New("XML Document mark is not supported yet")
    case AMF0_TYPED_OBJECT_MARKER: f = amf0TypedObjectDecoder
    case AMF0_ACMPLUS_OBJECT_MARKER: f = amf0AcmPlusObjectDecoder
    }
  } else {
    switch marker {
    case AMF3_UNDEFINED_MARKER: f = amf3UndefinedDecoder
    case AMF3_NULL_MARKER: f = amf3NullDecoder
    case AMF3_FALSE_MARKER: f = amf3FalseDecoder
    case AMF3_TRUE_MARKER: f = amf3TrueDecoder
    case AMF3_INTEGER_MARKER: f = amf3IntegerDecoder
    case AMF3_DOUBLE_MARKER: f = amf3DoubleDecoder
    case AMF3_STRING_MARKER: f = amf3StringDecoder
    case AMF3_XMLDOC_MARKER: err = errors.New("AMF3 XML Document mark is not supported yet")
    case AMF3_DATE_MARKER: err = errors.New("AMF3 Date mark is not supported yet")
    case AMF3_ARRAY_MARKER: f = amf3ArrayDecoder
    case AMF3_OBJECT_MARKER: f = amf3ObjectDecoder
    case AMF3_XML_MARKER: err = errors.New("AMF3 XML mark is not supported yet")
    case AMF3_BYTEARRAY_MARKER: err = errors.New("AMF3 Byte Array mark is not supported yet")
    }
  }
  return
}

//
// AMF0 Decoder
//

func amf0NumberDecoder(d *decodeState) (interface{}, error) {
  return readDouble(d)
}

func amf0BooleanDecoder(d *decodeState) (interface{}, error) {
  return readBoolean(d)
}

func amf0StringDecoder(d *decodeState) (interface{}, error) {
  return readUTF8(d)
}

func amf0ObjectDecoder(d *decodeState) (interface{}, error) {
  obj := make(AMF0Object)
  
  err := readObjectProperty(d, obj)
    
  return obj, err
}

func amf0NullDecoder(d *decodeState) (interface{}, error) {
  return nil, nil
}

func amf0UndefindedDecoder(d *decodeState) (interface{}, error) {
  return Undefined{}, nil
}

func amf0StrictArrayDecoder(d *decodeState) (interface{}, error) {
  count, err := readU32(d)
  if err != nil {
    return nil, err
  }
  
  arr := make([]interface{}, 0, count)
  for i := uint32(0); i < count; i++ {
    v, err := d.unmarshal()
    if err != nil {
      return nil, err
    }
    
    arr = append(arr, v)
  }
  return arr, nil
}

func amf0LongString(d *decodeState) (interface{}, error) {
  return readLongUTF8(d)
}

func amf0TypedObjectDecoder(d *decodeState) (interface{}, error) {
  className, err := readUTF8(d)
  if err != nil {
    return nil, err
  }
  
  obj := NewAMF0TypedObject(className)
  err = readObjectProperty(d, obj.values)
  return obj, err
}

func amf0AcmPlusObjectDecoder(d *decodeState) (interface{}, error) {
  version := d.version
  d.version = AMF3
  f, err := typeDecoder(d)
  if err != nil {
    return nil, err
  }
  obj, err := f(d)
  d.version = version
  return obj, err
}

//
// AMF3 Decoder
//

func amf3UndefinedDecoder(d *decodeState) (interface{}, error) {
  return Undefined{}, nil
}

func amf3NullDecoder(d *decodeState) (interface{}, error) {
  return nil, nil
}

func amf3FalseDecoder(d *decodeState) (interface{}, error) {
  return false, nil
}

func amf3TrueDecoder(d *decodeState) (interface{}, error) {
  return true, nil
}

func amf3IntegerDecoder(d *decodeState) (interface{}, error) {
  u29, err := readU29(d)
  return int32(u29), err
}

func amf3DoubleDecoder(d *decodeState) (interface{}, error) {
  return readDouble(d)
}

func amf3StringDecoder(d *decodeState) (interface{}, error) {
  return readUTF8Vr(d)
}

func amf3ArrayDecoder(d *decodeState) (interface{}, error) {
  length, err := readU29(d)
  if err != nil {
    return nil, err
  }
  
  if length & 0x01 == uint32(0) {
    v, err := d.getObjectRef(length >> 1)
    return v, err
  }
  
  length = length >> 1
  arr := NewAMF3Array(uint(length))
  for {
    k, v, err := readAssocValue(d)
    if err != nil {
      return nil, err
    }
    
    if k == "" && v == "" {
      break
    }
    
    arr.AddAssocValue(k, v)
  }
  
  for i := uint32(0); i < length; i++ {
    v, err := d.unmarshal()
    if err != nil {
      return nil, err
    }
    
    arr.AddDenseValue(v)
  }
  
  d.addObjectRef(arr)
  return arr, nil
}

func amf3ObjectDecoder(d *decodeState) (interface{}, error) {
  var obj *AMF3Object
  u29, err := readU29(d)
  if err != nil {
    return nil, err
  }

  if u29 & 0x01 == 0x00 {
    return nil, errors.New("Not support unmarshal for object reference")
  } else if u29 & 0x03 == 0x01 {
    obj, err = d.getTraitsRef(u29 >> 2)
    if err != nil {
      return nil, err
    }
    
    for _, k := range obj.keys {
      v, err := d.unmarshal()
      if err != nil {
        return nil, err
      }
      obj.AddValue(k, v)
    }
    if len(obj.keys) != len(obj.Values) {
      return nil, errors.New("The traits ref object's length of keys and values is not match")
    }
  } else if u29 & 0x07 == 0x07 {
    return nil, errors.New("Not support unmarshal for traits ext")
  } else if u29 & 0x0f == 0x03 || u29 & 0x0f == 0x0b {
    dyn := false
    if u29 & 0x08 == 0x08 {
      dyn = true
    }
    
    length := u29 >> 4
    className, err := readUTF8Vr(d)
    if err != nil {
      return nil, err
    }
    
    obj = NewAMF3Object(className, dyn)
    ks := make([]string, 0, length)
    vs := make([]interface{}, 0, length)
    for i := uint32(0); i < length; i++ {
      k, err := readUTF8Vr(d)
      if err != nil {
        return nil, err
      }
      ks = append(ks, k)
    }
    obj.keys = ks
    d.addTraitsRef(obj)
    for i := uint32(0); i < length; i++ {
      v, err := d.unmarshal()
      if err != nil {
        return nil, err
      }
      
      vs = append(vs, v)
    }
    
    if len(ks) != len(vs) || len(ks) != int(length) {
      return nil, errors.New("the length of object class members are not match")
    }
    
    for i := uint32(0); i < length; i++ {
      obj.AddValue(ks[i], vs[i])
    }
  } else {
    return nil, errors.New("Not support object unmarshal format")
  }
  
  if obj.Dyn {
    for {
      k, err := readUTF8Vr(d)
      if err != nil {
        return nil, err
      }
      if k == "" {
        break
      }
      v, err := d.unmarshal()
      if err != nil {
        return nil, err
      }
      obj.AddDynValue(k, v)
    }
  }
  return obj, nil
}
