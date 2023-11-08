# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: pulumi/resource.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import empty_pb2 as google_dot_protobuf_dot_empty__pb2
from google.protobuf import struct_pb2 as google_dot_protobuf_dot_struct__pb2
from . import provider_pb2 as pulumi_dot_provider__pb2
from . import alias_pb2 as pulumi_dot_alias__pb2
from . import source_pb2 as pulumi_dot_source__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x15pulumi/resource.proto\x12\tpulumirpc\x1a\x1bgoogle/protobuf/empty.proto\x1a\x1cgoogle/protobuf/struct.proto\x1a\x15pulumi/provider.proto\x1a\x12pulumi/alias.proto\x1a\x13pulumi/source.proto\"$\n\x16SupportsFeatureRequest\x12\n\n\x02id\x18\x01 \x01(\t\"-\n\x17SupportsFeatureResponse\x12\x12\n\nhasSupport\x18\x01 \x01(\x08\"\xe7\x03\n\x13ReadResourceRequest\x12\n\n\x02id\x18\x01 \x01(\t\x12\x0c\n\x04type\x18\x02 \x01(\t\x12\x0c\n\x04name\x18\x03 \x01(\t\x12\x0e\n\x06parent\x18\x04 \x01(\t\x12+\n\nproperties\x18\x05 \x01(\x0b\x32\x17.google.protobuf.Struct\x12\x14\n\x0c\x64\x65pendencies\x18\x06 \x03(\t\x12\x10\n\x08provider\x18\x07 \x01(\t\x12\x0f\n\x07version\x18\x08 \x01(\t\x12\x15\n\racceptSecrets\x18\t \x01(\x08\x12\x1f\n\x17\x61\x64\x64itionalSecretOutputs\x18\n \x03(\t\x12\x17\n\x0f\x61\x63\x63\x65ptResources\x18\x0c \x01(\x08\x12\x19\n\x11pluginDownloadURL\x18\r \x01(\t\x12L\n\x0fpluginChecksums\x18\x0f \x03(\x0b\x32\x33.pulumirpc.ReadResourceRequest.PluginChecksumsEntry\x12\x31\n\x0esourcePosition\x18\x0e \x01(\x0b\x32\x19.pulumirpc.SourcePosition\x1a\x36\n\x14PluginChecksumsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\x0c:\x02\x38\x01J\x04\x08\x0b\x10\x0cR\x07\x61liases\"P\n\x14ReadResourceResponse\x12\x0b\n\x03urn\x18\x01 \x01(\t\x12+\n\nproperties\x18\x02 \x01(\x0b\x32\x17.google.protobuf.Struct\"R\n\tParameter\x12\x0b\n\x03key\x18\x01 \x01(\t\x12%\n\x05value\x18\x02 \x01(\x0b\x32\x16.google.protobuf.Value\x12\x11\n\textension\x18\x03 \x01(\x08\"\xc1\n\n\x17RegisterResourceRequest\x12\x0c\n\x04type\x18\x01 \x01(\t\x12\x0c\n\x04name\x18\x02 \x01(\t\x12\x0e\n\x06parent\x18\x03 \x01(\t\x12\x0e\n\x06\x63ustom\x18\x04 \x01(\x08\x12\'\n\x06object\x18\x05 \x01(\x0b\x32\x17.google.protobuf.Struct\x12\x0f\n\x07protect\x18\x06 \x01(\x08\x12\x14\n\x0c\x64\x65pendencies\x18\x07 \x03(\t\x12\x10\n\x08provider\x18\x08 \x01(\t\x12Z\n\x14propertyDependencies\x18\t \x03(\x0b\x32<.pulumirpc.RegisterResourceRequest.PropertyDependenciesEntry\x12\x1b\n\x13\x64\x65leteBeforeReplace\x18\n \x01(\x08\x12\x0f\n\x07version\x18\x0b \x01(\t\x12\x15\n\rignoreChanges\x18\x0c \x03(\t\x12\x15\n\racceptSecrets\x18\r \x01(\x08\x12\x1f\n\x17\x61\x64\x64itionalSecretOutputs\x18\x0e \x03(\t\x12\x11\n\taliasURNs\x18\x0f \x03(\t\x12\x10\n\x08importId\x18\x10 \x01(\t\x12I\n\x0e\x63ustomTimeouts\x18\x11 \x01(\x0b\x32\x31.pulumirpc.RegisterResourceRequest.CustomTimeouts\x12\"\n\x1a\x64\x65leteBeforeReplaceDefined\x18\x12 \x01(\x08\x12\x1d\n\x15supportsPartialValues\x18\x13 \x01(\x08\x12\x0e\n\x06remote\x18\x14 \x01(\x08\x12\x17\n\x0f\x61\x63\x63\x65ptResources\x18\x15 \x01(\x08\x12\x44\n\tproviders\x18\x16 \x03(\x0b\x32\x31.pulumirpc.RegisterResourceRequest.ProvidersEntry\x12\x18\n\x10replaceOnChanges\x18\x17 \x03(\t\x12\x19\n\x11pluginDownloadURL\x18\x18 \x01(\t\x12P\n\x0fpluginChecksums\x18\x1e \x03(\x0b\x32\x37.pulumirpc.RegisterResourceRequest.PluginChecksumsEntry\x12\x16\n\x0eretainOnDelete\x18\x19 \x01(\x08\x12!\n\x07\x61liases\x18\x1a \x03(\x0b\x32\x10.pulumirpc.Alias\x12\x13\n\x0b\x64\x65letedWith\x18\x1b \x01(\t\x12\x12\n\naliasSpecs\x18\x1c \x01(\x08\x12\x31\n\x0esourcePosition\x18\x1d \x01(\x0b\x32\x19.pulumirpc.SourcePosition\x12\'\n\tparameter\x18\x1f \x01(\x0b\x32\x14.pulumirpc.Parameter\x1a$\n\x14PropertyDependencies\x12\x0c\n\x04urns\x18\x01 \x03(\t\x1a@\n\x0e\x43ustomTimeouts\x12\x0e\n\x06\x63reate\x18\x01 \x01(\t\x12\x0e\n\x06update\x18\x02 \x01(\t\x12\x0e\n\x06\x64\x65lete\x18\x03 \x01(\t\x1at\n\x19PropertyDependenciesEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\x46\n\x05value\x18\x02 \x01(\x0b\x32\x37.pulumirpc.RegisterResourceRequest.PropertyDependencies:\x02\x38\x01\x1a\x30\n\x0eProvidersEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\x1a\x36\n\x14PluginChecksumsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\x0c:\x02\x38\x01\"\xf7\x02\n\x18RegisterResourceResponse\x12\x0b\n\x03urn\x18\x01 \x01(\t\x12\n\n\x02id\x18\x02 \x01(\t\x12\'\n\x06object\x18\x03 \x01(\x0b\x32\x17.google.protobuf.Struct\x12\x0e\n\x06stable\x18\x04 \x01(\x08\x12\x0f\n\x07stables\x18\x05 \x03(\t\x12[\n\x14propertyDependencies\x18\x06 \x03(\x0b\x32=.pulumirpc.RegisterResourceResponse.PropertyDependenciesEntry\x1a$\n\x14PropertyDependencies\x12\x0c\n\x04urns\x18\x01 \x03(\t\x1au\n\x19PropertyDependenciesEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12G\n\x05value\x18\x02 \x01(\x0b\x32\x38.pulumirpc.RegisterResourceResponse.PropertyDependencies:\x02\x38\x01\"W\n\x1eRegisterResourceOutputsRequest\x12\x0b\n\x03urn\x18\x01 \x01(\t\x12(\n\x07outputs\x18\x02 \x01(\x0b\x32\x17.google.protobuf.Struct\"\xdd\x02\n\x15ResourceInvokeRequest\x12\x0b\n\x03tok\x18\x01 \x01(\t\x12%\n\x04\x61rgs\x18\x02 \x01(\x0b\x32\x17.google.protobuf.Struct\x12\x10\n\x08provider\x18\x03 \x01(\t\x12\x0f\n\x07version\x18\x04 \x01(\t\x12\x17\n\x0f\x61\x63\x63\x65ptResources\x18\x05 \x01(\x08\x12\x19\n\x11pluginDownloadURL\x18\x06 \x01(\t\x12N\n\x0fpluginChecksums\x18\x08 \x03(\x0b\x32\x35.pulumirpc.ResourceInvokeRequest.PluginChecksumsEntry\x12\x31\n\x0esourcePosition\x18\x07 \x01(\x0b\x32\x19.pulumirpc.SourcePosition\x1a\x36\n\x14PluginChecksumsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\x0c:\x02\x38\x01\x32\xd4\x04\n\x0fResourceMonitor\x12Z\n\x0fSupportsFeature\x12!.pulumirpc.SupportsFeatureRequest\x1a\".pulumirpc.SupportsFeatureResponse\"\x00\x12G\n\x06Invoke\x12 .pulumirpc.ResourceInvokeRequest\x1a\x19.pulumirpc.InvokeResponse\"\x00\x12O\n\x0cStreamInvoke\x12 .pulumirpc.ResourceInvokeRequest\x1a\x19.pulumirpc.InvokeResponse\"\x00\x30\x01\x12\x39\n\x04\x43\x61ll\x12\x16.pulumirpc.CallRequest\x1a\x17.pulumirpc.CallResponse\"\x00\x12Q\n\x0cReadResource\x12\x1e.pulumirpc.ReadResourceRequest\x1a\x1f.pulumirpc.ReadResourceResponse\"\x00\x12]\n\x10RegisterResource\x12\".pulumirpc.RegisterResourceRequest\x1a#.pulumirpc.RegisterResourceResponse\"\x00\x12^\n\x17RegisterResourceOutputs\x12).pulumirpc.RegisterResourceOutputsRequest\x1a\x16.google.protobuf.Empty\"\x00\x42\x34Z2github.com/pulumi/pulumi/sdk/v3/proto/go;pulumirpcb\x06proto3')

_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, globals())
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'pulumi.resource_pb2', globals())
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z2github.com/pulumi/pulumi/sdk/v3/proto/go;pulumirpc'
  _READRESOURCEREQUEST_PLUGINCHECKSUMSENTRY._options = None
  _READRESOURCEREQUEST_PLUGINCHECKSUMSENTRY._serialized_options = b'8\001'
  _REGISTERRESOURCEREQUEST_PROPERTYDEPENDENCIESENTRY._options = None
  _REGISTERRESOURCEREQUEST_PROPERTYDEPENDENCIESENTRY._serialized_options = b'8\001'
  _REGISTERRESOURCEREQUEST_PROVIDERSENTRY._options = None
  _REGISTERRESOURCEREQUEST_PROVIDERSENTRY._serialized_options = b'8\001'
  _REGISTERRESOURCEREQUEST_PLUGINCHECKSUMSENTRY._options = None
  _REGISTERRESOURCEREQUEST_PLUGINCHECKSUMSENTRY._serialized_options = b'8\001'
  _REGISTERRESOURCERESPONSE_PROPERTYDEPENDENCIESENTRY._options = None
  _REGISTERRESOURCERESPONSE_PROPERTYDEPENDENCIESENTRY._serialized_options = b'8\001'
  _RESOURCEINVOKEREQUEST_PLUGINCHECKSUMSENTRY._options = None
  _RESOURCEINVOKEREQUEST_PLUGINCHECKSUMSENTRY._serialized_options = b'8\001'
  _SUPPORTSFEATUREREQUEST._serialized_start=159
  _SUPPORTSFEATUREREQUEST._serialized_end=195
  _SUPPORTSFEATURERESPONSE._serialized_start=197
  _SUPPORTSFEATURERESPONSE._serialized_end=242
  _READRESOURCEREQUEST._serialized_start=245
  _READRESOURCEREQUEST._serialized_end=732
  _READRESOURCEREQUEST_PLUGINCHECKSUMSENTRY._serialized_start=663
  _READRESOURCEREQUEST_PLUGINCHECKSUMSENTRY._serialized_end=717
  _READRESOURCERESPONSE._serialized_start=734
  _READRESOURCERESPONSE._serialized_end=814
  _PARAMETER._serialized_start=816
  _PARAMETER._serialized_end=898
  _REGISTERRESOURCEREQUEST._serialized_start=901
  _REGISTERRESOURCEREQUEST._serialized_end=2246
  _REGISTERRESOURCEREQUEST_PROPERTYDEPENDENCIES._serialized_start=1920
  _REGISTERRESOURCEREQUEST_PROPERTYDEPENDENCIES._serialized_end=1956
  _REGISTERRESOURCEREQUEST_CUSTOMTIMEOUTS._serialized_start=1958
  _REGISTERRESOURCEREQUEST_CUSTOMTIMEOUTS._serialized_end=2022
  _REGISTERRESOURCEREQUEST_PROPERTYDEPENDENCIESENTRY._serialized_start=2024
  _REGISTERRESOURCEREQUEST_PROPERTYDEPENDENCIESENTRY._serialized_end=2140
  _REGISTERRESOURCEREQUEST_PROVIDERSENTRY._serialized_start=2142
  _REGISTERRESOURCEREQUEST_PROVIDERSENTRY._serialized_end=2190
  _REGISTERRESOURCEREQUEST_PLUGINCHECKSUMSENTRY._serialized_start=663
  _REGISTERRESOURCEREQUEST_PLUGINCHECKSUMSENTRY._serialized_end=717
  _REGISTERRESOURCERESPONSE._serialized_start=2249
  _REGISTERRESOURCERESPONSE._serialized_end=2624
  _REGISTERRESOURCERESPONSE_PROPERTYDEPENDENCIES._serialized_start=1920
  _REGISTERRESOURCERESPONSE_PROPERTYDEPENDENCIES._serialized_end=1956
  _REGISTERRESOURCERESPONSE_PROPERTYDEPENDENCIESENTRY._serialized_start=2507
  _REGISTERRESOURCERESPONSE_PROPERTYDEPENDENCIESENTRY._serialized_end=2624
  _REGISTERRESOURCEOUTPUTSREQUEST._serialized_start=2626
  _REGISTERRESOURCEOUTPUTSREQUEST._serialized_end=2713
  _RESOURCEINVOKEREQUEST._serialized_start=2716
  _RESOURCEINVOKEREQUEST._serialized_end=3065
  _RESOURCEINVOKEREQUEST_PLUGINCHECKSUMSENTRY._serialized_start=663
  _RESOURCEINVOKEREQUEST_PLUGINCHECKSUMSENTRY._serialized_end=717
  _RESOURCEMONITOR._serialized_start=3068
  _RESOURCEMONITOR._serialized_end=3664
# @@protoc_insertion_point(module_scope)
