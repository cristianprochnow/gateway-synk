package tests

import (
	"synk/gateway/app/util"
	"testing"
	"time"
)

func TestTimeNotEmpty(t *testing.T) {
	now := util.Now()

	if now == "" {
		t.Errorf("util.Now() returned empty value")
	}
}

func TestTimeIsRight(t *testing.T) {
	now := util.Now()
	nowToTest := time.Now().Format(time.RFC3339)

	if now != nowToTest {
		t.Errorf("util.Now with inconsistent values")
	}
}

func TestNowFormat(t *testing.T) {
	nowStr := util.Now()
	_, err := time.Parse(time.RFC3339, nowStr)

	if err != nil {
		t.Errorf("util.Now() returned an invalid time format. Got: %s, Error: %v", nowStr, err)
	}
}
