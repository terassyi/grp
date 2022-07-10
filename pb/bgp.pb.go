// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.2
// source: bgp.proto

package pb

import (
	empty "github.com/golang/protobuf/ptypes/empty"
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

type ResultCode int32

const (
	ResultCode_Ok    ResultCode = 0
	ResultCode_Error ResultCode = 1
)

// Enum value maps for ResultCode.
var (
	ResultCode_name = map[int32]string{
		0: "Ok",
		1: "Error",
	}
	ResultCode_value = map[string]int32{
		"Ok":    0,
		"Error": 1,
	}
)

func (x ResultCode) Enum() *ResultCode {
	p := new(ResultCode)
	*p = x
	return p
}

func (x ResultCode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ResultCode) Descriptor() protoreflect.EnumDescriptor {
	return file_bgp_proto_enumTypes[0].Descriptor()
}

func (ResultCode) Type() protoreflect.EnumType {
	return &file_bgp_proto_enumTypes[0]
}

func (x ResultCode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ResultCode.Descriptor instead.
func (ResultCode) EnumDescriptor() ([]byte, []int) {
	return file_bgp_proto_rawDescGZIP(), []int{0}
}

type GrpResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result  ResultCode `protobuf:"varint,1,opt,name=result,proto3,enum=grp.ResultCode" json:"result,omitempty"`
	Message string     `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *GrpResult) Reset() {
	*x = GrpResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bgp_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GrpResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GrpResult) ProtoMessage() {}

func (x *GrpResult) ProtoReflect() protoreflect.Message {
	mi := &file_bgp_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GrpResult.ProtoReflect.Descriptor instead.
func (*GrpResult) Descriptor() ([]byte, []int) {
	return file_bgp_proto_rawDescGZIP(), []int{0}
}

func (x *GrpResult) GetResult() ResultCode {
	if x != nil {
		return x.Result
	}
	return ResultCode_Ok
}

func (x *GrpResult) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type HealthRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *HealthRequest) Reset() {
	*x = HealthRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bgp_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HealthRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthRequest) ProtoMessage() {}

func (x *HealthRequest) ProtoReflect() protoreflect.Message {
	mi := &file_bgp_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HealthRequest.ProtoReflect.Descriptor instead.
func (*HealthRequest) Descriptor() ([]byte, []int) {
	return file_bgp_proto_rawDescGZIP(), []int{1}
}

type ShowRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ShowRequest) Reset() {
	*x = ShowRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bgp_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShowRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShowRequest) ProtoMessage() {}

func (x *ShowRequest) ProtoReflect() protoreflect.Message {
	mi := &file_bgp_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShowRequest.ProtoReflect.Descriptor instead.
func (*ShowRequest) Descriptor() ([]byte, []int) {
	return file_bgp_proto_rawDescGZIP(), []int{2}
}

type ShowResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	As       int32  `protobuf:"varint,1,opt,name=as,proto3" json:"as,omitempty"`
	Port     int32  `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	RouterId string `protobuf:"bytes,3,opt,name=routerId,proto3" json:"routerId,omitempty"`
}

func (x *ShowResponse) Reset() {
	*x = ShowResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bgp_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShowResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShowResponse) ProtoMessage() {}

func (x *ShowResponse) ProtoReflect() protoreflect.Message {
	mi := &file_bgp_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShowResponse.ProtoReflect.Descriptor instead.
func (*ShowResponse) Descriptor() ([]byte, []int) {
	return file_bgp_proto_rawDescGZIP(), []int{3}
}

func (x *ShowResponse) GetAs() int32 {
	if x != nil {
		return x.As
	}
	return 0
}

func (x *ShowResponse) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *ShowResponse) GetRouterId() string {
	if x != nil {
		return x.RouterId
	}
	return ""
}

type NeighborInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	As       uint32 `protobuf:"varint,1,opt,name=as,proto3" json:"as,omitempty"`
	Address  string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	Port     uint32 `protobuf:"varint,3,opt,name=port,proto3" json:"port,omitempty"`
	RouterId string `protobuf:"bytes,4,opt,name=routerId,proto3" json:"routerId,omitempty"`
}

func (x *NeighborInfo) Reset() {
	*x = NeighborInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bgp_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NeighborInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NeighborInfo) ProtoMessage() {}

func (x *NeighborInfo) ProtoReflect() protoreflect.Message {
	mi := &file_bgp_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NeighborInfo.ProtoReflect.Descriptor instead.
func (*NeighborInfo) Descriptor() ([]byte, []int) {
	return file_bgp_proto_rawDescGZIP(), []int{4}
}

func (x *NeighborInfo) GetAs() uint32 {
	if x != nil {
		return x.As
	}
	return 0
}

func (x *NeighborInfo) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *NeighborInfo) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *NeighborInfo) GetRouterId() string {
	if x != nil {
		return x.RouterId
	}
	return ""
}

type GetNeighborRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	As          uint32  `protobuf:"varint,1,opt,name=as,proto3" json:"as,omitempty"`
	PeerAddress *string `protobuf:"bytes,2,opt,name=peerAddress,proto3,oneof" json:"peerAddress,omitempty"`
	RouterId    *string `protobuf:"bytes,3,opt,name=routerId,proto3,oneof" json:"routerId,omitempty"`
}

func (x *GetNeighborRequest) Reset() {
	*x = GetNeighborRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bgp_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetNeighborRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetNeighborRequest) ProtoMessage() {}

func (x *GetNeighborRequest) ProtoReflect() protoreflect.Message {
	mi := &file_bgp_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetNeighborRequest.ProtoReflect.Descriptor instead.
func (*GetNeighborRequest) Descriptor() ([]byte, []int) {
	return file_bgp_proto_rawDescGZIP(), []int{5}
}

func (x *GetNeighborRequest) GetAs() uint32 {
	if x != nil {
		return x.As
	}
	return 0
}

func (x *GetNeighborRequest) GetPeerAddress() string {
	if x != nil && x.PeerAddress != nil {
		return *x.PeerAddress
	}
	return ""
}

func (x *GetNeighborRequest) GetRouterId() string {
	if x != nil && x.RouterId != nil {
		return *x.RouterId
	}
	return ""
}

type GetNeighborResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Neighbor *NeighborInfo `protobuf:"bytes,1,opt,name=neighbor,proto3" json:"neighbor,omitempty"`
}

func (x *GetNeighborResponse) Reset() {
	*x = GetNeighborResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bgp_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetNeighborResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetNeighborResponse) ProtoMessage() {}

func (x *GetNeighborResponse) ProtoReflect() protoreflect.Message {
	mi := &file_bgp_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetNeighborResponse.ProtoReflect.Descriptor instead.
func (*GetNeighborResponse) Descriptor() ([]byte, []int) {
	return file_bgp_proto_rawDescGZIP(), []int{6}
}

func (x *GetNeighborResponse) GetNeighbor() *NeighborInfo {
	if x != nil {
		return x.Neighbor
	}
	return nil
}

type ListNeighborRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ListNeighborRequest) Reset() {
	*x = ListNeighborRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bgp_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListNeighborRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListNeighborRequest) ProtoMessage() {}

func (x *ListNeighborRequest) ProtoReflect() protoreflect.Message {
	mi := &file_bgp_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListNeighborRequest.ProtoReflect.Descriptor instead.
func (*ListNeighborRequest) Descriptor() ([]byte, []int) {
	return file_bgp_proto_rawDescGZIP(), []int{7}
}

type ListNeighborResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Neighbors []*NeighborInfo `protobuf:"bytes,1,rep,name=neighbors,proto3" json:"neighbors,omitempty"`
}

func (x *ListNeighborResponse) Reset() {
	*x = ListNeighborResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bgp_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListNeighborResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListNeighborResponse) ProtoMessage() {}

func (x *ListNeighborResponse) ProtoReflect() protoreflect.Message {
	mi := &file_bgp_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListNeighborResponse.ProtoReflect.Descriptor instead.
func (*ListNeighborResponse) Descriptor() ([]byte, []int) {
	return file_bgp_proto_rawDescGZIP(), []int{8}
}

func (x *ListNeighborResponse) GetNeighbors() []*NeighborInfo {
	if x != nil {
		return x.Neighbors
	}
	return nil
}

type SetASRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	As int32 `protobuf:"varint,1,opt,name=as,proto3" json:"as,omitempty"`
}

func (x *SetASRequest) Reset() {
	*x = SetASRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bgp_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetASRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetASRequest) ProtoMessage() {}

func (x *SetASRequest) ProtoReflect() protoreflect.Message {
	mi := &file_bgp_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetASRequest.ProtoReflect.Descriptor instead.
func (*SetASRequest) Descriptor() ([]byte, []int) {
	return file_bgp_proto_rawDescGZIP(), []int{9}
}

func (x *SetASRequest) GetAs() int32 {
	if x != nil {
		return x.As
	}
	return 0
}

type RemoteASRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Addr string `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	As   int32  `protobuf:"varint,2,opt,name=as,proto3" json:"as,omitempty"`
}

func (x *RemoteASRequest) Reset() {
	*x = RemoteASRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bgp_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoteASRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoteASRequest) ProtoMessage() {}

func (x *RemoteASRequest) ProtoReflect() protoreflect.Message {
	mi := &file_bgp_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoteASRequest.ProtoReflect.Descriptor instead.
func (*RemoteASRequest) Descriptor() ([]byte, []int) {
	return file_bgp_proto_rawDescGZIP(), []int{10}
}

func (x *RemoteASRequest) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

func (x *RemoteASRequest) GetAs() int32 {
	if x != nil {
		return x.As
	}
	return 0
}

type RouterIdRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RouterId string `protobuf:"bytes,1,opt,name=routerId,proto3" json:"routerId,omitempty"`
}

func (x *RouterIdRequest) Reset() {
	*x = RouterIdRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bgp_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RouterIdRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RouterIdRequest) ProtoMessage() {}

func (x *RouterIdRequest) ProtoReflect() protoreflect.Message {
	mi := &file_bgp_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RouterIdRequest.ProtoReflect.Descriptor instead.
func (*RouterIdRequest) Descriptor() ([]byte, []int) {
	return file_bgp_proto_rawDescGZIP(), []int{11}
}

func (x *RouterIdRequest) GetRouterId() string {
	if x != nil {
		return x.RouterId
	}
	return ""
}

type NetworkRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Networks []string `protobuf:"bytes,1,rep,name=networks,proto3" json:"networks,omitempty"`
}

func (x *NetworkRequest) Reset() {
	*x = NetworkRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bgp_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NetworkRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NetworkRequest) ProtoMessage() {}

func (x *NetworkRequest) ProtoReflect() protoreflect.Message {
	mi := &file_bgp_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NetworkRequest.ProtoReflect.Descriptor instead.
func (*NetworkRequest) Descriptor() ([]byte, []int) {
	return file_bgp_proto_rawDescGZIP(), []int{12}
}

func (x *NetworkRequest) GetNetworks() []string {
	if x != nil {
		return x.Networks
	}
	return nil
}

var File_bgp_proto protoreflect.FileDescriptor

var file_bgp_proto_rawDesc = []byte{
	0x0a, 0x09, 0x62, 0x67, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x67, 0x72, 0x70,
	0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4e, 0x0a,
	0x09, 0x47, 0x72, 0x70, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x27, 0x0a, 0x06, 0x72, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0f, 0x2e, 0x67, 0x72, 0x70,
	0x2e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x06, 0x72, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x0f, 0x0a,
	0x0d, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x0d,
	0x0a, 0x0b, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x4e, 0x0a,
	0x0c, 0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a,
	0x02, 0x61, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x61, 0x73, 0x12, 0x12, 0x0a,
	0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x70, 0x6f, 0x72,
	0x74, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x49, 0x64, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x49, 0x64, 0x22, 0x68, 0x0a,
	0x0c, 0x4e, 0x65, 0x69, 0x67, 0x68, 0x62, 0x6f, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x0e, 0x0a,
	0x02, 0x61, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x02, 0x61, 0x73, 0x12, 0x18, 0x0a,
	0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x72,
	0x6f, 0x75, 0x74, 0x65, 0x72, 0x49, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72,
	0x6f, 0x75, 0x74, 0x65, 0x72, 0x49, 0x64, 0x22, 0x89, 0x01, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x4e,
	0x65, 0x69, 0x67, 0x68, 0x62, 0x6f, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e,
	0x0a, 0x02, 0x61, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x02, 0x61, 0x73, 0x12, 0x25,
	0x0a, 0x0b, 0x70, 0x65, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0b, 0x70, 0x65, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x88, 0x01, 0x01, 0x12, 0x1f, 0x0a, 0x08, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x49,
	0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52, 0x08, 0x72, 0x6f, 0x75, 0x74, 0x65,
	0x72, 0x49, 0x64, 0x88, 0x01, 0x01, 0x42, 0x0e, 0x0a, 0x0c, 0x5f, 0x70, 0x65, 0x65, 0x72, 0x41,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x42, 0x0b, 0x0a, 0x09, 0x5f, 0x72, 0x6f, 0x75, 0x74, 0x65,
	0x72, 0x49, 0x64, 0x22, 0x44, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x4e, 0x65, 0x69, 0x67, 0x68, 0x62,
	0x6f, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2d, 0x0a, 0x08, 0x6e, 0x65,
	0x69, 0x67, 0x68, 0x62, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x67,
	0x72, 0x70, 0x2e, 0x4e, 0x65, 0x69, 0x67, 0x68, 0x62, 0x6f, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52,
	0x08, 0x6e, 0x65, 0x69, 0x67, 0x68, 0x62, 0x6f, 0x72, 0x22, 0x15, 0x0a, 0x13, 0x4c, 0x69, 0x73,
	0x74, 0x4e, 0x65, 0x69, 0x67, 0x68, 0x62, 0x6f, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x22, 0x47, 0x0a, 0x14, 0x4c, 0x69, 0x73, 0x74, 0x4e, 0x65, 0x69, 0x67, 0x68, 0x62, 0x6f, 0x72,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2f, 0x0a, 0x09, 0x6e, 0x65, 0x69, 0x67,
	0x68, 0x62, 0x6f, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x67, 0x72,
	0x70, 0x2e, 0x4e, 0x65, 0x69, 0x67, 0x68, 0x62, 0x6f, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x09,
	0x6e, 0x65, 0x69, 0x67, 0x68, 0x62, 0x6f, 0x72, 0x73, 0x22, 0x1e, 0x0a, 0x0c, 0x53, 0x65, 0x74,
	0x41, 0x53, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x61, 0x73, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x61, 0x73, 0x22, 0x35, 0x0a, 0x0f, 0x52, 0x65, 0x6d,
	0x6f, 0x74, 0x65, 0x41, 0x53, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04,
	0x61, 0x64, 0x64, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x61, 0x64, 0x64, 0x72,
	0x12, 0x0e, 0x0a, 0x02, 0x61, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x61, 0x73,
	0x22, 0x2d, 0x0a, 0x0f, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x49, 0x64, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x49, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x49, 0x64, 0x22,
	0x2c, 0x0a, 0x0e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x1a, 0x0a, 0x08, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x08, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x2a, 0x1f, 0x0a,
	0x0a, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x06, 0x0a, 0x02, 0x4f,
	0x6b, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x10, 0x01, 0x32, 0xd2,
	0x03, 0x0a, 0x06, 0x42, 0x67, 0x70, 0x41, 0x70, 0x69, 0x12, 0x34, 0x0a, 0x06, 0x48, 0x65, 0x61,
	0x6c, 0x74, 0x68, 0x12, 0x12, 0x2e, 0x67, 0x72, 0x70, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12,
	0x2b, 0x0a, 0x04, 0x53, 0x68, 0x6f, 0x77, 0x12, 0x10, 0x2e, 0x67, 0x72, 0x70, 0x2e, 0x53, 0x68,
	0x6f, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x11, 0x2e, 0x67, 0x72, 0x70, 0x2e,
	0x53, 0x68, 0x6f, 0x77, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x40, 0x0a, 0x0b,
	0x47, 0x65, 0x74, 0x4e, 0x65, 0x69, 0x67, 0x68, 0x62, 0x6f, 0x72, 0x12, 0x17, 0x2e, 0x67, 0x72,
	0x70, 0x2e, 0x47, 0x65, 0x74, 0x4e, 0x65, 0x69, 0x67, 0x68, 0x62, 0x6f, 0x72, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x67, 0x72, 0x70, 0x2e, 0x47, 0x65, 0x74, 0x4e, 0x65,
	0x69, 0x67, 0x68, 0x62, 0x6f, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x43,
	0x0a, 0x0c, 0x4c, 0x69, 0x73, 0x74, 0x4e, 0x65, 0x69, 0x67, 0x68, 0x62, 0x6f, 0x72, 0x12, 0x18,
	0x2e, 0x67, 0x72, 0x70, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x4e, 0x65, 0x69, 0x67, 0x68, 0x62, 0x6f,
	0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x67, 0x72, 0x70, 0x2e, 0x4c,
	0x69, 0x73, 0x74, 0x4e, 0x65, 0x69, 0x67, 0x68, 0x62, 0x6f, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x32, 0x0a, 0x05, 0x53, 0x65, 0x74, 0x41, 0x53, 0x12, 0x11, 0x2e, 0x67,
	0x72, 0x70, 0x2e, 0x53, 0x65, 0x74, 0x41, 0x53, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x38, 0x0a, 0x08, 0x52, 0x6f, 0x75, 0x74, 0x65,
	0x72, 0x49, 0x64, 0x12, 0x14, 0x2e, 0x67, 0x72, 0x70, 0x2e, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x72,
	0x49, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x12, 0x38, 0x0a, 0x08, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x41, 0x53, 0x12, 0x14, 0x2e,
	0x67, 0x72, 0x70, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x41, 0x53, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x36, 0x0a, 0x07, 0x4e,
	0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x12, 0x13, 0x2e, 0x67, 0x72, 0x70, 0x2e, 0x4e, 0x65, 0x74,
	0x77, 0x6f, 0x72, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x42, 0x1c, 0x5a, 0x1a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x74, 0x65, 0x72, 0x61, 0x73, 0x73, 0x79, 0x69, 0x2f, 0x67, 0x72, 0x70, 0x2f, 0x70,
	0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_bgp_proto_rawDescOnce sync.Once
	file_bgp_proto_rawDescData = file_bgp_proto_rawDesc
)

func file_bgp_proto_rawDescGZIP() []byte {
	file_bgp_proto_rawDescOnce.Do(func() {
		file_bgp_proto_rawDescData = protoimpl.X.CompressGZIP(file_bgp_proto_rawDescData)
	})
	return file_bgp_proto_rawDescData
}

var file_bgp_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_bgp_proto_msgTypes = make([]protoimpl.MessageInfo, 13)
var file_bgp_proto_goTypes = []interface{}{
	(ResultCode)(0),              // 0: grp.ResultCode
	(*GrpResult)(nil),            // 1: grp.GrpResult
	(*HealthRequest)(nil),        // 2: grp.HealthRequest
	(*ShowRequest)(nil),          // 3: grp.ShowRequest
	(*ShowResponse)(nil),         // 4: grp.ShowResponse
	(*NeighborInfo)(nil),         // 5: grp.NeighborInfo
	(*GetNeighborRequest)(nil),   // 6: grp.GetNeighborRequest
	(*GetNeighborResponse)(nil),  // 7: grp.GetNeighborResponse
	(*ListNeighborRequest)(nil),  // 8: grp.ListNeighborRequest
	(*ListNeighborResponse)(nil), // 9: grp.ListNeighborResponse
	(*SetASRequest)(nil),         // 10: grp.SetASRequest
	(*RemoteASRequest)(nil),      // 11: grp.RemoteASRequest
	(*RouterIdRequest)(nil),      // 12: grp.RouterIdRequest
	(*NetworkRequest)(nil),       // 13: grp.NetworkRequest
	(*empty.Empty)(nil),          // 14: google.protobuf.Empty
}
var file_bgp_proto_depIdxs = []int32{
	0,  // 0: grp.GrpResult.result:type_name -> grp.ResultCode
	5,  // 1: grp.GetNeighborResponse.neighbor:type_name -> grp.NeighborInfo
	5,  // 2: grp.ListNeighborResponse.neighbors:type_name -> grp.NeighborInfo
	2,  // 3: grp.BgpApi.Health:input_type -> grp.HealthRequest
	3,  // 4: grp.BgpApi.Show:input_type -> grp.ShowRequest
	6,  // 5: grp.BgpApi.GetNeighbor:input_type -> grp.GetNeighborRequest
	8,  // 6: grp.BgpApi.ListNeighbor:input_type -> grp.ListNeighborRequest
	10, // 7: grp.BgpApi.SetAS:input_type -> grp.SetASRequest
	12, // 8: grp.BgpApi.RouterId:input_type -> grp.RouterIdRequest
	11, // 9: grp.BgpApi.RemoteAS:input_type -> grp.RemoteASRequest
	13, // 10: grp.BgpApi.Network:input_type -> grp.NetworkRequest
	14, // 11: grp.BgpApi.Health:output_type -> google.protobuf.Empty
	4,  // 12: grp.BgpApi.Show:output_type -> grp.ShowResponse
	7,  // 13: grp.BgpApi.GetNeighbor:output_type -> grp.GetNeighborResponse
	9,  // 14: grp.BgpApi.ListNeighbor:output_type -> grp.ListNeighborResponse
	14, // 15: grp.BgpApi.SetAS:output_type -> google.protobuf.Empty
	14, // 16: grp.BgpApi.RouterId:output_type -> google.protobuf.Empty
	14, // 17: grp.BgpApi.RemoteAS:output_type -> google.protobuf.Empty
	14, // 18: grp.BgpApi.Network:output_type -> google.protobuf.Empty
	11, // [11:19] is the sub-list for method output_type
	3,  // [3:11] is the sub-list for method input_type
	3,  // [3:3] is the sub-list for extension type_name
	3,  // [3:3] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_bgp_proto_init() }
func file_bgp_proto_init() {
	if File_bgp_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_bgp_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GrpResult); i {
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
		file_bgp_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HealthRequest); i {
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
		file_bgp_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShowRequest); i {
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
		file_bgp_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShowResponse); i {
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
		file_bgp_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NeighborInfo); i {
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
		file_bgp_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetNeighborRequest); i {
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
		file_bgp_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetNeighborResponse); i {
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
		file_bgp_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListNeighborRequest); i {
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
		file_bgp_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListNeighborResponse); i {
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
		file_bgp_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetASRequest); i {
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
		file_bgp_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RemoteASRequest); i {
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
		file_bgp_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RouterIdRequest); i {
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
		file_bgp_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NetworkRequest); i {
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
	file_bgp_proto_msgTypes[5].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_bgp_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   13,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_bgp_proto_goTypes,
		DependencyIndexes: file_bgp_proto_depIdxs,
		EnumInfos:         file_bgp_proto_enumTypes,
		MessageInfos:      file_bgp_proto_msgTypes,
	}.Build()
	File_bgp_proto = out.File
	file_bgp_proto_rawDesc = nil
	file_bgp_proto_goTypes = nil
	file_bgp_proto_depIdxs = nil
}
