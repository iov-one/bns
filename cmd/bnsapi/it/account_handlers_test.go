package it

import (
	"github.com/iov-one/bns/cmd/bnsapi/bnsapitest"
	"github.com/iov-one/bns/cmd/bnsapi/client"
	"github.com/iov-one/bns/cmd/bnsapi/handlers/account"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func TestAccountAccountsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	v, ok := os.LookupEnv("IT_TENDERMINT")
	if !ok {
		panic("set IT_TENDERMINT env var")
	}

	bnscli := client.NewHTTPBnsClient(v)
	h := account.AccountsHandler{Bns: bnscli}

	r, _ := http.NewRequest("GET", "/account/accounts?offset=1000", nil)
	rc := httptest.NewRecorder()
	h.ServeHTTP(rc, r)

	want, err := os.Open("account.offset.test.json")
	if err != nil {
		t.Fatal(err)
	}
	bnsapitest.AssertAPIResponseBasic(t, want, rc.Body)

	u := make(url.Values)
	u.Add("owner", "C1721181E83376EF978AA4A9A38A5E27C08C7BB2")
	r, _ = http.NewRequest("GET", "/account/accounts?" + u.Encode(), nil)
	rc = httptest.NewRecorder()
	h.ServeHTTP(rc, r)

	want, err = os.Open("account.owner.test.json")
	if err != nil {
		t.Fatal(err)
	}
	bnsapitest.AssertAPIResponseBasic(t, want, rc.Body)

	r, _ = http.NewRequest("GET", "/account/accounts/domain?offset=1000", nil)
	rc = httptest.NewRecorder()
	h.ServeHTTP(rc, r)

	want, err = os.Open("account.offset.test.json")
	if err != nil {
		t.Fatal(err)
	}
	bnsapitest.AssertAPIResponseBasic(t, want, rc.Body)
}
