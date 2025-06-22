package timewarrior

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type DateSuite struct {
	suite.Suite
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDateSuite(t *testing.T) {
	suite.Run(t, new(DateSuite))
}

func (suite *DateSuite) TestConvertDateStringToTime_OK() {
	now, err := time.Parse("2006-01-02", "2006-01-02")
	suite.NoError(err)

	tests := []struct {
		input    string
		expected string
	}{
		{"mon",        "2006-01-02"},
		{"tue",        "2006-01-03"},
		{"wed",        "2006-01-04"},
		{"thu",        "2006-01-05"},
		{"fri",        "2006-01-06"},
		{"sat",        "2006-01-07"},
		{"sun",        "2006-01-01"},
		{"today",      "2006-01-02"},
		{"now",        "2006-01-02"},
		{"yesterday",  "2006-01-01"},
		{"tomorrow",   "2006-01-03"},
		{"2006-01-02", "2006-01-02"},
		{"20060102",   "2006-01-02"},
	}
	for _, test := range tests {
		result, err := ConvertDateStringToTime(now, test.input)
		suite.NoError(err)
		suite.Equal(test.expected, result.Format("2006-01-02"))
	}
}
