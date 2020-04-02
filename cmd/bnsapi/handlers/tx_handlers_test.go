package handlers

import (
	"encoding/base64"
	"encoding/json"
	"github.com/iov-one/bns/cmd/bnsapi/client"
	bnsd "github.com/iov-one/weave/cmd/bnsd/app"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// Writing test case for tx submit is a lot of work. Easiest way to achieve this generating
//a tx bnscli using `print-out-tx-as-base64` branch
func TestTxSubmitHandlerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	v, ok := os.LookupEnv("IT_TENDERMINT")
	if !ok {
		panic("set IT_TENDERMINT env var")
	}

	bns := client.NewHTTPBnsClient(v)
	h := TxSubmitHandler{Bns: bns}

	strTx := "CgkaBwgBGgNJT1YSaBoiCiBlFOOe6QwUR2VxdWVatZGMJhAW4SANtqIW2xRKwwdusiJCCkAX5tsoKpRZ8sUbFPKF7WsgLfUlRggZEQrbSpTSyjlopMiQvmL1QwqGNwsl3jm3FyKdq9h4KhOb0MieXFIyTq0MmgM9CgIIARIU1cQd3zhuqcKWP+w3kw2/sy+DL/MaFNXEHd84bqnClj/sN5MNv7Mvgy/zIgsQgIjevgEaA0lPVg=="
	var tx bnsd.Tx
	hexTx, err := base64.StdEncoding.DecodeString(strTx)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Unmarshal(hexTx); err != nil {
		t.Fatal(err)
	}

	jsonTx, _ := json.Marshal(tx)
	t.Log(string(jsonTx))

	r, _ := http.NewRequest("POST", "/tx/submit", strings.NewReader(strTx))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("failed response: %d %s", w.Code, w.Body)
	}
	a, _ := ioutil.ReadAll(w.Body)
	log.Print(string(a))
}
