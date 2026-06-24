package primitives

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		wantPath  string
		wantFound bool
	}{
		{
			name:      "empty context",
			ctx:       context.Background(),
			wantPath:  "",
			wantFound: false,
		},
		{
			name:      "context with session dir path",
			ctx:       context.WithValue(context.Background(), sessionDirKey{}, "/path/to/session"),
			wantPath:  "/path/to/session",
			wantFound: true,
		},
		{
			name:      "context with incorrect type",
			ctx:       context.WithValue(context.Background(), sessionDirKey{}, 12345),
			wantPath:  "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotPath, gotFound := SessionDir(tt.ctx)
			assert.Equal(t, tt.wantPath, gotPath)
			assert.Equal(t, tt.wantFound, gotFound)
		})
	}
}

func TestSessionName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		wantName  string
		wantFound bool
	}{
		{
			name:      "empty context",
			ctx:       context.Background(),
			wantName:  "",
			wantFound: false,
		},
		{
			name:      "context with session name",
			ctx:       context.WithValue(context.Background(), sessionNameKey{}, "my-session"),
			wantName:  "my-session",
			wantFound: true,
		},
		{
			name:      "context with incorrect type",
			ctx:       context.WithValue(context.Background(), sessionNameKey{}, 12345),
			wantName:  "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotName, gotFound := SessionName(tt.ctx)
			assert.Equal(t, tt.wantName, gotName)
			assert.Equal(t, tt.wantFound, gotFound)
		})
	}
}

func TestSessionID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		wantID    string
		wantFound bool
	}{
		{
			name:      "empty context",
			ctx:       context.Background(),
			wantID:    "",
			wantFound: false,
		},
		{
			name:      "context with session id",
			ctx:       context.WithValue(context.Background(), sessionIDKey{}, "id-12345"),
			wantID:    "id-12345",
			wantFound: true,
		},
		{
			name:      "context with incorrect type",
			ctx:       context.WithValue(context.Background(), sessionIDKey{}, 12345),
			wantID:    "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotID, gotFound := SessionID(tt.ctx)
			assert.Equal(t, tt.wantID, gotID)
			assert.Equal(t, tt.wantFound, gotFound)
		})
	}
}

func TestArtifactDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		wantPath  string
		wantFound bool
	}{
		{
			name:      "empty context",
			ctx:       context.Background(),
			wantPath:  "",
			wantFound: false,
		},
		{
			name:      "context with artifact dir path",
			ctx:       context.WithValue(context.Background(), artifactDirKey{}, "/path/to/artifacts"),
			wantPath:  "/path/to/artifacts",
			wantFound: true,
		},
		{
			name:      "context with incorrect type",
			ctx:       context.WithValue(context.Background(), artifactDirKey{}, 12345),
			wantPath:  "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotPath, gotFound := ArtifactDir(tt.ctx)
			assert.Equal(t, tt.wantPath, gotPath)
			assert.Equal(t, tt.wantFound, gotFound)
		})
	}
}

func TestLogger(t *testing.T) {
	t.Parallel()

	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		ctx        context.Context
		wantLogger *slog.Logger
	}{
		{
			name:       "empty context fallback to default",
			ctx:        context.Background(),
			wantLogger: slog.Default(),
		},
		{
			name:       "context with logger",
			ctx:        context.WithValue(context.Background(), loggerKey{}, testLogger),
			wantLogger: testLogger,
		},
		{
			name:       "context with incorrect type fallback to default",
			ctx:        context.WithValue(context.Background(), loggerKey{}, "not-a-logger"),
			wantLogger: slog.Default(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotLogger := Logger(tt.ctx)
			assert.Equal(t, tt.wantLogger, gotLogger)
		})
	}
}
