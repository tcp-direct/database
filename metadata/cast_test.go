package metadata

import (
	"testing"

	"git.tcp.direct/tcp.direct/database/models"
)

func TestCast(t *testing.T) {
	mock := NewPlaceholder("mock")
	if IsCanonicalMetadata(mock) {
		t.Errorf("expected false, got true")
	}
	realM := NewMeta("yeet")
	if !IsCanonicalMetadata(realM) {
		t.Errorf("expected true, got false")
	}
	var generic models.Metadata
	generic = realM
	cast, err := CastToMetadata(generic)
	if err != nil {
		t.Fatalf("error casting: %v", err)
	}
	if cast != realM {
		t.Errorf("expected %v, got %v", realM, cast)
	}
	generic = mock
	cast, err = CastToMetadata(generic)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if cast != nil {
		t.Errorf("expected nil, got %T", cast)
	}
}
