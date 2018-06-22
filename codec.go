package cofire

import (
	proto "github.com/golang/protobuf/proto"
)

type RatingCodec struct{}

func (c *RatingCodec) Encode(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (c *RatingCodec) Decode(b []byte) (interface{}, error) {
	var v Rating
	return &v, proto.Unmarshal(b, &v)
}

type messageCodec struct{}

func (c *messageCodec) Encode(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (c *messageCodec) Decode(b []byte) (interface{}, error) {
	var v Message
	return &v, proto.Unmarshal(b, &v)
}

type UpdateCodec struct{}

func (c *UpdateCodec) Encode(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (c *UpdateCodec) Decode(b []byte) (interface{}, error) {
	var v Update
	return &v, proto.Unmarshal(b, &v)
}

type EntryCodec struct{}

func (c *EntryCodec) Encode(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (c *EntryCodec) Decode(b []byte) (interface{}, error) {
	var v Entry
	return &v, proto.Unmarshal(b, &v)
}
