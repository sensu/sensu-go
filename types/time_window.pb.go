// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: time_window.proto

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

// TimeWindowWhen defines the "when" attributes for time windows
type TimeWindowWhen struct {
	// Days is a hash of days
	Days TimeWindowDays `protobuf:"bytes,1,opt,name=days" json:"days"`
}

func (m *TimeWindowWhen) Reset()                    { *m = TimeWindowWhen{} }
func (m *TimeWindowWhen) String() string            { return proto.CompactTextString(m) }
func (*TimeWindowWhen) ProtoMessage()               {}
func (*TimeWindowWhen) Descriptor() ([]byte, []int) { return fileDescriptorTimeWindow, []int{0} }

func (m *TimeWindowWhen) GetDays() TimeWindowDays {
	if m != nil {
		return m.Days
	}
	return TimeWindowDays{}
}

// TimeWindowDays defines the days of a time window
type TimeWindowDays struct {
	All       []*TimeWindowTimeRange `protobuf:"bytes,1,rep,name=all" json:"all,omitempty"`
	Sunday    []*TimeWindowTimeRange `protobuf:"bytes,2,rep,name=sunday" json:"sunday,omitempty"`
	Monday    []*TimeWindowTimeRange `protobuf:"bytes,3,rep,name=monday" json:"monday,omitempty"`
	Tuesday   []*TimeWindowTimeRange `protobuf:"bytes,4,rep,name=tuesday" json:"tuesday,omitempty"`
	Wednesday []*TimeWindowTimeRange `protobuf:"bytes,5,rep,name=wednesday" json:"wednesday,omitempty"`
	Thursday  []*TimeWindowTimeRange `protobuf:"bytes,6,rep,name=thursday" json:"thursday,omitempty"`
	Friday    []*TimeWindowTimeRange `protobuf:"bytes,7,rep,name=friday" json:"friday,omitempty"`
	Saturday  []*TimeWindowTimeRange `protobuf:"bytes,8,rep,name=saturday" json:"saturday,omitempty"`
}

func (m *TimeWindowDays) Reset()                    { *m = TimeWindowDays{} }
func (m *TimeWindowDays) String() string            { return proto.CompactTextString(m) }
func (*TimeWindowDays) ProtoMessage()               {}
func (*TimeWindowDays) Descriptor() ([]byte, []int) { return fileDescriptorTimeWindow, []int{1} }

func (m *TimeWindowDays) GetAll() []*TimeWindowTimeRange {
	if m != nil {
		return m.All
	}
	return nil
}

func (m *TimeWindowDays) GetSunday() []*TimeWindowTimeRange {
	if m != nil {
		return m.Sunday
	}
	return nil
}

func (m *TimeWindowDays) GetMonday() []*TimeWindowTimeRange {
	if m != nil {
		return m.Monday
	}
	return nil
}

func (m *TimeWindowDays) GetTuesday() []*TimeWindowTimeRange {
	if m != nil {
		return m.Tuesday
	}
	return nil
}

func (m *TimeWindowDays) GetWednesday() []*TimeWindowTimeRange {
	if m != nil {
		return m.Wednesday
	}
	return nil
}

func (m *TimeWindowDays) GetThursday() []*TimeWindowTimeRange {
	if m != nil {
		return m.Thursday
	}
	return nil
}

func (m *TimeWindowDays) GetFriday() []*TimeWindowTimeRange {
	if m != nil {
		return m.Friday
	}
	return nil
}

func (m *TimeWindowDays) GetSaturday() []*TimeWindowTimeRange {
	if m != nil {
		return m.Saturday
	}
	return nil
}

// TimeWindowTimeRange defines the time ranges of a time
type TimeWindowTimeRange struct {
	// Begin is the time which the time window should begin, in the format
	// '3:00PM', which satisfies the time.Kitchen format
	Begin string `protobuf:"bytes,1,opt,name=begin,proto3" json:"begin"`
	// End is the time which the filter should end, in the format '3:00PM', which
	// satisfies the time.Kitchen format
	End string `protobuf:"bytes,2,opt,name=end,proto3" json:"end"`
}

func (m *TimeWindowTimeRange) Reset()                    { *m = TimeWindowTimeRange{} }
func (m *TimeWindowTimeRange) String() string            { return proto.CompactTextString(m) }
func (*TimeWindowTimeRange) ProtoMessage()               {}
func (*TimeWindowTimeRange) Descriptor() ([]byte, []int) { return fileDescriptorTimeWindow, []int{2} }

func (m *TimeWindowTimeRange) GetBegin() string {
	if m != nil {
		return m.Begin
	}
	return ""
}

func (m *TimeWindowTimeRange) GetEnd() string {
	if m != nil {
		return m.End
	}
	return ""
}

func init() {
	proto.RegisterType((*TimeWindowWhen)(nil), "sensu.types.TimeWindowWhen")
	proto.RegisterType((*TimeWindowDays)(nil), "sensu.types.TimeWindowDays")
	proto.RegisterType((*TimeWindowTimeRange)(nil), "sensu.types.TimeWindowTimeRange")
}
func (this *TimeWindowWhen) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TimeWindowWhen)
	if !ok {
		that2, ok := that.(TimeWindowWhen)
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
	if !this.Days.Equal(&that1.Days) {
		return false
	}
	return true
}
func (this *TimeWindowDays) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TimeWindowDays)
	if !ok {
		that2, ok := that.(TimeWindowDays)
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
	if len(this.All) != len(that1.All) {
		return false
	}
	for i := range this.All {
		if !this.All[i].Equal(that1.All[i]) {
			return false
		}
	}
	if len(this.Sunday) != len(that1.Sunday) {
		return false
	}
	for i := range this.Sunday {
		if !this.Sunday[i].Equal(that1.Sunday[i]) {
			return false
		}
	}
	if len(this.Monday) != len(that1.Monday) {
		return false
	}
	for i := range this.Monday {
		if !this.Monday[i].Equal(that1.Monday[i]) {
			return false
		}
	}
	if len(this.Tuesday) != len(that1.Tuesday) {
		return false
	}
	for i := range this.Tuesday {
		if !this.Tuesday[i].Equal(that1.Tuesday[i]) {
			return false
		}
	}
	if len(this.Wednesday) != len(that1.Wednesday) {
		return false
	}
	for i := range this.Wednesday {
		if !this.Wednesday[i].Equal(that1.Wednesday[i]) {
			return false
		}
	}
	if len(this.Thursday) != len(that1.Thursday) {
		return false
	}
	for i := range this.Thursday {
		if !this.Thursday[i].Equal(that1.Thursday[i]) {
			return false
		}
	}
	if len(this.Friday) != len(that1.Friday) {
		return false
	}
	for i := range this.Friday {
		if !this.Friday[i].Equal(that1.Friday[i]) {
			return false
		}
	}
	if len(this.Saturday) != len(that1.Saturday) {
		return false
	}
	for i := range this.Saturday {
		if !this.Saturday[i].Equal(that1.Saturday[i]) {
			return false
		}
	}
	return true
}
func (this *TimeWindowTimeRange) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TimeWindowTimeRange)
	if !ok {
		that2, ok := that.(TimeWindowTimeRange)
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
	if this.Begin != that1.Begin {
		return false
	}
	if this.End != that1.End {
		return false
	}
	return true
}
func (m *TimeWindowWhen) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TimeWindowWhen) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintTimeWindow(dAtA, i, uint64(m.Days.Size()))
	n1, err := m.Days.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	return i, nil
}

func (m *TimeWindowDays) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TimeWindowDays) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.All) > 0 {
		for _, msg := range m.All {
			dAtA[i] = 0xa
			i++
			i = encodeVarintTimeWindow(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.Sunday) > 0 {
		for _, msg := range m.Sunday {
			dAtA[i] = 0x12
			i++
			i = encodeVarintTimeWindow(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.Monday) > 0 {
		for _, msg := range m.Monday {
			dAtA[i] = 0x1a
			i++
			i = encodeVarintTimeWindow(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.Tuesday) > 0 {
		for _, msg := range m.Tuesday {
			dAtA[i] = 0x22
			i++
			i = encodeVarintTimeWindow(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.Wednesday) > 0 {
		for _, msg := range m.Wednesday {
			dAtA[i] = 0x2a
			i++
			i = encodeVarintTimeWindow(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.Thursday) > 0 {
		for _, msg := range m.Thursday {
			dAtA[i] = 0x32
			i++
			i = encodeVarintTimeWindow(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.Friday) > 0 {
		for _, msg := range m.Friday {
			dAtA[i] = 0x3a
			i++
			i = encodeVarintTimeWindow(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.Saturday) > 0 {
		for _, msg := range m.Saturday {
			dAtA[i] = 0x42
			i++
			i = encodeVarintTimeWindow(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *TimeWindowTimeRange) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TimeWindowTimeRange) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Begin) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintTimeWindow(dAtA, i, uint64(len(m.Begin)))
		i += copy(dAtA[i:], m.Begin)
	}
	if len(m.End) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintTimeWindow(dAtA, i, uint64(len(m.End)))
		i += copy(dAtA[i:], m.End)
	}
	return i, nil
}

func encodeVarintTimeWindow(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func NewPopulatedTimeWindowWhen(r randyTimeWindow, easy bool) *TimeWindowWhen {
	this := &TimeWindowWhen{}
	v1 := NewPopulatedTimeWindowDays(r, easy)
	this.Days = *v1
	if !easy && r.Intn(10) != 0 {
	}
	return this
}

func NewPopulatedTimeWindowDays(r randyTimeWindow, easy bool) *TimeWindowDays {
	this := &TimeWindowDays{}
	if r.Intn(10) != 0 {
		v2 := r.Intn(5)
		this.All = make([]*TimeWindowTimeRange, v2)
		for i := 0; i < v2; i++ {
			this.All[i] = NewPopulatedTimeWindowTimeRange(r, easy)
		}
	}
	if r.Intn(10) != 0 {
		v3 := r.Intn(5)
		this.Sunday = make([]*TimeWindowTimeRange, v3)
		for i := 0; i < v3; i++ {
			this.Sunday[i] = NewPopulatedTimeWindowTimeRange(r, easy)
		}
	}
	if r.Intn(10) != 0 {
		v4 := r.Intn(5)
		this.Monday = make([]*TimeWindowTimeRange, v4)
		for i := 0; i < v4; i++ {
			this.Monday[i] = NewPopulatedTimeWindowTimeRange(r, easy)
		}
	}
	if r.Intn(10) != 0 {
		v5 := r.Intn(5)
		this.Tuesday = make([]*TimeWindowTimeRange, v5)
		for i := 0; i < v5; i++ {
			this.Tuesday[i] = NewPopulatedTimeWindowTimeRange(r, easy)
		}
	}
	if r.Intn(10) != 0 {
		v6 := r.Intn(5)
		this.Wednesday = make([]*TimeWindowTimeRange, v6)
		for i := 0; i < v6; i++ {
			this.Wednesday[i] = NewPopulatedTimeWindowTimeRange(r, easy)
		}
	}
	if r.Intn(10) != 0 {
		v7 := r.Intn(5)
		this.Thursday = make([]*TimeWindowTimeRange, v7)
		for i := 0; i < v7; i++ {
			this.Thursday[i] = NewPopulatedTimeWindowTimeRange(r, easy)
		}
	}
	if r.Intn(10) != 0 {
		v8 := r.Intn(5)
		this.Friday = make([]*TimeWindowTimeRange, v8)
		for i := 0; i < v8; i++ {
			this.Friday[i] = NewPopulatedTimeWindowTimeRange(r, easy)
		}
	}
	if r.Intn(10) != 0 {
		v9 := r.Intn(5)
		this.Saturday = make([]*TimeWindowTimeRange, v9)
		for i := 0; i < v9; i++ {
			this.Saturday[i] = NewPopulatedTimeWindowTimeRange(r, easy)
		}
	}
	if !easy && r.Intn(10) != 0 {
	}
	return this
}

func NewPopulatedTimeWindowTimeRange(r randyTimeWindow, easy bool) *TimeWindowTimeRange {
	this := &TimeWindowTimeRange{}
	this.Begin = string(randStringTimeWindow(r))
	this.End = string(randStringTimeWindow(r))
	if !easy && r.Intn(10) != 0 {
	}
	return this
}

type randyTimeWindow interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneTimeWindow(r randyTimeWindow) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringTimeWindow(r randyTimeWindow) string {
	v10 := r.Intn(100)
	tmps := make([]rune, v10)
	for i := 0; i < v10; i++ {
		tmps[i] = randUTF8RuneTimeWindow(r)
	}
	return string(tmps)
}
func randUnrecognizedTimeWindow(r randyTimeWindow, maxFieldNumber int) (dAtA []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		dAtA = randFieldTimeWindow(dAtA, r, fieldNumber, wire)
	}
	return dAtA
}
func randFieldTimeWindow(dAtA []byte, r randyTimeWindow, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		dAtA = encodeVarintPopulateTimeWindow(dAtA, uint64(key))
		v11 := r.Int63()
		if r.Intn(2) == 0 {
			v11 *= -1
		}
		dAtA = encodeVarintPopulateTimeWindow(dAtA, uint64(v11))
	case 1:
		dAtA = encodeVarintPopulateTimeWindow(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		dAtA = encodeVarintPopulateTimeWindow(dAtA, uint64(key))
		ll := r.Intn(100)
		dAtA = encodeVarintPopulateTimeWindow(dAtA, uint64(ll))
		for j := 0; j < ll; j++ {
			dAtA = append(dAtA, byte(r.Intn(256)))
		}
	default:
		dAtA = encodeVarintPopulateTimeWindow(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return dAtA
}
func encodeVarintPopulateTimeWindow(dAtA []byte, v uint64) []byte {
	for v >= 1<<7 {
		dAtA = append(dAtA, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	dAtA = append(dAtA, uint8(v))
	return dAtA
}
func (m *TimeWindowWhen) Size() (n int) {
	var l int
	_ = l
	l = m.Days.Size()
	n += 1 + l + sovTimeWindow(uint64(l))
	return n
}

func (m *TimeWindowDays) Size() (n int) {
	var l int
	_ = l
	if len(m.All) > 0 {
		for _, e := range m.All {
			l = e.Size()
			n += 1 + l + sovTimeWindow(uint64(l))
		}
	}
	if len(m.Sunday) > 0 {
		for _, e := range m.Sunday {
			l = e.Size()
			n += 1 + l + sovTimeWindow(uint64(l))
		}
	}
	if len(m.Monday) > 0 {
		for _, e := range m.Monday {
			l = e.Size()
			n += 1 + l + sovTimeWindow(uint64(l))
		}
	}
	if len(m.Tuesday) > 0 {
		for _, e := range m.Tuesday {
			l = e.Size()
			n += 1 + l + sovTimeWindow(uint64(l))
		}
	}
	if len(m.Wednesday) > 0 {
		for _, e := range m.Wednesday {
			l = e.Size()
			n += 1 + l + sovTimeWindow(uint64(l))
		}
	}
	if len(m.Thursday) > 0 {
		for _, e := range m.Thursday {
			l = e.Size()
			n += 1 + l + sovTimeWindow(uint64(l))
		}
	}
	if len(m.Friday) > 0 {
		for _, e := range m.Friday {
			l = e.Size()
			n += 1 + l + sovTimeWindow(uint64(l))
		}
	}
	if len(m.Saturday) > 0 {
		for _, e := range m.Saturday {
			l = e.Size()
			n += 1 + l + sovTimeWindow(uint64(l))
		}
	}
	return n
}

func (m *TimeWindowTimeRange) Size() (n int) {
	var l int
	_ = l
	l = len(m.Begin)
	if l > 0 {
		n += 1 + l + sovTimeWindow(uint64(l))
	}
	l = len(m.End)
	if l > 0 {
		n += 1 + l + sovTimeWindow(uint64(l))
	}
	return n
}

func sovTimeWindow(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozTimeWindow(x uint64) (n int) {
	return sovTimeWindow(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *TimeWindowWhen) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTimeWindow
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
			return fmt.Errorf("proto: TimeWindowWhen: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TimeWindowWhen: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Days", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimeWindow
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
				return ErrInvalidLengthTimeWindow
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Days.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTimeWindow(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTimeWindow
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
func (m *TimeWindowDays) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTimeWindow
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
			return fmt.Errorf("proto: TimeWindowDays: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TimeWindowDays: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field All", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimeWindow
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
				return ErrInvalidLengthTimeWindow
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.All = append(m.All, &TimeWindowTimeRange{})
			if err := m.All[len(m.All)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sunday", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimeWindow
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
				return ErrInvalidLengthTimeWindow
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sunday = append(m.Sunday, &TimeWindowTimeRange{})
			if err := m.Sunday[len(m.Sunday)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Monday", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimeWindow
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
				return ErrInvalidLengthTimeWindow
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Monday = append(m.Monday, &TimeWindowTimeRange{})
			if err := m.Monday[len(m.Monday)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tuesday", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimeWindow
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
				return ErrInvalidLengthTimeWindow
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Tuesday = append(m.Tuesday, &TimeWindowTimeRange{})
			if err := m.Tuesday[len(m.Tuesday)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Wednesday", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimeWindow
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
				return ErrInvalidLengthTimeWindow
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Wednesday = append(m.Wednesday, &TimeWindowTimeRange{})
			if err := m.Wednesday[len(m.Wednesday)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Thursday", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimeWindow
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
				return ErrInvalidLengthTimeWindow
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Thursday = append(m.Thursday, &TimeWindowTimeRange{})
			if err := m.Thursday[len(m.Thursday)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Friday", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimeWindow
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
				return ErrInvalidLengthTimeWindow
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Friday = append(m.Friday, &TimeWindowTimeRange{})
			if err := m.Friday[len(m.Friday)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Saturday", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimeWindow
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
				return ErrInvalidLengthTimeWindow
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Saturday = append(m.Saturday, &TimeWindowTimeRange{})
			if err := m.Saturday[len(m.Saturday)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTimeWindow(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTimeWindow
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
func (m *TimeWindowTimeRange) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTimeWindow
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
			return fmt.Errorf("proto: TimeWindowTimeRange: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TimeWindowTimeRange: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Begin", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimeWindow
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
				return ErrInvalidLengthTimeWindow
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Begin = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field End", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTimeWindow
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
				return ErrInvalidLengthTimeWindow
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.End = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTimeWindow(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTimeWindow
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
func skipTimeWindow(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTimeWindow
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
					return 0, ErrIntOverflowTimeWindow
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
					return 0, ErrIntOverflowTimeWindow
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
				return 0, ErrInvalidLengthTimeWindow
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowTimeWindow
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
				next, err := skipTimeWindow(dAtA[start:])
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
	ErrInvalidLengthTimeWindow = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTimeWindow   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("time_window.proto", fileDescriptorTimeWindow) }

var fileDescriptorTimeWindow = []byte{
	// 367 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0x4f, 0x4a, 0xc3, 0x40,
	0x14, 0x87, 0x3b, 0x4d, 0xfa, 0x6f, 0x2a, 0x82, 0xe3, 0x26, 0x2a, 0x24, 0xa5, 0x6e, 0xba, 0x31,
	0x85, 0xba, 0x71, 0xa3, 0x48, 0xe8, 0x5e, 0x0c, 0x42, 0xc1, 0x8d, 0x24, 0x66, 0x9a, 0x06, 0x9a,
	0x99, 0x92, 0x99, 0xa1, 0xe4, 0x26, 0xe2, 0x09, 0x3c, 0x82, 0x47, 0xe8, 0xd2, 0x13, 0x04, 0x8d,
	0xbb, 0x9e, 0xc0, 0xa5, 0xcc, 0xc4, 0xda, 0x2e, 0x74, 0x31, 0x9b, 0xc9, 0xbc, 0xcc, 0xef, 0xfb,
	0x78, 0x3c, 0x1e, 0x3c, 0xe0, 0x49, 0x8a, 0x1f, 0x96, 0x09, 0x89, 0xe8, 0xd2, 0x5d, 0x64, 0x94,
	0x53, 0xd4, 0x65, 0x98, 0x30, 0xe1, 0xf2, 0x7c, 0x81, 0xd9, 0xf1, 0x59, 0x9c, 0xf0, 0x99, 0x08,
	0xdd, 0x47, 0x9a, 0x0e, 0x63, 0x1a, 0xd3, 0xa1, 0xca, 0x84, 0x62, 0xaa, 0x2a, 0x55, 0xa8, 0x5b,
	0xc5, 0xf6, 0x6f, 0xe0, 0xfe, 0x5d, 0x92, 0xe2, 0x89, 0xf2, 0x4d, 0x66, 0x98, 0xa0, 0x4b, 0x68,
	0x46, 0x41, 0xce, 0x2c, 0xd0, 0x03, 0x83, 0xee, 0xe8, 0xc4, 0xdd, 0x91, 0xbb, 0xdb, 0xe8, 0x38,
	0xc8, 0x99, 0xb7, 0xb7, 0x2a, 0x9c, 0xda, 0xba, 0x70, 0x14, 0xe0, 0xab, 0xb3, 0xff, 0x6c, 0xee,
	0x1a, 0x65, 0x0c, 0x5d, 0x40, 0x23, 0x98, 0xcf, 0x2d, 0xd0, 0x33, 0x06, 0xdd, 0x51, 0xef, 0x1f,
	0xa1, 0xbc, 0xf9, 0x01, 0x89, 0xb1, 0x67, 0xae, 0x0a, 0x07, 0xf8, 0x12, 0x41, 0x57, 0xb0, 0xc9,
	0x04, 0x89, 0x82, 0xdc, 0xaa, 0x6b, 0xc1, 0x3f, 0x94, 0xe4, 0x53, 0xaa, 0x78, 0x43, 0x8f, 0xaf,
	0x28, 0x74, 0x0d, 0x5b, 0x5c, 0x60, 0x26, 0x05, 0xa6, 0x96, 0x60, 0x83, 0xa1, 0x31, 0xec, 0x2c,
	0x71, 0x44, 0x2a, 0x47, 0x43, 0xcb, 0xb1, 0x05, 0x91, 0x07, 0xdb, 0x7c, 0x26, 0x32, 0x25, 0x69,
	0x6a, 0x49, 0x7e, 0x39, 0x39, 0x8b, 0x69, 0x96, 0x48, 0x43, 0x4b, 0x6f, 0x16, 0x15, 0x25, 0x7b,
	0x60, 0x01, 0x17, 0x99, 0x34, 0xb4, 0xf5, 0x7a, 0xd8, 0x70, 0xfd, 0x5b, 0x78, 0xf8, 0x47, 0x0c,
	0x39, 0xb0, 0x11, 0xe2, 0x38, 0x21, 0x6a, 0xe7, 0x3a, 0x5e, 0x67, 0x5d, 0x38, 0xd5, 0x0f, 0xbf,
	0xfa, 0xa0, 0x23, 0x68, 0x60, 0x12, 0x59, 0x75, 0xf5, 0xdc, 0x5a, 0x17, 0x8e, 0x2c, 0x7d, 0x79,
	0x78, 0xa7, 0x5f, 0x1f, 0x36, 0x78, 0x29, 0x6d, 0xf0, 0x5a, 0xda, 0x60, 0x55, 0xda, 0xe0, 0xad,
	0xb4, 0xc1, 0x7b, 0x69, 0x83, 0xa7, 0x4f, 0xbb, 0x76, 0xdf, 0x50, 0xbd, 0x85, 0x4d, 0xb5, 0xec,
	0xe7, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x36, 0xff, 0x53, 0x21, 0x3d, 0x03, 0x00, 0x00,
}
