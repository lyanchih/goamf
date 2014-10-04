package goamf

import (
  "errors"
)

func NewAmfPacket(version uint16) (*Packet, error) {
  if version != AMF0 && version != AMF3 {
    return nil, errors.New("AMF version must be 0 or 3 version")
  }
  
  return &Packet{version, make([]PacketHeader, 0), make([]PacketMessage, 0)}, nil
}

func (p *Packet) AddHeader(headerName string, mustUnderstand uint8, value interface{}) error {
  if len(headerName) >= AMF0_MAX_STRING_LEN {
    return errors.New("The length of header name is out of range")
  }
  
  p.Headers = append(p.Headers, PacketHeader{headerName, mustUnderstand, value})
  return nil
}

func (p *Packet) AddMessage(targetUri, responseUri string, value interface{}) error {
  if len(targetUri) >= AMF0_MAX_STRING_LEN {
    return errors.New("The length of target Uri is out of range")
  }
  
  if len(responseUri) >= AMF0_MAX_STRING_LEN {
    return errors.New("The length of response Uri is out of range")
  }
  
  if len(p.Messages) == cap(p.Messages) {
    tmpMessages := make([]PacketMessage,  len(p.Messages), len(p.Messages) << 1)
    copy(tmpMessages, p.Messages)
    p.Messages = tmpMessages
  }
  
  p.Messages = append(p.Messages, PacketMessage{targetUri, responseUri, value})
  return nil
}
