// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: keepalive.proto

package types

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// A KeepaliveRecord is a tuple of an Entity ID and the time at which the
// entity's keepalive will next expire.
type KeepaliveRecord struct {
	Environment  string `protobuf:"bytes,1,opt,name=environment,proto3" json:"environment,omitempty"`
	Organization string `protobuf:"bytes,2,opt,name=organization,proto3" json:"organization,omitempty"`
	EntityID     string `protobuf:"bytes,3,opt,name=entity_id,json=entityId,proto3" json:"entity_id,omitempty"`
	Time         int64  `protobuf:"varint,4,opt,name=time,proto3" json:"time"`
}

func (m *KeepaliveRecord) Reset()                    { *m = KeepaliveRecord{} }
func (m *KeepaliveRecord) String() string            { return proto.CompactTextString(m) }
func (*KeepaliveRecord) ProtoMessage()               {}
func (*KeepaliveRecord) Descriptor() ([]byte, []int) { return fileDescriptorKeepalive, []int{0} }

func (m *KeepaliveRecord) GetEnvironment() string {
	if m != nil {
		return m.Environment
	}
	return ""
}

func (m *KeepaliveRecord) GetOrganization() string {
	if m != nil {
		return m.Organization
	}
	return ""
}

func (m *KeepaliveRecord) GetEntityID() string {
	if m != nil {
		return m.EntityID
	}
	return ""
}

func (m *KeepaliveRecord) GetTime() int64 {
	if m != nil {
		return m.Time
	}
	return 0
}

func init() {
	proto.RegisterType((*KeepaliveRecord)(nil), "sensu.types.KeepaliveRecord")
}
func (this *KeepaliveRecord) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*KeepaliveRecord)
	if !ok {
		that2, ok := that.(KeepaliveRecord)
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
	if this.Environment != that1.Environment {
		return false
	}
	if this.Organization != that1.Organization {
		return false
	}
	if this.EntityID != that1.EntityID {
		return false
	}
	if this.Time != that1.Time {
		return false
	}
	return true
}
func (m *KeepaliveRecord) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *KeepaliveRecord) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Environment) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintKeepalive(dAtA, i, uint64(len(m.Environment)))
		i += copy(dAtA[i:], m.Environment)
	}
	if len(m.Organization) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintKeepalive(dAtA, i, uint64(len(m.Organization)))
		i += copy(dAtA[i:], m.Organization)
	}
	if len(m.EntityID) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintKeepalive(dAtA, i, uint64(len(m.EntityID)))
		i += copy(dAtA[i:], m.EntityID)
	}
	if m.Time != 0 {
		dAtA[i] = 0x20
		i++
		i = encodeVarintKeepalive(dAtA, i, uint64(m.Time))
	}
	return i, nil
}

func encodeVarintKeepalive(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func NewPopulatedKeepaliveRecord(r randyKeepalive, easy bool) *KeepaliveRecord {
	this := &KeepaliveRecord{}
	this.Environment = string(randStringKeepalive(r))
	this.Organization = string(randStringKeepalive(r))
	this.EntityID = string(randStringKeepalive(r))
	this.Time = int64(r.Int63())
	if r.Intn(2) == 0 {
		this.Time *= -1
	}
	if !easy && r.Intn(10) != 0 {
	}
	return this
}

type randyKeepalive interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneKeepalive(r randyKeepalive) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringKeepalive(r randyKeepalive) string {
	v1 := r.Intn(100)
	tmps := make([]rune, v1)
	for i := 0; i < v1; i++ {
		tmps[i] = randUTF8RuneKeepalive(r)
	}
	return string(tmps)
}
func randUnrecognizedKeepalive(r randyKeepalive, maxFieldNumber int) (dAtA []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		dAtA = randFieldKeepalive(dAtA, r, fieldNumber, wire)
	}
	return dAtA
}
func randFieldKeepalive(dAtA []byte, r randyKeepalive, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		dAtA = encodeVarintPopulateKeepalive(dAtA, uint64(key))
		v2 := r.Int63()
		if r.Intn(2) == 0 {
			v2 *= -1
		}
		dAtA = encodeVarintPopulateKeepalive(dAtA, uint64(v2))
	case 1:
		dAtA = encodeVarintPopulateKeepalive(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		dAtA = encodeVarintPopulateKeepalive(dAtA, uint64(key))
		ll := r.Intn(100)
		dAtA = encodeVarintPopulateKeepalive(dAtA, uint64(ll))
		for j := 0; j < ll; j++ {
			dAtA = append(dAtA, byte(r.Intn(256)))
		}
	default:
		dAtA = encodeVarintPopulateKeepalive(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return dAtA
}
func encodeVarintPopulateKeepalive(dAtA []byte, v uint64) []byte {
	for v >= 1<<7 {
		dAtA = append(dAtA, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	dAtA = append(dAtA, uint8(v))
	return dAtA
}
func (m *KeepaliveRecord) Size() (n int) {
	var l int
	_ = l
	l = len(m.Environment)
	if l > 0 {
		n += 1 + l + sovKeepalive(uint64(l))
	}
	l = len(m.Organization)
	if l > 0 {
		n += 1 + l + sovKeepalive(uint64(l))
	}
	l = len(m.EntityID)
	if l > 0 {
		n += 1 + l + sovKeepalive(uint64(l))
	}
	if m.Time != 0 {
		n += 1 + sovKeepalive(uint64(m.Time))
	}
	return n
}

func sovKeepalive(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozKeepalive(x uint64) (n int) {
	return sovKeepalive(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *KeepaliveRecord) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowKeepalive
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
			return fmt.Errorf("proto: KeepaliveRecord: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: KeepaliveRecord: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Environment", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowKeepalive
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
				return ErrInvalidLengthKeepalive
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Environment = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Organization", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowKeepalive
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
				return ErrInvalidLengthKeepalive
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Organization = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EntityID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowKeepalive
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
				return ErrInvalidLengthKeepalive
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.EntityID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Time", wireType)
			}
			m.Time = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowKeepalive
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Time |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipKeepalive(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthKeepalive
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipKeepalive(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowKeepalive
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
					return 0, ErrIntOverflowKeepalive
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
					return 0, ErrIntOverflowKeepalive
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
				return 0, ErrInvalidLengthKeepalive
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowKeepalive
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
				next, err := skipKeepalive(dAtA[start:])
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
	ErrInvalidLengthKeepalive = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowKeepalive   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("keepalive.proto", fileDescriptorKeepalive) }

var fileDescriptorKeepalive = []byte{
	// 243 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xcf, 0x4e, 0x4d, 0x2d,
	0x48, 0xcc, 0xc9, 0x2c, 0x4b, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x2e, 0x4e, 0xcd,
	0x2b, 0x2e, 0xd5, 0x2b, 0xa9, 0x2c, 0x48, 0x2d, 0x96, 0xd2, 0x4d, 0xcf, 0x2c, 0xc9, 0x28, 0x4d,
	0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0xcf, 0x4f, 0xcf, 0xd7, 0x07, 0xab, 0x49, 0x2a, 0x4d, 0x03,
	0xf3, 0xc0, 0x1c, 0x30, 0x0b, 0xa2, 0x57, 0x69, 0x01, 0x23, 0x17, 0xbf, 0x37, 0xcc, 0xbc, 0xa0,
	0xd4, 0xe4, 0xfc, 0xa2, 0x14, 0x21, 0x05, 0x2e, 0xee, 0xd4, 0xbc, 0xb2, 0xcc, 0xa2, 0xfc, 0xbc,
	0xdc, 0xd4, 0xbc, 0x12, 0x09, 0x46, 0x05, 0x46, 0x0d, 0xce, 0x20, 0x64, 0x21, 0x21, 0x25, 0x2e,
	0x9e, 0xfc, 0xa2, 0xf4, 0xc4, 0xbc, 0xcc, 0xaa, 0xc4, 0x92, 0xcc, 0xfc, 0x3c, 0x09, 0x26, 0xb0,
	0x12, 0x14, 0x31, 0x21, 0x4d, 0x2e, 0xce, 0xd4, 0xbc, 0x92, 0xcc, 0x92, 0xca, 0xf8, 0xcc, 0x14,
	0x09, 0x66, 0x90, 0x02, 0x27, 0x9e, 0x47, 0xf7, 0xe4, 0x39, 0x5c, 0xc1, 0x82, 0x9e, 0x2e, 0x41,
	0x1c, 0x10, 0x69, 0xcf, 0x14, 0x21, 0x19, 0x2e, 0x96, 0x92, 0xcc, 0xdc, 0x54, 0x09, 0x16, 0x05,
	0x46, 0x0d, 0x66, 0x27, 0x8e, 0x57, 0xf7, 0xe4, 0xc1, 0xfc, 0x20, 0x30, 0xe9, 0xa4, 0xfc, 0xe3,
	0xa1, 0x1c, 0xe3, 0x8a, 0x47, 0x72, 0x8c, 0x3b, 0x1e, 0xc9, 0x31, 0x9e, 0x78, 0x24, 0xc7, 0x78,
	0xe1, 0x91, 0x1c, 0xe3, 0x83, 0x47, 0x72, 0x8c, 0x33, 0x1e, 0xcb, 0x31, 0x44, 0xb1, 0x82, 0xbd,
	0x9d, 0xc4, 0x06, 0xf6, 0x8e, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x1c, 0x86, 0xd9, 0x32, 0x1d,
	0x01, 0x00, 0x00,
}
