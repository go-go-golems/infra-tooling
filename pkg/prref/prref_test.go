package prref

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		in   string
		want Ref
	}{
		{"https://github.com/go-go-golems/discord-bot/pull/9", Ref{Owner: "go-go-golems", Repo: "discord-bot", Number: 9}},
		{"go-go-golems/goja-git#2", Ref{Owner: "go-go-golems", Repo: "goja-git", Number: 2}},
	}
	for _, tt := range tests {
		got, err := Parse(tt.in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("Parse(%q)=%#v want %#v", tt.in, got, tt.want)
		}
	}
}
