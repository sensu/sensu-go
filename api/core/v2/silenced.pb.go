// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/sensu/sensu-go/api/core/v2/silenced.proto

package v2

import (
	bytes "bytes"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	github_com_golang_protobuf_proto "github.com/golang/protobuf/proto"
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

// Silenced is the representation of a silence entry.
type Silenced struct {
	// Metadata contains the name, namespace, labels and annotations of the
	// silenced
	ObjectMeta `protobuf:"bytes,1,opt,name=metadata,proto3,embedded=metadata" json:"metadata,omitempty"`
	// Expire is the number of seconds the entry will live
	Expire int64 `protobuf:"varint,2,opt,name=expire,proto3" json:"expire"`
	// ExpireOnResolve defaults to false, clears the entry on resolution when
	// set to true
	ExpireOnResolve bool `protobuf:"varint,3,opt,name=expire_on_resolve,json=expireOnResolve,proto3" json:"expire_on_resolve"`
	// Creator is the author of the silenced entry
	Creator string `protobuf:"bytes,4,opt,name=creator,proto3" json:"creator,omitempty"`
	// Check is the name of the check event to be silenced.
	Check string `protobuf:"bytes,5,opt,name=check,proto3" json:"check,omitempty"`
	// Reason is used to provide context to the entry
	Reason string `protobuf:"bytes,6,opt,name=reason,proto3" json:"reason,omitempty"`
	// Subscription is the name of the subscription to which the entry applies.
	Subscription string `protobuf:"bytes,7,opt,name=subscription,proto3" json:"subscription,omitempty"`
	// Begin is a timestamp at which the silenced entry takes effect.
	Begin int64 `protobuf:"varint,10,opt,name=begin,proto3" json:"begin"`
	// ExpireAt is a timestamp at which the silenced entry will expire.
	ExpireAt             int64    `protobuf:"varint,11,opt,name=expire_at,json=expireAt,proto3" json:"expire_at"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Silenced) Reset()         { *m = Silenced{} }
func (m *Silenced) String() string { return proto.CompactTextString(m) }
func (*Silenced) ProtoMessage()    {}
func (*Silenced) Descriptor() ([]byte, []int) {
	return fileDescriptor_768e9755d200fb5d, []int{0}
}
func (m *Silenced) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Silenced) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Silenced.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Silenced) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Silenced.Merge(m, src)
}
func (m *Silenced) XXX_Size() int {
	return m.Size()
}
func (m *Silenced) XXX_DiscardUnknown() {
	xxx_messageInfo_Silenced.DiscardUnknown(m)
}

var xxx_messageInfo_Silenced proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Silenced)(nil), "sensu.core.v2.Silenced")
}

func init() {
	proto.RegisterFile("github.com/sensu/sensu-go/api/core/v2/silenced.proto", fileDescriptor_768e9755d200fb5d)
}

var fileDescriptor_768e9755d200fb5d = []byte{
	// 420 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x91, 0x31, 0x6e, 0xd4, 0x40,
	0x14, 0x86, 0x77, 0xb2, 0xd9, 0x8d, 0x77, 0x42, 0x40, 0x8c, 0x28, 0x26, 0x11, 0x1a, 0x5b, 0x14,
	0xc8, 0x42, 0x60, 0x93, 0x0d, 0x15, 0x15, 0x58, 0xa2, 0x44, 0x91, 0x06, 0xd1, 0xd0, 0x44, 0xe3,
	0xc9, 0xc3, 0x19, 0xc8, 0x7a, 0xac, 0xf1, 0xac, 0x05, 0x37, 0xe0, 0x08, 0x94, 0x69, 0x90, 0x72,
	0x04, 0x8e, 0xb0, 0x65, 0x4e, 0x60, 0x81, 0xe9, 0xf6, 0x04, 0x94, 0x68, 0x67, 0x1c, 0xd8, 0x15,
	0x4d, 0x1a, 0xbf, 0xe7, 0xef, 0xff, 0xff, 0xe7, 0xf7, 0x64, 0xfc, 0xac, 0x50, 0xf6, 0x6c, 0x9e,
	0x27, 0x52, 0xcf, 0xd2, 0x1a, 0xca, 0x7a, 0xee, 0x9f, 0x4f, 0x0a, 0x9d, 0x8a, 0x4a, 0xa5, 0x52,
	0x1b, 0x48, 0x9b, 0x69, 0x5a, 0xab, 0x73, 0x28, 0x25, 0x9c, 0x26, 0x95, 0xd1, 0x56, 0x93, 0x3d,
	0x67, 0x4a, 0x56, 0x6a, 0xd2, 0x4c, 0x0f, 0xd6, 0x87, 0x14, 0xba, 0xd0, 0xa9, 0x73, 0xe5, 0xf3,
	0xf7, 0x2f, 0x9a, 0xc3, 0xe4, 0x28, 0x39, 0x74, 0xd0, 0x31, 0xd7, 0xf9, 0x21, 0x07, 0x4f, 0x6f,
	0xf6, 0xe9, 0x19, 0x58, 0xe1, 0x13, 0x0f, 0xbe, 0x0d, 0x71, 0xf0, 0xa6, 0xdf, 0x84, 0xbc, 0xc5,
	0xc1, 0x4a, 0x3a, 0x15, 0x56, 0x50, 0x14, 0xa1, 0x78, 0x77, 0xba, 0x9f, 0x6c, 0xac, 0x95, 0x1c,
	0xe7, 0x1f, 0x40, 0xda, 0xd7, 0x60, 0x45, 0xc6, 0x16, 0x6d, 0x38, 0xb8, 0x6a, 0x43, 0xb4, 0x6c,
	0x43, 0x72, 0x1d, 0x7b, 0xac, 0x67, 0xca, 0xc2, 0xac, 0xb2, 0x9f, 0xf9, 0xdf, 0x51, 0xe4, 0x21,
	0x1e, 0xc3, 0xa7, 0x4a, 0x19, 0xa0, 0x5b, 0x11, 0x8a, 0x87, 0xd9, 0xed, 0x85, 0x4f, 0xf5, 0x94,
	0xf7, 0x95, 0xbc, 0xc2, 0x77, 0x7d, 0x77, 0xa2, 0xcb, 0x13, 0x03, 0xb5, 0x3e, 0x6f, 0x80, 0x0e,
	0x23, 0x14, 0x07, 0xd9, 0x7e, 0x1f, 0xf9, 0xdf, 0xc0, 0xef, 0x78, 0x74, 0x5c, 0x72, 0x0f, 0x08,
	0xc3, 0x3b, 0xd2, 0x80, 0xb0, 0xda, 0xd0, 0xed, 0x08, 0xc5, 0x93, 0x6c, 0x7b, 0x15, 0xe6, 0xd7,
	0x90, 0xdc, 0xc3, 0x23, 0x79, 0x06, 0xf2, 0x23, 0x1d, 0xad, 0x54, 0xee, 0x5f, 0xc8, 0x7d, 0x3c,
	0x36, 0x20, 0x6a, 0x5d, 0xd2, 0xf1, 0x5a, 0xa8, 0x67, 0x24, 0xc6, 0xb7, 0xea, 0x79, 0x5e, 0x4b,
	0xa3, 0x2a, 0xab, 0x74, 0x49, 0x77, 0xd6, 0x3c, 0x1b, 0x0a, 0x09, 0xf1, 0x28, 0x87, 0x42, 0x95,
	0x14, 0xbb, 0x5b, 0x27, 0xcb, 0x36, 0xf4, 0x80, 0xfb, 0x42, 0x1e, 0xe1, 0x49, 0x7f, 0x84, 0xb0,
	0x74, 0xd7, 0x99, 0xf6, 0x96, 0x6d, 0xf8, 0x0f, 0xf2, 0xc0, 0xb7, 0x2f, 0xed, 0xf3, 0xe0, 0xcb,
	0x45, 0x38, 0xb8, 0xbc, 0x08, 0x51, 0x16, 0xfd, 0xfe, 0xc9, 0xd0, 0x65, 0xc7, 0xd0, 0xf7, 0x8e,
	0xa1, 0x45, 0xc7, 0xd0, 0x55, 0xc7, 0xd0, 0x8f, 0x8e, 0xa1, 0xaf, 0xbf, 0xd8, 0xe0, 0xdd, 0x56,
	0x33, 0xcd, 0xc7, 0xee, 0x87, 0x1e, 0xfd, 0x09, 0x00, 0x00, 0xff, 0xff, 0x27, 0x07, 0xa7, 0x6c,
	0x7f, 0x02, 0x00, 0x00,
}

func (this *Silenced) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Silenced)
	if !ok {
		that2, ok := that.(Silenced)
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
	if !this.ObjectMeta.Equal(&that1.ObjectMeta) {
		return false
	}
	if this.Expire != that1.Expire {
		return false
	}
	if this.ExpireOnResolve != that1.ExpireOnResolve {
		return false
	}
	if this.Creator != that1.Creator {
		return false
	}
	if this.Check != that1.Check {
		return false
	}
	if this.Reason != that1.Reason {
		return false
	}
	if this.Subscription != that1.Subscription {
		return false
	}
	if this.Begin != that1.Begin {
		return false
	}
	if this.ExpireAt != that1.ExpireAt {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}

type SilencedFace interface {
	Proto() github_com_golang_protobuf_proto.Message
	GetObjectMeta() ObjectMeta
	GetExpire() int64
	GetExpireOnResolve() bool
	GetCreator() string
	GetCheck() string
	GetReason() string
	GetSubscription() string
	GetBegin() int64
	GetExpireAt() int64
}

func (this *Silenced) Proto() github_com_golang_protobuf_proto.Message {
	return this
}

func (this *Silenced) TestProto() github_com_golang_protobuf_proto.Message {
	return NewSilencedFromFace(this)
}

func (this *Silenced) GetObjectMeta() ObjectMeta {
	return this.ObjectMeta
}

func (this *Silenced) GetExpire() int64 {
	return this.Expire
}

func (this *Silenced) GetExpireOnResolve() bool {
	return this.ExpireOnResolve
}

func (this *Silenced) GetCreator() string {
	return this.Creator
}

func (this *Silenced) GetCheck() string {
	return this.Check
}

func (this *Silenced) GetReason() string {
	return this.Reason
}

func (this *Silenced) GetSubscription() string {
	return this.Subscription
}

func (this *Silenced) GetBegin() int64 {
	return this.Begin
}

func (this *Silenced) GetExpireAt() int64 {
	return this.ExpireAt
}

func NewSilencedFromFace(that SilencedFace) *Silenced {
	this := &Silenced{}
	this.ObjectMeta = that.GetObjectMeta()
	this.Expire = that.GetExpire()
	this.ExpireOnResolve = that.GetExpireOnResolve()
	this.Creator = that.GetCreator()
	this.Check = that.GetCheck()
	this.Reason = that.GetReason()
	this.Subscription = that.GetSubscription()
	this.Begin = that.GetBegin()
	this.ExpireAt = that.GetExpireAt()
	return this
}

func (m *Silenced) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Silenced) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Silenced) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.ExpireAt != 0 {
		i = encodeVarintSilenced(dAtA, i, uint64(m.ExpireAt))
		i--
		dAtA[i] = 0x58
	}
	if m.Begin != 0 {
		i = encodeVarintSilenced(dAtA, i, uint64(m.Begin))
		i--
		dAtA[i] = 0x50
	}
	if len(m.Subscription) > 0 {
		i -= len(m.Subscription)
		copy(dAtA[i:], m.Subscription)
		i = encodeVarintSilenced(dAtA, i, uint64(len(m.Subscription)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.Reason) > 0 {
		i -= len(m.Reason)
		copy(dAtA[i:], m.Reason)
		i = encodeVarintSilenced(dAtA, i, uint64(len(m.Reason)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.Check) > 0 {
		i -= len(m.Check)
		copy(dAtA[i:], m.Check)
		i = encodeVarintSilenced(dAtA, i, uint64(len(m.Check)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Creator) > 0 {
		i -= len(m.Creator)
		copy(dAtA[i:], m.Creator)
		i = encodeVarintSilenced(dAtA, i, uint64(len(m.Creator)))
		i--
		dAtA[i] = 0x22
	}
	if m.ExpireOnResolve {
		i--
		if m.ExpireOnResolve {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x18
	}
	if m.Expire != 0 {
		i = encodeVarintSilenced(dAtA, i, uint64(m.Expire))
		i--
		dAtA[i] = 0x10
	}
	{
		size, err := m.ObjectMeta.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintSilenced(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintSilenced(dAtA []byte, offset int, v uint64) int {
	offset -= sovSilenced(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func NewPopulatedSilenced(r randySilenced, easy bool) *Silenced {
	this := &Silenced{}
	v1 := NewPopulatedObjectMeta(r, easy)
	this.ObjectMeta = *v1
	this.Expire = int64(r.Int63())
	if r.Intn(2) == 0 {
		this.Expire *= -1
	}
	this.ExpireOnResolve = bool(bool(r.Intn(2) == 0))
	this.Creator = string(randStringSilenced(r))
	this.Check = string(randStringSilenced(r))
	this.Reason = string(randStringSilenced(r))
	this.Subscription = string(randStringSilenced(r))
	this.Begin = int64(r.Int63())
	if r.Intn(2) == 0 {
		this.Begin *= -1
	}
	this.ExpireAt = int64(r.Int63())
	if r.Intn(2) == 0 {
		this.ExpireAt *= -1
	}
	if !easy && r.Intn(10) != 0 {
		this.XXX_unrecognized = randUnrecognizedSilenced(r, 12)
	}
	return this
}

type randySilenced interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneSilenced(r randySilenced) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringSilenced(r randySilenced) string {
	v2 := r.Intn(100)
	tmps := make([]rune, v2)
	for i := 0; i < v2; i++ {
		tmps[i] = randUTF8RuneSilenced(r)
	}
	return string(tmps)
}
func randUnrecognizedSilenced(r randySilenced, maxFieldNumber int) (dAtA []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		dAtA = randFieldSilenced(dAtA, r, fieldNumber, wire)
	}
	return dAtA
}
func randFieldSilenced(dAtA []byte, r randySilenced, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		dAtA = encodeVarintPopulateSilenced(dAtA, uint64(key))
		v3 := r.Int63()
		if r.Intn(2) == 0 {
			v3 *= -1
		}
		dAtA = encodeVarintPopulateSilenced(dAtA, uint64(v3))
	case 1:
		dAtA = encodeVarintPopulateSilenced(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		dAtA = encodeVarintPopulateSilenced(dAtA, uint64(key))
		ll := r.Intn(100)
		dAtA = encodeVarintPopulateSilenced(dAtA, uint64(ll))
		for j := 0; j < ll; j++ {
			dAtA = append(dAtA, byte(r.Intn(256)))
		}
	default:
		dAtA = encodeVarintPopulateSilenced(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return dAtA
}
func encodeVarintPopulateSilenced(dAtA []byte, v uint64) []byte {
	for v >= 1<<7 {
		dAtA = append(dAtA, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	dAtA = append(dAtA, uint8(v))
	return dAtA
}
func (m *Silenced) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.ObjectMeta.Size()
	n += 1 + l + sovSilenced(uint64(l))
	if m.Expire != 0 {
		n += 1 + sovSilenced(uint64(m.Expire))
	}
	if m.ExpireOnResolve {
		n += 2
	}
	l = len(m.Creator)
	if l > 0 {
		n += 1 + l + sovSilenced(uint64(l))
	}
	l = len(m.Check)
	if l > 0 {
		n += 1 + l + sovSilenced(uint64(l))
	}
	l = len(m.Reason)
	if l > 0 {
		n += 1 + l + sovSilenced(uint64(l))
	}
	l = len(m.Subscription)
	if l > 0 {
		n += 1 + l + sovSilenced(uint64(l))
	}
	if m.Begin != 0 {
		n += 1 + sovSilenced(uint64(m.Begin))
	}
	if m.ExpireAt != 0 {
		n += 1 + sovSilenced(uint64(m.ExpireAt))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovSilenced(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozSilenced(x uint64) (n int) {
	return sovSilenced(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Silenced) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSilenced
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
			return fmt.Errorf("proto: Silenced: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Silenced: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObjectMeta", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSilenced
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
				return ErrInvalidLengthSilenced
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthSilenced
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ObjectMeta.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Expire", wireType)
			}
			m.Expire = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSilenced
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Expire |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExpireOnResolve", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSilenced
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.ExpireOnResolve = bool(v != 0)
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Creator", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSilenced
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
				return ErrInvalidLengthSilenced
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSilenced
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Creator = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Check", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSilenced
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
				return ErrInvalidLengthSilenced
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSilenced
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Check = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Reason", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSilenced
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
				return ErrInvalidLengthSilenced
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSilenced
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Reason = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Subscription", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSilenced
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
				return ErrInvalidLengthSilenced
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSilenced
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Subscription = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 10:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Begin", wireType)
			}
			m.Begin = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSilenced
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Begin |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 11:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExpireAt", wireType)
			}
			m.ExpireAt = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSilenced
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ExpireAt |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipSilenced(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthSilenced
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
func skipSilenced(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowSilenced
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
					return 0, ErrIntOverflowSilenced
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
					return 0, ErrIntOverflowSilenced
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
				return 0, ErrInvalidLengthSilenced
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupSilenced
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthSilenced
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthSilenced        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowSilenced          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupSilenced = fmt.Errorf("proto: unexpected end of group")
)
