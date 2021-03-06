package handlers

import (
	"encoding/base64"
	"fmt"
	"github.com/iov-one/bns/cmd/bnsapi/client"
	"github.com/iov-one/bns/cmd/bnsapi/util"
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/cmd/bnsd/x/termdeposit"
	"github.com/iov-one/weave/errors"
	"log"
	"net/http"
	"strconv"
)

type ContractsHandler struct {
	Bns client.BnsClient
}

// ContractsHandler godoc
// @Summary Returns a list of bnsd/x/termdeposit entities.
// @Description The term deposit Contract are the contract defining the dates until which one can deposit.
// @Tags IOV token
// @Param offset query int false "Pagination offset"
// @Success 200 {object} handlers.MultipleObjectsResponse
// @Failure 404
// @Failure 500
// @Router /termdeposit/contracts [get]
func (h *ContractsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var offset []byte
	if q.Get("offset")!= "" {
		var err error
		offset, err = ExtractNumericID(q.Get("offset"))
		if err != nil && !errors.ErrEmpty.Is(err) {
			JSONErr(w, http.StatusBadRequest, "offset is in wrong format. send integer")
			return
		}
	}

	it := client.ABCIRangeQuery(r.Context(), h.Bns, "/depositcontracts", fmt.Sprintf("%x:", offset))

	objects := make([]util.KeyValue, 0, util.PaginationMaxItems)
fetchContracts:
	for {
		var c termdeposit.DepositContract
		switch key, err := it.Next(&c); {
		case err == nil:
			objects = append(objects, util.KeyValue{
				Key:   key,
				Value: &c,
			})
			if len(objects) == util.PaginationMaxItems {
				break fetchContracts
			}
		case errors.ErrIteratorDone.Is(err):
			break fetchContracts
		default:
			log.Printf("termdeposit contract ABCI query: %s", err)
			JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
	}

	JSONResp(w, http.StatusOK, MultipleObjectsResponse{
		Objects: objects,
	})
}

type DepositsHandler struct {
	Bns client.BnsClient
}

// DepositsHandler godoc
// @Summary Returns a list of bnsd/x/termdeposit Deposit entities (individual deposits).
// @Description At most one of the query parameters must exist (excluding offset).
// @Description The query may be filtered by Depositor, in which case it returns all the deposits from the Depositor.
// @Description The query may be filtered by Deposit Contract, in which case it returns all the deposits from this Contract.
// @Description The query may be filtered by Contract ID, in which case it returns the deposits from the Deposit Contract with this ID.
// @Tags IOV token
// @Param depositor query string false "Depositor address in bech32 (iov1c9eprq0gxdmwl9u25j568zj7ylqgc7ajyu8wxr) or hex(C1721181E83376EF978AA4A9A38A5E27C08C7BB2)"
// @Param contract query string false "Base64 encoded ID"
// @Param contract_id query int false "Integer encoded Contract ID"
// @Param offset query int false "Pagination offset"
// @Success 200 {object} handlers.MultipleObjectsResponse
// @Failure 404
// @Failure 500
// @Router /termdeposit/deposits [get]
func (h *DepositsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if !AtMostOne(q, "depositor", "contract_id", "contract") {
		JSONErr(w, http.StatusBadRequest, "At most one filter can be used at a time.")
		return
	}

	var offset []byte
	if q.Get("offset")!= "" {
		var err error
		offset, err = ExtractNumericID(q.Get("offset"))
		if err != nil && !errors.ErrEmpty.Is(err) {
			JSONErr(w, http.StatusBadRequest, "offset is in wrong format. send integer")
			return
		}
	}

	var it client.ABCIIterator
	if d := q.Get("depositor"); len(d) > 0 {
		rawAddr, err := weave.ParseAddress(d)
		if err != nil {
			JSONErr(w, http.StatusBadRequest, "Depositor address must be a valid address value..")
			return
		}
		end := NextKeyValue(rawAddr)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/deposits/depositor", fmt.Sprintf("%s:%x:%x", d, offset, end))
	} else if c := q.Get("contract_id"); len(c) > 0 {
		n, err := strconv.ParseInt(c, 10, 64)
		if err != nil {
			JSONErr(w, http.StatusBadGateway, "contract_id must be an integer contract sequence number.")
			return
		}
		cid := EncodeSequence(uint64(n))
		end := NextKeyValue(cid)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/deposits/contract", fmt.Sprintf("%x:%x:%x", cid, offset, end))
	} else if c := q.Get("contract"); len(c) > 0 {
		cid, err := base64.StdEncoding.DecodeString(c)
		if err != nil {
			JSONErr(w, http.StatusBadGateway, "Contract must be a base64 encoded contract key.")
			return
		}
		end := NextKeyValue(cid)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/deposits/contract", fmt.Sprintf("%x:%x:%x", cid, offset, end))
	} else {
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/deposits", fmt.Sprintf("%x:", offset))
	}

	objects := make([]util.KeyValue, 0, util.PaginationMaxItems)
fetchDeposits:
	for {
		var d termdeposit.Deposit
		switch key, err := it.Next(&d); {
		case err == nil:
			objects = append(objects, util.KeyValue{
				Key:   key,
				Value: &d,
			})
			if len(objects) == util.PaginationMaxItems {
				break fetchDeposits
			}
		case errors.ErrIteratorDone.Is(err):
			break fetchDeposits
		default:
			log.Printf("termdeposit deposit ABCI query: %s", err)
			JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
	}

	JSONResp(w, http.StatusOK, MultipleObjectsResponse{
		Objects: objects,
	})
}
