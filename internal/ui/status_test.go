package ui

import (
	"errors"
	"testing"
)

func TestSetWithError(t *testing.T) {
	status := NewStatus()
	errMsg := "something went wrong"
	status.Set("message", errors.New(errMsg))

	if status.kind != errorStatus {
		t.Errorf("Expected %d got %d", errorStatus, status.kind)
	}

	if status.msg != errMsg {
		t.Errorf("Expected %s got %s", errMsg, status.msg)
	}
}

func TestSetWithMsg(t *testing.T) {
	status := NewStatus()
	status.Set("message", nil)

	if status.kind != okStatus {
		t.Errorf("Expected %d got %d", errorStatus, status.kind)
	}

	if status.msg != "message" {
		t.Errorf("Expected %s got %s", "message", status.msg)
	}
}
