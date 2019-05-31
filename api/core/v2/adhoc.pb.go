// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: adhoc.proto

package v2

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"

import bytes "bytes"

import github_com_golang_protobuf_proto "github.com/golang/protobuf/proto"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type AdhocRequest struct {
	// Subscriptions is the list of entity subscriptions.
	Subscriptions []string `protobuf:"bytes,2,rep,name=subscriptions" json:"subscriptions,omitempty"`
	// Creator is the author of the adhoc request.
	Creator string `protobuf:"bytes,3,opt,name=creator,proto3" json:"creator,omitempty"`
	// Reason is used to provide context to the request.
	Reason string `protobuf:"bytes,4,opt,name=reason,proto3" json:"reason,omitempty"`
	// Metadata contains the name, namespace, labels and annotations of the AdhocCheck
	ObjectMeta           `protobuf:"bytes,5,opt,name=metadata,embedded=metadata" json:"metadata,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AdhocRequest) Reset()         { *m = AdhocRequest{} }
func (m *AdhocRequest) String() string { return proto.CompactTextString(m) }
func (*AdhocRequest) ProtoMessage()    {}
func (*AdhocRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_adhoc_5d62d46459ea6563, []int{0}
}
func (m *AdhocRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *AdhocRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_AdhocRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *AdhocRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AdhocRequest.Merge(dst, src)
}
func (m *AdhocRequest) XXX_Size() int {
	return m.Size()
}
func (m *AdhocRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_AdhocRequest.DiscardUnknown(m)
}

var xxx_messageInfo_AdhocRequest proto.InternalMessageInfo

func init() {
	proto.RegisterType((*AdhocRequest)(nil), "sensu.core.v2.AdhocRequest")
}
func (this *AdhocRequest) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*AdhocRequest)
	if !ok {
		that2, ok := that.(AdhocRequest)
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
	if len(this.Subscriptions) != len(that1.Subscriptions) {
		return false
	}
	for i := range this.Subscriptions {
		if this.Subscriptions[i] != that1.Subscriptions[i] {
			return false
		}
	}
	if this.Creator != that1.Creator {
		return false
	}
	if this.Reason != that1.Reason {
		return false
	}
	if !this.ObjectMeta.Equal(&that1.ObjectMeta) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}

type AdhocRequestFace interface {
	Proto() github_com_golang_protobuf_proto.Message
	GetSubscriptions() []string
	GetCreator() string
	GetReason() string
	GetObjectMeta() ObjectMeta
}

func (this *AdhocRequest) Proto() github_com_golang_protobuf_proto.Message {
	return this
}

func (this *AdhocRequest) TestProto() github_com_golang_protobuf_proto.Message {
	return NewAdhocRequestFromFace(this)
}

func (this *AdhocRequest) GetSubscriptions() []string {
	return this.Subscriptions
}

func (this *AdhocRequest) GetCreator() string {
	return this.Creator
}

func (this *AdhocRequest) GetReason() string {
	return this.Reason
}

func (this *AdhocRequest) GetObjectMeta() ObjectMeta {
	return this.ObjectMeta
}

func NewAdhocRequestFromFace(that AdhocRequestFace) *AdhocRequest {
	this := &AdhocRequest{}
	this.Subscriptions = that.GetSubscriptions()
	this.Creator = that.GetCreator()
	this.Reason = that.GetReason()
	this.ObjectMeta = that.GetObjectMeta()
	return this
}

func (m *AdhocRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AdhocRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Subscriptions) > 0 {
		for _, s := range m.Subscriptions {
			dAtA[i] = 0x12
			i++
			l = len(s)
			for l >= 1<<7 {
				dAtA[i] = uint8(uint64(l)&0x7f | 0x80)
				l >>= 7
				i++
			}
			dAtA[i] = uint8(l)
			i++
			i += copy(dAtA[i:], s)
		}
	}
	if len(m.Creator) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintAdhoc(dAtA, i, uint64(len(m.Creator)))
		i += copy(dAtA[i:], m.Creator)
	}
	if len(m.Reason) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintAdhoc(dAtA, i, uint64(len(m.Reason)))
		i += copy(dAtA[i:], m.Reason)
	}
	dAtA[i] = 0x2a
	i++
	i = encodeVarintAdhoc(dAtA, i, uint64(m.ObjectMeta.Size()))
	n1, err := m.ObjectMeta.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func encodeVarintAdhoc(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func NewPopulatedAdhocRequest(r randyAdhoc, easy bool) *AdhocRequest {
	this := &AdhocRequest{}
	v1 := r.Intn(10)
	this.Subscriptions = make([]string, v1)
	for i := 0; i < v1; i++ {
		this.Subscriptions[i] = string(randStringAdhoc(r))
	}
	this.Creator = string(randStringAdhoc(r))
	this.Reason = string(randStringAdhoc(r))
	v2 := NewPopulatedObjectMeta(r, easy)
	this.ObjectMeta = *v2
	if !easy && r.Intn(10) != 0 {
		this.XXX_unrecognized = randUnrecognizedAdhoc(r, 6)
	}
	return this
}

type randyAdhoc interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneAdhoc(r randyAdhoc) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringAdhoc(r randyAdhoc) string {
	v3 := r.Intn(100)
	tmps := make([]rune, v3)
	for i := 0; i < v3; i++ {
		tmps[i] = randUTF8RuneAdhoc(r)
	}
	return string(tmps)
}
func randUnrecognizedAdhoc(r randyAdhoc, maxFieldNumber int) (dAtA []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		dAtA = randFieldAdhoc(dAtA, r, fieldNumber, wire)
	}
	return dAtA
}
func randFieldAdhoc(dAtA []byte, r randyAdhoc, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		dAtA = encodeVarintPopulateAdhoc(dAtA, uint64(key))
		v4 := r.Int63()
		if r.Intn(2) == 0 {
			v4 *= -1
		}
		dAtA = encodeVarintPopulateAdhoc(dAtA, uint64(v4))
	case 1:
		dAtA = encodeVarintPopulateAdhoc(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		dAtA = encodeVarintPopulateAdhoc(dAtA, uint64(key))
		ll := r.Intn(100)
		dAtA = encodeVarintPopulateAdhoc(dAtA, uint64(ll))
		for j := 0; j < ll; j++ {
			dAtA = append(dAtA, byte(r.Intn(256)))
		}
	default:
		dAtA = encodeVarintPopulateAdhoc(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return dAtA
}
func encodeVarintPopulateAdhoc(dAtA []byte, v uint64) []byte {
	for v >= 1<<7 {
		dAtA = append(dAtA, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	dAtA = append(dAtA, uint8(v))
	return dAtA
}
func (m *AdhocRequest) Size() (n int) {
	var l int
	_ = l
	if len(m.Subscriptions) > 0 {
		for _, s := range m.Subscriptions {
			l = len(s)
			n += 1 + l + sovAdhoc(uint64(l))
		}
	}
	l = len(m.Creator)
	if l > 0 {
		n += 1 + l + sovAdhoc(uint64(l))
	}
	l = len(m.Reason)
	if l > 0 {
		n += 1 + l + sovAdhoc(uint64(l))
	}
	l = m.ObjectMeta.Size()
	n += 1 + l + sovAdhoc(uint64(l))
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovAdhoc(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozAdhoc(x uint64) (n int) {
	return sovAdhoc(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *AdhocRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAdhoc
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: AdhocRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AdhocRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Subscriptions", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAdhoc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAdhoc
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Subscriptions = append(m.Subscriptions, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Creator", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAdhoc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAdhoc
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Creator = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Reason", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAdhoc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAdhoc
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Reason = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObjectMeta", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAdhoc
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAdhoc
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ObjectMeta.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAdhoc(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthAdhoc
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
func skipAdhoc(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowAdhoc
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
					return 0, ErrIntOverflowAdhoc
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAdhoc
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
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthAdhoc
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowAdhoc
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipAdhoc(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthAdhoc = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowAdhoc   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("adhoc.proto", fileDescriptor_adhoc_5d62d46459ea6563) }

var fileDescriptor_adhoc_5d62d46459ea6563 = []byte{
	// 284 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x50, 0x3d, 0x4e, 0xc3, 0x30,
	0x14, 0xce, 0x6b, 0x4b, 0x29, 0x2e, 0x5d, 0x3c, 0x85, 0x0a, 0xd9, 0x11, 0x53, 0x85, 0xc0, 0x95,
	0xc2, 0xc6, 0x46, 0x76, 0x84, 0x14, 0x89, 0x85, 0xcd, 0x71, 0x4d, 0x1a, 0xa4, 0xc4, 0xc1, 0x76,
	0x2a, 0x71, 0x03, 0x8e, 0xc0, 0xd8, 0xb1, 0x47, 0xe0, 0x08, 0x19, 0xbb, 0xb2, 0x44, 0x10, 0x36,
	0x4e, 0xc0, 0x88, 0x9a, 0xd2, 0xaa, 0xdd, 0xbe, 0x7f, 0xd9, 0x0f, 0xf5, 0xf9, 0x64, 0xaa, 0x04,
	0xcb, 0xb5, 0xb2, 0x0a, 0x0f, 0x8c, 0xcc, 0x4c, 0xc1, 0x84, 0xd2, 0x92, 0xcd, 0xfc, 0xe1, 0x65,
	0x9c, 0xd8, 0x69, 0x11, 0x31, 0xa1, 0xd2, 0x71, 0xac, 0x62, 0x35, 0x6e, 0x52, 0x51, 0xf1, 0xd8,
	0xb0, 0x86, 0x34, 0x68, 0xdd, 0x1e, 0xa2, 0x54, 0x5a, 0xbe, 0xc6, 0x67, 0x1f, 0x80, 0x8e, 0x6f,
	0x56, 0xcb, 0xa1, 0x7c, 0x2e, 0xa4, 0xb1, 0xf8, 0x1c, 0x0d, 0x4c, 0x11, 0x19, 0xa1, 0x93, 0xdc,
	0x26, 0x2a, 0x33, 0x6e, 0xcb, 0x6b, 0x8f, 0x8e, 0x82, 0x4e, 0x59, 0x51, 0x08, 0xf7, 0x2d, 0x4c,
	0xd0, 0xa1, 0xd0, 0x92, 0x5b, 0xa5, 0xdd, 0xb6, 0x07, 0xdb, 0xd4, 0x46, 0xc4, 0xa7, 0xa8, 0xab,
	0x25, 0x37, 0x2a, 0x73, 0x3b, 0x3b, 0xf6, 0xbf, 0x86, 0xef, 0x51, 0x6f, 0xf5, 0x90, 0x09, 0xb7,
	0xdc, 0x3d, 0xf0, 0x60, 0xd4, 0xf7, 0x4f, 0xd8, 0xde, 0xbf, 0xd8, 0x5d, 0xf4, 0x24, 0x85, 0xbd,
	0x95, 0x96, 0x07, 0xa4, 0xac, 0xa8, 0xb3, 0xac, 0x28, 0xfc, 0x54, 0x14, 0x6f, 0x6a, 0x17, 0x2a,
	0x4d, 0xac, 0x4c, 0x73, 0xfb, 0x12, 0x6e, 0xa7, 0xae, 0x7b, 0xaf, 0x73, 0xea, 0x2c, 0xe6, 0x14,
	0x02, 0xef, 0xf7, 0x8b, 0xc0, 0xa2, 0x26, 0xf0, 0x5e, 0x13, 0x28, 0x6b, 0x02, 0xcb, 0x9a, 0xc0,
	0x67, 0x4d, 0xe0, 0xed, 0x9b, 0x38, 0x0f, 0xad, 0x99, 0x1f, 0x75, 0x9b, 0x23, 0x5c, 0xfd, 0x05,
	0x00, 0x00, 0xff, 0xff, 0xe5, 0x80, 0x86, 0x90, 0x5d, 0x01, 0x00, 0x00,
}
