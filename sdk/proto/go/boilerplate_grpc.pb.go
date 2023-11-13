// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.20.1
// source: pulumi/boilerplate.proto

package pulumirpc

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

// BoilerplateClient is the client API for Boilerplate service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BoilerplateClient interface {
	// CreatePackage creates a new Pulumi package - provider.
	CreatePackage(ctx context.Context, in *CreatePackageRequest, opts ...grpc.CallOption) (*CreatePackageResponse, error)
}

type boilerplateClient struct {
	cc grpc.ClientConnInterface
}

func NewBoilerplateClient(cc grpc.ClientConnInterface) BoilerplateClient {
	return &boilerplateClient{cc}
}

func (c *boilerplateClient) CreatePackage(ctx context.Context, in *CreatePackageRequest, opts ...grpc.CallOption) (*CreatePackageResponse, error) {
	out := new(CreatePackageResponse)
	err := c.cc.Invoke(ctx, "/pulumirpc.Boilerplate/CreatePackage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BoilerplateServer is the server API for Boilerplate service.
// All implementations must embed UnimplementedBoilerplateServer
// for forward compatibility
type BoilerplateServer interface {
	// CreatePackage creates a new Pulumi package - provider.
	CreatePackage(context.Context, *CreatePackageRequest) (*CreatePackageResponse, error)
	mustEmbedUnimplementedBoilerplateServer()
}

// UnimplementedBoilerplateServer must be embedded to have forward compatible implementations.
type UnimplementedBoilerplateServer struct {
}

func (UnimplementedBoilerplateServer) CreatePackage(context.Context, *CreatePackageRequest) (*CreatePackageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreatePackage not implemented")
}
func (UnimplementedBoilerplateServer) mustEmbedUnimplementedBoilerplateServer() {}

// UnsafeBoilerplateServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BoilerplateServer will
// result in compilation errors.
type UnsafeBoilerplateServer interface {
	mustEmbedUnimplementedBoilerplateServer()
}

func RegisterBoilerplateServer(s grpc.ServiceRegistrar, srv BoilerplateServer) {
	s.RegisterService(&Boilerplate_ServiceDesc, srv)
}

func _Boilerplate_CreatePackage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreatePackageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BoilerplateServer).CreatePackage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pulumirpc.Boilerplate/CreatePackage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BoilerplateServer).CreatePackage(ctx, req.(*CreatePackageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Boilerplate_ServiceDesc is the grpc.ServiceDesc for Boilerplate service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Boilerplate_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pulumirpc.Boilerplate",
	HandlerType: (*BoilerplateServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreatePackage",
			Handler:    _Boilerplate_CreatePackage_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pulumi/boilerplate.proto",
}
