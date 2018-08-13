// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: filter.proto

package types

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"

import bytes "bytes"

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

// EventFilter is a filter specification.
type EventFilter struct {
	// Name is the unique identifier for a filter
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Action specifies to allow/deny events to continue through the pipeline
	Action string `protobuf:"bytes,2,opt,name=action,proto3" json:"action,omitempty"`
	// Statements is an array of boolean expressions that are &&'d together
	// to determine if the event matches this filter.
	Statements []string `protobuf:"bytes,3,rep,name=statements" json:"statements"`
	// Environment indicates to which env a filter belongs to
	Environment string `protobuf:"bytes,4,opt,name=environment,proto3" json:"environment,omitempty"`
	// Organization indicates to which org a filter belongs to
	Organization string `protobuf:"bytes,5,opt,name=organization,proto3" json:"organization,omitempty"`
	// When indicates a TimeWindowWhen that a filter uses to filter by days & times
	When                 *TimeWindowWhen `protobuf:"bytes,6,opt,name=when" json:"when,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *EventFilter) Reset()         { *m = EventFilter{} }
func (m *EventFilter) String() string { return proto.CompactTextString(m) }
func (*EventFilter) ProtoMessage()    {}
func (*EventFilter) Descriptor() ([]byte, []int) {
	return fileDescriptor_filter_ee73b8fb45db20c9, []int{0}
}
func (m *EventFilter) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventFilter) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventFilter.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *EventFilter) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventFilter.Merge(dst, src)
}
func (m *EventFilter) XXX_Size() int {
	return m.Size()
}
func (m *EventFilter) XXX_DiscardUnknown() {
	xxx_messageInfo_EventFilter.DiscardUnknown(m)
}

var xxx_messageInfo_EventFilter proto.InternalMessageInfo

func (m *EventFilter) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *EventFilter) GetAction() string {
	if m != nil {
		return m.Action
	}
	return ""
}

func (m *EventFilter) GetStatements() []string {
	if m != nil {
		return m.Statements
	}
	return nil
}

func (m *EventFilter) GetEnvironment() string {
	if m != nil {
		return m.Environment
	}
	return ""
}

func (m *EventFilter) GetOrganization() string {
	if m != nil {
		return m.Organization
	}
	return ""
}

func (m *EventFilter) GetWhen() *TimeWindowWhen {
	if m != nil {
		return m.When
	}
	return nil
}

func init() {
	proto.RegisterType((*EventFilter)(nil), "sensu.types.EventFilter")
}
func (this *EventFilter) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*EventFilter)
	if !ok {
		that2, ok := that.(EventFilter)
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
	if this.Action != that1.Action {
		return false
	}
	if len(this.Statements) != len(that1.Statements) {
		return false
	}
	for i := range this.Statements {
		if this.Statements[i] != that1.Statements[i] {
			return false
		}
	}
	if this.Environment != that1.Environment {
		return false
	}
	if this.Organization != that1.Organization {
		return false
	}
	if !this.When.Equal(that1.When) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (m *EventFilter) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventFilter) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintFilter(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if len(m.Action) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintFilter(dAtA, i, uint64(len(m.Action)))
		i += copy(dAtA[i:], m.Action)
	}
	if len(m.Statements) > 0 {
		for _, s := range m.Statements {
			dAtA[i] = 0x1a
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
	if len(m.Environment) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintFilter(dAtA, i, uint64(len(m.Environment)))
		i += copy(dAtA[i:], m.Environment)
	}
	if len(m.Organization) > 0 {
		dAtA[i] = 0x2a
		i++
		i = encodeVarintFilter(dAtA, i, uint64(len(m.Organization)))
		i += copy(dAtA[i:], m.Organization)
	}
	if m.When != nil {
		dAtA[i] = 0x32
		i++
		i = encodeVarintFilter(dAtA, i, uint64(m.When.Size()))
		n1, err := m.When.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func encodeVarintFilter(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func NewPopulatedEventFilter(r randyFilter, easy bool) *EventFilter {
	this := &EventFilter{}
	this.Name = string(randStringFilter(r))
	this.Action = string(randStringFilter(r))
	v1 := r.Intn(10)
	this.Statements = make([]string, v1)
	for i := 0; i < v1; i++ {
		this.Statements[i] = string(randStringFilter(r))
	}
	this.Environment = string(randStringFilter(r))
	this.Organization = string(randStringFilter(r))
	if r.Intn(10) != 0 {
		this.When = NewPopulatedTimeWindowWhen(r, easy)
	}
	if !easy && r.Intn(10) != 0 {
		this.XXX_unrecognized = randUnrecognizedFilter(r, 7)
	}
	return this
}

type randyFilter interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneFilter(r randyFilter) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringFilter(r randyFilter) string {
	v2 := r.Intn(100)
	tmps := make([]rune, v2)
	for i := 0; i < v2; i++ {
		tmps[i] = randUTF8RuneFilter(r)
	}
	return string(tmps)
}
func randUnrecognizedFilter(r randyFilter, maxFieldNumber int) (dAtA []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		dAtA = randFieldFilter(dAtA, r, fieldNumber, wire)
	}
	return dAtA
}
func randFieldFilter(dAtA []byte, r randyFilter, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		dAtA = encodeVarintPopulateFilter(dAtA, uint64(key))
		v3 := r.Int63()
		if r.Intn(2) == 0 {
			v3 *= -1
		}
		dAtA = encodeVarintPopulateFilter(dAtA, uint64(v3))
	case 1:
		dAtA = encodeVarintPopulateFilter(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		dAtA = encodeVarintPopulateFilter(dAtA, uint64(key))
		ll := r.Intn(100)
		dAtA = encodeVarintPopulateFilter(dAtA, uint64(ll))
		for j := 0; j < ll; j++ {
			dAtA = append(dAtA, byte(r.Intn(256)))
		}
	default:
		dAtA = encodeVarintPopulateFilter(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return dAtA
}
func encodeVarintPopulateFilter(dAtA []byte, v uint64) []byte {
	for v >= 1<<7 {
		dAtA = append(dAtA, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	dAtA = append(dAtA, uint8(v))
	return dAtA
}
func (m *EventFilter) Size() (n int) {
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovFilter(uint64(l))
	}
	l = len(m.Action)
	if l > 0 {
		n += 1 + l + sovFilter(uint64(l))
	}
	if len(m.Statements) > 0 {
		for _, s := range m.Statements {
			l = len(s)
			n += 1 + l + sovFilter(uint64(l))
		}
	}
	l = len(m.Environment)
	if l > 0 {
		n += 1 + l + sovFilter(uint64(l))
	}
	l = len(m.Organization)
	if l > 0 {
		n += 1 + l + sovFilter(uint64(l))
	}
	if m.When != nil {
		l = m.When.Size()
		n += 1 + l + sovFilter(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovFilter(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozFilter(x uint64) (n int) {
	return sovFilter(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *EventFilter) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowFilter
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
			return fmt.Errorf("proto: EventFilter: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventFilter: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFilter
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
				return ErrInvalidLengthFilter
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Action", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFilter
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
				return ErrInvalidLengthFilter
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Action = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Statements", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFilter
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
				return ErrInvalidLengthFilter
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Statements = append(m.Statements, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Environment", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFilter
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
				return ErrInvalidLengthFilter
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Environment = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Organization", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFilter
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
				return ErrInvalidLengthFilter
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Organization = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field When", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFilter
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
				return ErrInvalidLengthFilter
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.When == nil {
				m.When = &TimeWindowWhen{}
			}
			if err := m.When.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipFilter(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthFilter
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
func skipFilter(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowFilter
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
					return 0, ErrIntOverflowFilter
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
					return 0, ErrIntOverflowFilter
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
				return 0, ErrInvalidLengthFilter
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowFilter
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
				next, err := skipFilter(dAtA[start:])
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
	ErrInvalidLengthFilter = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowFilter   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("filter.proto", fileDescriptor_filter_ee73b8fb45db20c9) }

var fileDescriptor_filter_ee73b8fb45db20c9 = []byte{
	// 281 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x90, 0x41, 0x4a, 0xc3, 0x40,
	0x14, 0x86, 0x1d, 0xdb, 0x06, 0x3a, 0x29, 0x82, 0xb3, 0x90, 0x50, 0x61, 0x0c, 0x75, 0x93, 0x8d,
	0x13, 0xd0, 0x1b, 0x14, 0xf4, 0x00, 0x41, 0x28, 0xb8, 0x91, 0xa4, 0xbe, 0x26, 0x03, 0xe6, 0x4d,
	0x49, 0x5e, 0x1a, 0xf4, 0x24, 0x1e, 0xc1, 0x23, 0x78, 0x04, 0x97, 0x9e, 0x40, 0x6a, 0xdc, 0x79,
	0x02, 0x97, 0xe2, 0xab, 0x8b, 0xb8, 0xfb, 0xbf, 0x8f, 0xff, 0xcd, 0x63, 0x9e, 0x9c, 0xac, 0xec,
	0x3d, 0x41, 0x65, 0xd6, 0x95, 0x23, 0xa7, 0xfc, 0x1a, 0xb0, 0x6e, 0x0c, 0x3d, 0xac, 0xa1, 0x9e,
	0x9e, 0xe5, 0x96, 0x8a, 0x26, 0x33, 0x4b, 0x57, 0xc6, 0xb9, 0xcb, 0x5d, 0xcc, 0x9d, 0xac, 0x59,
	0x31, 0x31, 0x70, 0xda, 0xcd, 0x4e, 0x0f, 0xc9, 0x96, 0x70, 0xdb, 0x5a, 0xbc, 0x73, 0xed, 0x4e,
	0xcd, 0xb6, 0x42, 0xfa, 0x97, 0x1b, 0x40, 0xba, 0xe2, 0x25, 0x4a, 0xc9, 0x21, 0xa6, 0x25, 0x04,
	0x22, 0x14, 0xd1, 0x38, 0xe1, 0xac, 0x8e, 0xa4, 0x97, 0x2e, 0xc9, 0x3a, 0x0c, 0xf6, 0xd9, 0xfe,
	0x91, 0x32, 0x52, 0xd6, 0x94, 0x12, 0x94, 0x80, 0x54, 0x07, 0x83, 0x70, 0x10, 0x8d, 0xe7, 0x07,
	0x5f, 0xef, 0x27, 0x3d, 0x9b, 0xf4, 0xb2, 0x0a, 0xa5, 0x0f, 0xb8, 0xb1, 0x95, 0xc3, 0x5f, 0x0e,
	0x86, 0xfc, 0x58, 0x5f, 0xa9, 0x99, 0x9c, 0xb8, 0x2a, 0x4f, 0xd1, 0x3e, 0xa6, 0xbc, 0x6f, 0xc4,
	0x95, 0x7f, 0x4e, 0xc5, 0x72, 0xd8, 0x16, 0x80, 0x81, 0x17, 0x8a, 0xc8, 0x3f, 0x3f, 0x36, 0xbd,
	0x7b, 0x98, 0x6b, 0x5b, 0xc2, 0x82, 0xbf, 0xb7, 0x28, 0x00, 0x13, 0x2e, 0xce, 0x4f, 0xbf, 0x3f,
	0xb4, 0x78, 0xee, 0xb4, 0x78, 0xe9, 0xb4, 0x78, 0xed, 0xb4, 0x78, 0xeb, 0xb4, 0xd8, 0x76, 0x5a,
	0x3c, 0x7d, 0xea, 0xbd, 0x9b, 0x11, 0x4f, 0x66, 0x1e, 0x9f, 0xe3, 0xe2, 0x27, 0x00, 0x00, 0xff,
	0xff, 0x91, 0x6e, 0x00, 0xe7, 0x6d, 0x01, 0x00, 0x00,
}
