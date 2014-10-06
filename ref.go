package goamf

import (
  "errors"
)

type refStore struct {
  stringRef []string
  objectRef []interface{}
  traitsRef []*AMF3Object
}

func (ref *refStore) addStringRef(str string) {
  if ref == nil {
    return
  }
  
  if ref.stringRef == nil {
    ref.stringRef = make([]string, 0, 1)
  }
  ref.stringRef = append(ref.stringRef, str)
}

func (ref *refStore) findStringRef(str string) (uint32, bool) {
  if ref == nil || ref.stringRef == nil {
    return 0, false
  }
  
  stringRef := ref.stringRef
  length := uint32(len(stringRef))
  for index := uint32(0); index < length; index++ {
    if str == stringRef[index] {
      return index, true
    }
  }
  
  return 0, false
}

func (ref *refStore) getStringRef(index uint32) (string, error) {
  if ref == nil {
    return "", errors.New("Ref store is nil")
  }
  
  if index >= uint32(len(ref.stringRef)) {
    return "", errors.New("The index of string ref is out of range")
  }
  
  return ref.stringRef[index], nil
}

func (ref *refStore) addObjectRef(v interface{}) {
  if ref == nil {
    return
  }
  
  if ref.objectRef == nil {
    ref.objectRef = make([]interface{}, 0, 1)
  }
  ref.objectRef = append(ref.objectRef, v)
}

func (ref *refStore) findObjectRef(v interface{}) (uint32, bool) {
  if ref == nil || ref.objectRef == nil {
    return 0, false
  }
  
  objectRef := ref.objectRef
  length := uint32(len(objectRef))
  for index := uint32(0); index < length; index++ {
    if v == objectRef[index] {
      return index, true
    }
  }
  
  return 0, false
}

func (ref *refStore) getObjectRef(index uint32) (interface{}, error) {
  if ref == nil {
    return "", errors.New("Ref store is nil")
  }
  if index >= uint32(len(ref.objectRef)) {
    return "", errors.New("The index of object ref is out of range")
  }
  
  return ref.objectRef[index], nil
}

func (ref *refStore) addTraitsRef(obj *AMF3Object) {
  if ref == nil {
  }
  
  if ref.traitsRef == nil {
    ref.traitsRef = make([]*AMF3Object, 0, 1)
  }
  ref.traitsRef = append(ref.traitsRef, obj)
}

func (ref *refStore) findTraitsRef(obj *AMF3Object) (uint32, bool) {
  if ref == nil || ref.traitsRef == nil {
    return 0, false
  }
  
  traitsRef := ref.traitsRef
  length := uint32(len(traitsRef))
  for index := uint32(0); index < length; index++ {
    if obj == traitsRef[index] {
      return index, true
    }
  }
  
  return 0, false
}

func (ref *refStore) getTraitsRef(index uint32) (*AMF3Object, error) {
  if ref == nil {
    return nil, errors.New("Ref store is nil")
  }
  if index >= uint32(len(ref.traitsRef)) {
    return nil, errors.New("The index of traits ref is out of range")
  }

  obj := ref.traitsRef[index]
  return &AMF3Object{
    obj.ClassName,
    obj.Dyn,
    obj.keys,
    make(map[string]interface{}, len(obj.keys)),
    make(map[string]interface{}),
    true,
    uint32(index),
  }, nil
}
