package slice

import "testing"

func TestContainsString(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		value string
		want  bool
	}{
		{
			name:  "value-found",
			slice: []string{"a", "b", "c", "d"},
			value: "c",
			want:  true,
		},
		{
			name:  "value-missing",
			slice: []string{"a", "b", "c", "d"},
			value: "f",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsString(tt.slice, tt.value); got != tt.want {
				t.Errorf("ContainsString() = %v, want %v", got, tt.want)
			}
		})
	}
}
