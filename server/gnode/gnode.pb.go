// Code generated by protoc-gen-go.
// source: server/gnode/gnode.proto
// DO NOT EDIT!

/*
Package gnode is a generated protocol buffer package.

It is generated from these files:
	server/gnode/gnode.proto

It has these top-level messages:
	AntMes
	AntRet
	PingRet
	AskConnectionRequest
	StoreFileRequest
	StoreFileRet
	EmptyRet
	RetrieveFileRequest
	RetrieveFileRet
*/
package gnode

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type AntMes struct {
	Id           string   `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Origin       string   `protobuf:"bytes,2,opt,name=origin" json:"origin,omitempty"`
	FromClient   string   `protobuf:"bytes,3,opt,name=from_client,json=fromClient" json:"from_client,omitempty"`
	Target       string   `protobuf:"bytes,4,opt,name=target" json:"target,omitempty"`
	IsAnswer     bool     `protobuf:"varint,5,opt,name=isAnswer" json:"isAnswer,omitempty"`
	ReturnAnswer bool     `protobuf:"varint,6,opt,name=return_answer,json=returnAnswer" json:"return_answer,omitempty"`
	IsPathWriter bool     `protobuf:"varint,7,opt,name=is_path_writer,json=isPathWriter" json:"is_path_writer,omitempty"`
	OriginId     string   `protobuf:"bytes,8,opt,name=origin_id,json=originId" json:"origin_id,omitempty"`
	Path         []string `protobuf:"bytes,9,rep,name=path" json:"path,omitempty"`
	PathIndex    int32    `protobuf:"varint,10,opt,name=path_index,json=pathIndex" json:"path_index,omitempty"`
	Function     string   `protobuf:"bytes,11,opt,name=function" json:"function,omitempty"`
	Args         []string `protobuf:"bytes,12,rep,name=args" json:"args,omitempty"`
	TransferId   string   `protobuf:"bytes,13,opt,name=transfer_id,json=transferId" json:"transfer_id,omitempty"`
	Order        int64    `protobuf:"varint,14,opt,name=order" json:"order,omitempty"`
	NbBlockTotal int64    `protobuf:"varint,15,opt,name=nb_block_total,json=nbBlockTotal" json:"nb_block_total,omitempty"`
	NbBlock      int64    `protobuf:"varint,16,opt,name=nb_block,json=nbBlock" json:"nb_block,omitempty"`
	Size         int64    `protobuf:"varint,17,opt,name=size" json:"size,omitempty"`
	TargetedPath string   `protobuf:"bytes,18,opt,name=targeted_path,json=targetedPath" json:"targeted_path,omitempty"`
	NoBlocking   bool     `protobuf:"varint,19,opt,name=no_blocking,json=noBlocking" json:"no_blocking,omitempty"`
	Debug        bool     `protobuf:"varint,20,opt,name=debug" json:"debug,omitempty"`
	AnswerWait   bool     `protobuf:"varint,21,opt,name=answer_wait,json=answerWait" json:"answer_wait,omitempty"`
	ErrorMes     string   `protobuf:"bytes,22,opt,name=errorMes" json:"errorMes,omitempty"`
	NbThread     int32    `protobuf:"varint,23,opt,name=nb_thread,json=nbThread" json:"nb_thread,omitempty"`
	Thread       int32    `protobuf:"varint,24,opt,name=thread" json:"thread,omitempty"`
	Duplicate    int32    `protobuf:"varint,25,opt,name=duplicate" json:"duplicate,omitempty"`
	Eof          bool     `protobuf:"varint,26,opt,name=eof" json:"eof,omitempty"`
	UserName     string   `protobuf:"bytes,27,opt,name=user_name,json=userName" json:"user_name,omitempty"`
	UserToken    string   `protobuf:"bytes,28,opt,name=user_token,json=userToken" json:"user_token,omitempty"`
	Data         []byte   `protobuf:"bytes,29,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *AntMes) Reset()                    { *m = AntMes{} }
func (m *AntMes) String() string            { return proto.CompactTextString(m) }
func (*AntMes) ProtoMessage()               {}
func (*AntMes) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *AntMes) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *AntMes) GetOrigin() string {
	if m != nil {
		return m.Origin
	}
	return ""
}

func (m *AntMes) GetFromClient() string {
	if m != nil {
		return m.FromClient
	}
	return ""
}

func (m *AntMes) GetTarget() string {
	if m != nil {
		return m.Target
	}
	return ""
}

func (m *AntMes) GetIsAnswer() bool {
	if m != nil {
		return m.IsAnswer
	}
	return false
}

func (m *AntMes) GetReturnAnswer() bool {
	if m != nil {
		return m.ReturnAnswer
	}
	return false
}

func (m *AntMes) GetIsPathWriter() bool {
	if m != nil {
		return m.IsPathWriter
	}
	return false
}

func (m *AntMes) GetOriginId() string {
	if m != nil {
		return m.OriginId
	}
	return ""
}

func (m *AntMes) GetPath() []string {
	if m != nil {
		return m.Path
	}
	return nil
}

func (m *AntMes) GetPathIndex() int32 {
	if m != nil {
		return m.PathIndex
	}
	return 0
}

func (m *AntMes) GetFunction() string {
	if m != nil {
		return m.Function
	}
	return ""
}

func (m *AntMes) GetArgs() []string {
	if m != nil {
		return m.Args
	}
	return nil
}

func (m *AntMes) GetTransferId() string {
	if m != nil {
		return m.TransferId
	}
	return ""
}

func (m *AntMes) GetOrder() int64 {
	if m != nil {
		return m.Order
	}
	return 0
}

func (m *AntMes) GetNbBlockTotal() int64 {
	if m != nil {
		return m.NbBlockTotal
	}
	return 0
}

func (m *AntMes) GetNbBlock() int64 {
	if m != nil {
		return m.NbBlock
	}
	return 0
}

func (m *AntMes) GetSize() int64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *AntMes) GetTargetedPath() string {
	if m != nil {
		return m.TargetedPath
	}
	return ""
}

func (m *AntMes) GetNoBlocking() bool {
	if m != nil {
		return m.NoBlocking
	}
	return false
}

func (m *AntMes) GetDebug() bool {
	if m != nil {
		return m.Debug
	}
	return false
}

func (m *AntMes) GetAnswerWait() bool {
	if m != nil {
		return m.AnswerWait
	}
	return false
}

func (m *AntMes) GetErrorMes() string {
	if m != nil {
		return m.ErrorMes
	}
	return ""
}

func (m *AntMes) GetNbThread() int32 {
	if m != nil {
		return m.NbThread
	}
	return 0
}

func (m *AntMes) GetThread() int32 {
	if m != nil {
		return m.Thread
	}
	return 0
}

func (m *AntMes) GetDuplicate() int32 {
	if m != nil {
		return m.Duplicate
	}
	return 0
}

func (m *AntMes) GetEof() bool {
	if m != nil {
		return m.Eof
	}
	return false
}

func (m *AntMes) GetUserName() string {
	if m != nil {
		return m.UserName
	}
	return ""
}

func (m *AntMes) GetUserToken() string {
	if m != nil {
		return m.UserToken
	}
	return ""
}

func (m *AntMes) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type AntRet struct {
	Ack bool   `protobuf:"varint,1,opt,name=ack" json:"ack,omitempty"`
	Id  string `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
}

func (m *AntRet) Reset()                    { *m = AntRet{} }
func (m *AntRet) String() string            { return proto.CompactTextString(m) }
func (*AntRet) ProtoMessage()               {}
func (*AntRet) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *AntRet) GetAck() bool {
	if m != nil {
		return m.Ack
	}
	return false
}

func (m *AntRet) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type PingRet struct {
	Host         string `protobuf:"bytes,1,opt,name=host" json:"host,omitempty"`
	Name         string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	NbNode       int32  `protobuf:"varint,3,opt,name=nb_node,json=nbNode" json:"nb_node,omitempty"`
	NbDuplicate  int32  `protobuf:"varint,4,opt,name=nb_duplicate,json=nbDuplicate" json:"nb_duplicate,omitempty"`
	ClientNumber int32  `protobuf:"varint,5,opt,name=client_number,json=clientNumber" json:"client_number,omitempty"`
}

func (m *PingRet) Reset()                    { *m = PingRet{} }
func (m *PingRet) String() string            { return proto.CompactTextString(m) }
func (*PingRet) ProtoMessage()               {}
func (*PingRet) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *PingRet) GetHost() string {
	if m != nil {
		return m.Host
	}
	return ""
}

func (m *PingRet) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *PingRet) GetNbNode() int32 {
	if m != nil {
		return m.NbNode
	}
	return 0
}

func (m *PingRet) GetNbDuplicate() int32 {
	if m != nil {
		return m.NbDuplicate
	}
	return 0
}

func (m *PingRet) GetClientNumber() int32 {
	if m != nil {
		return m.ClientNumber
	}
	return 0
}

type AskConnectionRequest struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Host string `protobuf:"bytes,2,opt,name=host" json:"host,omitempty"`
	Ip   string `protobuf:"bytes,3,opt,name=ip" json:"ip,omitempty"`
}

func (m *AskConnectionRequest) Reset()                    { *m = AskConnectionRequest{} }
func (m *AskConnectionRequest) String() string            { return proto.CompactTextString(m) }
func (*AskConnectionRequest) ProtoMessage()               {}
func (*AskConnectionRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *AskConnectionRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *AskConnectionRequest) GetHost() string {
	if m != nil {
		return m.Host
	}
	return ""
}

func (m *AskConnectionRequest) GetIp() string {
	if m != nil {
		return m.Ip
	}
	return ""
}

type StoreFileRequest struct {
	ClientId     string   `protobuf:"bytes,1,opt,name=client_id,json=clientId" json:"client_id,omitempty"`
	Name         string   `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Path         string   `protobuf:"bytes,3,opt,name=path" json:"path,omitempty"`
	NbBlockTotal int64    `protobuf:"varint,4,opt,name=nb_block_total,json=nbBlockTotal" json:"nb_block_total,omitempty"`
	NbBlock      int64    `protobuf:"varint,5,opt,name=nb_block,json=nbBlock" json:"nb_block,omitempty"`
	BlockSize    int64    `protobuf:"varint,6,opt,name=blockSize" json:"blockSize,omitempty"`
	TransferId   string   `protobuf:"bytes,7,opt,name=transfer_id,json=transferId" json:"transfer_id,omitempty"`
	Key          string   `protobuf:"bytes,8,opt,name=key" json:"key,omitempty"`
	UserName     string   `protobuf:"bytes,9,opt,name=user_name,json=userName" json:"user_name,omitempty"`
	UserToken    string   `protobuf:"bytes,10,opt,name=user_token,json=userToken" json:"user_token,omitempty"`
	Metadata     []string `protobuf:"bytes,11,rep,name=metadata" json:"metadata,omitempty"`
}

func (m *StoreFileRequest) Reset()                    { *m = StoreFileRequest{} }
func (m *StoreFileRequest) String() string            { return proto.CompactTextString(m) }
func (*StoreFileRequest) ProtoMessage()               {}
func (*StoreFileRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *StoreFileRequest) GetClientId() string {
	if m != nil {
		return m.ClientId
	}
	return ""
}

func (m *StoreFileRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *StoreFileRequest) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *StoreFileRequest) GetNbBlockTotal() int64 {
	if m != nil {
		return m.NbBlockTotal
	}
	return 0
}

func (m *StoreFileRequest) GetNbBlock() int64 {
	if m != nil {
		return m.NbBlock
	}
	return 0
}

func (m *StoreFileRequest) GetBlockSize() int64 {
	if m != nil {
		return m.BlockSize
	}
	return 0
}

func (m *StoreFileRequest) GetTransferId() string {
	if m != nil {
		return m.TransferId
	}
	return ""
}

func (m *StoreFileRequest) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *StoreFileRequest) GetUserName() string {
	if m != nil {
		return m.UserName
	}
	return ""
}

func (m *StoreFileRequest) GetUserToken() string {
	if m != nil {
		return m.UserToken
	}
	return ""
}

func (m *StoreFileRequest) GetMetadata() []string {
	if m != nil {
		return m.Metadata
	}
	return nil
}

type StoreFileRet struct {
	TransferId string `protobuf:"bytes,1,opt,name=transfer_id,json=transferId" json:"transfer_id,omitempty"`
	BlockSize  int64  `protobuf:"varint,2,opt,name=block_size,json=blockSize" json:"block_size,omitempty"`
}

func (m *StoreFileRet) Reset()                    { *m = StoreFileRet{} }
func (m *StoreFileRet) String() string            { return proto.CompactTextString(m) }
func (*StoreFileRet) ProtoMessage()               {}
func (*StoreFileRet) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *StoreFileRet) GetTransferId() string {
	if m != nil {
		return m.TransferId
	}
	return ""
}

func (m *StoreFileRet) GetBlockSize() int64 {
	if m != nil {
		return m.BlockSize
	}
	return 0
}

type EmptyRet struct {
}

func (m *EmptyRet) Reset()                    { *m = EmptyRet{} }
func (m *EmptyRet) String() string            { return proto.CompactTextString(m) }
func (*EmptyRet) ProtoMessage()               {}
func (*EmptyRet) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type RetrieveFileRequest struct {
	ClientId  string `protobuf:"bytes,1,opt,name=client_id,json=clientId" json:"client_id,omitempty"`
	Name      string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	NbThread  int32  `protobuf:"varint,3,opt,name=nbThread" json:"nbThread,omitempty"`
	Thread    int32  `protobuf:"varint,4,opt,name=thread" json:"thread,omitempty"`
	Duplicate int32  `protobuf:"varint,5,opt,name=duplicate" json:"duplicate,omitempty"`
	UserName  string `protobuf:"bytes,6,opt,name=user_name,json=userName" json:"user_name,omitempty"`
	UserToken string `protobuf:"bytes,7,opt,name=user_token,json=userToken" json:"user_token,omitempty"`
	BlockList string `protobuf:"bytes,8,opt,name=block_list,json=blockList" json:"block_list,omitempty"`
}

func (m *RetrieveFileRequest) Reset()                    { *m = RetrieveFileRequest{} }
func (m *RetrieveFileRequest) String() string            { return proto.CompactTextString(m) }
func (*RetrieveFileRequest) ProtoMessage()               {}
func (*RetrieveFileRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *RetrieveFileRequest) GetClientId() string {
	if m != nil {
		return m.ClientId
	}
	return ""
}

func (m *RetrieveFileRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *RetrieveFileRequest) GetNbThread() int32 {
	if m != nil {
		return m.NbThread
	}
	return 0
}

func (m *RetrieveFileRequest) GetThread() int32 {
	if m != nil {
		return m.Thread
	}
	return 0
}

func (m *RetrieveFileRequest) GetDuplicate() int32 {
	if m != nil {
		return m.Duplicate
	}
	return 0
}

func (m *RetrieveFileRequest) GetUserName() string {
	if m != nil {
		return m.UserName
	}
	return ""
}

func (m *RetrieveFileRequest) GetUserToken() string {
	if m != nil {
		return m.UserToken
	}
	return ""
}

func (m *RetrieveFileRequest) GetBlockList() string {
	if m != nil {
		return m.BlockList
	}
	return ""
}

type RetrieveFileRet struct {
	TransferId string `protobuf:"bytes,1,opt,name=transfer_id,json=transferId" json:"transfer_id,omitempty"`
	BlockSize  int64  `protobuf:"varint,2,opt,name=block_size,json=blockSize" json:"block_size,omitempty"`
}

func (m *RetrieveFileRet) Reset()                    { *m = RetrieveFileRet{} }
func (m *RetrieveFileRet) String() string            { return proto.CompactTextString(m) }
func (*RetrieveFileRet) ProtoMessage()               {}
func (*RetrieveFileRet) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *RetrieveFileRet) GetTransferId() string {
	if m != nil {
		return m.TransferId
	}
	return ""
}

func (m *RetrieveFileRet) GetBlockSize() int64 {
	if m != nil {
		return m.BlockSize
	}
	return 0
}

func init() {
	proto.RegisterType((*AntMes)(nil), "gnode.AntMes")
	proto.RegisterType((*AntRet)(nil), "gnode.AntRet")
	proto.RegisterType((*PingRet)(nil), "gnode.PingRet")
	proto.RegisterType((*AskConnectionRequest)(nil), "gnode.AskConnectionRequest")
	proto.RegisterType((*StoreFileRequest)(nil), "gnode.StoreFileRequest")
	proto.RegisterType((*StoreFileRet)(nil), "gnode.StoreFileRet")
	proto.RegisterType((*EmptyRet)(nil), "gnode.EmptyRet")
	proto.RegisterType((*RetrieveFileRequest)(nil), "gnode.RetrieveFileRequest")
	proto.RegisterType((*RetrieveFileRet)(nil), "gnode.RetrieveFileRet")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for GNodeService service

type GNodeServiceClient interface {
	ExecuteFunction(ctx context.Context, in *AntMes, opts ...grpc.CallOption) (*AntRet, error)
	GetClientStream(ctx context.Context, opts ...grpc.CallOption) (GNodeService_GetClientStreamClient, error)
	AskConnection(ctx context.Context, in *AskConnectionRequest, opts ...grpc.CallOption) (*PingRet, error)
	Ping(ctx context.Context, in *AntMes, opts ...grpc.CallOption) (*PingRet, error)
	StoreFile(ctx context.Context, in *StoreFileRequest, opts ...grpc.CallOption) (*StoreFileRet, error)
	RetrieveFile(ctx context.Context, in *RetrieveFileRequest, opts ...grpc.CallOption) (*RetrieveFileRet, error)
}

type gNodeServiceClient struct {
	cc *grpc.ClientConn
}

func NewGNodeServiceClient(cc *grpc.ClientConn) GNodeServiceClient {
	return &gNodeServiceClient{cc}
}

func (c *gNodeServiceClient) ExecuteFunction(ctx context.Context, in *AntMes, opts ...grpc.CallOption) (*AntRet, error) {
	out := new(AntRet)
	err := grpc.Invoke(ctx, "/gnode.GNodeService/ExecuteFunction", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gNodeServiceClient) GetClientStream(ctx context.Context, opts ...grpc.CallOption) (GNodeService_GetClientStreamClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_GNodeService_serviceDesc.Streams[0], c.cc, "/gnode.GNodeService/GetClientStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &gNodeServiceGetClientStreamClient{stream}
	return x, nil
}

type GNodeService_GetClientStreamClient interface {
	Send(*AntMes) error
	Recv() (*AntMes, error)
	grpc.ClientStream
}

type gNodeServiceGetClientStreamClient struct {
	grpc.ClientStream
}

func (x *gNodeServiceGetClientStreamClient) Send(m *AntMes) error {
	return x.ClientStream.SendMsg(m)
}

func (x *gNodeServiceGetClientStreamClient) Recv() (*AntMes, error) {
	m := new(AntMes)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *gNodeServiceClient) AskConnection(ctx context.Context, in *AskConnectionRequest, opts ...grpc.CallOption) (*PingRet, error) {
	out := new(PingRet)
	err := grpc.Invoke(ctx, "/gnode.GNodeService/AskConnection", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gNodeServiceClient) Ping(ctx context.Context, in *AntMes, opts ...grpc.CallOption) (*PingRet, error) {
	out := new(PingRet)
	err := grpc.Invoke(ctx, "/gnode.GNodeService/Ping", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gNodeServiceClient) StoreFile(ctx context.Context, in *StoreFileRequest, opts ...grpc.CallOption) (*StoreFileRet, error) {
	out := new(StoreFileRet)
	err := grpc.Invoke(ctx, "/gnode.GNodeService/StoreFile", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gNodeServiceClient) RetrieveFile(ctx context.Context, in *RetrieveFileRequest, opts ...grpc.CallOption) (*RetrieveFileRet, error) {
	out := new(RetrieveFileRet)
	err := grpc.Invoke(ctx, "/gnode.GNodeService/RetrieveFile", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for GNodeService service

type GNodeServiceServer interface {
	ExecuteFunction(context.Context, *AntMes) (*AntRet, error)
	GetClientStream(GNodeService_GetClientStreamServer) error
	AskConnection(context.Context, *AskConnectionRequest) (*PingRet, error)
	Ping(context.Context, *AntMes) (*PingRet, error)
	StoreFile(context.Context, *StoreFileRequest) (*StoreFileRet, error)
	RetrieveFile(context.Context, *RetrieveFileRequest) (*RetrieveFileRet, error)
}

func RegisterGNodeServiceServer(s *grpc.Server, srv GNodeServiceServer) {
	s.RegisterService(&_GNodeService_serviceDesc, srv)
}

func _GNodeService_ExecuteFunction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AntMes)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GNodeServiceServer).ExecuteFunction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gnode.GNodeService/ExecuteFunction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GNodeServiceServer).ExecuteFunction(ctx, req.(*AntMes))
	}
	return interceptor(ctx, in, info, handler)
}

func _GNodeService_GetClientStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GNodeServiceServer).GetClientStream(&gNodeServiceGetClientStreamServer{stream})
}

type GNodeService_GetClientStreamServer interface {
	Send(*AntMes) error
	Recv() (*AntMes, error)
	grpc.ServerStream
}

type gNodeServiceGetClientStreamServer struct {
	grpc.ServerStream
}

func (x *gNodeServiceGetClientStreamServer) Send(m *AntMes) error {
	return x.ServerStream.SendMsg(m)
}

func (x *gNodeServiceGetClientStreamServer) Recv() (*AntMes, error) {
	m := new(AntMes)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _GNodeService_AskConnection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AskConnectionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GNodeServiceServer).AskConnection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gnode.GNodeService/AskConnection",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GNodeServiceServer).AskConnection(ctx, req.(*AskConnectionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GNodeService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AntMes)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GNodeServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gnode.GNodeService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GNodeServiceServer).Ping(ctx, req.(*AntMes))
	}
	return interceptor(ctx, in, info, handler)
}

func _GNodeService_StoreFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StoreFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GNodeServiceServer).StoreFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gnode.GNodeService/StoreFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GNodeServiceServer).StoreFile(ctx, req.(*StoreFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GNodeService_RetrieveFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RetrieveFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GNodeServiceServer).RetrieveFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gnode.GNodeService/RetrieveFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GNodeServiceServer).RetrieveFile(ctx, req.(*RetrieveFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _GNodeService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gnode.GNodeService",
	HandlerType: (*GNodeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ExecuteFunction",
			Handler:    _GNodeService_ExecuteFunction_Handler,
		},
		{
			MethodName: "AskConnection",
			Handler:    _GNodeService_AskConnection_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _GNodeService_Ping_Handler,
		},
		{
			MethodName: "StoreFile",
			Handler:    _GNodeService_StoreFile_Handler,
		},
		{
			MethodName: "RetrieveFile",
			Handler:    _GNodeService_RetrieveFile_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetClientStream",
			Handler:       _GNodeService_GetClientStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "server/gnode/gnode.proto",
}

func init() { proto.RegisterFile("server/gnode/gnode.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 952 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xa4, 0x56, 0x5b, 0x6e, 0xe3, 0xc6,
	0x12, 0x35, 0x65, 0x3d, 0xc8, 0x32, 0xfd, 0xb8, 0x6d, 0x5f, 0xbb, 0x47, 0xb6, 0x11, 0x85, 0x13,
	0x20, 0x42, 0x3e, 0x66, 0x82, 0x04, 0xf9, 0x0a, 0x10, 0xc0, 0xf3, 0x84, 0x81, 0xd8, 0x98, 0xd0,
	0x06, 0xe6, 0x93, 0x20, 0xc5, 0xb2, 0xdc, 0x90, 0xd4, 0x54, 0x9a, 0x2d, 0x7b, 0x26, 0xcb, 0xc8,
	0x32, 0xb2, 0x8b, 0x6c, 0x25, 0x5b, 0xc8, 0x06, 0x82, 0xaa, 0x26, 0x35, 0x7a, 0x38, 0xce, 0xc7,
	0xfc, 0x08, 0x55, 0xa7, 0xaa, 0x9b, 0xa7, 0x4f, 0xd7, 0x21, 0x05, 0xb2, 0x44, 0x73, 0x87, 0xe6,
	0xf9, 0x50, 0x17, 0x39, 0xba, 0xdf, 0x67, 0x53, 0x53, 0xd8, 0x42, 0xb4, 0x38, 0x89, 0xfe, 0x68,
	0x43, 0xfb, 0x4c, 0xdb, 0x0b, 0x2c, 0xc5, 0x0e, 0x34, 0x54, 0x2e, 0xbd, 0x9e, 0xd7, 0x0f, 0xe2,
	0x86, 0xca, 0xc5, 0x21, 0xb4, 0x0b, 0xa3, 0x86, 0x4a, 0xcb, 0x06, 0x63, 0x55, 0x26, 0xbe, 0x80,
	0xad, 0x1b, 0x53, 0x4c, 0x92, 0xc1, 0x58, 0xa1, 0xb6, 0x72, 0x93, 0x8b, 0x40, 0xd0, 0x4b, 0x46,
	0x68, 0xa1, 0x4d, 0xcd, 0x10, 0xad, 0x6c, 0xba, 0x85, 0x2e, 0x13, 0x5d, 0xf0, 0x55, 0x79, 0xa6,
	0xcb, 0x7b, 0x34, 0xb2, 0xd5, 0xf3, 0xfa, 0x7e, 0x3c, 0xcf, 0xc5, 0x53, 0xd8, 0x36, 0x68, 0x67,
	0x46, 0x27, 0xa9, 0x6b, 0x68, 0x73, 0x43, 0xe8, 0xc0, 0xaa, 0xe9, 0x2b, 0xd8, 0x51, 0x65, 0x32,
	0x4d, 0xed, 0x6d, 0x72, 0x6f, 0x94, 0x45, 0x23, 0x3b, 0xae, 0x4b, 0x95, 0xef, 0x52, 0x7b, 0xfb,
	0x9e, 0x31, 0x71, 0x0c, 0x81, 0x63, 0x9a, 0xa8, 0x5c, 0xfa, 0xcc, 0xc0, 0x77, 0xc0, 0x79, 0x2e,
	0x04, 0x34, 0x69, 0xbd, 0x0c, 0x7a, 0x9b, 0xfd, 0x20, 0xe6, 0x58, 0x9c, 0x02, 0xf0, 0x9e, 0x4a,
	0xe7, 0xf8, 0x41, 0x42, 0xcf, 0xeb, 0xb7, 0xe2, 0x80, 0x90, 0x73, 0x02, 0x88, 0xf6, 0xcd, 0x4c,
	0x0f, 0xac, 0x2a, 0xb4, 0xdc, 0x72, 0xdb, 0xd5, 0x39, 0x6d, 0x97, 0x9a, 0x61, 0x29, 0x43, 0xb7,
	0x1d, 0xc5, 0xa4, 0x8f, 0x35, 0xa9, 0x2e, 0x6f, 0xd0, 0x10, 0x83, 0x6d, 0xa7, 0x4f, 0x0d, 0x9d,
	0xe7, 0xe2, 0x00, 0x5a, 0x85, 0xc9, 0xd1, 0xc8, 0x9d, 0x9e, 0xd7, 0xdf, 0x8c, 0x5d, 0x42, 0x87,
	0xd3, 0x59, 0x92, 0x8d, 0x8b, 0xc1, 0x28, 0xb1, 0x85, 0x4d, 0xc7, 0x72, 0x97, 0xcb, 0xa1, 0xce,
	0x5e, 0x10, 0x78, 0x4d, 0x98, 0x78, 0x02, 0x7e, 0xdd, 0x25, 0xf7, 0xb8, 0xde, 0xa9, 0xea, 0xc4,
	0xa5, 0x54, 0xbf, 0xa1, 0xfc, 0x1f, 0xc3, 0x1c, 0x93, 0xac, 0x4e, 0x7c, 0xcc, 0x59, 0x37, 0x29,
	0x98, 0x4d, 0x58, 0x83, 0x24, 0x1b, 0x11, 0xd6, 0x85, 0xdb, 0x53, 0xe9, 0xa1, 0xdc, 0x67, 0x4d,
	0x41, 0x17, 0x2f, 0x2a, 0x84, 0x08, 0xe7, 0x98, 0xcd, 0x86, 0xf2, 0x80, 0x4b, 0x2e, 0xa1, 0x65,
	0xee, 0xae, 0x92, 0xfb, 0x54, 0x59, 0xf9, 0x7f, 0xb7, 0xcc, 0x41, 0xef, 0x53, 0xc5, 0xf7, 0x8d,
	0xc6, 0x14, 0xe6, 0x02, 0x4b, 0x79, 0xe8, 0x84, 0xab, 0x73, 0xba, 0x24, 0x9d, 0x25, 0xf6, 0xd6,
	0x60, 0x9a, 0xcb, 0x23, 0x96, 0xdc, 0xd7, 0xd9, 0x35, 0xe7, 0x3c, 0x40, 0xae, 0x22, 0xb9, 0x52,
	0x65, 0xe2, 0x04, 0x82, 0x7c, 0x36, 0x1d, 0xab, 0x41, 0x6a, 0x51, 0x3e, 0x71, 0xf7, 0x34, 0x07,
	0xc4, 0x1e, 0x6c, 0x62, 0x71, 0x23, 0xbb, 0xcc, 0x83, 0x42, 0x7a, 0xc8, 0xac, 0x44, 0x93, 0xe8,
	0x74, 0x82, 0xf2, 0xd8, 0x31, 0x20, 0xe0, 0x32, 0x9d, 0x20, 0xdd, 0x3a, 0x17, 0x6d, 0x31, 0x42,
	0x2d, 0x4f, 0xb8, 0xca, 0xed, 0xd7, 0x04, 0x90, 0x9a, 0x79, 0x6a, 0x53, 0x79, 0xda, 0xf3, 0xfa,
	0x61, 0xcc, 0x71, 0xf4, 0x0d, 0x7b, 0x25, 0x46, 0x4b, 0xcf, 0x4a, 0x07, 0x23, 0x36, 0x8b, 0x1f,
	0x53, 0x58, 0xb9, 0xa7, 0x51, 0xbb, 0x27, 0xfa, 0xdd, 0x83, 0xce, 0x3b, 0xa5, 0x87, 0xd4, 0x2d,
	0xa0, 0x79, 0x5b, 0x94, 0xb6, 0xf2, 0x16, 0xc7, 0x84, 0x31, 0x2d, 0xb7, 0x82, 0x63, 0x71, 0x04,
	0x1d, 0x9d, 0x25, 0xe4, 0x4b, 0x76, 0x55, 0x2b, 0x6e, 0xeb, 0xec, 0xb2, 0xc8, 0x51, 0x7c, 0x09,
	0xa1, 0xce, 0x92, 0x4f, 0x67, 0x6f, 0x72, 0x75, 0x4b, 0x67, 0xaf, 0xe6, 0xa7, 0x7f, 0x0a, 0xdb,
	0xce, 0x90, 0x89, 0x9e, 0x4d, 0xb2, 0xca, 0x61, 0xad, 0x38, 0x74, 0xe0, 0x25, 0x63, 0xd1, 0x25,
	0x1c, 0x9c, 0x95, 0xa3, 0x97, 0x85, 0xd6, 0xc8, 0xf3, 0x1b, 0xe3, 0xaf, 0x33, 0x5c, 0x20, 0xe3,
	0x2d, 0x90, 0xa9, 0x49, 0x37, 0x16, 0x48, 0xd3, 0x21, 0xa7, 0x95, 0xe3, 0x1b, 0x6a, 0x1a, 0xfd,
	0xd9, 0x80, 0xbd, 0x2b, 0x5b, 0x18, 0x7c, 0xa3, 0xc6, 0x58, 0x6f, 0x76, 0x0c, 0x41, 0xc5, 0x64,
	0xfe, 0x3a, 0xf1, 0x1d, 0xe0, 0xfc, 0xb7, 0x76, 0xec, 0xda, 0x93, 0x6e, 0x5f, 0xe7, 0xc9, 0x75,
	0x37, 0x34, 0xff, 0xc3, 0x0d, 0xad, 0x65, 0x37, 0x9c, 0x40, 0xc0, 0xf8, 0x15, 0x59, 0xa2, 0xcd,
	0xb5, 0x4f, 0xc0, 0xaa, 0x47, 0x3b, 0x6b, 0x1e, 0xdd, 0x83, 0xcd, 0x11, 0x7e, 0xac, 0x5e, 0x1f,
	0x14, 0x2e, 0x0f, 0x53, 0xf0, 0xe8, 0x30, 0xc1, 0xea, 0x30, 0x75, 0xc1, 0x9f, 0xa0, 0x4d, 0x79,
	0xa0, 0xb6, 0xf8, 0x55, 0x31, 0xcf, 0xa3, 0x4b, 0x08, 0x17, 0x24, 0xb4, 0xab, 0xd4, 0xbc, 0x35,
	0x6a, 0xa7, 0x00, 0x4e, 0x17, 0x76, 0x7b, 0x63, 0xe5, 0x68, 0x11, 0x80, 0xff, 0x7a, 0x32, 0xb5,
	0x1f, 0x63, 0xb4, 0xd1, 0xdf, 0x1e, 0xec, 0xc7, 0x68, 0x8d, 0xc2, 0xbb, 0xcf, 0xbb, 0xa2, 0x2e,
	0xcc, 0xdd, 0x59, 0x8d, 0xe6, 0x43, 0x6e, 0x6d, 0xfe, 0xbb, 0x5b, 0x5b, 0xab, 0x6e, 0x5d, 0x92,
	0xb3, 0xfd, 0xa8, 0x9c, 0x9d, 0x55, 0x39, 0xe7, 0x0a, 0x8c, 0x55, 0x69, 0xab, 0x3b, 0x72, 0x0a,
	0xfc, 0xac, 0x4a, 0x1b, 0xfd, 0x02, 0xbb, 0xcb, 0x87, 0xfe, 0x6c, 0x51, 0xbf, 0xfb, 0xab, 0x01,
	0xe1, 0x5b, 0xb2, 0xe2, 0x15, 0x9a, 0x3b, 0x35, 0x40, 0xf1, 0x1c, 0x76, 0x5f, 0x7f, 0xc0, 0xc1,
	0xcc, 0xe2, 0x9b, 0xfa, 0x5b, 0xb0, 0xfd, 0xcc, 0x7d, 0x5f, 0xdd, 0xe7, 0xb4, 0xbb, 0x90, 0x12,
	0x83, 0x1f, 0x60, 0xf7, 0x2d, 0x5a, 0xf7, 0x85, 0xbc, 0xb2, 0x06, 0xd3, 0xc9, 0x23, 0x0b, 0x2e,
	0xb0, 0x8c, 0x36, 0xfa, 0xde, 0xb7, 0x9e, 0xf8, 0x09, 0xb6, 0x97, 0x1c, 0x2b, 0x8e, 0xeb, 0xae,
	0x07, 0x7c, 0xdc, 0xdd, 0xa9, 0x8a, 0xd5, 0x8b, 0x27, 0xda, 0x10, 0x5f, 0x43, 0x93, 0x92, 0xd5,
	0x67, 0xad, 0x37, 0xfe, 0x08, 0xc1, 0x7c, 0x0c, 0xc5, 0x51, 0x55, 0x5e, 0xf5, 0x76, 0x77, 0x7f,
	0xbd, 0x40, 0x8b, 0x5f, 0x41, 0xb8, 0xa8, 0xb8, 0xe8, 0x56, 0x6d, 0x0f, 0xcc, 0x5e, 0xf7, 0xf0,
	0xc1, 0x9a, 0x8d, 0x36, 0xb2, 0x36, 0xff, 0x33, 0xf9, 0xfe, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff,
	0xab, 0xaa, 0x15, 0xfc, 0xb5, 0x08, 0x00, 0x00,
}
