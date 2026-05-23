package domain

import "testing"

func TestOperationType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		op       OperationType
		expected bool
	}{
		{"ADD 合法", OperationAdd, true},
		{"MOD 合法", OperationMod, true},
		{"DEL 合法", OperationDel, true},
		{"空字符串不合法", OperationType(""), false},
		{"非法值不合法", OperationType("UPDATE"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.op.IsValid(); got != tt.expected {
				t.Errorf("OperationType.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}
