// compile with:
// protoc --go_out=. --go_opt=paths=source_relative challenge.proto

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.22.2
// source: challenge.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// A challenge represents a clients intention to connect as a player. When
// initiating a connection, a client will issue a "getchallege" message to the
// server. The server will respond with one of these messages which includes
// a challenge number, which is a sort of session id, and the protocol versions
// supported by the server. Original q2 servers only supported protocol 34, but
// r1q2 supports up to 35 and q2pro supports up to 36
type Challenge struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A session number included in the subsequent "connect" message
	Number int32 `protobuf:"varint,1,opt,name=number,proto3" json:"number,omitempty"`
	// All the protocol version supported by the server
	Protocols []int32 `protobuf:"varint,2,rep,packed,name=protocols,proto3" json:"protocols,omitempty"`
}

func (x *Challenge) Reset() {
	*x = Challenge{}
	if protoimpl.UnsafeEnabled {
		mi := &file_challenge_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Challenge) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Challenge) ProtoMessage() {}

func (x *Challenge) ProtoReflect() protoreflect.Message {
	mi := &file_challenge_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Challenge.ProtoReflect.Descriptor instead.
func (*Challenge) Descriptor() ([]byte, []int) {
	return file_challenge_proto_rawDescGZIP(), []int{0}
}

func (x *Challenge) GetNumber() int32 {
	if x != nil {
		return x.Number
	}
	return 0
}

func (x *Challenge) GetProtocols() []int32 {
	if x != nil {
		return x.Protocols
	}
	return nil
}

var File_challenge_proto protoreflect.FileDescriptor

var file_challenge_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x63, 0x68, 0x61, 0x6c, 0x6c, 0x65, 0x6e, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x41, 0x0a, 0x09, 0x43, 0x68, 0x61, 0x6c,
	0x6c, 0x65, 0x6e, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x1c, 0x0a,
	0x09, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x05,
	0x52, 0x09, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x42, 0x26, 0x5a, 0x24, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74,
	0x66, 0x6c, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x2f, 0x6c, 0x69, 0x62, 0x71, 0x32, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_challenge_proto_rawDescOnce sync.Once
	file_challenge_proto_rawDescData = file_challenge_proto_rawDesc
)

func file_challenge_proto_rawDescGZIP() []byte {
	file_challenge_proto_rawDescOnce.Do(func() {
		file_challenge_proto_rawDescData = protoimpl.X.CompressGZIP(file_challenge_proto_rawDescData)
	})
	return file_challenge_proto_rawDescData
}

var file_challenge_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_challenge_proto_goTypes = []interface{}{
	(*Challenge)(nil), // 0: proto.Challenge
}
var file_challenge_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_challenge_proto_init() }
func file_challenge_proto_init() {
	if File_challenge_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_challenge_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Challenge); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_challenge_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_challenge_proto_goTypes,
		DependencyIndexes: file_challenge_proto_depIdxs,
		MessageInfos:      file_challenge_proto_msgTypes,
	}.Build()
	File_challenge_proto = out.File
	file_challenge_proto_rawDesc = nil
	file_challenge_proto_goTypes = nil
	file_challenge_proto_depIdxs = nil
}