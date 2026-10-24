package mlp

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

type cvec struct {
	Name string `json:"name"`
	Hex  string `json:"hex"`
}
type cfile struct {
	Vectors []cvec `json:"vectors"`
}

func TestConformanceVectors(t *testing.T) {
	data, err := os.ReadFile("conformance/vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var cf cfile
	if err := json.Unmarshal(data, &cf); err != nil {
		t.Fatal(err)
	}
	for _, v := range cf.Vectors {
		b, _ := hex.DecodeString(v.Hex)
		f, n, err := ParseFragment(b)
		if err != nil {
			t.Fatalf("%s: %v", v.Name, err)
		}
		if n != len(b) {
			t.Fatalf("%s: consumed %d != %d", v.Name, n, len(b))
		}
		_ = f
	}
}
