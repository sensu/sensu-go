package timeutil

import (
	"runtime"
	"testing"
	"time"

	v2 "github.com/sensu/core/v2"
)

func TestDateToTime(t *testing.T) {
	baseTime, _ := time.Parse(time.RFC3339, "2018-05-10T15:04:00Z")

	// Our test cases
	tests := []struct {
		name	string
		str	string
		want	time.Time
		wantErr	bool
	}{
		{
			name:	"RFC3339 UTC",
			str:	"2018-05-10T15:04:00Z",
			want:	baseTime,
		},
		{
			name:	"RFC3339 with numeric zone offset",
			str:	"2018-05-10T07:04:00-08:00",
			want:	baseTime,
		},
		{
			name:	"RFC3339 with space delimiter",
			str:	"2018-05-10 07:04:00 -08:00",
			want:	baseTime,
		},
		{
			name:	"legacy UTC",
			str:	"May 10 2018 3:04PM UTC",
			want:	baseTime,
		},
		{
			name:		"unknown format",
			str:		"Mon Jan _2 15:04:05 2006",
			wantErr:	true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dateToTime(tt.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("dateToTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.want.Equal(got) {
				t.Errorf("dateToTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKitchenToTime(t *testing.T) {
	baseTime, err := time.ParseInLocation(time.Kitchen, "3:04PM", time.UTC)
	if err != nil {
		t.Fatal(err)
	}

	// Our test cases
	tests := []struct {
		name		string
		str		string
		skipWindows	bool	// Canonical timezones are not supported on Windows
		want		time.Time
		wantErr		bool
	}{
		{
			name:	"24-hour kitchen UTC",
			str:	"15:04 UTC",
			want:	baseTime,
		},
		//{
		//	name:        "24-hour kitchen with canonical timezone",
		//	str:         "07:04 America/Vancouver",
		//	skipWindows: true,
		//	want:        baseTime,
		//},
		{
			name:	"24-hour kitchen with numeric zone offset",
			str:	"07:04 -08:00",
			want:	baseTime,
		},
		{
			name:		"24-hour kitchen with unknown location",
			str:		"07:04 foo",
			wantErr:	true,
		},
		{
			name:	"12-hour kitchen UTC",
			str:	"3:04PM UTC",
			want:	baseTime,
		},
		//{
		//	name:        "12-hour kitchen with canonical timezone",
		//	str:         "10:04AM America/Montreal",
		//	skipWindows: true,
		//	want:        baseTime,
		//},
		{
			name:		"12-hour kitchen with unknown location",
			str:		"10:04AM foo",
			wantErr:	true,
		},
		{
			name:		"unknown format",
			str:		"15:04:05.000000000",
			wantErr:	true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if runtime.GOOS == "windows" && tt.skipWindows {
				return
			}

			got, err := kitchenToTime(tt.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("kitchenToTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.want.Equal(got) {
				t.Errorf("kitchenToTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToUnix(t *testing.T) {
	tests := []struct {
		name	string
		str	string
		now	bool
		want	int64
		wantErr	bool
	}{
		{
			name:	"RFC3339",
			str:	"2018-05-10T15:04:00Z",
			want:	1525964640,
		},
		{
			name:	"0 value",
			str:	"0",
			now:	true,
		},
		{
			name:	"now value",
			str:	"now",
			now:	true,
		},
		{
			name:		"unknown value",
			str:		"3:04PM",
			wantErr:	true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ConvertToUnix(tc.str)
			if (err != nil) != tc.wantErr {
				t.Errorf("ConvertToUnix() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if tc.now {
				// Remove the last two digits of the unix timestamp when comparing the
				// current timestamp, in case the test took more than 1 second (i.e.
				// 1526399179 vs 1526399180)
				if time.Now().Unix()/100 != got/100 {
					t.Errorf("ConvertToUnix() = %v, want ~ %v", got, time.Now().Unix())
					return
				}
				return
			}
			if got != tc.want {
				t.Errorf("ConvertToUnix() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestConvertToUTC(t *testing.T) {
	b, err := time.ParseInLocation(time.Kitchen, "3:04PM", time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	begin := b.Format(time.Kitchen)
	e, err := time.ParseInLocation(time.Kitchen, "4:04PM", time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	end := e.Format(time.Kitchen)

	tests := []struct {
		name		string
		window		*v2.TimeWindowTimeRange
		skipWindows	bool	// Canonical timezones are not supported on Windows
		wantBegin	string
		wantEnd		string
		wantErr		bool
	}{
		{
			name:	"12-hour kitchen UTC",
			window: &v2.TimeWindowTimeRange{
				Begin:	"3:04PM UTC",
				End:	"4:04PM UTC",
			},
			wantBegin:	begin,
			wantEnd:	end,
		},
		//{
		//	name: "12-hour kitchen canonical timezone",
		//	window: &types.TimeWindowTimeRange{
		//		Begin: "7:04AM America/Vancouver",
		//		End:   "8:04AM America/Vancouver",
		//	},
		//	skipWindows: true,
		//	wantBegin:   begin,
		//	wantEnd:     end,
		//},
		{
			name:	"24-hour kitchen numeric timezone",
			window: &v2.TimeWindowTimeRange{
				Begin:	"07:04 -08:00",
				End:	"08:04 -08:00",
			},
			wantBegin:	begin,
			wantEnd:	end,
		},
		{
			name:	"invalid begin",
			window: &v2.TimeWindowTimeRange{
				Begin:	"15:04:00.000000000",
				End:	"08:04 -08:00",
			},
			wantBegin:	"15:04:00.000000000",
			wantEnd:	"08:04 -08:00",
			wantErr:	true,
		},
		{
			name:	"invalid end",
			window: &v2.TimeWindowTimeRange{
				Begin:	"07:04 -08:00",
				End:	"16:04:00.000000000",
			},
			wantBegin:	"07:04 -08:00",
			wantEnd:	"16:04:00.000000000",
			wantErr:	true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if runtime.GOOS == "windows" && tc.skipWindows {
				return
			}

			if err := ConvertToUTC(tc.window); (err != nil) != tc.wantErr {
				t.Errorf("ConvertToUTC() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if tc.window.Begin != tc.wantBegin {
				t.Errorf("ConvertToUTC() = %v, want begin %v", tc.window.Begin, tc.wantBegin)
				return
			}

			if tc.window.End != tc.wantEnd {
				t.Errorf("ConvertToUTC() = %v, want end %v", tc.window.End, tc.wantEnd)
			}
		})
	}
}

func TestHumanTimestamp(t *testing.T) {
	tests := []struct {
		name		string
		timestamp	int64
		wantNA		bool
	}{
		{
			name:		"valid timestamp",
			timestamp:	1525964640,
			wantNA:		false,
		},
		{
			name:		"zero timestamp",
			timestamp:	0,
			wantNA:		true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := HumanTimestamp(tc.timestamp); (got == "N/A") != tc.wantNA {
				t.Errorf("HumanTimestamp() length = %v, wanted N/A? %v", len(got), tc.wantNA)
			}
		})
	}
}
