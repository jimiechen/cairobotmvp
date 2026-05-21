package commonlib

import "testing"

func TestErrorCodeValues(t *testing.T) {
	tests := []struct {
		name     string
		actual   int
		expected int
	}{
		{"CodeSuccess", CodeSuccess, 10200},
		{"CodeBadRequest", CodeBadRequest, 10400},
		{"CodeUnauthorized", CodeUnauthorized, 10401},
		{"CodeNotFound", CodeNotFound, 10404},
		{"CodeInternalError", CodeInternalError, 10500},
		{"CodeTarsNotImplemented", CodeTarsNotImplemented, 10501},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, tt.actual)
			}
		})
	}
}

func TestErrorCodeUniqueness(t *testing.T) {
	codes := map[int]string{
		CodeSuccess:            "CodeSuccess",
		CodeBadRequest:         "CodeBadRequest",
		CodeUnauthorized:       "CodeUnauthorized",
		CodeNotFound:           "CodeNotFound",
		CodeInternalError:      "CodeInternalError",
		CodeTarsNotImplemented: "CodeTarsNotImplemented",
	}

	if len(codes) != 6 {
		t.Fatalf("expected 6 unique codes, got %d", len(codes))
	}
}
