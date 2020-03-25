package bnsapitest

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iov-one/bns/cmd/bnsapi/client"
	"github.com/iov-one/bns/cmd/bnsapi/util"
	"github.com/iov-one/weave"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func NewAbciQueryResponse(t testing.TB, keys [][]byte, models []weave.Persistent) client.AbciQueryResponse {
	t.Helper()
	k, v := util.SerializePairs(t, keys, models)

	return client.AbciQueryResponse{
		Response: client.AbciQueryResponseResponse{
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

	result := client.AbciQueryResponse{
		Response: client.AbciQueryResponseResponse{
			Key:   []byte("foo"),
			Value: []byte("bar"),
		},
	}
	bns := BnsClientMock{GetResults: map[string]client.AbciQueryResponse{
		"/foo": result,
	}}
	var response client.AbciQueryResponse
	if err := bns.Get(context.Background(), "/foo", &response); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(response, result) {
		t.Fatalf("unexpected response: %+v", response)
	}
}

type BnsClientMock struct {
	GetResults  map[string]client.AbciQueryResponse
	PostResults map[string]map[string]client.AbciQueryResponse
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

func (mock *BnsClientMock) Post(ctx context.Context, path string, data []byte, dest interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	hexData := strings.ToUpper(hex.EncodeToString(data))
	resp, ok := mock.PostResults[path][hexData]
	if !ok {
		raw, _ := url.PathUnescape(path)
		return fmt.Errorf("no result declared in mock for %q %q (%q)", path, hexData, raw)
	}

	v := reflect.ValueOf(dest)
	// Below panics if cannot be fullfilled. User did something wrong and
	// this is test so panic is acceptable.
	src := reflect.ValueOf(resp)
	v.Elem().Set(src)

	return mock.Err
}
