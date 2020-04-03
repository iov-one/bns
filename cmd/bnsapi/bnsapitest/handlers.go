package bnsapitest

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	md "github.com/iov-one/bns/cmd/bnsapi/models"
	"github.com/iov-one/bns/cmd/bnsapi/util"
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/app"
	rpctypes "github.com/tendermint/tendermint/rpc/lib/types"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func NewAbciQueryResponse(t testing.TB, keys [][]byte, m []weave.Persistent) md.AbciQueryResponse {
	t.Helper()
	k, v := SerializePairs(t, keys, m)

	return md.AbciQueryResponse{
		Response: md.AbciQueryResponseResponse{
			Key:   k,
			Value: v,
		},
	}
}

func AssertAPIResponse(t testing.TB, w *httptest.ResponseRecorder, want []util.KeyValue) {
	t.Helper()

	if w.Code != http.StatusOK {
		t.Log(w.Body)
		t.Fatalf("response code %d", w.Code)
	}

	var payload struct {
		Objects json.RawMessage
	}
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("cannot decode JSON serialized body: %s", err)
	}

	// We cannot unmarshal returned JSON because KeyValue structure does
	// not declare what type Value is. Instead of comparing Go objects,
	// compare JSON output. We know what is the expected JSON content for
	// given KeyValue collection.
	rawGot := []byte(payload.Objects)

	rawWant, err := json.MarshalIndent(want, "", "\t")
	if err != nil {
		t.Fatalf("cannot JSON serialize expected result: %s", err)
	}

	// Because rawGot is part of a bigger JSON message its indentation
	// differs. Indentation is not relevant so it can be removed for
	// comparison.
	if !bytes.Equal(removeTabs(rawGot), removeTabs(rawWant)) {
		t.Logf("want JSON response:\n%s", string(rawWant))
		t.Logf("got JSON response:\n%s", string(rawGot))
		t.Fatal("unexpected response")
	}
}

func removeTabs(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte("\t"), []byte(""))
}

func TestBnsClientMock(t *testing.T) {
	// Just to be sure, test the mock.

	result := md.AbciQueryResponse{
		Response: md.AbciQueryResponseResponse{
			Key:   []byte("foo"),
			Value: []byte("bar"),
		},
	}
	bns := BnsClientMock{GetResults: map[string]md.AbciQueryResponse{
		"/foo": result,
	}}
	var response md.AbciQueryResponse
	if err := bns.Get(context.Background(), "/foo", &response); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(response, result) {
		t.Fatalf("unexpected response: %+v", response)
	}
}

type BnsClientMock struct {
	GetResults  map[string]md.AbciQueryResponse
	PostResults map[string]map[string]md.AbciQueryResponse
	Err         error
}

func (mock *BnsClientMock) Get(ctx context.Context, path string, dest interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	resp, ok := mock.GetResults[path]
	if !ok {
		raw, _ := url.PathUnescape(path)
		return fmt.Errorf("no result declared in mock for %q (%q)", path, raw)
	}

	v := reflect.ValueOf(dest)
	// Below panics if cannot be fullfilled. User did something wrong and
	// this is test so panic is acceptable.
	src := reflect.ValueOf(resp)
	v.Elem().Set(src)

	return mock.Err
}

func (mock *BnsClientMock) Post(ctx context.Context, data []byte, dest interface{}) error {
	var req rpctypes.RPCRequest
	err := json.Unmarshal(data, &req)
	if err != nil {
		return err
	}

	type params struct {
		Path string `json:"path"`
		Data string `json:"data"`
	}

	var p params
	_ = json.Unmarshal(req.Params, &p)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	resp, ok := mock.PostResults[p.Path][p.Data]
	if !ok {
		raw, _ := url.PathUnescape(p.Path)
		return fmt.Errorf("no result declared in mock for %q %q (%q)", p.Path, p.Data, raw)
	}

	v := reflect.ValueOf(dest)
	// Below panics if cannot be fullfilled. User did something wrong and
	// this is test so panic is acceptable.
	src := reflect.ValueOf(resp)
	v.Elem().Set(src)

	return mock.Err
}

func AssertAPIResponseBasic(t testing.TB, want, got io.Reader) {
	t.Helper()

	var w json.RawMessage
	if err := json.NewDecoder(want).Decode(&w); err != nil {
		t.Fatalf("cannot decode JSON serialized body: %s", err)
	}
	var g json.RawMessage
	if err := json.NewDecoder(got).Decode(&g); err != nil {
		t.Fatalf("cannot decode JSON serialized body: %s", err)
	}

	w1, _ := json.MarshalIndent(w, "", "    ")
	g1, _ := json.MarshalIndent(g, "", "    ")

	if !bytes.Equal(w1, g1) {
		t.Logf("want JSON response:\n%s", w1)
		t.Logf("got JSON response:\n%s", g1)
		t.Fatal("unexpected response")
	}
}

func SerializePairs(t testing.TB, keys [][]byte, models []weave.Persistent) ([]byte, []byte) {
	t.Helper()

	if len(keys) != len(models) {
		t.Fatalf("keys and models length must be the same: %d != %d", len(keys), len(models))
	}

	kset := app.ResultSet{
		Results: keys,
	}
	kraw, err := kset.Marshal()
	if err != nil {
		t.Fatalf("cannot marshal keys: %s", err)
	}

	var values [][]byte
	for i, m := range models {
		raw, err := m.Marshal()
		if err != nil {
			t.Fatalf("cannot marshal %d model: %s", i, err)
		}
		values = append(values, raw)
	}
	vset := app.ResultSet{
		Results: values,
	}
	vraw, err := vset.Marshal()
	if err != nil {
		t.Fatalf("cannot marshal values: %s", err)
	}

	return kraw, vraw
}

func SequenceID(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}
