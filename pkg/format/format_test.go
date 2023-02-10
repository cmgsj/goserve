package format_test

import (
	"goserve/pkg/format"
	"math"
	"testing"
	"time"
)

func TestThousandsSeparator(t *testing.T) {
	type testcase struct {
		input int
		want  string
	}
	testcases := []testcase{
		{input: 0, want: "0"},
		{input: 1, want: "1"},
		{input: 11, want: "11"},
		{input: 111, want: "111"},
		{input: 1111, want: "1,111"},
		{input: 11111, want: "11,111"},
		{input: 111111, want: "111,111"},
		{input: 1111111, want: "1,111,111"},
		{input: 11111111, want: "11,111,111"},
		{input: 111111111, want: "111,111,111"},
		{input: 1111111111, want: "1,111,111,111"},
		{input: math.MaxInt32, want: "2,147,483,647"},
		{input: math.MaxInt64, want: "9,223,372,036,854,775,807"},
	}
	for _, tc := range testcases {
		got := format.ThousandsSeparator(tc.input)
		if got != tc.want {
			t.Errorf("ThousandsSeparator(%d) = %s, want %s", tc.input, got, tc.want)
		}
	}
}

func TestFileSize(t *testing.T) {
	type testcase struct {
		input int64
		want  string
	}
	testcases := []testcase{
		{input: 0, want: "0.00B"},
		{input: 1, want: "1.00B"},
		{input: 1024 - 1, want: "1023.00B"},
		{input: 1024, want: "1.00KB"},
		{input: 1024 * 1024 / 2, want: "512.00KB"},
		{input: 1024 * 1024, want: "1.00MB"},
		{input: 1024 * 1024 * 1024 / 2, want: "512.00MB"},
		{input: 1024 * 1024 * 1024, want: "1.00GB"},
		{input: 1024 * 1024 * 1024 * 1024 / 2, want: "512.00GB"},
	}
	for _, tc := range testcases {
		got := format.FileSize(tc.input)
		if got != tc.want {
			t.Errorf("FileSize(%d) = %s, want %s", tc.input, got, tc.want)
		}
	}
}

func TestTimeDuration(t *testing.T) {
	type testcase struct {
		input time.Duration
		want  string
	}
	testcases := []testcase{
		{input: 0, want: "0.00ns"},
		{input: 1, want: "1.00ns"},
		{input: 1000 / 2, want: "500.00ns"},
		{input: 1000, want: "1.00µs"},
		{input: 1000 * 1000 / 2, want: "500.00µs"},
		{input: 1000 * 1000, want: "1.00ms"},
		{input: 1000 * 1000 * 1000 / 2, want: "500.00ms"},
		{input: 1000 * 1000 * 1000, want: "1.00s"},
		{input: 1000 * 1000 * 1000 * 60 / 2, want: "30.00s"},
		{input: 1000 * 1000 * 1000 * 60, want: "1.00min"},
		{input: 1000 * 1000 * 1000 * 60 * 60 / 2, want: "30.00min"},
	}
	for _, tc := range testcases {
		got := format.TimeDuration(tc.input)
		if got != tc.want {
			t.Errorf("TimeDuration(%d) = %s, want %s", tc.input, got, tc.want)
		}
	}
}
