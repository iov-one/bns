package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/iov-one/bns/cmd/bnsapi/bnsapitest"
	"github.com/iov-one/bns/cmd/bnsapi/models"
	"github.com/iov-one/bns/cmd/bnsapi/util"
	rpctypes "github.com/tendermint/tendermint/rpc/lib/types"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
)

func TestABCIKeyQuery(t *testing.T) {
	// Run a fake Tendermint API server that will answer to only expected
	// query requests.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		type fullBody struct {
			Rpc       float32         `json:"json-rpc"`
			Method    string          `json:"method"`
			BodyParam abciQueryParams `json:"params"`
		}

		var fullBodyParam fullBody
		err = json.Unmarshal(body, &fullBodyParam)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		bodyParam := fullBodyParam.BodyParam

		switch {
		case bodyParam.Path == "/myentity" && bodyParam.Data == "656E746974796B6579":
			writeServerResponse(t, w, [][]byte{
				[]byte("0001"),
			}, []weave.Persistent{
				&persistentMock{Raw: []byte("content")},
			})
		case bodyParam.Path == "/myentity":
			writeServerResponse(t, w, nil, nil)
		default:
			t.Fatalf("unknown condition: %q", bodyParam)
		}

	}))
	defer srv.Close()

	bns := NewHTTPBnsClient(srv.URL)

	model := persistentMock{Raw: []byte("content")}
	dest := models.KeyModel{
		Model: &model,
	}
	if err := ABCIKeyQuery(context.Background(), bns, "/myentity", []byte("entitykey"), &dest); err != nil {
		t.Fatalf("cannot get by key: %v", err)
	}

	if err := ABCIKeyQuery(context.Background(), bns, "/myentity", []byte("xxxxx"), &dest); !errors.ErrNotFound.Is(err) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestABCIKeyQueryIter(t *testing.T) {
	keys := [][]byte{
		[]byte("0001"),
		[]byte("0002"),
		[]byte("0003"),
	}
	values := []weave.Persistent{
		&persistentMock{Raw: []byte("1")},
		&persistentMock{Raw: []byte("2")},
		&persistentMock{Raw: []byte("3")},
	}

	// Run a fake Tendermint API server that will answer to only expected
	// query requests.

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		type fullBody struct {
			Rpc       float32         `json:"json-rpc"`
			Method    string          `json:"method"`
			BodyParam abciQueryParams `json:"params"`
		}

		var fullBodyParam fullBody
		err = json.Unmarshal(body, &fullBodyParam)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		bodyParam := fullBodyParam.BodyParam

		switch {
		case bodyParam.Path == "/myentity" && bodyParam.Data == "656E746974796B6579":
			writeServerResponse(t, w, keys, values)
		case bodyParam.Path == "/myentity":
			writeServerResponse(t, w, nil, nil)
		default:
			t.Fatalf("unknown condition: %q", bodyParam)
		}

	}))
	defer srv.Close()

	bns := NewHTTPBnsClient(srv.URL)

	it := ABCIKeyQueryIter(context.Background(), bns, "/myentity", []byte("entitykey"))
	objects := make([]util.KeyValue, 0, util.PaginationMaxItems)
iterate:
	for {
		m := &persistentMock{}
		switch key, err := it.Next(m); {
		case err == nil:
			objects = append(objects, util.KeyValue{
				Key:   key,
				Value: m,
			})
			if len(objects) == util.PaginationMaxItems {
				break iterate
			}
		case errors.ErrIteratorDone.Is(err):
			break iterate
		default:
			log.Fatalf("ABCI query: %s", err)
		}
	}

	var resKeys [][]byte
	var resValues []weave.Persistent
	for _, o := range objects {
		resKeys = append(resKeys, o.Key)
		resValues = append(resValues, o.Value)
	}

	if !reflect.DeepEqual(resKeys, keys) {
		for i, k := range keys {
			t.Logf("key %2d: %q", i, k)
		}
		t.Fatalf("unexpected %d keys", len(keys))
	}

	if !reflect.DeepEqual(values, resValues) {
		for i, k := range keys {
			t.Logf("value %2d: %q", i, k)
		}
		t.Fatalf("unexpected %d values", len(values))
	}

	it = ABCIKeyQueryIter(context.Background(), bns, "/myentity", []byte("xxx"))
	if _, err := it.Next(nil); !errors.ErrNotFound.Is(err) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestBnsClientDo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/foo" {
			t.Fatalf("unexpected path: %q", r.URL)
		}
		_, _ = io.WriteString(w, `
			{
				"result": "a result"
			}
		`)
	}))
	defer srv.Close()

	bns := NewHTTPBnsClient(srv.URL)

	var result string
	if err := bns.Get(context.Background(), "/foo", &result); err != nil {
		t.Fatalf("get: %s", err)
	}
	if result != "a result" {
		t.Fatalf("unexpected result: %q", result)
	}
}

func TestABCIFullRangeQuery(t *testing.T) {
	// Run a fake Tendermint API server that will answer to only expected
	// query requests.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		type fullBody struct {
			Rpc       float32         `json:"json-rpc"`
			Method    string          `json:"method"`
			BodyParam abciQueryParams `json:"params"`
		}

		var fullBodyParam fullBody
		err = json.Unmarshal(body, &fullBodyParam)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		bodyParam := fullBodyParam.BodyParam

		switch {
		case bodyParam.Path == "/myquery?range" && bodyParam.Data == "":
			writeServerResponse(t, w, [][]byte{
				[]byte("0001"),
				[]byte("0002"),
				[]byte("0003"),
			}, []weave.Persistent{
				&persistentMock{Raw: []byte("1")},
				&persistentMock{Raw: []byte("2")},
				&persistentMock{Raw: []byte("3")},
			})
		case bodyParam.Path == "/myquery?range" && bodyParam.Data == "33303330333033333A":
			writeServerResponse(t, w, [][]byte{
				[]byte("0003"), // Filter is inclusive.
				[]byte("0004"),
			}, []weave.Persistent{
				&persistentMock{Raw: []byte("3")},
				&persistentMock{Raw: []byte("4")},
			})
		case bodyParam.Path == "/myquery?range" && bodyParam.Data == "33303330333033343A":
			writeServerResponse(t, w, [][]byte{
				[]byte("0004"), // Filter is inclusive.
			}, []weave.Persistent{
				&persistentMock{Raw: []byte("4")},
			})
		default:
			t.Logf("data: %q", bodyParam.Data)
			t.Errorf("not supported request: %q", r.URL)
			http.Error(w, "not supported", http.StatusNotImplemented)
		}
	}))
	defer srv.Close()

	bns := NewHTTPBnsClient(srv.URL)
	it := ABCIFullRangeQuery(context.Background(), bns, "/myquery", "")

	var keys [][]byte
consumeIterator:
	for {
		switch key, err := it.Next(ignoreModel{}); {
		case err == nil:
			keys = append(keys, key)
		case errors.ErrIteratorDone.Is(err):
			break consumeIterator
		default:
			t.Fatalf("iterator failed: %s", err)
		}

	}

	// ABCIFullRangeQuery iterator must return all available keys in the
	// right order and each key only once. We do not check values because
	// we ignore them in this test.
	wantKeys := [][]byte{
		[]byte("0001"),
		[]byte("0002"),
		[]byte("0003"),
		[]byte("0004"),
	}

	if !reflect.DeepEqual(wantKeys, keys) {
		for i, k := range keys {
			t.Logf("key %2d: %q", i, k)
		}
		t.Fatalf("unexpected %d keys", len(keys))
	}
}

func TestABCIPrefixQuery(t *testing.T) {
	// Run a fake Tendermint API server that will answer to only expected
	// query requests.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rpc rpctypes.RPCRequest
		if err := json.NewDecoder(r.Body).Decode(&rpc); err != nil {
			t.Fatalf("unexpected path: %q", r.URL)
		}

		var params abciQueryParams
		if err := json.Unmarshal(rpc.Params, &params); err != nil {
			t.Fatalf("unexpected path: %q", r.URL)
		}

		switch {
		case params.Path == "/myquery?prefix" && params.Data == "":
			writeServerResponse(t, w, [][]byte{
				[]byte("0001"),
				[]byte("0002"),
				[]byte("0003"),
			}, []weave.Persistent{
				&persistentMock{Raw: []byte("1")},
				&persistentMock{Raw: []byte("2")},
				&persistentMock{Raw: []byte("3")},
			})
		case params.Path == "/myquery?prefix" && params.Data == "0001":
			writeServerResponse(t, w, [][]byte{
				[]byte("0001"),
				[]byte("0002"),
				[]byte("0003"),
			}, []weave.Persistent{
				&persistentMock{Raw: []byte("1")},
				&persistentMock{Raw: []byte("2")},
				&persistentMock{Raw: []byte("3")},
			})
		default:
			t.Logf("path: %q", params.Path)
			t.Logf("data: %q", params.Data)
			t.Errorf("not supported request: %q", r.URL)
			http.Error(w, "not supported", http.StatusNotImplemented)
		}
	}))
	defer srv.Close()

	bns := NewHTTPBnsClient(srv.URL)
	it := ABCIPrefixQuery(context.Background(), bns, "/myquery", []byte(""))

	var keys [][]byte
consumeIterator:
	for {
		switch key, err := it.Next(ignoreModel{}); {
		case err == nil:
			keys = append(keys, key)
		case errors.ErrIteratorDone.Is(err):
			break consumeIterator
		default:
			t.Fatalf("iterator failed: %s", err)
		}

	}

	// ABCIFullRangeQuery iterator must return all available keys in the
	// right order and each key only once. We do not check values because
	// we ignore them in this test.
	wantKeys := [][]byte{
		[]byte("0001"),
		[]byte("0002"),
		[]byte("0003"),
	}

	if !reflect.DeepEqual(wantKeys, keys) {
		for i, k := range keys {
			t.Logf("key %2d: %q", i, k)
		}
		t.Fatalf("unexpected %d keys", len(keys))
	}
}

// ignoreModel is a stub. Its unmarshal is a no-op. Use it together with an
// iterator if you do not care about the result unloading.
type ignoreModel struct {
	orm.Model
}

func (ignoreModel) Unmarshal([]byte) error { return nil }

func writeServerResponse(t testing.TB, w http.ResponseWriter, keys [][]byte, models []weave.Persistent) {
	t.Helper()

	k, v := bnsapitest.SerializePairs(t, keys, models)

	// Minimal acceptable by our code jsonrpc response.
	type dict map[string]interface{}
	payload := dict{
		"result": dict{
			"response": dict{
				"key":   k,
				"value": v,
			},
		},
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("cannot write response: %s", err)
	}

}

type persistentMock struct {
	orm.Model

	Raw []byte
	Err error
}

func (m *persistentMock) Unmarshal(raw []byte) error {
	if m.Raw != nil && !bytes.Equal(m.Raw, raw) {
		return fmt.Errorf("want %q, got %q", m.Raw, raw)
	}
	m.Raw = raw
	return m.Err
}

func (m *persistentMock) Marshal() ([]byte, error) {
	return m.Raw, m.Err
}
