package scp

import (
	"github.com/ImSingee/tt"
	"testing"
)

func TestConvertFileModeToPermString(t *testing.T) {
	tt.AssertEqual(t, "0777", ConvertFileModeToPermString(0o777))
	tt.AssertEqual(t, "0755", ConvertFileModeToPermString(0o755))
	tt.AssertEqual(t, "0644", ConvertFileModeToPermString(0o644))
	tt.AssertEqual(t, "0000", ConvertFileModeToPermString(0))
	tt.AssertEqual(t, "0777", ConvertFileModeToPermString(0o1777))
}
