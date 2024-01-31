// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: upshot/state/v1/query.proto

package statev1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Query_Params_FullMethodName                     = "/upshot.state.v1.Query/Params"
	Query_GetLastRewardsUpdate_FullMethodName       = "/upshot.state.v1.Query/GetLastRewardsUpdate"
	Query_GetAccumulatedEpochRewards_FullMethodName = "/upshot.state.v1.Query/GetAccumulatedEpochRewards"
	Query_GetNextTopicId_FullMethodName             = "/upshot.state.v1.Query/GetNextTopicId"
	Query_GetTopic_FullMethodName                   = "/upshot.state.v1.Query/GetTopic"
	Query_GetWeight_FullMethodName                  = "/upshot.state.v1.Query/GetWeight"
	Query_GetAllInferences_FullMethodName           = "/upshot.state.v1.Query/GetAllInferences"
	Query_GetInferencesToScore_FullMethodName       = "/upshot.state.v1.Query/GetInferencesToScore"
	Query_GetTotalStake_FullMethodName              = "/upshot.state.v1.Query/GetTotalStake"
	Query_GetWorkerNodeRegistration_FullMethodName  = "/upshot.state.v1.Query/GetWorkerNodeRegistration"
)

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type QueryClient interface {
	// Params returns the module parameters.
	Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
	GetLastRewardsUpdate(ctx context.Context, in *QueryLastRewardsUpdateRequest, opts ...grpc.CallOption) (*QueryLastRewardsUpdateResponse, error)
	GetAccumulatedEpochRewards(ctx context.Context, in *QueryAccumulatedEpochRewardsRequest, opts ...grpc.CallOption) (*QueryAccumulatedEpochRewardsResponse, error)
	GetNextTopicId(ctx context.Context, in *QueryNextTopicIdRequest, opts ...grpc.CallOption) (*QueryNextTopicIdResponse, error)
	GetTopic(ctx context.Context, in *QueryTopicRequest, opts ...grpc.CallOption) (*QueryTopicResponse, error)
	GetWeight(ctx context.Context, in *QueryWeightRequest, opts ...grpc.CallOption) (*QueryWeightResponse, error)
	GetAllInferences(ctx context.Context, in *QueryInferenceRequest, opts ...grpc.CallOption) (*QueryInferenceResponse, error)
	GetInferencesToScore(ctx context.Context, in *QueryInferencesToScoreRequest, opts ...grpc.CallOption) (*QueryInferencesToScoreResponse, error)
	GetTotalStake(ctx context.Context, in *QueryTotalStakeRequest, opts ...grpc.CallOption) (*QueryTotalStakeResponse, error)
	GetWorkerNodeRegistration(ctx context.Context, in *QueryRegisteredWorkerNodesRequest, opts ...grpc.CallOption) (*QueryRegisteredWorkerNodesResponse, error)
}

type queryClient struct {
	cc grpc.ClientConnInterface
}

func NewQueryClient(cc grpc.ClientConnInterface) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error) {
	out := new(QueryParamsResponse)
	err := c.cc.Invoke(ctx, Query_Params_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) GetLastRewardsUpdate(ctx context.Context, in *QueryLastRewardsUpdateRequest, opts ...grpc.CallOption) (*QueryLastRewardsUpdateResponse, error) {
	out := new(QueryLastRewardsUpdateResponse)
	err := c.cc.Invoke(ctx, Query_GetLastRewardsUpdate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) GetAccumulatedEpochRewards(ctx context.Context, in *QueryAccumulatedEpochRewardsRequest, opts ...grpc.CallOption) (*QueryAccumulatedEpochRewardsResponse, error) {
	out := new(QueryAccumulatedEpochRewardsResponse)
	err := c.cc.Invoke(ctx, Query_GetAccumulatedEpochRewards_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) GetNextTopicId(ctx context.Context, in *QueryNextTopicIdRequest, opts ...grpc.CallOption) (*QueryNextTopicIdResponse, error) {
	out := new(QueryNextTopicIdResponse)
	err := c.cc.Invoke(ctx, Query_GetNextTopicId_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) GetTopic(ctx context.Context, in *QueryTopicRequest, opts ...grpc.CallOption) (*QueryTopicResponse, error) {
	out := new(QueryTopicResponse)
	err := c.cc.Invoke(ctx, Query_GetTopic_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) GetWeight(ctx context.Context, in *QueryWeightRequest, opts ...grpc.CallOption) (*QueryWeightResponse, error) {
	out := new(QueryWeightResponse)
	err := c.cc.Invoke(ctx, Query_GetWeight_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) GetAllInferences(ctx context.Context, in *QueryInferenceRequest, opts ...grpc.CallOption) (*QueryInferenceResponse, error) {
	out := new(QueryInferenceResponse)
	err := c.cc.Invoke(ctx, Query_GetAllInferences_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) GetInferencesToScore(ctx context.Context, in *QueryInferencesToScoreRequest, opts ...grpc.CallOption) (*QueryInferencesToScoreResponse, error) {
	out := new(QueryInferencesToScoreResponse)
	err := c.cc.Invoke(ctx, Query_GetInferencesToScore_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) GetTotalStake(ctx context.Context, in *QueryTotalStakeRequest, opts ...grpc.CallOption) (*QueryTotalStakeResponse, error) {
	out := new(QueryTotalStakeResponse)
	err := c.cc.Invoke(ctx, Query_GetTotalStake_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) GetWorkerNodeRegistration(ctx context.Context, in *QueryRegisteredWorkerNodesRequest, opts ...grpc.CallOption) (*QueryRegisteredWorkerNodesResponse, error) {
	out := new(QueryRegisteredWorkerNodesResponse)
	err := c.cc.Invoke(ctx, Query_GetWorkerNodeRegistration_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
// All implementations must embed UnimplementedQueryServer
// for forward compatibility
type QueryServer interface {
	// Params returns the module parameters.
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	GetLastRewardsUpdate(context.Context, *QueryLastRewardsUpdateRequest) (*QueryLastRewardsUpdateResponse, error)
	GetAccumulatedEpochRewards(context.Context, *QueryAccumulatedEpochRewardsRequest) (*QueryAccumulatedEpochRewardsResponse, error)
	GetNextTopicId(context.Context, *QueryNextTopicIdRequest) (*QueryNextTopicIdResponse, error)
	GetTopic(context.Context, *QueryTopicRequest) (*QueryTopicResponse, error)
	GetWeight(context.Context, *QueryWeightRequest) (*QueryWeightResponse, error)
	GetAllInferences(context.Context, *QueryInferenceRequest) (*QueryInferenceResponse, error)
	GetInferencesToScore(context.Context, *QueryInferencesToScoreRequest) (*QueryInferencesToScoreResponse, error)
	GetTotalStake(context.Context, *QueryTotalStakeRequest) (*QueryTotalStakeResponse, error)
	GetWorkerNodeRegistration(context.Context, *QueryRegisteredWorkerNodesRequest) (*QueryRegisteredWorkerNodesResponse, error)
	mustEmbedUnimplementedQueryServer()
}

// UnimplementedQueryServer must be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (UnimplementedQueryServer) Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Params not implemented")
}
func (UnimplementedQueryServer) GetLastRewardsUpdate(context.Context, *QueryLastRewardsUpdateRequest) (*QueryLastRewardsUpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLastRewardsUpdate not implemented")
}
func (UnimplementedQueryServer) GetAccumulatedEpochRewards(context.Context, *QueryAccumulatedEpochRewardsRequest) (*QueryAccumulatedEpochRewardsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccumulatedEpochRewards not implemented")
}
func (UnimplementedQueryServer) GetNextTopicId(context.Context, *QueryNextTopicIdRequest) (*QueryNextTopicIdResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNextTopicId not implemented")
}
func (UnimplementedQueryServer) GetTopic(context.Context, *QueryTopicRequest) (*QueryTopicResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTopic not implemented")
}
func (UnimplementedQueryServer) GetWeight(context.Context, *QueryWeightRequest) (*QueryWeightResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWeight not implemented")
}
func (UnimplementedQueryServer) GetAllInferences(context.Context, *QueryInferenceRequest) (*QueryInferenceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllInferences not implemented")
}
func (UnimplementedQueryServer) GetInferencesToScore(context.Context, *QueryInferencesToScoreRequest) (*QueryInferencesToScoreResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInferencesToScore not implemented")
}
func (UnimplementedQueryServer) GetTotalStake(context.Context, *QueryTotalStakeRequest) (*QueryTotalStakeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTotalStake not implemented")
}
func (UnimplementedQueryServer) GetWorkerNodeRegistration(context.Context, *QueryRegisteredWorkerNodesRequest) (*QueryRegisteredWorkerNodesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWorkerNodeRegistration not implemented")
}
func (UnimplementedQueryServer) mustEmbedUnimplementedQueryServer() {}

// UnsafeQueryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to QueryServer will
// result in compilation errors.
type UnsafeQueryServer interface {
	mustEmbedUnimplementedQueryServer()
}

func RegisterQueryServer(s grpc.ServiceRegistrar, srv QueryServer) {
	s.RegisterService(&Query_ServiceDesc, srv)
}

func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Params(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_Params_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_GetLastRewardsUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryLastRewardsUpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).GetLastRewardsUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_GetLastRewardsUpdate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).GetLastRewardsUpdate(ctx, req.(*QueryLastRewardsUpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_GetAccumulatedEpochRewards_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryAccumulatedEpochRewardsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).GetAccumulatedEpochRewards(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_GetAccumulatedEpochRewards_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).GetAccumulatedEpochRewards(ctx, req.(*QueryAccumulatedEpochRewardsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_GetNextTopicId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryNextTopicIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).GetNextTopicId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_GetNextTopicId_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).GetNextTopicId(ctx, req.(*QueryNextTopicIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_GetTopic_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryTopicRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).GetTopic(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_GetTopic_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).GetTopic(ctx, req.(*QueryTopicRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_GetWeight_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryWeightRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).GetWeight(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_GetWeight_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).GetWeight(ctx, req.(*QueryWeightRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_GetAllInferences_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryInferenceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).GetAllInferences(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_GetAllInferences_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).GetAllInferences(ctx, req.(*QueryInferenceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_GetInferencesToScore_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryInferencesToScoreRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).GetInferencesToScore(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_GetInferencesToScore_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).GetInferencesToScore(ctx, req.(*QueryInferencesToScoreRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_GetTotalStake_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryTotalStakeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).GetTotalStake(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_GetTotalStake_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).GetTotalStake(ctx, req.(*QueryTotalStakeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_GetWorkerNodeRegistration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRegisteredWorkerNodesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).GetWorkerNodeRegistration(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_GetWorkerNodeRegistration_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).GetWorkerNodeRegistration(ctx, req.(*QueryRegisteredWorkerNodesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Query_ServiceDesc is the grpc.ServiceDesc for Query service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Query_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "upshot.state.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Params",
			Handler:    _Query_Params_Handler,
		},
		{
			MethodName: "GetLastRewardsUpdate",
			Handler:    _Query_GetLastRewardsUpdate_Handler,
		},
		{
			MethodName: "GetAccumulatedEpochRewards",
			Handler:    _Query_GetAccumulatedEpochRewards_Handler,
		},
		{
			MethodName: "GetNextTopicId",
			Handler:    _Query_GetNextTopicId_Handler,
		},
		{
			MethodName: "GetTopic",
			Handler:    _Query_GetTopic_Handler,
		},
		{
			MethodName: "GetWeight",
			Handler:    _Query_GetWeight_Handler,
		},
		{
			MethodName: "GetAllInferences",
			Handler:    _Query_GetAllInferences_Handler,
		},
		{
			MethodName: "GetInferencesToScore",
			Handler:    _Query_GetInferencesToScore_Handler,
		},
		{
			MethodName: "GetTotalStake",
			Handler:    _Query_GetTotalStake_Handler,
		},
		{
			MethodName: "GetWorkerNodeRegistration",
			Handler:    _Query_GetWorkerNodeRegistration_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "upshot/state/v1/query.proto",
}
