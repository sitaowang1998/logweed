// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.3
// source: skiplist.proto

package skiplist

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

type SkipListProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StartLevels []*SkipListElementReference `protobuf:"bytes,1,rep,name=start_levels,json=startLevels,proto3" json:"start_levels,omitempty"`
	EndLevels   []*SkipListElementReference `protobuf:"bytes,2,rep,name=end_levels,json=endLevels,proto3" json:"end_levels,omitempty"`
	MaxNewLevel int32                       `protobuf:"varint,3,opt,name=max_new_level,json=maxNewLevel,proto3" json:"max_new_level,omitempty"`
	MaxLevel    int32                       `protobuf:"varint,4,opt,name=max_level,json=maxLevel,proto3" json:"max_level,omitempty"`
}

func (x *SkipListProto) Reset() {
	*x = SkipListProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_skiplist_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SkipListProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SkipListProto) ProtoMessage() {}

func (x *SkipListProto) ProtoReflect() protoreflect.Message {
	mi := &file_skiplist_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SkipListProto.ProtoReflect.Descriptor instead.
func (*SkipListProto) Descriptor() ([]byte, []int) {
	return file_skiplist_proto_rawDescGZIP(), []int{0}
}

func (x *SkipListProto) GetStartLevels() []*SkipListElementReference {
	if x != nil {
		return x.StartLevels
	}
	return nil
}

func (x *SkipListProto) GetEndLevels() []*SkipListElementReference {
	if x != nil {
		return x.EndLevels
	}
	return nil
}

func (x *SkipListProto) GetMaxNewLevel() int32 {
	if x != nil {
		return x.MaxNewLevel
	}
	return 0
}

func (x *SkipListProto) GetMaxLevel() int32 {
	if x != nil {
		return x.MaxLevel
	}
	return 0
}

type SkipListElementReference struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ElementPointer int64  `protobuf:"varint,1,opt,name=element_pointer,json=elementPointer,proto3" json:"element_pointer,omitempty"`
	Key            []byte `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *SkipListElementReference) Reset() {
	*x = SkipListElementReference{}
	if protoimpl.UnsafeEnabled {
		mi := &file_skiplist_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SkipListElementReference) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SkipListElementReference) ProtoMessage() {}

func (x *SkipListElementReference) ProtoReflect() protoreflect.Message {
	mi := &file_skiplist_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SkipListElementReference.ProtoReflect.Descriptor instead.
func (*SkipListElementReference) Descriptor() ([]byte, []int) {
	return file_skiplist_proto_rawDescGZIP(), []int{1}
}

func (x *SkipListElementReference) GetElementPointer() int64 {
	if x != nil {
		return x.ElementPointer
	}
	return 0
}

func (x *SkipListElementReference) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

type SkipListElement struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    int64                       `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Next  []*SkipListElementReference `protobuf:"bytes,2,rep,name=next,proto3" json:"next,omitempty"`
	Level int32                       `protobuf:"varint,3,opt,name=level,proto3" json:"level,omitempty"`
	Key   []byte                      `protobuf:"bytes,4,opt,name=key,proto3" json:"key,omitempty"`
	Value []byte                      `protobuf:"bytes,5,opt,name=value,proto3" json:"value,omitempty"`
	Prev  *SkipListElementReference   `protobuf:"bytes,6,opt,name=prev,proto3" json:"prev,omitempty"`
}

func (x *SkipListElement) Reset() {
	*x = SkipListElement{}
	if protoimpl.UnsafeEnabled {
		mi := &file_skiplist_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SkipListElement) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SkipListElement) ProtoMessage() {}

func (x *SkipListElement) ProtoReflect() protoreflect.Message {
	mi := &file_skiplist_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SkipListElement.ProtoReflect.Descriptor instead.
func (*SkipListElement) Descriptor() ([]byte, []int) {
	return file_skiplist_proto_rawDescGZIP(), []int{2}
}

func (x *SkipListElement) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *SkipListElement) GetNext() []*SkipListElementReference {
	if x != nil {
		return x.Next
	}
	return nil
}

func (x *SkipListElement) GetLevel() int32 {
	if x != nil {
		return x.Level
	}
	return 0
}

func (x *SkipListElement) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *SkipListElement) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

func (x *SkipListElement) GetPrev() *SkipListElementReference {
	if x != nil {
		return x.Prev
	}
	return nil
}

type NameBatchData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Names [][]byte `protobuf:"bytes,1,rep,name=names,proto3" json:"names,omitempty"`
}

func (x *NameBatchData) Reset() {
	*x = NameBatchData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_skiplist_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NameBatchData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NameBatchData) ProtoMessage() {}

func (x *NameBatchData) ProtoReflect() protoreflect.Message {
	mi := &file_skiplist_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NameBatchData.ProtoReflect.Descriptor instead.
func (*NameBatchData) Descriptor() ([]byte, []int) {
	return file_skiplist_proto_rawDescGZIP(), []int{3}
}

func (x *NameBatchData) GetNames() [][]byte {
	if x != nil {
		return x.Names
	}
	return nil
}

var File_skiplist_proto protoreflect.FileDescriptor

var file_skiplist_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x73, 0x6b, 0x69, 0x70, 0x6c, 0x69, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x08, 0x73, 0x6b, 0x69, 0x70, 0x6c, 0x69, 0x73, 0x74, 0x22, 0xda, 0x01, 0x0a, 0x0d, 0x53,
	0x6b, 0x69, 0x70, 0x4c, 0x69, 0x73, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x45, 0x0a, 0x0c,
	0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x22, 0x2e, 0x73, 0x6b, 0x69, 0x70, 0x6c, 0x69, 0x73, 0x74, 0x2e, 0x53, 0x6b,
	0x69, 0x70, 0x4c, 0x69, 0x73, 0x74, 0x45, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x66,
	0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x0b, 0x73, 0x74, 0x61, 0x72, 0x74, 0x4c, 0x65, 0x76,
	0x65, 0x6c, 0x73, 0x12, 0x41, 0x0a, 0x0a, 0x65, 0x6e, 0x64, 0x5f, 0x6c, 0x65, 0x76, 0x65, 0x6c,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x73, 0x6b, 0x69, 0x70, 0x6c, 0x69,
	0x73, 0x74, 0x2e, 0x53, 0x6b, 0x69, 0x70, 0x4c, 0x69, 0x73, 0x74, 0x45, 0x6c, 0x65, 0x6d, 0x65,
	0x6e, 0x74, 0x52, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x09, 0x65, 0x6e, 0x64,
	0x4c, 0x65, 0x76, 0x65, 0x6c, 0x73, 0x12, 0x22, 0x0a, 0x0d, 0x6d, 0x61, 0x78, 0x5f, 0x6e, 0x65,
	0x77, 0x5f, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x6d,
	0x61, 0x78, 0x4e, 0x65, 0x77, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x12, 0x1b, 0x0a, 0x09, 0x6d, 0x61,
	0x78, 0x5f, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x6d,
	0x61, 0x78, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x22, 0x55, 0x0a, 0x18, 0x53, 0x6b, 0x69, 0x70, 0x4c,
	0x69, 0x73, 0x74, 0x45, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x66, 0x65, 0x72, 0x65,
	0x6e, 0x63, 0x65, 0x12, 0x27, 0x0a, 0x0f, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x70,
	0x6f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0e, 0x65, 0x6c,
	0x65, 0x6d, 0x65, 0x6e, 0x74, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x22, 0xcf,
	0x01, 0x0a, 0x0f, 0x53, 0x6b, 0x69, 0x70, 0x4c, 0x69, 0x73, 0x74, 0x45, 0x6c, 0x65, 0x6d, 0x65,
	0x6e, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x36, 0x0a, 0x04, 0x6e, 0x65, 0x78, 0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x22, 0x2e, 0x73, 0x6b, 0x69, 0x70, 0x6c, 0x69, 0x73, 0x74, 0x2e, 0x53, 0x6b, 0x69, 0x70,
	0x4c, 0x69, 0x73, 0x74, 0x45, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x66, 0x65, 0x72,
	0x65, 0x6e, 0x63, 0x65, 0x52, 0x04, 0x6e, 0x65, 0x78, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x65,
	0x76, 0x65, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6c, 0x65, 0x76, 0x65, 0x6c,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x36, 0x0a, 0x04, 0x70, 0x72, 0x65, 0x76,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x73, 0x6b, 0x69, 0x70, 0x6c, 0x69, 0x73,
	0x74, 0x2e, 0x53, 0x6b, 0x69, 0x70, 0x4c, 0x69, 0x73, 0x74, 0x45, 0x6c, 0x65, 0x6d, 0x65, 0x6e,
	0x74, 0x52, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x04, 0x70, 0x72, 0x65, 0x76,
	0x22, 0x25, 0x0a, 0x0d, 0x4e, 0x61, 0x6d, 0x65, 0x42, 0x61, 0x74, 0x63, 0x68, 0x44, 0x61, 0x74,
	0x61, 0x12, 0x14, 0x0a, 0x05, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c,
	0x52, 0x05, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x42, 0x33, 0x5a, 0x31, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x65, 0x61, 0x77, 0x65, 0x65, 0x64, 0x66, 0x73, 0x2f,
	0x73, 0x65, 0x61, 0x77, 0x65, 0x65, 0x64, 0x66, 0x73, 0x2f, 0x77, 0x65, 0x65, 0x64, 0x2f, 0x75,
	0x74, 0x69, 0x6c, 0x2f, 0x73, 0x6b, 0x69, 0x70, 0x6c, 0x69, 0x73, 0x74, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_skiplist_proto_rawDescOnce sync.Once
	file_skiplist_proto_rawDescData = file_skiplist_proto_rawDesc
)

func file_skiplist_proto_rawDescGZIP() []byte {
	file_skiplist_proto_rawDescOnce.Do(func() {
		file_skiplist_proto_rawDescData = protoimpl.X.CompressGZIP(file_skiplist_proto_rawDescData)
	})
	return file_skiplist_proto_rawDescData
}

var file_skiplist_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_skiplist_proto_goTypes = []interface{}{
	(*SkipListProto)(nil),            // 0: skiplist.SkipListProto
	(*SkipListElementReference)(nil), // 1: skiplist.SkipListElementReference
	(*SkipListElement)(nil),          // 2: skiplist.SkipListElement
	(*NameBatchData)(nil),            // 3: skiplist.NameBatchData
}
var file_skiplist_proto_depIdxs = []int32{
	1, // 0: skiplist.SkipListProto.start_levels:type_name -> skiplist.SkipListElementReference
	1, // 1: skiplist.SkipListProto.end_levels:type_name -> skiplist.SkipListElementReference
	1, // 2: skiplist.SkipListElement.next:type_name -> skiplist.SkipListElementReference
	1, // 3: skiplist.SkipListElement.prev:type_name -> skiplist.SkipListElementReference
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_skiplist_proto_init() }
func file_skiplist_proto_init() {
	if File_skiplist_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_skiplist_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SkipListProto); i {
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
		file_skiplist_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SkipListElementReference); i {
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
		file_skiplist_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SkipListElement); i {
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
		file_skiplist_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NameBatchData); i {
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
			RawDescriptor: file_skiplist_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_skiplist_proto_goTypes,
		DependencyIndexes: file_skiplist_proto_depIdxs,
		MessageInfos:      file_skiplist_proto_msgTypes,
	}.Build()
	File_skiplist_proto = out.File
	file_skiplist_proto_rawDesc = nil
	file_skiplist_proto_goTypes = nil
	file_skiplist_proto_depIdxs = nil
}
