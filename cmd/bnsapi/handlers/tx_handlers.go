package handlers

import (
	"encoding/base64"
	"encoding/json"
	"github.com/iov-one/bns/cmd/bnsapi/client"
	rpctypes "github.com/tendermint/tendermint/rpc/lib/types"
	"io/ioutil"
	"log"
	"net/http"
)

type TxSubmitHandler struct {
	Bns client.BnsClient
}

// TxSubmitHandler
// @Summary Submit transaction
// @Description Submit transaction to the blockchain
// @Tags Transaction
// @Accept plain
// @Param tx body string true "base64 encoded transaction"
// @Success 200
// @Failure 404
// @Redirect 303
// @Router /tx/submit [post]
func (h *TxSubmitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tx, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JSONErr(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}
	strTx := string(tx)
	log.Print(strTx)
	_, err = base64.StdEncoding.DecodeString(strTx)
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "send base64 tx")
		return
	}

	params := struct {
		Tx string `json:"tx"`
	}{
		Tx: strTx,
	}

	p, err := json.Marshal(params)
	if err != nil {
		JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	request := rpctypes.NewRPCRequest(rpctypes.JSONRPCIntID(1), "broadcast_tx_sync", p)
	req, err := json.Marshal(request)
	if err != nil {
		JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	// We do not care about payload, proxy all!
	var payload json.RawMessage
	if err := h.Bns.Post(r.Context(), req, &payload); err != nil {
		log.Printf("Tx submit error: %s", err)
		JSONErr(w, http.StatusBadGateway, http.StatusText(http.StatusBadGateway))
		return
	}

	JSONResp(w, http.StatusOK, payload)
}
