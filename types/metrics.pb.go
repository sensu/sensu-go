// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: metrics.proto

package types

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"

import encoding_binary "encoding/binary"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// A Metrics is an event metrics payload specification.
type Metrics struct {
	// Handlers is a list of handlers for the metric points.
	Handlers []string `protobuf:"bytes,1,rep,name=handlers" json:"handlers"`
	// Points is a list of metric points (measurements).
	Points []*MetricPoint `protobuf:"bytes,2,rep,name=points" json:"points"`
}

func (m *Metrics) Reset()                    { *m = Metrics{} }
func (m *Metrics) String() string            { return proto.CompactTextString(m) }
func (*Metrics) ProtoMessage()               {}
func (*Metrics) Descriptor() ([]byte, []int) { return fileDescriptorMetrics, []int{0} }

func (m *Metrics) GetHandlers() []string {
	if m != nil {
		return m.Handlers
	}
	return nil
}

func (m *Metrics) GetPoints() []*MetricPoint {
	if m != nil {
		return m.Points
	}
	return nil
}

// A MetricPoint represents a single measurement.
type MetricPoint struct {
	// The metric point name.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The metric point value.
	Value float64 `protobuf:"fixed64,2,opt,name=value,proto3" json:"value"`
	// The metric point timestamp, time in nanoseconds since the Epoch.
	Timestamp int64 `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp"`
	// Tags is a list of metric tags (dimensions).
	Tags []*MetricTag `protobuf:"bytes,4,rep,name=tags" json:"tags"`
}

func (m *MetricPoint) Reset()                    { *m = MetricPoint{} }
func (m *MetricPoint) String() string            { return proto.CompactTextString(m) }
func (*MetricPoint) ProtoMessage()               {}
func (*MetricPoint) Descriptor() ([]byte, []int) { return fileDescriptorMetrics, []int{1} }

func (m *MetricPoint) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *MetricPoint) GetValue() float64 {
	if m != nil {
		return m.Value
	}
	return 0
}

func (m *MetricPoint) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *MetricPoint) GetTags() []*MetricTag {
	if m != nil {
		return m.Tags
	}
	return nil
}

// A MetricTag adds a dimension to a metric point.
type MetricTag struct {
	// The metric tag name.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The metric tag value.
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *MetricTag) Reset()                    { *m = MetricTag{} }
func (m *MetricTag) String() string            { return proto.CompactTextString(m) }
func (*MetricTag) ProtoMessage()               {}
func (*MetricTag) Descriptor() ([]byte, []int) { return fileDescriptorMetrics, []int{2} }

func (m *MetricTag) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *MetricTag) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

func init() {
	proto.RegisterType((*Metrics)(nil), "sensu.types.Metrics")
	proto.RegisterType((*MetricPoint)(nil), "sensu.types.MetricPoint")
	proto.RegisterType((*MetricTag)(nil), "sensu.types.MetricTag")
}
func (this *Metrics) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*Metrics)
	if !ok {
		that2, ok := that.(Metrics)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if len(this.Handlers) != len(that1.Handlers) {
		return false
	}
	for i := range this.Handlers {
		if this.Handlers[i] != that1.Handlers[i] {
			return false
		}
	}
	if len(this.Points) != len(that1.Points) {
		return false
	}
	for i := range this.Points {
		if !this.Points[i].Equal(that1.Points[i]) {
			return false
		}
	}
	return true
}
func (this *MetricPoint) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*MetricPoint)
	if !ok {
		that2, ok := that.(MetricPoint)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if this.Name != that1.Name {
		return false
	}
	if this.Value != that1.Value {
		return false
	}
	if this.Timestamp != that1.Timestamp {
		return false
	}
	if len(this.Tags) != len(that1.Tags) {
		return false
	}
	for i := range this.Tags {
		if !this.Tags[i].Equal(that1.Tags[i]) {
			return false
		}
	}
	return true
}
func (this *MetricTag) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*MetricTag)
	if !ok {
		that2, ok := that.(MetricTag)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if this.Name != that1.Name {
		return false
	}
	if this.Value != that1.Value {
		return false
	}
	return true
}
func (m *Metrics) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Metrics) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Handlers) > 0 {
		for _, s := range m.Handlers {
			dAtA[i] = 0xa
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
	if len(m.Points) > 0 {
		for _, msg := range m.Points {
			dAtA[i] = 0x12
			i++
			i = encodeVarintMetrics(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *MetricPoint) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MetricPoint) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintMetrics(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if m.Value != 0 {
		dAtA[i] = 0x11
		i++
		encoding_binary.LittleEndian.PutUint64(dAtA[i:], uint64(math.Float64bits(float64(m.Value))))
		i += 8
	}
	if m.Timestamp != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintMetrics(dAtA, i, uint64(m.Timestamp))
	}
	if len(m.Tags) > 0 {
		for _, msg := range m.Tags {
			dAtA[i] = 0x22
			i++
			i = encodeVarintMetrics(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *MetricTag) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MetricTag) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintMetrics(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if len(m.Value) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintMetrics(dAtA, i, uint64(len(m.Value)))
		i += copy(dAtA[i:], m.Value)
	}
	return i, nil
}

func encodeVarintMetrics(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func NewPopulatedMetrics(r randyMetrics, easy bool) *Metrics {
	this := &Metrics{}
	v1 := r.Intn(10)
	this.Handlers = make([]string, v1)
	for i := 0; i < v1; i++ {
		this.Handlers[i] = string(randStringMetrics(r))
	}
	if r.Intn(10) != 0 {
		v2 := r.Intn(5)
		this.Points = make([]*MetricPoint, v2)
		for i := 0; i < v2; i++ {
			this.Points[i] = NewPopulatedMetricPoint(r, easy)
		}
	}
	if !easy && r.Intn(10) != 0 {
	}
	return this
}

func NewPopulatedMetricPoint(r randyMetrics, easy bool) *MetricPoint {
	this := &MetricPoint{}
	this.Name = string(randStringMetrics(r))
	this.Value = float64(r.Float64())
	if r.Intn(2) == 0 {
		this.Value *= -1
	}
	this.Timestamp = int64(r.Int63())
	if r.Intn(2) == 0 {
		this.Timestamp *= -1
	}
	if r.Intn(10) != 0 {
		v3 := r.Intn(5)
		this.Tags = make([]*MetricTag, v3)
		for i := 0; i < v3; i++ {
			this.Tags[i] = NewPopulatedMetricTag(r, easy)
		}
	}
	if !easy && r.Intn(10) != 0 {
	}
	return this
}

func NewPopulatedMetricTag(r randyMetrics, easy bool) *MetricTag {
	this := &MetricTag{}
	this.Name = string(randStringMetrics(r))
	this.Value = string(randStringMetrics(r))
	if !easy && r.Intn(10) != 0 {
	}
	return this
}

type randyMetrics interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneMetrics(r randyMetrics) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringMetrics(r randyMetrics) string {
	v4 := r.Intn(100)
	tmps := make([]rune, v4)
	for i := 0; i < v4; i++ {
		tmps[i] = randUTF8RuneMetrics(r)
	}
	return string(tmps)
}
func randUnrecognizedMetrics(r randyMetrics, maxFieldNumber int) (dAtA []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		dAtA = randFieldMetrics(dAtA, r, fieldNumber, wire)
	}
	return dAtA
}
func randFieldMetrics(dAtA []byte, r randyMetrics, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		dAtA = encodeVarintPopulateMetrics(dAtA, uint64(key))
		v5 := r.Int63()
		if r.Intn(2) == 0 {
			v5 *= -1
		}
		dAtA = encodeVarintPopulateMetrics(dAtA, uint64(v5))
	case 1:
		dAtA = encodeVarintPopulateMetrics(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		dAtA = encodeVarintPopulateMetrics(dAtA, uint64(key))
		ll := r.Intn(100)
		dAtA = encodeVarintPopulateMetrics(dAtA, uint64(ll))
		for j := 0; j < ll; j++ {
			dAtA = append(dAtA, byte(r.Intn(256)))
		}
	default:
		dAtA = encodeVarintPopulateMetrics(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return dAtA
}
func encodeVarintPopulateMetrics(dAtA []byte, v uint64) []byte {
	for v >= 1<<7 {
		dAtA = append(dAtA, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	dAtA = append(dAtA, uint8(v))
	return dAtA
}
func (m *Metrics) Size() (n int) {
	var l int
	_ = l
	if len(m.Handlers) > 0 {
		for _, s := range m.Handlers {
			l = len(s)
			n += 1 + l + sovMetrics(uint64(l))
		}
	}
	if len(m.Points) > 0 {
		for _, e := range m.Points {
			l = e.Size()
			n += 1 + l + sovMetrics(uint64(l))
		}
	}
	return n
}

func (m *MetricPoint) Size() (n int) {
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovMetrics(uint64(l))
	}
	if m.Value != 0 {
		n += 9
	}
	if m.Timestamp != 0 {
		n += 1 + sovMetrics(uint64(m.Timestamp))
	}
	if len(m.Tags) > 0 {
		for _, e := range m.Tags {
			l = e.Size()
			n += 1 + l + sovMetrics(uint64(l))
		}
	}
	return n
}

func (m *MetricTag) Size() (n int) {
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovMetrics(uint64(l))
	}
	l = len(m.Value)
	if l > 0 {
		n += 1 + l + sovMetrics(uint64(l))
	}
	return n
}

func sovMetrics(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozMetrics(x uint64) (n int) {
	return sovMetrics(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Metrics) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMetrics
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
			return fmt.Errorf("proto: Metrics: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Metrics: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Handlers", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetrics
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
				return ErrInvalidLengthMetrics
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Handlers = append(m.Handlers, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Points", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetrics
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
				return ErrInvalidLengthMetrics
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Points = append(m.Points, &MetricPoint{})
			if err := m.Points[len(m.Points)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMetrics(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMetrics
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
func (m *MetricPoint) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMetrics
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
			return fmt.Errorf("proto: MetricPoint: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MetricPoint: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetrics
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
				return ErrInvalidLengthMetrics
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			v = uint64(encoding_binary.LittleEndian.Uint64(dAtA[iNdEx:]))
			iNdEx += 8
			m.Value = float64(math.Float64frombits(v))
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
			}
			m.Timestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetrics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Timestamp |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tags", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetrics
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
				return ErrInvalidLengthMetrics
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Tags = append(m.Tags, &MetricTag{})
			if err := m.Tags[len(m.Tags)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMetrics(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMetrics
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
func (m *MetricTag) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMetrics
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
			return fmt.Errorf("proto: MetricTag: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MetricTag: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetrics
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
				return ErrInvalidLengthMetrics
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetrics
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
				return ErrInvalidLengthMetrics
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Value = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMetrics(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMetrics
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
func skipMetrics(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMetrics
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
					return 0, ErrIntOverflowMetrics
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
					return 0, ErrIntOverflowMetrics
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
				return 0, ErrInvalidLengthMetrics
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowMetrics
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
				next, err := skipMetrics(dAtA[start:])
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
	ErrInvalidLengthMetrics = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMetrics   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("metrics.proto", fileDescriptorMetrics) }

var fileDescriptorMetrics = []byte{
	// 316 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0xbf, 0x4e, 0xf3, 0x30,
	0x14, 0xc5, 0xbf, 0xdb, 0xb4, 0xfd, 0x1a, 0x97, 0x2e, 0x16, 0x42, 0x16, 0x83, 0x1d, 0x95, 0xc5,
	0x12, 0x22, 0x95, 0xf8, 0xb3, 0x31, 0x79, 0x47, 0x42, 0x16, 0x13, 0x9b, 0x5b, 0x4c, 0x1a, 0xa9,
	0x89, 0x43, 0xed, 0x20, 0xf1, 0x26, 0xbc, 0x00, 0x12, 0x8f, 0xc0, 0x23, 0x30, 0xf2, 0x04, 0x11,
	0x84, 0x2d, 0x4f, 0xc0, 0x88, 0x70, 0xaa, 0xb6, 0x43, 0x17, 0xdf, 0x73, 0x8f, 0x8f, 0xaf, 0x7f,
	0xba, 0x68, 0x94, 0x69, 0xb7, 0x4c, 0x67, 0x36, 0x2e, 0x96, 0xc6, 0x19, 0x3c, 0xb4, 0x3a, 0xb7,
	0x65, 0xec, 0x9e, 0x0a, 0x6d, 0x0f, 0x4f, 0x92, 0xd4, 0xcd, 0xcb, 0x69, 0x3c, 0x33, 0xd9, 0x24,
	0x31, 0x89, 0x99, 0xf8, 0xcc, 0xb4, 0xbc, 0xf7, 0x9d, 0x6f, 0xbc, 0x6a, 0xdf, 0x8e, 0x1f, 0xd0,
	0xff, 0xab, 0x76, 0x18, 0xe6, 0x68, 0x30, 0x57, 0xf9, 0xdd, 0x42, 0x2f, 0x2d, 0x81, 0x28, 0xe0,
	0xa1, 0xd8, 0x6b, 0x2a, 0xb6, 0xf6, 0xe4, 0x5a, 0xe1, 0x4b, 0xd4, 0x2f, 0x4c, 0x9a, 0x3b, 0x4b,
	0x3a, 0x51, 0xc0, 0x87, 0xa7, 0x24, 0xde, 0x22, 0x88, 0xdb, 0x79, 0xd7, 0x7f, 0x01, 0x81, 0x9a,
	0x8a, 0xad, 0xb2, 0x72, 0x55, 0xc7, 0x2f, 0x80, 0x86, 0x5b, 0x19, 0x8c, 0x51, 0x37, 0x57, 0x99,
	0x26, 0x10, 0x01, 0x0f, 0xa5, 0xd7, 0x98, 0xa1, 0xde, 0xa3, 0x5a, 0x94, 0x9a, 0x74, 0x22, 0xe0,
	0x20, 0xc2, 0xa6, 0x62, 0xad, 0x21, 0xdb, 0x82, 0x8f, 0x51, 0xe8, 0xd2, 0x4c, 0x5b, 0xa7, 0xb2,
	0x82, 0x04, 0x11, 0xf0, 0x40, 0x8c, 0x9a, 0x8a, 0x6d, 0x4c, 0xb9, 0x91, 0xf8, 0x1c, 0x75, 0x9d,
	0x4a, 0x2c, 0xe9, 0x7a, 0xda, 0x83, 0x1d, 0xb4, 0x37, 0x2a, 0x11, 0x83, 0xa6, 0x62, 0x3e, 0x27,
	0xfd, 0x39, 0xbe, 0x40, 0xe1, 0xfa, 0x72, 0x27, 0xe4, 0xfe, 0x36, 0x64, 0xb8, 0x22, 0x13, 0x47,
	0x3f, 0x5f, 0x14, 0x5e, 0x6b, 0x0a, 0x6f, 0x35, 0x85, 0xf7, 0x9a, 0xc2, 0x47, 0x4d, 0xe1, 0xb3,
	0xa6, 0xf0, 0xfc, 0x4d, 0xff, 0xdd, 0xf6, 0xfc, 0xaf, 0xd3, 0xbe, 0xdf, 0xfe, 0xd9, 0x6f, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x13, 0x3c, 0xce, 0x9c, 0xca, 0x01, 0x00, 0x00,
}
