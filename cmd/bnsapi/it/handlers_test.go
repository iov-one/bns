package it

import (
	"encoding/hex"
	"github.com/iov-one/bns/cmd/bnsapi/bnsapitest"
	"github.com/iov-one/bns/cmd/bnsapi/client"
	"github.com/iov-one/bns/cmd/bnsapi/handlers"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestCashBalanceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	v, ok := os.LookupEnv("IT_TENDERMINT")
	if !ok {
		panic("set IT_TENDERMINT env var")
	}

	bnscli := client.NewHTTPBnsClient(v)
	h := handlers.CashBalanceHandler{Bns: bnscli}

	u := make(url.Values)
	u.Add("address", "C68D1B4F0E39ADED977159C577BE6F9F46700292")
	apiPath := "/cash/balances?" + u.Encode()
	r, _ := http.NewRequest("GET", apiPath, nil)

	rc := httptest.NewRecorder()
	h.ServeHTTP(rc, r)

	want := strings.NewReader(`{
            "metadata": {
                "schema": 1
            },
            "coins": [
                {
                    "whole": 1000,
                    "ticker": "IOV"
                }
            ]
        }`)
	bnsapitest.AssertAPIResponseBasic(t, want, rc.Body)

	u = make(url.Values)
	u.Add("address", "tiov1c6x3kncw8xk7m9m3t8zh00n0nar8qq5jh7cv0j")
	apiPath = "/cash/balances?" + u.Encode()
	r, _ = http.NewRequest("GET", apiPath, nil)

	rc = httptest.NewRecorder()
	h.ServeHTTP(rc, r)

	want = strings.NewReader(`{
            "metadata": {
                "schema": 1
            },
            "coins": [
                {
                    "whole": 1000,
                    "ticker": "IOV"
                }
            ]
        }`)
	bnsapitest.AssertAPIResponseBasic(t, want, rc.Body)
}

func TestMultisigContractsHandlerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	v, ok := os.LookupEnv("IT_TENDERMINT")
	if !ok {
		panic("set IT_TENDERMINT env var")
	}

	bnscli := client.NewHTTPBnsClient(v)
	h := handlers.MultisigContractsHandler{Bns: bnscli}

	r, _ := http.NewRequest("GET", "/multisig/contracts", nil)
	rc := httptest.NewRecorder()
	h.ServeHTTP(rc, r)

	want, err := os.Open("multisigcontract.test.json")
	if err != nil {
		t.Fatal(err)
	}
	bnsapitest.AssertAPIResponseBasic(t, want, rc.Body)

	// test with prefix
	u := make(url.Values)
	u.Add("prefix", `1`)
	rURL := "/multisig/contracts?" + u.Encode()
	r, _ = http.NewRequest("GET", rURL, nil)
	rc = httptest.NewRecorder()
	h.ServeHTTP(rc, r)

	want, err = os.Open("multisigcontractoffset.test.json")
	if err != nil {
		t.Fatal(err)
	}
	bnsapitest.AssertAPIResponseBasic(t, want, rc.Body)
}

func TestEscrowEscrowsIntegrationTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	v, ok := os.LookupEnv("IT_TENDERMINT")
	if !ok {
		panic("set IT_TENDERMINT env var")
	}

	bnscli := client.NewHTTPBnsClient(v)
	h := handlers.EscrowEscrowsHandler{Bns: bnscli}

	u := make(url.Values)
	offset := hex.EncodeToString(bnsapitest.SequenceID(63))
	u.Add("offset", offset)
	_ = u.Encode()
	r, _ := http.NewRequest("GET", "/escrow/escrows?offset=63", nil)
	rc := httptest.NewRecorder()
	h.ServeHTTP(rc, r)

	want, err := os.Open("escrowoffset.test.json")
	if err != nil {
		t.Fatal(err)
	}
	bnsapitest.AssertAPIResponseBasic(t, want, rc.Body)
}

func TestMsgFeeHandlerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	v, ok := os.LookupEnv("IT_TENDERMINT")
	if !ok {
		panic("set IT_TENDERMINT env var")
	}

	bnscli := client.NewHTTPBnsClient(v)
	h := handlers.MsgFeeHandler{Bns: bnscli}

	r, _ := http.NewRequest("GET", "/msgfee/msgfees", nil)
	rc := httptest.NewRecorder()
	h.ServeHTTP(rc, r)

	want, err := os.Open("msgfee.test.json")
	if err != nil {
		t.Fatal(err)
	}
	bnsapitest.AssertAPIResponseBasic(t, want, rc.Body)
}

func TestGovProposalHandlerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	v, ok := os.LookupEnv("IT_TENDERMINT")
	if !ok {
		panic("set IT_TENDERMINT env var")
	}

	bnscli := client.NewHTTPBnsClient(v)
	h := handlers.GovProposalsHandler{Bns: bnscli}

	r, _ := http.NewRequest("GET", "/gov/proposals", nil)
	rc := httptest.NewRecorder()
	h.ServeHTTP(rc, r)

	var want io.Reader
	want, err := os.Open("govproposal.test.json")
	if err != nil {
		t.Fatal(err)
	}
	bnsapitest.AssertAPIResponseBasic(t, want, rc.Body)

	u := make(url.Values)
	author := "C1721181E83376EF978AA4A9A38A5E27C08C7BB2"
	u.Add("author", author)
	r, _ = http.NewRequest("GET", "/gov/proposals?" + u.Encode(), nil)
	rc = httptest.NewRecorder()
	h.ServeHTTP(rc, r)

	want, err = os.Open("govproposal_author.test.json")
	if err != nil {
		t.Fatal(err)
	}

	bnsapitest.AssertAPIResponseBasic(t, want, rc.Body)

	u = make(url.Values)
	offset := "40"
	u.Add("offset", offset)
	r, _ = http.NewRequest("GET", "/gov/proposals?" + u.Encode(), nil)
	rc = httptest.NewRecorder()
	h.ServeHTTP(rc, r)

	want, err = os.Open("govproposal_offset.test.json")
	if err != nil {
		t.Fatal(err)
	}

	bnsapitest.AssertAPIResponseBasic(t, want, rc.Body)
}

func TestGovVotesHandlerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	v, ok := os.LookupEnv("IT_TENDERMINT")
	if !ok {
		panic("set IT_TENDERMINT env var")
	}

	bnscli := client.NewHTTPBnsClient(v)
	h := handlers.GovProposalsHandler{Bns: bnscli}

	r, _ := http.NewRequest("GET", "/gov/votes", nil)
	rc := httptest.NewRecorder()
	h.ServeHTTP(rc, r)

	var want io.Reader
	want, err := os.Open("govvotes.test.json")
	if err != nil {
		t.Fatal(err)
	}
	bnsapitest.AssertAPIResponseBasic(t, want, rc.Body)

	u := make(url.Values)
	elector := "C1721181E83376EF97" +
		"8AA4A9A38A5E27C08C7BB2"
	u.Add("elector", elector)
	r, _ = http.NewRequest("GET", "/gov/votes?" + u.Encode(), nil)
	rc = httptest.NewRecorder()
	h.ServeHTTP(rc, r)

	want, err = os.Open("govvotes_elector.test.json")
	if err != nil {
		t.Fatal(err)
	}
	// TODO these two responses return same responses
	bnsapitest.AssertAPIResponseBasic(t, want, rc.Body)

}
