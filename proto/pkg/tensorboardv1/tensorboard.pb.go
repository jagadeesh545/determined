// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// source: determined/tensorboard/v1/tensorboard.proto

package tensorboardv1

import (
	containerv1 "github.com/determined-ai/determined/proto/pkg/containerv1"
	taskv1 "github.com/determined-ai/determined/proto/pkg/taskv1"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options"
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

// Tensorboard is a tensorboard instance in a containerized environment.
type Tensorboard struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The id of the tensorboard.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// The description of the tensorboard.
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	// The state of the tensorboard.
	State taskv1.State `protobuf:"varint,3,opt,name=state,proto3,enum=determined.task.v1.State" json:"state,omitempty"`
	// The time the tensorboard was started.
	StartTime *timestamp.Timestamp `protobuf:"bytes,4,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	// The container running the tensorboard.
	Container *containerv1.Container `protobuf:"bytes,6,opt,name=container,proto3" json:"container,omitempty"`
	// The experiment ids loaded into this tensorboard instance.
	ExperimentIds []int32 `protobuf:"varint,7,rep,packed,name=experiment_ids,json=experimentIds,proto3" json:"experiment_ids,omitempty"`
	// The trial ids loaded into this tensorboard instance.
	TrialIds []int32 `protobuf:"varint,8,rep,packed,name=trial_ids,json=trialIds,proto3" json:"trial_ids,omitempty"`
	// The display name of the user that created the tensorboard.
	DisplayName string `protobuf:"bytes,15,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	// The id of the user that created the tensorboard.
	UserId int32 `protobuf:"varint,16,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	// The username of the user that created the tensorboard.
	Username string `protobuf:"bytes,10,opt,name=username,proto3" json:"username,omitempty"`
	// The service address.
	ServiceAddress string `protobuf:"bytes,11,opt,name=service_address,json=serviceAddress,proto3" json:"service_address,omitempty"`
	// The name of the resource pool the command was created in
	ResourcePool string `protobuf:"bytes,12,opt,name=resource_pool,json=resourcePool,proto3" json:"resource_pool,omitempty"`
	// The exit status;
	ExitStatus string `protobuf:"bytes,13,opt,name=exit_status,json=exitStatus,proto3" json:"exit_status,omitempty"`
	// The associated job id.
	JobId string `protobuf:"bytes,14,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
}

func (x *Tensorboard) Reset() {
	*x = Tensorboard{}
	if protoimpl.UnsafeEnabled {
		mi := &file_determined_tensorboard_v1_tensorboard_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tensorboard) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tensorboard) ProtoMessage() {}

func (x *Tensorboard) ProtoReflect() protoreflect.Message {
	mi := &file_determined_tensorboard_v1_tensorboard_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tensorboard.ProtoReflect.Descriptor instead.
func (*Tensorboard) Descriptor() ([]byte, []int) {
	return file_determined_tensorboard_v1_tensorboard_proto_rawDescGZIP(), []int{0}
}

func (x *Tensorboard) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Tensorboard) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Tensorboard) GetState() taskv1.State {
	if x != nil {
		return x.State
	}
	return taskv1.State_STATE_UNSPECIFIED
}

func (x *Tensorboard) GetStartTime() *timestamp.Timestamp {
	if x != nil {
		return x.StartTime
	}
	return nil
}

func (x *Tensorboard) GetContainer() *containerv1.Container {
	if x != nil {
		return x.Container
	}
	return nil
}

func (x *Tensorboard) GetExperimentIds() []int32 {
	if x != nil {
		return x.ExperimentIds
	}
	return nil
}

func (x *Tensorboard) GetTrialIds() []int32 {
	if x != nil {
		return x.TrialIds
	}
	return nil
}

func (x *Tensorboard) GetDisplayName() string {
	if x != nil {
		return x.DisplayName
	}
	return ""
}

func (x *Tensorboard) GetUserId() int32 {
	if x != nil {
		return x.UserId
	}
	return 0
}

func (x *Tensorboard) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *Tensorboard) GetServiceAddress() string {
	if x != nil {
		return x.ServiceAddress
	}
	return ""
}

func (x *Tensorboard) GetResourcePool() string {
	if x != nil {
		return x.ResourcePool
	}
	return ""
}

func (x *Tensorboard) GetExitStatus() string {
	if x != nil {
		return x.ExitStatus
	}
	return ""
}

func (x *Tensorboard) GetJobId() string {
	if x != nil {
		return x.JobId
	}
	return ""
}

var File_determined_tensorboard_v1_tensorboard_proto protoreflect.FileDescriptor

var file_determined_tensorboard_v1_tensorboard_proto_rawDesc = []byte{
	0x0a, 0x2b, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2f, 0x74, 0x65, 0x6e,
	0x73, 0x6f, 0x72, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x65, 0x6e, 0x73,
	0x6f, 0x72, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x19, 0x64,
	0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72,
	0x62, 0x6f, 0x61, 0x72, 0x64, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2c, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x73, 0x77, 0x61, 0x67, 0x67, 0x65, 0x72, 0x2f, 0x6f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x27, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69,
	0x6e, 0x65, 0x64, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x2f, 0x76, 0x31,
	0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1d, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2f, 0x74, 0x61, 0x73,
	0x6b, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x61, 0x73, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0xe2, 0x04, 0x0a, 0x0b, 0x54, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x2f, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x19, 0x2e, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e, 0x74, 0x61,
	0x73, 0x6b, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x12, 0x39, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x40, 0x0a,
	0x09, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x22, 0x2e, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e, 0x63, 0x6f,
	0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x61,
	0x69, 0x6e, 0x65, 0x72, 0x52, 0x09, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x12,
	0x25, 0x0a, 0x0e, 0x65, 0x78, 0x70, 0x65, 0x72, 0x69, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64,
	0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x05, 0x52, 0x0d, 0x65, 0x78, 0x70, 0x65, 0x72, 0x69, 0x6d,
	0x65, 0x6e, 0x74, 0x49, 0x64, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x74, 0x72, 0x69, 0x61, 0x6c, 0x5f,
	0x69, 0x64, 0x73, 0x18, 0x08, 0x20, 0x03, 0x28, 0x05, 0x52, 0x08, 0x74, 0x72, 0x69, 0x61, 0x6c,
	0x49, 0x64, 0x73, 0x12, 0x21, 0x0a, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x5f, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x69, 0x73, 0x70, 0x6c,
	0x61, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69,
	0x64, 0x18, 0x10, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12,
	0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x27, 0x0a, 0x0f, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x0b,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x41, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x12, 0x23, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x5f, 0x70, 0x6f, 0x6f, 0x6c, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x72, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x50, 0x6f, 0x6f, 0x6c, 0x12, 0x1f, 0x0a, 0x0b, 0x65, 0x78, 0x69,
	0x74, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x65, 0x78, 0x69, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x15, 0x0a, 0x06, 0x6a, 0x6f,
	0x62, 0x5f, 0x69, 0x64, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6a, 0x6f, 0x62, 0x49,
	0x64, 0x3a, 0x51, 0x92, 0x41, 0x4e, 0x0a, 0x4c, 0xd2, 0x01, 0x02, 0x69, 0x64, 0xd2, 0x01, 0x0b,
	0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0xd2, 0x01, 0x0a, 0x73, 0x74,
	0x61, 0x72, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0xd2, 0x01, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65,
	0xd2, 0x01, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0xd2, 0x01, 0x06, 0x6a, 0x6f,
	0x62, 0x5f, 0x69, 0x64, 0xd2, 0x01, 0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f,
	0x70, 0x6f, 0x6f, 0x6c, 0x42, 0x3d, 0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2d, 0x61, 0x69,
	0x2f, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x62, 0x6f, 0x61, 0x72,
	0x64, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_determined_tensorboard_v1_tensorboard_proto_rawDescOnce sync.Once
	file_determined_tensorboard_v1_tensorboard_proto_rawDescData = file_determined_tensorboard_v1_tensorboard_proto_rawDesc
)

func file_determined_tensorboard_v1_tensorboard_proto_rawDescGZIP() []byte {
	file_determined_tensorboard_v1_tensorboard_proto_rawDescOnce.Do(func() {
		file_determined_tensorboard_v1_tensorboard_proto_rawDescData = protoimpl.X.CompressGZIP(file_determined_tensorboard_v1_tensorboard_proto_rawDescData)
	})
	return file_determined_tensorboard_v1_tensorboard_proto_rawDescData
}

var file_determined_tensorboard_v1_tensorboard_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_determined_tensorboard_v1_tensorboard_proto_goTypes = []interface{}{
	(*Tensorboard)(nil),           // 0: determined.tensorboard.v1.Tensorboard
	(taskv1.State)(0),             // 1: determined.task.v1.State
	(*timestamp.Timestamp)(nil),   // 2: google.protobuf.Timestamp
	(*containerv1.Container)(nil), // 3: determined.container.v1.Container
}
var file_determined_tensorboard_v1_tensorboard_proto_depIdxs = []int32{
	1, // 0: determined.tensorboard.v1.Tensorboard.state:type_name -> determined.task.v1.State
	2, // 1: determined.tensorboard.v1.Tensorboard.start_time:type_name -> google.protobuf.Timestamp
	3, // 2: determined.tensorboard.v1.Tensorboard.container:type_name -> determined.container.v1.Container
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_determined_tensorboard_v1_tensorboard_proto_init() }
func file_determined_tensorboard_v1_tensorboard_proto_init() {
	if File_determined_tensorboard_v1_tensorboard_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_determined_tensorboard_v1_tensorboard_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tensorboard); i {
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
			RawDescriptor: file_determined_tensorboard_v1_tensorboard_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_determined_tensorboard_v1_tensorboard_proto_goTypes,
		DependencyIndexes: file_determined_tensorboard_v1_tensorboard_proto_depIdxs,
		MessageInfos:      file_determined_tensorboard_v1_tensorboard_proto_msgTypes,
	}.Build()
	File_determined_tensorboard_v1_tensorboard_proto = out.File
	file_determined_tensorboard_v1_tensorboard_proto_rawDesc = nil
	file_determined_tensorboard_v1_tensorboard_proto_goTypes = nil
	file_determined_tensorboard_v1_tensorboard_proto_depIdxs = nil
}
