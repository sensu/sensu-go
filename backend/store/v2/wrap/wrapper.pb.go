// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/sensu/sensu-go/backend/store/v2/wrap/wrapper.proto

package wrap

import (
	bytes "bytes"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/golang/protobuf/proto"
	v2 "github.com/sensu/core/v2"
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

// Encoding is the serialization encoding of the wrapped value.
type Encoding int32

const (
	Encoding_json     Encoding = 0
	Encoding_protobuf Encoding = 1
)

var Encoding_name = map[int32]string{
	0: "json",
	1: "protobuf",
}

var Encoding_value = map[string]int32{
	"json":     0,
	"protobuf": 1,
}

func (x Encoding) String() string {
	return proto.EnumName(Encoding_name, int32(x))
}

func (Encoding) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_0d211efcc0f41ca5, []int{0}
}

// Compression is the compression algorithm used to compress the wrapped
// value.
type Compression int32

const (
	Compression_none   Compression = 0
	Compression_snappy Compression = 1
)

var Compression_name = map[int32]string{
	0: "none",
	1: "snappy",
}

var Compression_value = map[string]int32{
	"none":   0,
	"snappy": 1,
}

func (x Compression) String() string {
	return proto.EnumName(Compression_name, int32(x))
}

func (Compression) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_0d211efcc0f41ca5, []int{1}
}

// Wrapper represents a serialized resource for storage purposes.
type Wrapper struct {
	// TypeMeta contains the type and the API version of the resource.
	TypeMeta *v2.TypeMeta `protobuf:"bytes,1,opt,name=TypeMeta,proto3" json:"TypeMeta,omitempty"`
	// Encoding is the type of serialization used.
	Encoding Encoding `protobuf:"varint,2,opt,name=encoding,proto3,enum=backend.store.wrap.Encoding" json:"encoding,omitempty"`
	// Compression is the type of compression used.
	Compression Compression `protobuf:"varint,3,opt,name=compression,proto3,enum=backend.store.wrap.Compression" json:"compression,omitempty"`
	// Value contains the encoded resource value
	Value                []byte   `protobuf:"bytes,4,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Wrapper) Reset()         { *m = Wrapper{} }
func (m *Wrapper) String() string { return proto.CompactTextString(m) }
func (*Wrapper) ProtoMessage()    {}
func (*Wrapper) Descriptor() ([]byte, []int) {
	return fileDescriptor_0d211efcc0f41ca5, []int{0}
}
func (m *Wrapper) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Wrapper) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Wrapper.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Wrapper) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Wrapper.Merge(m, src)
}
func (m *Wrapper) XXX_Size() int {
	return m.Size()
}
func (m *Wrapper) XXX_DiscardUnknown() {
	xxx_messageInfo_Wrapper.DiscardUnknown(m)
}

var xxx_messageInfo_Wrapper proto.InternalMessageInfo

func (m *Wrapper) GetTypeMeta() *v2.TypeMeta {
	if m != nil {
		return m.TypeMeta
	}
	return nil
}

func (m *Wrapper) GetEncoding() Encoding {
	if m != nil {
		return m.Encoding
	}
	return Encoding_json
}

func (m *Wrapper) GetCompression() Compression {
	if m != nil {
		return m.Compression
	}
	return Compression_none
}

func (m *Wrapper) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func init() {
	proto.RegisterEnum("backend.store.wrap.Encoding", Encoding_name, Encoding_value)
	proto.RegisterEnum("backend.store.wrap.Compression", Compression_name, Compression_value)
	proto.RegisterType((*Wrapper)(nil), "backend.store.wrap.Wrapper")
}

func init() {
	proto.RegisterFile("github.com/sensu/sensu-go/backend/store/v2/wrap/wrapper.proto", fileDescriptor_0d211efcc0f41ca5)
}

var fileDescriptor_0d211efcc0f41ca5 = []byte{
	// 344 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x90, 0x41, 0x4b, 0xfb, 0x30,
	0x18, 0xc6, 0x97, 0xff, 0xbf, 0xce, 0x92, 0x0d, 0x29, 0x41, 0xb0, 0x0c, 0xc9, 0xc6, 0xbc, 0x8c,
	0x81, 0x89, 0xeb, 0x3c, 0xe8, 0x41, 0xd0, 0x89, 0x47, 0x2f, 0x45, 0x10, 0xbc, 0xa5, 0x5d, 0xac,
	0x55, 0x97, 0x84, 0xa6, 0xad, 0xec, 0x9b, 0xf8, 0x11, 0xfc, 0x28, 0x3b, 0x7a, 0xf3, 0x26, 0x5a,
	0xbf, 0x84, 0x47, 0x69, 0xba, 0xcd, 0x81, 0x7a, 0x79, 0x49, 0xde, 0xf7, 0xf7, 0x3c, 0x79, 0xf2,
	0xc2, 0xa3, 0x28, 0x4e, 0x6f, 0xb2, 0x80, 0x84, 0x72, 0x42, 0x35, 0x17, 0x3a, 0xab, 0xea, 0x6e,
	0x24, 0x69, 0xc0, 0xc2, 0x3b, 0x2e, 0xc6, 0x54, 0xa7, 0x32, 0xe1, 0x34, 0xf7, 0xe8, 0x43, 0xc2,
	0x94, 0x29, 0x8a, 0x27, 0x44, 0x25, 0x32, 0x95, 0x08, 0xcd, 0x21, 0x62, 0x20, 0x52, 0x0e, 0x5b,
	0xfb, 0x2b, 0x96, 0x91, 0x8c, 0x24, 0x35, 0x68, 0x90, 0x5d, 0x1f, 0xe7, 0x03, 0x32, 0x24, 0x03,
	0xd3, 0x34, 0x3d, 0x73, 0xaa, 0x9c, 0x5a, 0x7b, 0x7f, 0x07, 0x61, 0x2a, 0xa6, 0xe1, 0x3c, 0xc3,
	0x84, 0xa7, 0xac, 0x52, 0x74, 0x5f, 0x00, 0x5c, 0xbf, 0xac, 0xd2, 0xa0, 0x43, 0x68, 0x5f, 0x4c,
	0x15, 0x3f, 0xe7, 0x29, 0x73, 0x41, 0x07, 0xf4, 0x1a, 0xde, 0x16, 0x31, 0x7a, 0x52, 0x0a, 0x49,
	0xee, 0x91, 0xc5, 0x78, 0x64, 0xcd, 0x5e, 0xdb, 0xc0, 0x5f, 0xe2, 0xe8, 0x00, 0xda, 0x5c, 0x84,
	0x72, 0x1c, 0x8b, 0xc8, 0xfd, 0xd7, 0x01, 0xbd, 0x0d, 0x6f, 0x9b, 0xfc, 0xfc, 0x15, 0x39, 0x9b,
	0x33, 0xfe, 0x92, 0x46, 0x27, 0xb0, 0x11, 0xca, 0x89, 0x4a, 0xb8, 0xd6, 0xb1, 0x14, 0xee, 0x7f,
	0x23, 0x6e, 0xff, 0x26, 0x3e, 0xfd, 0xc6, 0xfc, 0x55, 0x0d, 0xda, 0x84, 0x6b, 0x39, 0xbb, 0xcf,
	0xb8, 0x6b, 0x75, 0x40, 0xaf, 0xe9, 0x57, 0x97, 0x7e, 0x17, 0xda, 0x8b, 0xe7, 0x90, 0x0d, 0xad,
	0x5b, 0x2d, 0x85, 0x53, 0x43, 0x4d, 0x68, 0x2f, 0x36, 0xe9, 0x80, 0xfe, 0x0e, 0x6c, 0xac, 0xb8,
	0x96, 0x98, 0x90, 0x82, 0x3b, 0x35, 0x04, 0x61, 0x5d, 0x0b, 0xa6, 0xd4, 0xd4, 0x01, 0x23, 0xfc,
	0xf9, 0x8e, 0xc1, 0x53, 0x81, 0xc1, 0xac, 0xc0, 0xe0, 0xb9, 0xc0, 0xe0, 0xad, 0xc0, 0xe0, 0xf1,
	0x03, 0xd7, 0xae, 0xac, 0x32, 0x57, 0x50, 0x37, 0x86, 0xc3, 0xaf, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x92, 0x2b, 0xae, 0xaf, 0x06, 0x02, 0x00, 0x00,
}

func (this *Wrapper) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Wrapper)
	if !ok {
		that2, ok := that.(Wrapper)
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
	if !this.TypeMeta.Equal(that1.TypeMeta) {
		return false
	}
	if this.Encoding != that1.Encoding {
		return false
	}
	if this.Compression != that1.Compression {
		return false
	}
	if !bytes.Equal(this.Value, that1.Value) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (m *Wrapper) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Wrapper) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Wrapper) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Value) > 0 {
		i -= len(m.Value)
		copy(dAtA[i:], m.Value)
		i = encodeVarintWrapper(dAtA, i, uint64(len(m.Value)))
		i--
		dAtA[i] = 0x22
	}
	if m.Compression != 0 {
		i = encodeVarintWrapper(dAtA, i, uint64(m.Compression))
		i--
		dAtA[i] = 0x18
	}
	if m.Encoding != 0 {
		i = encodeVarintWrapper(dAtA, i, uint64(m.Encoding))
		i--
		dAtA[i] = 0x10
	}
	if m.TypeMeta != nil {
		{
			size, err := m.TypeMeta.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintWrapper(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintWrapper(dAtA []byte, offset int, v uint64) int {
	offset -= sovWrapper(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func NewPopulatedWrapper(r randyWrapper, easy bool) *Wrapper {
	this := &Wrapper{}
	if r.Intn(5) != 0 {
		this.TypeMeta = v2.NewPopulatedTypeMeta(r, easy)
	}
	this.Encoding = Encoding([]int32{0, 1}[r.Intn(2)])
	this.Compression = Compression([]int32{0, 1}[r.Intn(2)])
	v1 := r.Intn(100)
	this.Value = make([]byte, v1)
	for i := 0; i < v1; i++ {
		this.Value[i] = byte(r.Intn(256))
	}
	if !easy && r.Intn(10) != 0 {
		this.XXX_unrecognized = randUnrecognizedWrapper(r, 5)
	}
	return this
}

type randyWrapper interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneWrapper(r randyWrapper) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringWrapper(r randyWrapper) string {
	v2 := r.Intn(100)
	tmps := make([]rune, v2)
	for i := 0; i < v2; i++ {
		tmps[i] = randUTF8RuneWrapper(r)
	}
	return string(tmps)
}
func randUnrecognizedWrapper(r randyWrapper, maxFieldNumber int) (dAtA []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		dAtA = randFieldWrapper(dAtA, r, fieldNumber, wire)
	}
	return dAtA
}
func randFieldWrapper(dAtA []byte, r randyWrapper, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		dAtA = encodeVarintPopulateWrapper(dAtA, uint64(key))
		v3 := r.Int63()
		if r.Intn(2) == 0 {
			v3 *= -1
		}
		dAtA = encodeVarintPopulateWrapper(dAtA, uint64(v3))
	case 1:
		dAtA = encodeVarintPopulateWrapper(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		dAtA = encodeVarintPopulateWrapper(dAtA, uint64(key))
		ll := r.Intn(100)
		dAtA = encodeVarintPopulateWrapper(dAtA, uint64(ll))
		for j := 0; j < ll; j++ {
			dAtA = append(dAtA, byte(r.Intn(256)))
		}
	default:
		dAtA = encodeVarintPopulateWrapper(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return dAtA
}
func encodeVarintPopulateWrapper(dAtA []byte, v uint64) []byte {
	for v >= 1<<7 {
		dAtA = append(dAtA, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	dAtA = append(dAtA, uint8(v))
	return dAtA
}
func (m *Wrapper) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.TypeMeta != nil {
		l = m.TypeMeta.Size()
		n += 1 + l + sovWrapper(uint64(l))
	}
	if m.Encoding != 0 {
		n += 1 + sovWrapper(uint64(m.Encoding))
	}
	if m.Compression != 0 {
		n += 1 + sovWrapper(uint64(m.Compression))
	}
	l = len(m.Value)
	if l > 0 {
		n += 1 + l + sovWrapper(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovWrapper(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozWrapper(x uint64) (n int) {
	return sovWrapper(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Wrapper) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowWrapper
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
			return fmt.Errorf("proto: Wrapper: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Wrapper: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TypeMeta", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWrapper
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
				return ErrInvalidLengthWrapper
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthWrapper
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.TypeMeta == nil {
				m.TypeMeta = &v2.TypeMeta{}
			}
			if err := m.TypeMeta.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Encoding", wireType)
			}
			m.Encoding = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWrapper
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Encoding |= Encoding(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Compression", wireType)
			}
			m.Compression = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWrapper
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Compression |= Compression(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowWrapper
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthWrapper
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthWrapper
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Value = append(m.Value[:0], dAtA[iNdEx:postIndex]...)
			if m.Value == nil {
				m.Value = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipWrapper(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthWrapper
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthWrapper
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
func skipWrapper(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowWrapper
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
					return 0, ErrIntOverflowWrapper
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
					return 0, ErrIntOverflowWrapper
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
				return 0, ErrInvalidLengthWrapper
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupWrapper
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthWrapper
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthWrapper        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowWrapper          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupWrapper = fmt.Errorf("proto: unexpected end of group")
)