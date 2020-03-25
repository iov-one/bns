package account

import (
	"encoding/json"
	"github.com/iov-one/bns/cmd/bnsapi/bnsapitest"
	_ "github.com/iov-one/bns/cmd/bnsapi/bnsapitest"
	"github.com/iov-one/bns/cmd/bnsapi/client"
	"github.com/iov-one/bns/cmd/bnsapi/util"
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/cmd/bnsd/x/account"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAccountAccountDetailHandler(t *testing.T) {
	bns := &bnsapitest.BnsClientMock{
		PostResults: map[string]map[string]client.AbciQueryResponse{
			"/accounts": {
				"666F6F2A626172": bnsapitest.NewAbciQueryResponse(t,
					[][]byte{
						[]byte("foo*bar"),
					},
					[]weave.Persistent{
						&account.Account{
							Name:   "foo",
							Domain: "bar",
						},
					}),
			},
		},
	}

	h := DetailHandler{Bns: bns}

	reqBody := `{ "json-rpc": 2.0, "method": "abci_query", "params": { "path": "/accounts", "data": "2F666F6F2F626172"}}`

	r, _ := http.NewRequest("POST", "/something/xyz/foo*bar", strings.NewReader(reqBody))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("failed response: %d %s", w.Code, w.Body)
	}

	var acc account.Account
	if err := json.NewDecoder(w.Body).Decode(&acc); err != nil {
		t.Fatalf("cannot decode JSON response: %s", err)
	}
	if acc.Name != "foo" || acc.Domain != "bar" {
		t.Fatalf("unexpected response: %+v", acc)
	}
}

func TestAccountAccountssHandler(t *testing.T) {
	bns := &bnsapitest.BnsClientMock{
		GetResults: map[string]client.AbciQueryResponse{
			"/abci_query?data=%22%3A%22&path=%22%2Faccounts%3Frange%22": bnsapitest.NewAbciQueryResponse(t,
				[][]byte{
					[]byte("first"),
					[]byte("second"),
				},
				[]weave.Persistent{
					&account.Account{Name: "first", Domain: "adomain"},
					&account.Account{Name: "second", Domain: "adomain"},
				}),
		},
	}
	h := AccountsHandler{Bns: bns}

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	bnsapitest.AssertAPIResponse(t, w, []util.KeyValue{
		{
			Key:   []byte("first"),
			Value: &account.Account{Name: "first", Domain: "adomain"},
		},
		{
			Key:   []byte("second"),
			Value: &account.Account{Name: "second", Domain: "adomain"},
		},
	})
}

func TestAccountAccountssHandlerOffsetAndFilter(t *testing.T) {
	bns := &bnsapitest.BnsClientMock{
		GetResults: map[string]client.AbciQueryResponse{
			"/abci_query?data=%2261646f6d61696e%3A36363639373237333734%3A61646f6d61696f%22&path=%22%2Faccounts%2Fdomain%3Frange%22": bnsapitest.NewAbciQueryResponse(t, nil, nil),
		},
	}
	h := AccountsHandler{Bns: bns}

	r, _ := http.NewRequest("GET", "/?offset=6669727374&domain=adomain", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	bnsapitest.AssertAPIResponse(t, w, []util.KeyValue{})
}

func TestAccountDomainsHandler(t *testing.T) {
	bns := &bnsapitest.BnsClientMock{
		GetResults: map[string]client.AbciQueryResponse{
			"/abci_query?data=%22%3A%22&path=%22%2Fdomains%3Frange%22": bnsapitest.NewAbciQueryResponse(t,
				[][]byte{
					[]byte("first"),
					[]byte("second"),
				},
				[]weave.Persistent{
					&account.Domain{Domain: "f"},
					&account.Domain{Domain: "s"},
				}),
			"/abci_query?data=%227365636f6e64%3A%22&path=%22%2Fdomains%3Frange%22": bnsapitest.NewAbciQueryResponse(t, nil, nil),
		},
	}
	h := DomainsHandler{Bns: bns}

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	bnsapitest.AssertAPIResponse(t, w, []util.KeyValue{
		{
			Key:   []byte("first"),
			Value: &account.Domain{Domain: "f"},
		},
		{
			Key:   []byte("second"),
			Value: &account.Domain{Domain: "s"},
		},
	})
}
