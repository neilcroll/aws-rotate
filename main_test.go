package main

import "testing"

func TestDetermineUpdateMode(t *testing.T) {
	tests := []struct {
		name       string
		concurrent bool
		sequential bool
		want       updateMode
	}{
		{
			name:       "defaults to concurrent when no flags set",
			concurrent: false,
			sequential: false,
			want:       modeConcurrent,
		},
		{
			name:       "sequential flag forces sequential mode",
			concurrent: false,
			sequential: true,
			want:       modeSequential,
		},
		{
			name:       "concurrent flag forces concurrent mode",
			concurrent: true,
			sequential: false,
			want:       modeConcurrent,
		},
		{
			name:       "concurrent wins when both flags provided",
			concurrent: true,
			sequential: true,
			want:       modeConcurrent,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := determineUpdateMode(tt.concurrent, tt.sequential)
			if got != tt.want {
				t.Fatalf("determineUpdateMode(%v, %v) = %v, want %v", tt.concurrent, tt.sequential, got, tt.want)
			}
		})
	}
}

func TestRunProfileUpdates(t *testing.T) {
	t.Parallel()
	var concurrentCalled, sequentialCalled bool
	concurrentFn := func(_ []Profile, _, _ string) {
		concurrentCalled = true
	}
	sequentialFn := func(_ []Profile, _, _ string) {
		sequentialCalled = true
	}

	runProfileUpdates(modeConcurrent, nil, "", "", concurrentFn, sequentialFn)
	if !concurrentCalled || sequentialCalled {
		t.Fatalf("modeConcurrent should call concurrentFn only; got concurrent=%v sequential=%v", concurrentCalled, sequentialCalled)
	}

	concurrentCalled = false
	sequentialCalled = false

	runProfileUpdates(modeSequential, nil, "", "", concurrentFn, sequentialFn)
	if !sequentialCalled || concurrentCalled {
		t.Fatalf("modeSequential should call sequentialFn only; got concurrent=%v sequential=%v", concurrentCalled, sequentialCalled)
	}
}
