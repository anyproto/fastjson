package fastjson

import (
	"encoding/json"
	"fmt"
	"testing"
	"unicode/utf8"
)

func assertJSONStringRoundTrip(t *testing.T, s string) {
	t.Helper()
	var a Arena
	for _, object := range []bool{false, true} {
		v := a.NewString(s)
		if object {
			v = a.NewObject()
			v.Set(s, a.NewString(s))
		}
		wire := v.MarshalTo([]byte("prefix:"))[7:]
		if !json.Valid(wire) {
			t.Fatalf("invalid JSON for %q: %q", s, wire)
		}
		if object {
			var decoded map[string]string
			if err := json.Unmarshal(wire, &decoded); err != nil {
				t.Fatal(err)
			}
			if len(decoded) != 1 || decoded[s] != s {
				t.Fatalf("object round trip: %q", decoded)
			}
		} else {
			var decoded string
			if err := json.Unmarshal(wire, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded != s {
				t.Fatalf("string round trip: got %q want %q", decoded, s)
			}
		}
	}
}

func TestJSONStringEscapes(t *testing.T) {
	for _, s := range []string{"", "plain", "\n\U000e0067", "quoted \"😀\"\\\U000e007f", "café\n日本語\u2028\u2029\u007f"} {
		assertJSONStringRoundTrip(t, s)
	}
	for c := rune(0); c <= 0x9f; c++ {
		t.Run(fmt.Sprintf("U+%04X", c), func(t *testing.T) {
			assertJSONStringRoundTrip(t, "before\n"+string(c)+"\U000e0067after")
		})
	}
}

func FuzzJSONStringEscapes(f *testing.F) {
	for _, s := range []string{"", "\n\U000e0067", "\a\v\x00\x7f", "quote\"slash\\"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) {
			t.Skip()
		}
		assertJSONStringRoundTrip(t, s)
	})
}
