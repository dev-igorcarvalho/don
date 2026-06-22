package primitives

import (
	"testing"
)

func TestClaudeResult_Failure(t *testing.T) {
	tests := []struct {
		name    string
		r       *ClaudeResult
		wantErr bool
		errText string
	}{
		{
			name: "no error",
			r: &ClaudeResult{
				IsError: false,
				Result:  "success message",
			},
			wantErr: false,
		},
		{
			name: "has error",
			r: &ClaudeResult{
				IsError: true,
				Result:  "something went wrong",
			},
			wantErr: true,
			errText: "something went wrong",
		},
		{
			name: "has error with empty result",
			r: &ClaudeResult{
				IsError: true,
				Result:  "",
			},
			wantErr: true,
			errText: "",
		},
		{
			name:    "nil receiver",
			r:       nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.r.Failure()
			if (err != nil) != tt.wantErr {
				t.Errorf("ClaudeResult.Failure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errText {
				t.Errorf("ClaudeResult.Failure() error = %v, wantErrText %v", err.Error(), tt.errText)
			}
		})
	}
}
