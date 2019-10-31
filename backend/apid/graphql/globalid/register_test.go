package globalid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterLookup(t *testing.T) {
	register := NewRegister()
	usersTranslator := commonTranslator{}
	register.translators["users"] = usersTranslator
	checksTranslator := commonTranslator{}
	register.translators["checks"] = checksTranslator

	testCases := []struct {
		gid     string
		want    Decoder
		wantErr bool
	}{
		{"srn:sdfasdfasdf:123", nil, true},
		{"srn:users:123", usersTranslator, false},
		{"srn:checks:is-google-up", checksTranslator, false},
	}
	for _, tc := range testCases {
		t.Run(tc.gid, func(t *testing.T) {
			components, _ := Parse(tc.gid)
			decoder, err := register.Lookup(*components)
			assert.Equal(t, tc.want, decoder)
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestReverseLookup(t *testing.T) {
	usersTranslator := commonTranslator{
		name: "users",
		isResponsibleFunc: func(record interface{}) bool {
			return record.(string) == "users"
		},
	}
	checksTranslator := commonTranslator{
		name: "checks",
		isResponsibleFunc: func(record interface{}) bool {
			return record.(string) == "checks"
		},
	}
	register := NewRegister()
	register.translators["users"] = usersTranslator
	register.translators["checks"] = checksTranslator

	testCases := []struct {
		record  interface{}
		wantRes bool
		wantErr bool
	}{
		{"not-going-to-find", false, true},
		{"users", true, false},
		{"checks", true, false},
	}
	for _, tc := range testCases {
		t.Run(tc.record.(string), func(t *testing.T) {
			encoder, err := register.ReverseLookup(tc.record)
			assert.Equal(t, tc.wantErr, err != nil)
			assert.Equal(t, tc.wantRes, encoder != nil)
		})
	}
}
