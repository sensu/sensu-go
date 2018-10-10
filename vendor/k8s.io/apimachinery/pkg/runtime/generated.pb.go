/*
Copyright (c) 2018 Sensu Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/ // Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/sensu/sensu-go/vendor/k8s.io/apimachinery/pkg/runtime/generated.proto

package runtime

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import strings "strings"
import reflect "reflect"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

func (m *RawExtension) Reset()      { *m = RawExtension{} }
func (*RawExtension) ProtoMessage() {}
func (*RawExtension) Descriptor() ([]byte, []int) {
	return fileDescriptor_generated_231e40ffafc6d527, []int{0}
}
func (m *RawExtension) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RawExtension) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *RawExtension) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RawExtension.Merge(dst, src)
}
func (m *RawExtension) XXX_Size() int {
	return m.Size()
}
func (m *RawExtension) XXX_DiscardUnknown() {
	xxx_messageInfo_RawExtension.DiscardUnknown(m)
}

var xxx_messageInfo_RawExtension proto.InternalMessageInfo

func (m *TypeMeta) Reset()      { *m = TypeMeta{} }
func (*TypeMeta) ProtoMessage() {}
func (*TypeMeta) Descriptor() ([]byte, []int) {
	return fileDescriptor_generated_231e40ffafc6d527, []int{1}
}
func (m *TypeMeta) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TypeMeta) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *TypeMeta) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TypeMeta.Merge(dst, src)
}
func (m *TypeMeta) XXX_Size() int {
	return m.Size()
}
func (m *TypeMeta) XXX_DiscardUnknown() {
	xxx_messageInfo_TypeMeta.DiscardUnknown(m)
}

var xxx_messageInfo_TypeMeta proto.InternalMessageInfo

func (m *Unknown) Reset()      { *m = Unknown{} }
func (*Unknown) ProtoMessage() {}
func (*Unknown) Descriptor() ([]byte, []int) {
	return fileDescriptor_generated_231e40ffafc6d527, []int{2}
}
func (m *Unknown) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Unknown) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *Unknown) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Unknown.Merge(dst, src)
}
func (m *Unknown) XXX_Size() int {
	return m.Size()
}
func (m *Unknown) XXX_DiscardUnknown() {
	xxx_messageInfo_Unknown.DiscardUnknown(m)
}

var xxx_messageInfo_Unknown proto.InternalMessageInfo

func init() {
	proto.RegisterType((*RawExtension)(nil), "k8s.io.apimachinery.pkg.runtime.RawExtension")
	proto.RegisterType((*TypeMeta)(nil), "k8s.io.apimachinery.pkg.runtime.TypeMeta")
	proto.RegisterType((*Unknown)(nil), "k8s.io.apimachinery.pkg.runtime.Unknown")
}
func (m *RawExtension) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RawExtension) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Raw != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintGenerated(dAtA, i, uint64(len(m.Raw)))
		i += copy(dAtA[i:], m.Raw)
	}
	return i, nil
}

func (m *TypeMeta) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TypeMeta) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.APIVersion)))
	i += copy(dAtA[i:], m.APIVersion)
	dAtA[i] = 0x12
	i++
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.Kind)))
	i += copy(dAtA[i:], m.Kind)
	return i, nil
}

func (m *Unknown) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Unknown) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintGenerated(dAtA, i, uint64(m.TypeMeta.Size()))
	n1, err := m.TypeMeta.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	if m.Raw != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintGenerated(dAtA, i, uint64(len(m.Raw)))
		i += copy(dAtA[i:], m.Raw)
	}
	dAtA[i] = 0x1a
	i++
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.ContentEncoding)))
	i += copy(dAtA[i:], m.ContentEncoding)
	dAtA[i] = 0x22
	i++
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.ContentType)))
	i += copy(dAtA[i:], m.ContentType)
	return i, nil
}

func encodeVarintGenerated(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *RawExtension) Size() (n int) {
	var l int
	_ = l
	if m.Raw != nil {
		l = len(m.Raw)
		n += 1 + l + sovGenerated(uint64(l))
	}
	return n
}

func (m *TypeMeta) Size() (n int) {
	var l int
	_ = l
	l = len(m.APIVersion)
	n += 1 + l + sovGenerated(uint64(l))
	l = len(m.Kind)
	n += 1 + l + sovGenerated(uint64(l))
	return n
}

func (m *Unknown) Size() (n int) {
	var l int
	_ = l
	l = m.TypeMeta.Size()
	n += 1 + l + sovGenerated(uint64(l))
	if m.Raw != nil {
		l = len(m.Raw)
		n += 1 + l + sovGenerated(uint64(l))
	}
	l = len(m.ContentEncoding)
	n += 1 + l + sovGenerated(uint64(l))
	l = len(m.ContentType)
	n += 1 + l + sovGenerated(uint64(l))
	return n
}

func sovGenerated(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozGenerated(x uint64) (n int) {
	return sovGenerated(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *RawExtension) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&RawExtension{`,
		`Raw:` + valueToStringGenerated(this.Raw) + `,`,
		`}`,
	}, "")
	return s
}
func (this *TypeMeta) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&TypeMeta{`,
		`APIVersion:` + fmt.Sprintf("%v", this.APIVersion) + `,`,
		`Kind:` + fmt.Sprintf("%v", this.Kind) + `,`,
		`}`,
	}, "")
	return s
}
func (this *Unknown) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Unknown{`,
		`TypeMeta:` + strings.Replace(strings.Replace(this.TypeMeta.String(), "TypeMeta", "TypeMeta", 1), `&`, ``, 1) + `,`,
		`Raw:` + valueToStringGenerated(this.Raw) + `,`,
		`ContentEncoding:` + fmt.Sprintf("%v", this.ContentEncoding) + `,`,
		`ContentType:` + fmt.Sprintf("%v", this.ContentType) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringGenerated(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *RawExtension) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
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
			return fmt.Errorf("proto: RawExtension: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RawExtension: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Raw", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Raw = append(m.Raw[:0], dAtA[iNdEx:postIndex]...)
			if m.Raw == nil {
				m.Raw = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGenerated
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
func (m *TypeMeta) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
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
			return fmt.Errorf("proto: TypeMeta: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TypeMeta: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field APIVersion", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
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
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.APIVersion = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Kind", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
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
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Kind = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGenerated
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
func (m *Unknown) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
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
			return fmt.Errorf("proto: Unknown: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Unknown: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TypeMeta", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
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
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.TypeMeta.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Raw", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Raw = append(m.Raw[:0], dAtA[iNdEx:postIndex]...)
			if m.Raw == nil {
				m.Raw = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContentEncoding", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
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
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContentEncoding = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContentType", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
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
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContentType = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGenerated
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
func skipGenerated(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenerated
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
					return 0, ErrIntOverflowGenerated
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
					return 0, ErrIntOverflowGenerated
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
				return 0, ErrInvalidLengthGenerated
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowGenerated
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
				next, err := skipGenerated(dAtA[start:])
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
	ErrInvalidLengthGenerated = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenerated   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/sensu/sensu-go/vendor/k8s.io/apimachinery/pkg/runtime/generated.proto", fileDescriptor_generated_231e40ffafc6d527)
}

var fileDescriptor_generated_231e40ffafc6d527 = []byte{
	// 381 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x8f, 0x3f, 0x8f, 0xd3, 0x30,
	0x18, 0xc6, 0xe3, 0x6b, 0xa5, 0x1e, 0x6e, 0xa5, 0x43, 0x66, 0x20, 0x30, 0x38, 0xa7, 0x4e, 0xdc,
	0x50, 0x5b, 0x3a, 0x09, 0x89, 0xf5, 0x72, 0xba, 0x01, 0x21, 0xa4, 0x93, 0xc5, 0x1f, 0x89, 0x09,
	0x5f, 0x62, 0x7c, 0x56, 0xd4, 0xd7, 0x51, 0xe2, 0x10, 0xba, 0xf1, 0x11, 0xf8, 0x58, 0x1d, 0x3b,
	0x76, 0xaa, 0x68, 0xf8, 0x10, 0xac, 0xa8, 0xae, 0x5b, 0x02, 0x0c, 0x2c, 0x51, 0xfc, 0x3e, 0xcf,
	0xef, 0x79, 0xdf, 0x07, 0xdf, 0x6a, 0xe3, 0xee, 0x9b, 0x3b, 0x96, 0xd9, 0x39, 0xaf, 0x15, 0xd4,
	0xcd, 0xfe, 0x3b, 0xd3, 0x96, 0x7f, 0x56, 0x90, 0xdb, 0x8a, 0x17, 0x2f, 0x6a, 0x66, 0x2c, 0x97,
	0xa5, 0x99, 0xcb, 0xec, 0xde, 0x80, 0xaa, 0x16, 0xbc, 0x2c, 0x34, 0xaf, 0x1a, 0x70, 0x66, 0xae,
	0xb8, 0x56, 0xa0, 0x2a, 0xe9, 0x54, 0xce, 0xca, 0xca, 0x3a, 0x4b, 0x92, 0x3d, 0xc0, 0xfa, 0x00,
	0x2b, 0x0b, 0xcd, 0x02, 0xf0, 0x74, 0xd6, 0x5b, 0xa9, 0xad, 0xb6, 0xdc, 0x73, 0x77, 0xcd, 0x27,
	0xff, 0xf2, 0x0f, 0xff, 0xb7, 0xcf, 0x9b, 0x5e, 0xe0, 0x89, 0x90, 0xed, 0xcd, 0x17, 0xa7, 0xa0,
	0x36, 0x16, 0xc8, 0x13, 0x3c, 0xa8, 0x64, 0x1b, 0xa3, 0x73, 0xf4, 0x6c, 0x92, 0x8e, 0xba, 0x4d,
	0x32, 0x10, 0xb2, 0x15, 0xbb, 0xd9, 0xf4, 0x23, 0x3e, 0x7d, 0xb3, 0x28, 0xd5, 0x6b, 0xe5, 0x24,
	0xb9, 0xc4, 0x58, 0x96, 0xe6, 0x9d, 0xaa, 0x76, 0x90, 0x77, 0x3f, 0x48, 0xc9, 0x72, 0x93, 0x44,
	0xdd, 0x26, 0xc1, 0x57, 0xb7, 0x2f, 0x83, 0x22, 0x7a, 0x2e, 0x72, 0x8e, 0x87, 0x85, 0x81, 0x3c,
	0x3e, 0xf1, 0xee, 0x49, 0x70, 0x0f, 0x5f, 0x19, 0xc8, 0x85, 0x57, 0xa6, 0x3f, 0x11, 0x1e, 0xbd,
	0x85, 0x02, 0x6c, 0x0b, 0xe4, 0x3d, 0x3e, 0x75, 0x61, 0x9b, 0xcf, 0x1f, 0x5f, 0x5e, 0xb0, 0xff,
	0x74, 0x67, 0x87, 0xf3, 0xd2, 0x87, 0x21, 0xfc, 0x78, 0xb0, 0x38, 0x86, 0x1d, 0x1a, 0x9e, 0xfc,
	0xdb, 0x90, 0x5c, 0xe1, 0xb3, 0xcc, 0x82, 0x53, 0xe0, 0x6e, 0x20, 0xb3, 0xb9, 0x01, 0x1d, 0x0f,
	0xfc, 0xb1, 0x8f, 0x43, 0xde, 0xd9, 0xf5, 0x9f, 0xb2, 0xf8, 0xdb, 0x4f, 0x9e, 0xe3, 0x71, 0x18,
	0xed, 0x56, 0xc7, 0x43, 0x8f, 0x3f, 0x0a, 0xf8, 0xf8, 0xfa, 0xb7, 0x24, 0xfa, 0xbe, 0x74, 0xb6,
	0xdc, 0xd2, 0x68, 0xb5, 0xa5, 0xd1, 0x7a, 0x4b, 0xa3, 0xaf, 0x1d, 0x45, 0xcb, 0x8e, 0xa2, 0x55,
	0x47, 0xd1, 0xba, 0xa3, 0xe8, 0x7b, 0x47, 0xd1, 0xb7, 0x1f, 0x34, 0xfa, 0x30, 0x0a, 0x45, 0x7f,
	0x05, 0x00, 0x00, 0xff, 0xff, 0x81, 0x28, 0xb8, 0x50, 0x58, 0x02, 0x00, 0x00,
}
