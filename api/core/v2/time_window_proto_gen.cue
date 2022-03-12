package v2

// TimeWindowWhen defines the "when" attributes for time windows
#TimeWindowWhen: {
	// Days is a hash of days
	days?: #TimeWindowDays @protobuf(1,TimeWindowDays,#"(gogoproto.jsontag)="days""#,"(gogoproto.nullable)=false")
}

// TimeWindowDays defines the days of a time window
#TimeWindowDays: {
	all?: [...#TimeWindowTimeRange] @protobuf(1,TimeWindowTimeRange,"(gogoproto.nullable)")
	sunday?: [...#TimeWindowTimeRange] @protobuf(2,TimeWindowTimeRange,"(gogoproto.nullable)")
	monday?: [...#TimeWindowTimeRange] @protobuf(3,TimeWindowTimeRange,"(gogoproto.nullable)")
	tuesday?: [...#TimeWindowTimeRange] @protobuf(4,TimeWindowTimeRange,"(gogoproto.nullable)")
	wednesday?: [...#TimeWindowTimeRange] @protobuf(5,TimeWindowTimeRange,"(gogoproto.nullable)")
	thursday?: [...#TimeWindowTimeRange] @protobuf(6,TimeWindowTimeRange,"(gogoproto.nullable)")
	friday?: [...#TimeWindowTimeRange] @protobuf(7,TimeWindowTimeRange,"(gogoproto.nullable)")
	saturday?: [...#TimeWindowTimeRange] @protobuf(8,TimeWindowTimeRange,"(gogoproto.nullable)")
}

// TimeWindowTimeRange defines the time ranges of a time
#TimeWindowTimeRange: {
	// Begin is the time which the time window should begin, in the format
	// '3:00PM', which satisfies the time.Kitchen format
	begin?: string @protobuf(1,string,#"(gogoproto.jsontag)="begin""#)

	// End is the time which the filter should end, in the format '3:00PM', which
	// satisfies the time.Kitchen format
	end?: string @protobuf(2,string,#"(gogoproto.jsontag)="end""#)
}
