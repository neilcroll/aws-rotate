package main

import (
	"testing"

	"aws-rotate/internal/rotator"
)

func TestDetermineUpdateMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		concurrent bool
		sequential bool
		want       rotator.UpdateMode
	}{
		{
			name:       "defaults to concurrent",
			concurrent: false,
			sequential: false,
			want:       rotator.ModeConcurrent,
		},
		{
			name:       "sequential flag takes precedence",
			concurrent: false,
			sequential: true,
			want:       rotator.ModeSequential,
		},
		{
			name:       "concurrent flag keeps concurrent mode",
			concurrent: true,
			sequential: false,
			want:       rotator.ModeConcurrent,
		},
		{
			name:       "both flags defaults to concurrent",
			concurrent: true,
			sequential: true,
			want:       rotator.ModeConcurrent,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := determineUpdateMode(tt.concurrent, tt.sequential); got != tt.want {
				t.Fatalf("determineUpdateMode(%v, %v) = %v, want %v", tt.concurrent, tt.sequential, got, tt.want)
			}
		})
	}
}
