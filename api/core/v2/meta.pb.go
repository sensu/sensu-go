// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: meta.proto

package v2

import (
	bytes "bytes"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/golang/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// ObjectMeta is metadata all persisted objects have.
type ObjectMeta struct {
	// Name must be unique within a namespace. Name is primarily intended for creation
	// idempotence and configuration definition.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty" yaml: "name,omitempty"`
	// Namespace defines a logical grouping of objects within which each object name must
	// be unique.
	Namespace string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty" yaml: "namespace,omitempty"`
	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May also be used in filters and token
	// substitution.
	// TODO: Link to Sensu documentation.
	// More info: http://kubernetes.io/docs/user-guide/labels
	Labels map[string]string `protobuf:"bytes,3,rep,name=labels,proto3" json:"labels,omitempty" yaml: ",labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// TODO: Link to Sensu documentation.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	Annotations          map[string]string `protobuf:"bytes,4,rep,name=annotations,proto3" json:"annotations,omitempty" yaml: "annotations,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *ObjectMeta) Reset()         { *m = ObjectMeta{} }
func (m *ObjectMeta) String() string { return proto.CompactTextString(m) }
func (*ObjectMeta) ProtoMessage()    {}
func (*ObjectMeta) Descriptor() ([]byte, []int) {
	return fileDescriptor_3b5ea8fe65782bcc, []int{0}
}
func (m *ObjectMeta) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ObjectMeta) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ObjectMeta.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ObjectMeta) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ObjectMeta.Merge(m, src)
}
func (m *ObjectMeta) XXX_Size() int {
	return m.Size()
}
func (m *ObjectMeta) XXX_DiscardUnknown() {
	xxx_messageInfo_ObjectMeta.DiscardUnknown(m)
}

var xxx_messageInfo_ObjectMeta proto.InternalMessageInfo

func (m *ObjectMeta) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ObjectMeta) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
}

func (m *ObjectMeta) GetLabels() map[string]string {
	if m != nil {
		return m.Labels
	}
	return nil
}

func (m *ObjectMeta) GetAnnotations() map[string]string {
	if m != nil {
		return m.Annotations
	}
	return nil
}

// TypeMeta is information that can be used to resolve a data type
type TypeMeta struct {
	// Type is the type name of the data type
	Type string `protobuf:"bytes,1,opt,name=Type,proto3" json:"type" yaml: "type,omitempty"`
	// APIVersion is the APIVersion of the data type
	APIVersion           string   `protobuf:"bytes,2,opt,name=APIVersion,proto3" json:"api_version" yaml: "api_version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TypeMeta) Reset()         { *m = TypeMeta{} }
func (m *TypeMeta) String() string { return proto.CompactTextString(m) }
func (*TypeMeta) ProtoMessage()    {}
func (*TypeMeta) Descriptor() ([]byte, []int) {
	return fileDescriptor_3b5ea8fe65782bcc, []int{1}
}
func (m *TypeMeta) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TypeMeta) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TypeMeta.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TypeMeta) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TypeMeta.Merge(m, src)
}
func (m *TypeMeta) XXX_Size() int {
	return m.Size()
}
func (m *TypeMeta) XXX_DiscardUnknown() {
	xxx_messageInfo_TypeMeta.DiscardUnknown(m)
}

var xxx_messageInfo_TypeMeta proto.InternalMessageInfo

func (m *TypeMeta) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *TypeMeta) GetAPIVersion() string {
	if m != nil {
		return m.APIVersion
	}
	return ""
}

func init() {
	proto.RegisterType((*ObjectMeta)(nil), "sensu.core.v2.ObjectMeta")
	proto.RegisterMapType((map[string]string)(nil), "sensu.core.v2.ObjectMeta.AnnotationsEntry")
	proto.RegisterMapType((map[string]string)(nil), "sensu.core.v2.ObjectMeta.LabelsEntry")
	proto.RegisterType((*TypeMeta)(nil), "sensu.core.v2.TypeMeta")
}

func init() { proto.RegisterFile("meta.proto", fileDescriptor_3b5ea8fe65782bcc) }

var fileDescriptor_3b5ea8fe65782bcc = []byte{
	// 448 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x52, 0xcd, 0x8a, 0xd4, 0x40,
	0x10, 0xb6, 0x67, 0xe2, 0xe2, 0xd6, 0xa0, 0x0c, 0xed, 0x0f, 0xe3, 0x88, 0xc9, 0xd0, 0x20, 0x2c,
	0x32, 0xf4, 0x3a, 0xb3, 0xb2, 0xe8, 0x1c, 0x64, 0x77, 0xc0, 0x83, 0xa0, 0xb8, 0x84, 0xc1, 0x83,
	0x17, 0xe9, 0x84, 0x76, 0x8c, 0x26, 0xe9, 0x90, 0x74, 0x02, 0x79, 0x03, 0x1f, 0xc0, 0x83, 0x07,
	0x1f, 0xc0, 0x47, 0xf0, 0x11, 0x3c, 0xfa, 0x04, 0x41, 0xe3, 0x2d, 0x47, 0x4f, 0x1e, 0x25, 0x9d,
	0x60, 0x3a, 0x83, 0x1e, 0xf6, 0x94, 0xaa, 0xaf, 0xaa, 0xbe, 0xaf, 0xaa, 0xf3, 0x01, 0x04, 0x5c,
	0x32, 0x1a, 0xc5, 0x42, 0x0a, 0x7c, 0x39, 0xe1, 0x61, 0x92, 0x52, 0x57, 0xc4, 0x9c, 0x66, 0xcb,
	0xe9, 0xfd, 0xad, 0x27, 0xdf, 0xa4, 0x0e, 0x75, 0x45, 0x70, 0xb8, 0x15, 0x5b, 0x71, 0xa8, 0xba,
	0x9c, 0xf4, 0xf5, 0x49, 0xb6, 0xa0, 0x47, 0x74, 0xa1, 0x40, 0x85, 0xa9, 0xa8, 0x21, 0x21, 0x9f,
	0x0c, 0x80, 0xe7, 0xce, 0x5b, 0xee, 0xca, 0x67, 0x5c, 0x32, 0x7c, 0x02, 0x46, 0xc8, 0x02, 0x3e,
	0x41, 0x33, 0x74, 0xb0, 0xbf, 0x9e, 0x57, 0x85, 0x75, 0xa5, 0xce, 0xe7, 0x22, 0xf0, 0x24, 0x0f,
	0x22, 0x99, 0xff, 0x2a, 0xac, 0x1b, 0x39, 0x0b, 0xfc, 0xd5, 0x8c, 0xf4, 0x0b, 0xc4, 0x56, 0x93,
	0x78, 0x03, 0xfb, 0xf5, 0x37, 0x89, 0x98, 0xcb, 0x27, 0x03, 0x45, 0x73, 0x5c, 0x15, 0xd6, 0xd5,
	0xbf, 0x60, 0x8f, 0xeb, 0x96, 0xc6, 0xb5, 0x53, 0x25, 0x76, 0x47, 0x84, 0x23, 0xd8, 0xf3, 0x99,
	0xc3, 0xfd, 0x64, 0x32, 0x9c, 0x0d, 0x0f, 0x46, 0xcb, 0x3b, 0xb4, 0x77, 0x3c, 0xed, 0x4e, 0xa0,
	0x4f, 0x55, 0xdf, 0xe3, 0x50, 0xc6, 0xf9, 0x7a, 0x51, 0x15, 0xd6, 0xb8, 0x19, 0xec, 0xc9, 0xde,
	0x6c, 0x65, 0xe7, 0xbb, 0x35, 0x62, 0xb7, 0x3a, 0xf8, 0x3d, 0x82, 0x11, 0x0b, 0x43, 0x21, 0x99,
	0xf4, 0x44, 0x98, 0x4c, 0x0c, 0xa5, 0x7b, 0xf7, 0xff, 0xba, 0xa7, 0x5d, 0x73, 0x23, 0xbe, 0xaa,
	0x0a, 0xeb, 0xba, 0x46, 0xd1, 0xdb, 0xe0, 0x76, 0xbb, 0xc1, 0x3f, 0xeb, 0xc4, 0xd6, 0xa5, 0xa7,
	0x0f, 0x61, 0xa4, 0x1d, 0x85, 0xc7, 0x30, 0x7c, 0xc7, 0xf3, 0xe6, 0x17, 0xd9, 0x75, 0x88, 0xaf,
	0xc1, 0xc5, 0x8c, 0xf9, 0x69, 0xfb, 0xde, 0x76, 0x93, 0xac, 0x06, 0x0f, 0xd0, 0xf4, 0x11, 0x8c,
	0x77, 0xf7, 0x3a, 0xcf, 0x3c, 0xf9, 0x80, 0xe0, 0xd2, 0x26, 0x8f, 0xb8, 0x32, 0xc7, 0x31, 0x18,
	0x75, 0xdc, 0x9a, 0x83, 0x54, 0x85, 0x65, 0xc8, 0x3c, 0xe2, 0x9a, 0x25, 0xea, 0xb4, 0x67, 0x89,
	0xba, 0x1f, 0x9f, 0x01, 0x9c, 0x9e, 0x3d, 0x79, 0xc1, 0xe3, 0xc4, 0x13, 0x61, 0xeb, 0x89, 0x7b,
	0x55, 0x61, 0x8d, 0x58, 0xe4, 0xbd, 0xca, 0x1a, 0x58, 0x7f, 0x92, 0x0e, 0xd5, 0xb9, 0x34, 0x8e,
	0xf5, 0xec, 0xf7, 0x0f, 0x13, 0x7d, 0x2e, 0x4d, 0xf4, 0xa5, 0x34, 0xd1, 0xd7, 0xd2, 0x44, 0xdf,
	0x4a, 0x13, 0x7d, 0x2f, 0x4d, 0xf4, 0xf1, 0xa7, 0x79, 0xe1, 0xe5, 0x20, 0x5b, 0x3a, 0x7b, 0xca,
	0xde, 0x47, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff, 0x28, 0x89, 0x03, 0xe1, 0x31, 0x03, 0x00, 0x00,
}

func (this *ObjectMeta) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ObjectMeta)
	if !ok {
		that2, ok := that.(ObjectMeta)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Name != that1.Name {
		return false
	}
	if this.Namespace != that1.Namespace {
		return false
	}
	if len(this.Labels) != len(that1.Labels) {
		return false
	}
	for i := range this.Labels {
		if this.Labels[i] != that1.Labels[i] {
			return false
		}
	}
	if len(this.Annotations) != len(that1.Annotations) {
		return false
	}
	for i := range this.Annotations {
		if this.Annotations[i] != that1.Annotations[i] {
			return false
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *TypeMeta) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TypeMeta)
	if !ok {
		that2, ok := that.(TypeMeta)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Type != that1.Type {
		return false
	}
	if this.APIVersion != that1.APIVersion {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (m *ObjectMeta) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ObjectMeta) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ObjectMeta) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Annotations) > 0 {
		for k := range m.Annotations {
			v := m.Annotations[k]
			baseI := i
			i -= len(v)
			copy(dAtA[i:], v)
			i = encodeVarintMeta(dAtA, i, uint64(len(v)))
			i--
			dAtA[i] = 0x12
			i -= len(k)
			copy(dAtA[i:], k)
			i = encodeVarintMeta(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = encodeVarintMeta(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.Labels) > 0 {
		for k := range m.Labels {
			v := m.Labels[k]
			baseI := i
			i -= len(v)
			copy(dAtA[i:], v)
			i = encodeVarintMeta(dAtA, i, uint64(len(v)))
			i--
			dAtA[i] = 0x12
			i -= len(k)
			copy(dAtA[i:], k)
			i = encodeVarintMeta(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = encodeVarintMeta(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.Namespace) > 0 {
		i -= len(m.Namespace)
		copy(dAtA[i:], m.Namespace)
		i = encodeVarintMeta(dAtA, i, uint64(len(m.Namespace)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintMeta(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *TypeMeta) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TypeMeta) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TypeMeta) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.APIVersion) > 0 {
		i -= len(m.APIVersion)
		copy(dAtA[i:], m.APIVersion)
		i = encodeVarintMeta(dAtA, i, uint64(len(m.APIVersion)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Type) > 0 {
		i -= len(m.Type)
		copy(dAtA[i:], m.Type)
		i = encodeVarintMeta(dAtA, i, uint64(len(m.Type)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintMeta(dAtA []byte, offset int, v uint64) int {
	offset -= sovMeta(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func NewPopulatedObjectMeta(r randyMeta, easy bool) *ObjectMeta {
	this := &ObjectMeta{}
	this.Name = string(randStringMeta(r))
	this.Namespace = string(randStringMeta(r))
	if r.Intn(5) != 0 {
		v1 := r.Intn(10)
		this.Labels = make(map[string]string)
		for i := 0; i < v1; i++ {
			this.Labels[randStringMeta(r)] = randStringMeta(r)
		}
	}
	if r.Intn(5) != 0 {
		v2 := r.Intn(10)
		this.Annotations = make(map[string]string)
		for i := 0; i < v2; i++ {
			this.Annotations[randStringMeta(r)] = randStringMeta(r)
		}
	}
	if !easy && r.Intn(10) != 0 {
		this.XXX_unrecognized = randUnrecognizedMeta(r, 5)
	}
	return this
}

func NewPopulatedTypeMeta(r randyMeta, easy bool) *TypeMeta {
	this := &TypeMeta{}
	this.Type = string(randStringMeta(r))
	this.APIVersion = string(randStringMeta(r))
	if !easy && r.Intn(10) != 0 {
		this.XXX_unrecognized = randUnrecognizedMeta(r, 3)
	}
	return this
}

type randyMeta interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneMeta(r randyMeta) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringMeta(r randyMeta) string {
	v3 := r.Intn(100)
	tmps := make([]rune, v3)
	for i := 0; i < v3; i++ {
		tmps[i] = randUTF8RuneMeta(r)
	}
	return string(tmps)
}
func randUnrecognizedMeta(r randyMeta, maxFieldNumber int) (dAtA []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		dAtA = randFieldMeta(dAtA, r, fieldNumber, wire)
	}
	return dAtA
}
func randFieldMeta(dAtA []byte, r randyMeta, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		dAtA = encodeVarintPopulateMeta(dAtA, uint64(key))
		v4 := r.Int63()
		if r.Intn(2) == 0 {
			v4 *= -1
		}
		dAtA = encodeVarintPopulateMeta(dAtA, uint64(v4))
	case 1:
		dAtA = encodeVarintPopulateMeta(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		dAtA = encodeVarintPopulateMeta(dAtA, uint64(key))
		ll := r.Intn(100)
		dAtA = encodeVarintPopulateMeta(dAtA, uint64(ll))
		for j := 0; j < ll; j++ {
			dAtA = append(dAtA, byte(r.Intn(256)))
		}
	default:
		dAtA = encodeVarintPopulateMeta(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return dAtA
}
func encodeVarintPopulateMeta(dAtA []byte, v uint64) []byte {
	for v >= 1<<7 {
		dAtA = append(dAtA, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	dAtA = append(dAtA, uint8(v))
	return dAtA
}
func (m *ObjectMeta) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovMeta(uint64(l))
	}
	l = len(m.Namespace)
	if l > 0 {
		n += 1 + l + sovMeta(uint64(l))
	}
	if len(m.Labels) > 0 {
		for k, v := range m.Labels {
			_ = k
			_ = v
			mapEntrySize := 1 + len(k) + sovMeta(uint64(len(k))) + 1 + len(v) + sovMeta(uint64(len(v)))
			n += mapEntrySize + 1 + sovMeta(uint64(mapEntrySize))
		}
	}
	if len(m.Annotations) > 0 {
		for k, v := range m.Annotations {
			_ = k
			_ = v
			mapEntrySize := 1 + len(k) + sovMeta(uint64(len(k))) + 1 + len(v) + sovMeta(uint64(len(v)))
			n += mapEntrySize + 1 + sovMeta(uint64(mapEntrySize))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *TypeMeta) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Type)
	if l > 0 {
		n += 1 + l + sovMeta(uint64(l))
	}
	l = len(m.APIVersion)
	if l > 0 {
		n += 1 + l + sovMeta(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovMeta(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMeta(x uint64) (n int) {
	return sovMeta(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ObjectMeta) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMeta
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ObjectMeta: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ObjectMeta: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeta
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMeta
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMeta
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Namespace", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeta
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMeta
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMeta
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Namespace = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Labels", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeta
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMeta
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMeta
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Labels == nil {
				m.Labels = make(map[string]string)
			}
			var mapkey string
			var mapvalue string
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowMeta
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					wire |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowMeta
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthMeta
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return ErrInvalidLengthMeta
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var stringLenmapvalue uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowMeta
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapvalue |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapvalue := int(stringLenmapvalue)
					if intStringLenmapvalue < 0 {
						return ErrInvalidLengthMeta
					}
					postStringIndexmapvalue := iNdEx + intStringLenmapvalue
					if postStringIndexmapvalue < 0 {
						return ErrInvalidLengthMeta
					}
					if postStringIndexmapvalue > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = string(dAtA[iNdEx:postStringIndexmapvalue])
					iNdEx = postStringIndexmapvalue
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipMeta(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if skippy < 0 {
						return ErrInvalidLengthMeta
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.Labels[mapkey] = mapvalue
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Annotations", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeta
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMeta
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMeta
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Annotations == nil {
				m.Annotations = make(map[string]string)
			}
			var mapkey string
			var mapvalue string
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowMeta
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					wire |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowMeta
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthMeta
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return ErrInvalidLengthMeta
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var stringLenmapvalue uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowMeta
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapvalue |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapvalue := int(stringLenmapvalue)
					if intStringLenmapvalue < 0 {
						return ErrInvalidLengthMeta
					}
					postStringIndexmapvalue := iNdEx + intStringLenmapvalue
					if postStringIndexmapvalue < 0 {
						return ErrInvalidLengthMeta
					}
					if postStringIndexmapvalue > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = string(dAtA[iNdEx:postStringIndexmapvalue])
					iNdEx = postStringIndexmapvalue
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipMeta(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if skippy < 0 {
						return ErrInvalidLengthMeta
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.Annotations[mapkey] = mapvalue
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMeta(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMeta
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthMeta
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *TypeMeta) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMeta
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TypeMeta: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TypeMeta: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeta
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMeta
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMeta
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Type = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field APIVersion", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeta
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMeta
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMeta
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.APIVersion = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMeta(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMeta
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthMeta
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipMeta(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMeta
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMeta
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMeta
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthMeta
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMeta
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMeta
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMeta        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMeta          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMeta = fmt.Errorf("proto: unexpected end of group")
)
