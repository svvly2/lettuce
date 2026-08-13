package main

import (
	"testing"
	"time"
)

func TestRemoveFinishedJobAfter(t *testing.T) {
	finishedAt := time.Now()
	state := newAppState()
	state.Queue = []UploadJob{{ID: "1", SourceAssetID: "1", Status: "failed", finishedAt: finishedAt}}

	state.removeFinishedJobAfter("1", finishedAt, time.Millisecond)
	if len(state.Queue) != 0 {
		t.Fatalf("expected finished job to be removed, got %d jobs", len(state.Queue))
	}
}

func TestOldTimerDoesNotRemoveReplacementJob(t *testing.T) {
	oldFinishedAt := time.Now()
	state := newAppState()
	state.Queue = []UploadJob{{ID: "1", SourceAssetID: "1", Status: "uploading"}}

	state.removeFinishedJobAfter("1", oldFinishedAt, time.Millisecond)
	if len(state.Queue) != 1 || state.Queue[0].Status != "uploading" {
		t.Fatal("an old cleanup timer removed the replacement job")
	}
}
