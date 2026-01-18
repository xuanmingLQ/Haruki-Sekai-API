package orderedmsgpack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/iancoleman/orderedmap"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/vmihailenco/msgpack/v5/msgpcode"
)

type OrderedMsgpackDecoder struct {
	msgpack.Decoder
}

func (d *OrderedMsgpackDecoder) DecodeInterface() (interface{}, error) {
	c, err := d.PeekCode()
	if err != nil {
		return nil, err
	}
	if msgpcode.IsFixedNum(c) {
		return decodeInt(d)
	}
	if msgpcode.IsFixedMap(c) {
		return d.DecodeMap()
	}
	if msgpcode.IsFixedArray(c) {
		return d.DecodeSlice()
	}
	if msgpcode.IsFixedString(c) {
		return d.DecodeString()
	}
	switch c {
	case msgpcode.Nil:
		return nil, d.DecodeNil()
	case msgpcode.False, msgpcode.True:
		return d.DecodeBool()
	case msgpcode.Float, msgpcode.Double:
		return decodefloat(d)
	case msgpcode.Uint8, msgpcode.Uint16, msgpcode.Uint32, msgpcode.Uint64:
		return decodeUint(d)
	case msgpcode.Int8, msgpcode.Int16, msgpcode.Int32, msgpcode.Int64:
		return decodeInt(d)
	case msgpcode.Str8, msgpcode.Str16, msgpcode.Str32,
		msgpcode.Bin8, msgpcode.Bin16, msgpcode.Bin32:
		return d.DecodeString()
	case msgpcode.Array16, msgpcode.Array32:
		return d.DecodeSlice()
	case msgpcode.Map16, msgpcode.Map32:
		return d.DecodeMap()
	}
	var v any
	if err := d.Decoder.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}
func (d *OrderedMsgpackDecoder) DecodeMap() (*orderedmap.OrderedMap, error) {
	n, err := d.DecodeMapLen()
	if err != nil {
		return nil, err
	}
	om := orderedmap.New()
	om.SetEscapeHTML(false)
	for i := 0; i < n; i++ {
		k, err := d.DecodeInterface()
		if err != nil {
			return nil, err
		}
		v, err := d.DecodeInterface()
		if err != nil {
			return nil, err
		}
		var key string
		if jn, ok := k.(json.Number); ok {
			key = jn.String()
		} else {
			key = fmt.Sprint(k)
		}
		om.Set(key, v)
	}
	return om, nil
}
func (d *OrderedMsgpackDecoder) DecodeSlice() ([]interface{}, error) {
	n, err := d.DecodeArrayLen()
	if err != nil {
		return nil, err
	}
	s := make([]interface{}, n)
	for i := 0; i < n; i++ {
		v, err := d.DecodeInterface()
		if err != nil {
			return nil, err
		}
		s[i] = v
	}
	return s, nil
}
func decodeOrderedMap(d *msgpack.Decoder) (interface{}, error) {
	od := new(OrderedMsgpackDecoder)
	od.Decoder = *d
	return od.DecodeMap()
}
func decodeUint(d *OrderedMsgpackDecoder) (json.Number, error) {
	u, err := d.DecodeUint64()
	if err != nil {
		return "", err
	}
	return json.Number(strconv.FormatUint(u, 10)), nil
}

func decodeInt(d *OrderedMsgpackDecoder) (json.Number, error) {
	i, err := d.DecodeInt64()
	if err != nil {
		return "", err
	}
	return json.Number(strconv.FormatInt(i, 10)), nil
}

func decodefloat(d *OrderedMsgpackDecoder) (json.Number, error) {
	f, err := d.DecodeFloat64()
	if err != nil {
		return "", err
	}
	raw := strconv.FormatFloat(f, 'f', 17, 64)
	if strings.Contains(raw, ".") {
		trimmed := strings.TrimRight(raw, "0")
		if strings.HasSuffix(trimmed, ".") {
			trimmed += "0"
		}
		raw = trimmed
	} else {
		raw += ".0"
	}
	return json.Number(raw), nil
}
func NewOrderedMsgpackDecoder(r io.Reader) *OrderedMsgpackDecoder {
	d := new(OrderedMsgpackDecoder)
	d.Reset(r)
	d.SetMapDecoder(decodeOrderedMap)
	return d
}
func TestMa(t *testing.T) {
	type testStruct struct {
		A string
		B []interface{}
		C string
		D struct {
			DEFS  string
			CCEGG int
			PPSLF bool
			EGSDG []interface{}
		}
	}
	v := testStruct{
		A: "123",
		B: []interface{}{
			"1234",
			true,
			nil,
			testStruct{
				A: "pou",
				B: nil,
				C: "1231244",
				D: struct {
					DEFS  string
					CCEGG int
					PPSLF bool
					EGSDG []interface{}
				}{
					DEFS:  "",
					CCEGG: 1234,
					PPSLF: false,
					EGSDG: []interface{}{1, 23, 4, "23"},
				},
			},
		},
		C: "wwweere",
		D: struct {
			DEFS  string
			CCEGG int
			PPSLF bool
			EGSDG []interface{}
		}{
			DEFS:  "124545",
			CCEGG: 0,
			PPSLF: true,
			EGSDG: nil,
		},
	}
	b, err := msgpack.Marshal(v)
	d := NewOrderedMsgpackDecoder(bytes.NewReader(b))
	om, err := d.DecodeMap()
	if err != nil {
		t.Fatal(err)
	}
	j, err := json.Marshal(om)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(string(j))
}
