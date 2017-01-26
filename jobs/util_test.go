package jobs

import "testing"

func Test_removeAccents(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"no accent", "normal", "normal"},
		{"with accent", "café", "cafe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeAccents(tt.arg); got != tt.want {
				t.Errorf("removeAccents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hash(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"test1", "something", "1af17e73721dbe0c40011b82ed4bb1a7dbe3ce29"},
		{"test2", "Another Thing!", "7855b7a4593be53562cd2702f3a1ea24f2d7a52f"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hash(tt.arg); got != tt.want {
				t.Errorf("hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createID(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"test1", []string{"a", "b", "c"}, "e088d9e3f737c091378fe8494936b16d51eb42ee"},
		{"test2", []string{"This", "is an", "ID!"}, "35b573bc871f9ef897f95b15561f40f5377a3c12"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createID(tt.args...); got != tt.want {
				t.Errorf("createID() = %v, want %v", got, tt.want)
			}
		})
	}
}
