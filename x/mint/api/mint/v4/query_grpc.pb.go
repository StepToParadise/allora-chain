// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: mint/v4/query.proto

package mintv4

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	QueryService_Params_FullMethodName       = "/mint.v4.QueryService/Params"
	QueryService_Inflation_FullMethodName    = "/mint.v4.QueryService/Inflation"
	QueryService_EmissionInfo_FullMethodName = "/mint.v4.QueryService/EmissionInfo"
)

// QueryServiceClient is the client API for QueryService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// QueryService provides defines the gRPC querier service.
type QueryServiceClient interface {
	// Params returns the total set of minting parameters.
	Params(ctx context.Context, in *QueryServiceParamsRequest, opts ...grpc.CallOption) (*QueryServiceParamsResponse, error)
	// Inflation returns the current minting inflation value.
	Inflation(ctx context.Context, in *QueryServiceInflationRequest, opts ...grpc.CallOption) (*QueryServiceInflationResponse, error)
	EmissionInfo(ctx context.Context, in *QueryServiceEmissionInfoRequest, opts ...grpc.CallOption) (*QueryServiceEmissionInfoResponse, error)
}

type queryServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewQueryServiceClient(cc grpc.ClientConnInterface) QueryServiceClient {
	return &queryServiceClient{cc}
}

func (c *queryServiceClient) Params(ctx context.Context, in *QueryServiceParamsRequest, opts ...grpc.CallOption) (*QueryServiceParamsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(QueryServiceParamsResponse)
	err := c.cc.Invoke(ctx, QueryService_Params_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryServiceClient) Inflation(ctx context.Context, in *QueryServiceInflationRequest, opts ...grpc.CallOption) (*QueryServiceInflationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(QueryServiceInflationResponse)
	err := c.cc.Invoke(ctx, QueryService_Inflation_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryServiceClient) EmissionInfo(ctx context.Context, in *QueryServiceEmissionInfoRequest, opts ...grpc.CallOption) (*QueryServiceEmissionInfoResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(QueryServiceEmissionInfoResponse)
	err := c.cc.Invoke(ctx, QueryService_EmissionInfo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServiceServer is the server API for QueryService service.
// All implementations must embed UnimplementedQueryServiceServer
// for forward compatibility.
//
// QueryService provides defines the gRPC querier service.
type QueryServiceServer interface {
	// Params returns the total set of minting parameters.
	Params(context.Context, *QueryServiceParamsRequest) (*QueryServiceParamsResponse, error)
	// Inflation returns the current minting inflation value.
	Inflation(context.Context, *QueryServiceInflationRequest) (*QueryServiceInflationResponse, error)
	EmissionInfo(context.Context, *QueryServiceEmissionInfoRequest) (*QueryServiceEmissionInfoResponse, error)
	mustEmbedUnimplementedQueryServiceServer()
}

// UnimplementedQueryServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedQueryServiceServer struct{}

func (UnimplementedQueryServiceServer) Params(context.Context, *QueryServiceParamsRequest) (*QueryServiceParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Params not implemented")
}
func (UnimplementedQueryServiceServer) Inflation(context.Context, *QueryServiceInflationRequest) (*QueryServiceInflationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Inflation not implemented")
}
func (UnimplementedQueryServiceServer) EmissionInfo(context.Context, *QueryServiceEmissionInfoRequest) (*QueryServiceEmissionInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EmissionInfo not implemented")
}
func (UnimplementedQueryServiceServer) mustEmbedUnimplementedQueryServiceServer() {}
func (UnimplementedQueryServiceServer) testEmbeddedByValue()                      {}

// UnsafeQueryServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to QueryServiceServer will
// result in compilation errors.
type UnsafeQueryServiceServer interface {
	mustEmbedUnimplementedQueryServiceServer()
}

func RegisterQueryServiceServer(s grpc.ServiceRegistrar, srv QueryServiceServer) {
	// If the following call pancis, it indicates UnimplementedQueryServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&QueryService_ServiceDesc, srv)
}

func _QueryService_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryServiceParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServiceServer).Params(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: QueryService_Params_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServiceServer).Params(ctx, req.(*QueryServiceParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QueryService_Inflation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryServiceInflationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServiceServer).Inflation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: QueryService_Inflation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServiceServer).Inflation(ctx, req.(*QueryServiceInflationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QueryService_EmissionInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryServiceEmissionInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServiceServer).EmissionInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: QueryService_EmissionInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServiceServer).EmissionInfo(ctx, req.(*QueryServiceEmissionInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// QueryService_ServiceDesc is the grpc.ServiceDesc for QueryService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var QueryService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mint.v4.QueryService",
	HandlerType: (*QueryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Params",
			Handler:    _QueryService_Params_Handler,
		},
		{
			MethodName: "Inflation",
			Handler:    _QueryService_Inflation_Handler,
		},
		{
			MethodName: "EmissionInfo",
			Handler:    _QueryService_EmissionInfo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mint/v4/query.proto",
}
