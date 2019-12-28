package streamql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

type model struct {
	A             string           `db:"a" json:"a"`
	B             int              `db:"b" json:"b"`
	C             bool             `db:"c" json:"c"`
	PStringSetP   *string          `db:"p_string_p" json:"pStringP"`
	PStringSetNP  *string          `db:"p_string_np" json:"pStringNP"`
	StringSlice   []string         `db:"s_slice" json:"sSlice"`
	StringPSlice  *[]string        `db:"s_pslice" json:"sPslice"`
	StringNPSlice *[]string        `db:"s_npslice" json:"sNpslice"`
	Bytes         []byte           `db:"bytes" json:"bytes"`
	RBytes        json.RawMessage  `db:"r_bytes" json:"rBytes"`
	RPBytes       *json.RawMessage `db:"r_pbytes" json:"rPbytes"`
	RNPBytes      *json.RawMessage `db:"r_npbytes" json:"rNpbytes"`
}
type testNexter struct {
	rows    int
	currRow int
}

func (tn *testNexter) Next() bool {
	if tn.rows > tn.currRow {
		tn.currRow++
		return true
	}
	return false
}

func (tn *testNexter) Columns() ([]string, error) {
	return []string{
			"a", "b", "c",
			"p_string_p", "p_string_np",
			"s_slice", "s_pslice", "s_npslice",
			"bytes", "r_bytes",
			"r_pbytes", "r_npbytes",
		},
		nil
}

func (tn *testNexter) Scan(v ...interface{}) error {
	v[0] = "This is A"
	v[1] = 100
	v[2] = true
	s := "Pointer String coming at you"
	v[3] = &s
	v[4] = "NPointer String coming at you"
	v[5] = []string{"1", "2", "andre 3000"}
	strslice := []string{"100", "200", "big boi"}
	v[6] = &strslice
	var npslice []string
	v[7] = npslice // needs to be the right type but still nil
	v[8] = make([]byte, 0)
	v[9] = []byte("{}")
	var rbytes = json.RawMessage([]byte("{}"))
	v[10] = &rbytes
	var nprbytes json.RawMessage
	v[11] = nprbytes

	return nil
}

func encode(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

func TestThing(t *testing.T) {
	b := &bytes.Buffer{}
	tn := &testNexter{rows: 3}
	if err := Stream(tn, &model{}, encode, b); err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}
	d := json.NewDecoder(b)

	for {
		m := &model{}
		if err := d.Decode(m); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if m.A != "This is A" {
			t.Fatal("fail")
		}
	}
}

func BenchmarkThing(b *testing.B) {
	w := &bytes.Buffer{}
	tn := &testNexter{rows: 3000}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Stream(tn, &model{}, encode, w); err != nil {
			fmt.Println(err)
			b.Fatal(err)
		}
	}
}
