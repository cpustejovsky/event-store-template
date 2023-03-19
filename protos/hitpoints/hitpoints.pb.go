// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.12.4
// source: protos/hitpoints/levels.proto

package hitpoints

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

// PlayerCharacterHitPoints records changes hit points
// for the player character with CharacterName
// along with a Note as to the reason
type PlayerCharacterHitPoints struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CharacterName      string `protobuf:"bytes,1,opt,name=CharacterName,proto3" json:"CharacterName,omitempty"`
	CharacterHitPoints int32  `protobuf:"varint,2,opt,name=CharacterHitPoints,proto3" json:"CharacterHitPoints,omitempty"`
	Note               string `protobuf:"bytes,3,opt,name=Note,proto3" json:"Note,omitempty"`
}

func (x *PlayerCharacterHitPoints) Reset() {
	*x = PlayerCharacterHitPoints{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_hitpoints_hitpoints_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PlayerCharacterHitPoints) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlayerCharacterHitPoints) ProtoMessage() {}

func (x *PlayerCharacterHitPoints) ProtoReflect() protoreflect.Message {
	mi := &file_protos_hitpoints_hitpoints_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PlayerCharacterHitPoints.ProtoReflect.Descriptor instead.
func (*PlayerCharacterHitPoints) Descriptor() ([]byte, []int) {
	return file_protos_hitpoints_hitpoints_proto_rawDescGZIP(), []int{0}
}

func (x *PlayerCharacterHitPoints) GetCharacterName() string {
	if x != nil {
		return x.CharacterName
	}
	return ""
}

func (x *PlayerCharacterHitPoints) GetCharacterHitPoints() int32 {
	if x != nil {
		return x.CharacterHitPoints
	}
	return 0
}

func (x *PlayerCharacterHitPoints) GetNote() string {
	if x != nil {
		return x.Note
	}
	return ""
}

var File_protos_hitpoints_hitpoints_proto protoreflect.FileDescriptor

var file_protos_hitpoints_hitpoints_proto_rawDesc = []byte{
	0x0a, 0x20, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x68, 0x69, 0x74, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x73, 0x2f, 0x68, 0x69, 0x74, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x09, 0x68, 0x69, 0x74, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x22, 0x84, 0x01,
	0x0a, 0x18, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x43, 0x68, 0x61, 0x72, 0x61, 0x63, 0x74, 0x65,
	0x72, 0x48, 0x69, 0x74, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x12, 0x24, 0x0a, 0x0d, 0x43, 0x68,
	0x61, 0x72, 0x61, 0x63, 0x74, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0d, 0x43, 0x68, 0x61, 0x72, 0x61, 0x63, 0x74, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x2e, 0x0a, 0x12, 0x43, 0x68, 0x61, 0x72, 0x61, 0x63, 0x74, 0x65, 0x72, 0x48, 0x69, 0x74,
	0x50, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x12, 0x43, 0x68,
	0x61, 0x72, 0x61, 0x63, 0x74, 0x65, 0x72, 0x48, 0x69, 0x74, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x73,
	0x12, 0x12, 0x0a, 0x04, 0x4e, 0x6f, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x4e, 0x6f, 0x74, 0x65, 0x42, 0x36, 0x5a, 0x34, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x63, 0x70, 0x75, 0x73, 0x74, 0x65, 0x6a, 0x6f, 0x76, 0x73, 0x6b, 0x79, 0x2f,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x2d, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x73, 0x2f, 0x68, 0x69, 0x74, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protos_hitpoints_hitpoints_proto_rawDescOnce sync.Once
	file_protos_hitpoints_hitpoints_proto_rawDescData = file_protos_hitpoints_hitpoints_proto_rawDesc
)

func file_protos_hitpoints_hitpoints_proto_rawDescGZIP() []byte {
	file_protos_hitpoints_hitpoints_proto_rawDescOnce.Do(func() {
		file_protos_hitpoints_hitpoints_proto_rawDescData = protoimpl.X.CompressGZIP(file_protos_hitpoints_hitpoints_proto_rawDescData)
	})
	return file_protos_hitpoints_hitpoints_proto_rawDescData
}

var file_protos_hitpoints_hitpoints_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_protos_hitpoints_hitpoints_proto_goTypes = []interface{}{
	(*PlayerCharacterHitPoints)(nil), // 0: hitpoints.PlayerCharacterHitPoints
}
var file_protos_hitpoints_hitpoints_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_protos_hitpoints_hitpoints_proto_init() }
func file_protos_hitpoints_hitpoints_proto_init() {
	if File_protos_hitpoints_hitpoints_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protos_hitpoints_hitpoints_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PlayerCharacterHitPoints); i {
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
			RawDescriptor: file_protos_hitpoints_hitpoints_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_hitpoints_hitpoints_proto_goTypes,
		DependencyIndexes: file_protos_hitpoints_hitpoints_proto_depIdxs,
		MessageInfos:      file_protos_hitpoints_hitpoints_proto_msgTypes,
	}.Build()
	File_protos_hitpoints_hitpoints_proto = out.File
	file_protos_hitpoints_hitpoints_proto_rawDesc = nil
	file_protos_hitpoints_hitpoints_proto_goTypes = nil
	file_protos_hitpoints_hitpoints_proto_depIdxs = nil
}
