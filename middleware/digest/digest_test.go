package digest

import "testing"

func TestDigestHashAlgorithms(t *testing.T) {
	d := newDigestAuth(Config{Username: "user", Password: "password"})

	tests := []struct {
		algorithm string
		want      string
	}{
		{algorithm: "MD5", want: "5d41402abc4b2a76b9719d911017c592"},
		{algorithm: "SHA-256", want: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{algorithm: "SHA-512-256", want: "e30d87cfa2a75db545eac4d61baf970366a8357c7f72fa95b52d0accb698f13a"},
	}

	for _, tt := range tests {
		t.Run(tt.algorithm, func(t *testing.T) {
			if got := d.hashString("hello", tt.algorithm); got != tt.want {
				t.Fatalf("hashString() = %q, want %q", got, tt.want)
			}
			if got := d.hashString("hello", tt.algorithm+"-sess"); got != tt.want {
				t.Fatalf("hashString() for sess variant = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSupportsAlgorithm(t *testing.T) {
	d := newDigestAuth(Config{})

	for _, algorithm := range []string{"", "MD5", "MD5-sess", "SHA-256", "SHA-256-sess", "SHA-512-256", "SHA-512-256-sess"} {
		if !d.supportsAlgorithm(algorithm) {
			t.Errorf("expected %q to be supported", algorithm)
		}
	}
	if d.supportsAlgorithm("SHA-512") {
		t.Error("expected SHA-512 to be rejected")
	}
}
