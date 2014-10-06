package goamf

import (
  "errors"
  "encoding/binary"
)

func readU8(r Reader) (num uint8, err error) {
  err = binary.Read(r, binary.BigEndian, &num)
  return
}

func writeU8(w Writer, num uint8) error {
  return binary.Write(w, binary.BigEndian, &num)
}

func readU16(r Reader) (num uint16, err error) {
  err = binary.Read(r, binary.BigEndian, &num)
  return
}

func writeU16(w Writer, num uint16) error {
  return binary.Write(w, binary.BigEndian, &num)
}

func readU32(r Reader) (num uint32, err error) {
  err = binary.Read(r, binary.BigEndian, &num)
  return
}

func writeU32(w Writer, num uint32) error {
  return binary.Write(w, binary.BigEndian, &num)
}

//
//
//  AMF0 Reader
//
//

func readDouble(r Reader) (num float64, err error) {
  err = binary.Read(r, binary.BigEndian, &num)
  return
}

func readBoolean(r Reader) (bool, error) {
  b, err := r.ReadByte()
  if err != nil {
    return false, err
  }
  
  return b != 0x00, err
}

func readUTF8(r Reader) (string, error) {
  var length uint16
  err := binary.Read(r, binary.BigEndian, &length)
  if err != nil {
    return "", err
  } else if length == 0 {
    return "", nil
  }
  
  data := make([]byte, length)
  _, err = r.Read(data)
  if err != nil {
    return "", err
  }
  return string(data), nil
}

func readLongUTF8(r Reader) (string, error) {
  var length uint32
  err := binary.Read(r, binary.BigEndian, &length)
  if err != nil {
    return "", err
  } else if length == 0 {
    return "", nil
  }
  
  data := make([]byte, length)
  _, err = r.Read(data)
  if err != nil {
    return "", err
  }
  return string(data), nil
}

func readStrictArray(r Reader) ([]interface{}, error) {
  var count uint32
  err := binary.Read(r, binary.BigEndian, &count)
  if err != nil {
    return nil, err
  }

  arr := make([]interface{}, 0, count)
  for i := uint32(0); i < count; i++ {
  }
  
  return arr, nil
}

func readObjectProperty(d *decodeState, values map[string]interface{}) error {
  for {
    k, err := readUTF8(d)
    if err != nil {
      return err
    }
    
    if k == "" {
      mark, err := d.ReadByte()
      if mark == byte(AMF0_OBJECT_END_MARKER) {
        return nil
      } else if err == nil {
        err = errors.New("Can not find AMF0_OBJECT_END_MARKER")
      }
      
      if err != nil {
        return err
      }
    }
    
    v, err := d.unmarshal()
    if err != nil {
      return err
    }
    values[k] = v
  }
  return nil
}

//
//
//  AMF0 Writer
//
//

func writeDouble(w Writer, num float64) (n int, err error) {
  err = w.WriteByte(AMF0_NUMBER_MARKER)
  if err != nil {
    return 0, err
  }
  
  err = binary.Write(w, binary.BigEndian, num)
  if err != nil {
    return 1, err
  }
  return 9, nil
}

func writeBoolean(w Writer, b bool) (n int, err error) {
  err = w.WriteByte(AMF0_BOOLEAN_MARKER)
  if err != nil {
    return 1, err
  }
  
  if b {
    err = w.WriteByte(0x01)
  } else {
    err = w.WriteByte(0x00)
  }
  
  if err != nil {
    return 1, err
  }
  return 2, err
}

func writeAMF0String(w Writer, str string) (n int, err error) {
  length := uint32(len(str))
  if length > 0xffff {
    err = w.WriteByte(AMF0_LONG_STRING_MARKER)
    if err != nil {
      return 0, err
    }
    
    return writeLongUTF8(w, str)
  }
  
  err = w.WriteByte(AMF0_STRING_MARKER)
  if err != nil {
    return 0, err
  }
  return writeUTF8(w, str)
}

func writeUTF8(w Writer, str string) (n int, err error) {
  length := uint16(len(str))
  
  err = binary.Write(w, binary.BigEndian, &length)
  if err != nil {
    return 1, err
  }
  
  n, err = w.Write([]byte(str))
  return n+2, err
}

func writeLongUTF8(w Writer, str string) (n int, err error) {
  length := uint32(len(str))
  
  err = binary.Write(w, binary.BigEndian, &length)
  if err != nil {
    return 1, err
  }
  
  n, err = w.Write([]byte(str))
  return n+2, err
}

func writeAMF0EmptyUTF8(w Writer) error {
  data := uint16(AMF0_EMPTY_UTF8)
  return binary.Write(w, binary.BigEndian, &data)
}

func writeStrictArray(e *encodeState, arr []interface{}) error {
  err := e.WriteByte(byte(AMF0_STRICT_ARRAY_MARKER))
  if err != nil {
    return err
  }
  
  count := uint32(len(arr))
  err = binary.Write(e, binary.BigEndian, &count)
  if err != nil {
    return err
  }
  
  for _, v := range arr {
    err = e.marshal(v)
    if err != nil {
      return err
    }
  }
  return nil
}

//
//
//  AMF3 Reader
//
//

func readU29(r Reader) (num uint32, err error) {
  var b byte
  for i := 0; i < 3; i++ {
    b, err = r.ReadByte()
    if err != nil {
      return
    }
    num = (num << 7) + uint32(b & 0x7f)
    if (b & 0x80) == 0 {
      return num, nil
    }
  }
  
  b, err = r.ReadByte()
  if err != nil {
    return 0, err
  }
  return ((num << 8) + uint32(b)), nil
}

func readInteger(r Reader) (uint32, error) {
  return readU29(r)
}

func readUTF8Vr(d *decodeState) (string, error) {
  u29, err := readU29(d)
  if err != nil {
    return "", err
  }
  
  if (u29 & 0x01) == 0 {
    index := u29 >> 1
    return d.getStringRef(index)
  }
  
  length := u29 >> 1
  if length == 0 {
    return "", nil
  }
  
  data := make([]byte, length)
  _, err = d.Read(data)
  if err != nil {
    return "", err
  }
  
  str := string(data)
  d.addStringRef(str)
  return str, nil
}

func readAssocValue(d *decodeState) (k string, v interface{}, err error) {
  k, err = readUTF8Vr(d)
  if err != nil {
    return "", "", err
  }
  
  if k == "" {
    return "", "", nil
  }
  
  v, err = d.unmarshal()
  return
}

//
//
//  AMF3 Writer
//
//

func writeU29(w Writer, num uint32) (n int, err error) {
  if num <= 0x0000007f {
    err = w.WriteByte(byte(num))
    if err != nil {
      return 0, err
    }
    return 1, nil
  } else if num <= 0x00003fff {
    return w.Write([]byte{byte(num>>7 | 0x80), byte(num & 0x7f)})
  } else if num <= 0x001fffff {
    return w.Write([]byte{byte(num>>14 | 0x80), byte(num>>7 & 0x7f | 0x80), byte(num & 0x7f)})
  } else if num <= 0x3fffffff {
    return w.Write([]byte{byte(num>>22 | 0x80), byte(num>>15 & 0x7f | 0x80), byte(num>>8 & 0x7f | 0x80), byte(num)})
  }
  return 0, errors.New("out of range")
}

func writeU29Ref(w Writer, index uint32) error {
  index = index << 1 & 0x00
  _, err := writeU29(w, index)
  return err
}

func writeTraitsRef(w Writer, index uint32) error {
  index = index << 2 | 0x01
  _, err := writeU29(w, index)
  return err
}

func writeInteger(w Writer, num uint32) (int, error) {
  err := w.WriteByte(AMF3_INTEGER_MARKER)
  if err != nil {
    return 0, err
  }
  
  return writeU29(w, num)
}

func writeTrueOrFalse(w Writer, b bool) (n int, err error) {
  if b {
    err = w.WriteByte(AMF3_TRUE_MARKER)
  } else {
    err = w.WriteByte(AMF3_FALSE_MARKER)
  }

  if err != nil {
    return 0, err
  }
  return 1, nil
}

func writeUTF8Vr(e *encodeState, str string) (n int, err error){
  if str == "" {
    err := writeAMF3EmptyUTF8(e)
    return 1, err
  }
  
  if index, find := e.findStringRef(str); find {
    return writeUTF8Ref(e, index)
  }
  
  n, err = writeAMF3UTF8(e, str)
  if err != nil {
    return 0, err
  }
  
  e.addStringRef(str)
  return
}

func writeUTF8Ref(w Writer, index uint32) (n int, err error){
  index = index << 1
  return writeU29(w, index)
}

func writeAMF3UTF8(w Writer, str string) (n int, err error){
  length := uint32(len(str))
  m, err := writeU29(w, length << 1 | 0x01)
  if err != nil {
    return 0, err
  }
  
  
  n, err = w.Write([]byte(str))
  return n+m, err
}

func writeAMF3EmptyUTF8(w Writer) error {
  return w.WriteByte(AMF3_UTF8_EMPTY)
}

func writeAssocValue(e *encodeState, k string, v interface{}) (err error) {
  _, err = writeUTF8Vr(e, k)
  if err != nil {
    return err
  }
  
  err = e.marshal(v)
  return err
}
