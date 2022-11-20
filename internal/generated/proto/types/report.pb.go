// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        (unknown)
// source: types/report.proto

package types

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Report struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Value:
	//	*Report_ByCategories_
	Value isReport_Value `protobuf_oneof:"value"`
}

func (x *Report) Reset() {
	*x = Report{}
	if protoimpl.UnsafeEnabled {
		mi := &file_types_report_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Report) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Report) ProtoMessage() {}

func (x *Report) ProtoReflect() protoreflect.Message {
	mi := &file_types_report_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Report.ProtoReflect.Descriptor instead.
func (*Report) Descriptor() ([]byte, []int) {
	return file_types_report_proto_rawDescGZIP(), []int{0}
}

func (m *Report) GetValue() isReport_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (x *Report) GetByCategories() *Report_ByCategories {
	if x, ok := x.GetValue().(*Report_ByCategories_); ok {
		return x.ByCategories
	}
	return nil
}

type isReport_Value interface {
	isReport_Value()
}

type Report_ByCategories_ struct {
	ByCategories *Report_ByCategories `protobuf:"bytes,10,opt,name=by_categories,json=byCategories,proto3,oneof"`
}

func (*Report_ByCategories_) isReport_Value() {}

type Report_ByCategories struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value map[string]*Decimal `protobuf:"bytes,1,rep,name=value,proto3" json:"value,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Report_ByCategories) Reset() {
	*x = Report_ByCategories{}
	if protoimpl.UnsafeEnabled {
		mi := &file_types_report_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Report_ByCategories) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Report_ByCategories) ProtoMessage() {}

func (x *Report_ByCategories) ProtoReflect() protoreflect.Message {
	mi := &file_types_report_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Report_ByCategories.ProtoReflect.Descriptor instead.
func (*Report_ByCategories) Descriptor() ([]byte, []int) {
	return file_types_report_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Report_ByCategories) GetValue() map[string]*Decimal {
	if x != nil {
		return x.Value
	}
	return nil
}

var File_types_report_proto protoreflect.FileDescriptor

var file_types_report_proto_rawDesc = []byte{
	0x0a, 0x12, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x74, 0x79, 0x70, 0x65, 0x73, 0x1a, 0x13, 0x74, 0x79, 0x70,
	0x65, 0x73, 0x2f, 0x64, 0x65, 0x63, 0x69, 0x6d, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xf2, 0x01, 0x0a, 0x06, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x41, 0x0a, 0x0d, 0x62,
	0x79, 0x5f, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x18, 0x0a, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x72,
	0x74, 0x2e, 0x42, 0x79, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x48, 0x00,
	0x52, 0x0c, 0x62, 0x79, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x1a, 0x95,
	0x01, 0x0a, 0x0c, 0x42, 0x79, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x12,
	0x3b, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x25,
	0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x42, 0x79,
	0x43, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x1a, 0x48, 0x0a, 0x0a,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x24, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x74, 0x79,
	0x70, 0x65, 0x73, 0x2e, 0x44, 0x65, 0x63, 0x69, 0x6d, 0x61, 0x6c, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x07, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x4a,
	0x04, 0x08, 0x01, 0x10, 0x0a, 0x42, 0x47, 0x5a, 0x45, 0x67, 0x69, 0x74, 0x6c, 0x61, 0x62, 0x2e,
	0x6f, 0x7a, 0x6f, 0x6e, 0x2e, 0x64, 0x65, 0x76, 0x2f, 0x6d, 0x72, 0x2e, 0x65, 0x73, 0x6b, 0x6f,
	0x76, 0x31, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x67, 0x72, 0x61, 0x6d, 0x2d, 0x62, 0x6f, 0x74, 0x2f,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74,
	0x65, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_types_report_proto_rawDescOnce sync.Once
	file_types_report_proto_rawDescData = file_types_report_proto_rawDesc
)

func file_types_report_proto_rawDescGZIP() []byte {
	file_types_report_proto_rawDescOnce.Do(func() {
		file_types_report_proto_rawDescData = protoimpl.X.CompressGZIP(file_types_report_proto_rawDescData)
	})
	return file_types_report_proto_rawDescData
}

var file_types_report_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_types_report_proto_goTypes = []interface{}{
	(*Report)(nil),              // 0: types.Report
	(*Report_ByCategories)(nil), // 1: types.Report.ByCategories
	nil,                         // 2: types.Report.ByCategories.ValueEntry
	(*Decimal)(nil),             // 3: types.Decimal
}
var file_types_report_proto_depIdxs = []int32{
	1, // 0: types.Report.by_categories:type_name -> types.Report.ByCategories
	2, // 1: types.Report.ByCategories.value:type_name -> types.Report.ByCategories.ValueEntry
	3, // 2: types.Report.ByCategories.ValueEntry.value:type_name -> types.Decimal
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_types_report_proto_init() }
func file_types_report_proto_init() {
	if File_types_report_proto != nil {
		return
	}
	file_types_decimal_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_types_report_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Report); i {
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
		file_types_report_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Report_ByCategories); i {
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
	file_types_report_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Report_ByCategories_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_types_report_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_types_report_proto_goTypes,
		DependencyIndexes: file_types_report_proto_depIdxs,
		MessageInfos:      file_types_report_proto_msgTypes,
	}.Build()
	File_types_report_proto = out.File
	file_types_report_proto_rawDesc = nil
	file_types_report_proto_goTypes = nil
	file_types_report_proto_depIdxs = nil
}
