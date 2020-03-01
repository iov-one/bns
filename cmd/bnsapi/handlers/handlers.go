package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/iov-one/bns/cmd/bnsapi/models"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/iov-one/bns/cmd/bnsapi/client"
	"github.com/iov-one/bns/cmd/bnsapi/util"
	"github.com/iov-one/weave/x/cash"

	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/gconf"
	"github.com/iov-one/weave/x/escrow"
	"github.com/iov-one/weave/x/gov"
	"github.com/iov-one/weave/x/multisig"
)

type GovProposalsHandler struct {
	Bns client.BnsClient
}

// GovProposalsHandler godoc
// @Summary Returns a list of x/gov Votes entities.
// @Description At most one of the query parameters must exist(excluding offset)
// @Tags Governance
// @Param author query string false "Author address"
// @Param electorate query string false "Base64 encoded electorate ID"
// @Param elector query string false "Base64 encoded Elector ID"
// @Param electorate_id query int false "Integer Electorate ID"
// @Param offset query string false "Pagination offset"
// @Success 200 {object} handlers.MultipleObjectsResponse
// @Failure 404
// @Failure 400
// @Failure 500
// @Router /gov/proposals [get]
func (h *GovProposalsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if !AtMostOne(q, "author", "electorate", "electorate_id") {
		JSONErr(w, http.StatusBadRequest, "At most one filter can be used at a time.")
		return
	}

	var it client.ABCIIterator
	offset := ExtractIDFromKey(q.Get("offset"))
	if e := q.Get("electorate"); len(e) > 0 {
		rawAddr, err := base64.StdEncoding.DecodeString(e)
		if err != nil {
			JSONErr(w, http.StatusBadRequest, "electorate address must be a base64 encoded value.")
			return
		}
		end := NextKeyValue(rawAddr)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/proposals/electorate", fmt.Sprintf("%x:%x:%x", rawAddr, offset, end))
	} else if e := q.Get("electorate_id"); len(e) > 0 {
		n, err := strconv.ParseInt(e, 10, 64)
		if err != nil {
			JSONErr(w, http.StatusBadGateway, "electorate_id must be an integer contract sequence number.")
			return
		}
		start := EncodeSequence(uint64(n))
		end := NextKeyValue(start)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/proposals/electorate", fmt.Sprintf("%x:%x:%x", start, offset, end))
	} else if s := q.Get("author"); len(s) > 0 {
		rawAddr, err := WeaveAddressFromQuery(s)
		if err != nil {
			JSONErr(w, http.StatusBadRequest, "author address must be a valid address value.")
			return
		}
		end := NextKeyValue(rawAddr)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/proposals/author", fmt.Sprintf("%x:%x:%x", rawAddr, offset, end))
	} else {
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/proposals", fmt.Sprintf("%x:", offset))
	}

	objects := make([]KeyValue, 0, PaginationMaxItems)
fetchProposals:
	for {
		var p gov.Proposal
		switch key, err := it.Next(&p); {
		case err == nil:
			objects = append(objects, KeyValue{
				Key:   key,
				Value: &p,
			})
			if len(objects) == PaginationMaxItems {
				break fetchProposals
			}
		case errors.ErrIteratorDone.Is(err):
			break fetchProposals
		default:
			log.Printf("gov proposals ABCI query: %s", err)
			JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
	}

	JSONResp(w, http.StatusOK, MultipleObjectsResponse{
		Objects: objects,
	})
}

type GovVotesHandler struct {
	Bns client.BnsClient
}

// GovVotesHandler godoc
// @Summary Returns a list of Votes made on the governance.
// @Description At most one of the query parameters must exist(excluding offset)
// @Tags Governance
// @Param proposal query string false "Base64 encoded Proposal ID"
// @Param proposal_id query int false "Integer encoded Proposal ID"
// @Param elector query string false "Base64 encoded Elector ID"
// @Param elector_id query int false "Integer encoded Elector ID"
// @Param offset query string false "Pagination offset"
// @Success 200 {object} handlers.MultipleObjectsResponse
// @Failure 404
// @Failure 400
// @Failure 500
// @Router /gov/votes [get]
func (h *GovVotesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if !AtMostOne(q, "proposal", "proposal_id", "elector", "elector_id") {
		JSONErr(w, http.StatusBadRequest, "At most one filter can be used at a time.")
		return
	}

	var it client.ABCIIterator
	offset := ExtractIDFromKey(q.Get("offset"))
	if e := q.Get("elector"); len(e) > 0 {
		rawAddr, err := WeaveAddressFromQuery(e)
		if err != nil {
			JSONErr(w, http.StatusBadRequest, "elector ID address must be a valid address value..")
			return
		}
		end := NextKeyValue(rawAddr)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/votes/electors", fmt.Sprintf("%x:%x:%x", rawAddr, offset, end))
	} else if e := q.Get("elector_id"); len(e) > 0 {
		// TODO - is elector the same as electorate?
		n, err := strconv.ParseInt(e, 10, 64)
		if err != nil {
			JSONErr(w, http.StatusBadGateway, "elector_id must be an integer contract sequence number.")
			return
		}
		start := EncodeSequence(uint64(n))
		end := NextKeyValue(start)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/votes/electors", fmt.Sprintf("%x:%x:%x", start, offset, end))
	} else if p := q.Get("proposal"); len(p) > 0 {
		rawAddr, err := WeaveAddressFromQuery(p)
		if err != nil {
			JSONErr(w, http.StatusBadRequest, "proposal ID address must be a valid address value..")
			return
		}
		end := NextKeyValue(rawAddr)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/votes/proposals", fmt.Sprintf("%x:%x:%x", rawAddr, offset, end))
	} else if p := q.Get("proposal_id"); len(p) > 0 {
		n, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			JSONErr(w, http.StatusBadGateway, "proposal_id must be an integer contract sequence number.")
			return
		}
		start := EncodeSequence(uint64(n))
		end := NextKeyValue(start)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/votes/proposals", fmt.Sprintf("%x:%x:%x", start, offset, end))
	} else {
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/votes", fmt.Sprintf("%x:", offset))
	}

	objects := make([]KeyValue, 0, PaginationMaxItems)
fetchVotes:
	for {
		var v gov.Vote
		switch key, err := it.Next(&v); {
		case err == nil:
			objects = append(objects, KeyValue{
				Key:   key,
				Value: &v,
			})
			if len(objects) == PaginationMaxItems {
				break fetchVotes
			}
		case errors.ErrIteratorDone.Is(err):
			break fetchVotes
		default:
			log.Printf("gov votes ABCI query: %s", err)
			JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
	}

	JSONResp(w, http.StatusOK, MultipleObjectsResponse{
		Objects: objects,
	})
}

type EscrowEscrowsHandler struct {
	Bns client.BnsClient
}

// EscrowEscrowsHandler godoc
// @Summary Returns a list of all the smart contract Escrows.
// @Description At most one of the query parameters must exist(excluding offset)
// @Tags IOV token
// @Param offset query string false "Iteration offset"
// @Param source query string false "Source address"
// @Param destination query string false "Destination address"
// @Success 200 {object} handlers.MultipleObjectsResponse
// @Failure 404
// @Failure 400
// @Failure 500
// @Router /escrow/escrows [get]
func (h *EscrowEscrowsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if !AtMostOne(q, "source", "destination") {
		JSONErr(w, http.StatusBadRequest, "At most one filter can be used at a time.")
		return
	}

	var it client.ABCIIterator
	offset := ExtractIDFromKey(q.Get("offset"))
	if d := q.Get("destination"); len(d) > 0 {
		rawAddr, err := WeaveAddressFromQuery(d)
		if err != nil {
			JSONErr(w, http.StatusBadRequest, "Destination address must be a valid address value..")
			return
		}
		end := NextKeyValue(rawAddr)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/escrows/destination", fmt.Sprintf("%x:%x:%x", rawAddr, offset, end))
	} else if s := q.Get("source"); len(s) > 0 {
		rawAddr, err := WeaveAddressFromQuery(s)
		if err != nil {
			JSONErr(w, http.StatusBadRequest, "Source address must be a valid address value..")
			return
		}
		end := NextKeyValue(rawAddr)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/escrows/source", fmt.Sprintf("%x:%x:%x", rawAddr, offset, end))
	} else {
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/escrows", fmt.Sprintf("%x:", offset))
	}

	objects := make([]KeyValue, 0, PaginationMaxItems)
fetchEscrows:
	for {
		var e escrow.Escrow
		switch key, err := it.Next(&e); {
		case err == nil:
			objects = append(objects, KeyValue{
				Key:   key,
				Value: &e,
			})
			if len(objects) == PaginationMaxItems {
				break fetchEscrows
			}
		case errors.ErrIteratorDone.Is(err):
			break fetchEscrows
		default:
			log.Printf("escrow ABCI query: %s", err)
			JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
	}

	JSONResp(w, http.StatusOK, MultipleObjectsResponse{
		Objects: objects,
	})
}

type MultisigContractsHandler struct {
	Bns client.BnsClient
}

// MultisigContractsHandler godoc
// @Summary Returns a list of all the multisig Contracts.
// @Description At most one of the query parameters must exist(excluding offset)
// @Tags IOV token
// @Param offset query string false "Pagination offset"
// @Success 200 {object} handlers.MultipleObjectsResponse
// @Failure 404
// @Failure 500
// @Router /multisig/contracts [get]
func (h *MultisigContractsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	offset := ExtractIDFromKey(r.URL.Query().Get("offset"))
	it := client.ABCIRangeQuery(r.Context(), h.Bns, "/contracts", fmt.Sprintf("%x:", offset))

	objects := make([]KeyValue, 0, PaginationMaxItems)
fetchContracts:
	for {
		var c multisig.Contract
		switch key, err := it.Next(&c); {
		case err == nil:
			objects = append(objects, KeyValue{
				Key:   key,
				Value: &c,
			})
			if len(objects) == PaginationMaxItems {
				break fetchContracts
			}
		case errors.ErrIteratorDone.Is(err):
			break fetchContracts
		default:
			log.Printf("multisig contract ABCI query: %s", err)
			JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
	}

	JSONResp(w, http.StatusOK, MultipleObjectsResponse{
		Objects: objects,
	})
}

type GconfHandler struct {
	Bns   client.BnsClient
	Confs map[string]func() gconf.Configuration
}

func (h *GconfHandler) knownConfigurations() []string {
	known := make([]string, 0, len(h.Confs))
	for name := range h.Confs {
		known = append(known, name)
	}
	sort.Strings(known)
	return known
}

// GconfHandler godoc
// @Summary Get configuration with extension name
// @Tags Status
// @Param extensionName path string true "Extension name"
// @Success 200 {object} gconf.Configuration
// @Failure 404
// @Failure 500
// @Router /gconf/{extensionName} [get]
func (h *GconfHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	extensionName := LastChunk(r.URL.Path)
	if extensionName == "" {
		JSONErr(w, http.StatusNotFound,
			fmt.Sprintf("Extension name must be provided. Supported extensions are %q", h.knownConfigurations()))
		return
	}

	var conf gconf.Configuration
	if fn, ok := h.Confs[extensionName]; ok {
		conf = fn()
	} else {
		log.Printf("extension %q gconf configuration entity unknown to gconf handler", extensionName)
		JSONErr(w, http.StatusNotFound,
			fmt.Sprintf("Configuration not registered for browsing. Supported extensions are %q", h.knownConfigurations()))
		return
	}

	res := models.KeyModel{
		Model: conf,
	}
	switch err := client.ABCIKeyQuery(r.Context(), h.Bns, "/gconf", []byte(extensionName), &res); {
	case err == nil:
		JSONResp(w, http.StatusOK, res)
	case errors.ErrNotFound.Is(err):
		JSONErr(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
	default:
		log.Printf("gconf ABCI query: %s", err)
		JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

type InfoHandler struct{}

// InfoHandler godoc
// @Summary Returns information about this instance of `bnsapi`.
// @Tags Status
// @Success 200
// @Router /info/ [get]
func (h *InfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	JSONResp(w, http.StatusOK, struct {
		BuildHash    string `json:"build_hash"`
		BuildVersion string `json:"build_version"`
	}{
		BuildHash:    util.BuildHash,
		BuildVersion: util.BuildVersion,
	})
}

type BlocksHandler struct {
	Bns client.BnsClient
}

// BlocksHandler godoc
// @Summary Get block details by height
// @Description get block detail by blockHeight
// @Tags Status
// @Param blockHeight path int true "Block Height"
// @Success 200
// @Failure 404
// @Redirect 303
// @Router /blocks/{blockHeight} [get]
func (h *BlocksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	heightStr := LastChunk(r.URL.Path)
	if heightStr == "" {
		JSONRedirect(w, http.StatusSeeOther, "/blocks/1")
		return
	}
	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		JSONErr(w, http.StatusNotFound, "block height must be a number")
		return
	}

	// We do not care about payload, proxy all!
	var payload json.RawMessage
	if err := h.Bns.Get(r.Context(), fmt.Sprintf("/block?height=%d", height), &payload); err != nil {
		log.Printf("Bns block height info: %s", err)
		JSONErr(w, http.StatusBadGateway, http.StatusText(http.StatusBadGateway))
		return
	}
	JSONResp(w, http.StatusOK, payload)
}

// DefaultHandler is used to handle the request that no other handler wants.
type DefaultHandler struct{}

var wEndpoint = []string{
	"/account/accounts/?domainKey=_&ownerKey=_",
	"/account/domains/?admin=_&offset=_",
	"/account/accounts/{accountKey}",
	"/cash/balances?address=_[OR]offset=_",
	"/username/resolve/{username}",
	"/username/owner/{ownerAddress}",
	"/escrow/escrows/?source=_&destination=_&offset=_",
	"/multisig/contracts/?offset=_",
	"/termdeposit/contracts/?offset=_",
	"/termdeposit/deposits/?depositor=_&contract=_&contract_id=?_offset=_",
}

var withoutParamEndpoint = []string{
	"/info/",
	"/gov/proposals",
	"/gov/votes",
	"/blocks/{blockHeight}",
	"/gconf/{extensionName}",
}

type endpoints struct {
	WithParam    []string
	WithoutParam []string
}

var availableEndpointsTempl = template.Must(template.New("").Parse(`
<h1>Available endpoints with query parameters:</h1>

{{range .WithParam}}
<a href="{{ .}}">{{ .}}</a></br>
{{end}}

<h1>Available endpoints without parameters:</h1>

{{range .WithoutParam}}
<a href="{{ .}}">{{ .}}</a></br>
{{end}}

<h1>Swagger documentation: </h1>
<a href="/docs">/docs</a></br>
`))

func (h *DefaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// No trailing slash.
	if len(r.URL.Path) > 1 && r.URL.Path[len(r.URL.Path)-1] == '/' {
		path := strings.TrimRight(r.URL.Path, "/")
		JSONRedirect(w, http.StatusPermanentRedirect, path)
		return
	}

	eps := endpoints{
		WithParam:    wEndpoint,
		WithoutParam: withoutParamEndpoint}

	if err := availableEndpointsTempl.Execute(w, eps); err != nil {
		log.Print(err)
		JSONErr(w, http.StatusInternalServerError, "template error")
	}
}

type CashBalanceHandler struct {
	Bns client.BnsClient
}

// CashBalanceHandler godoc
// @Summary returns balance in IOV Token of the given iov address
// @Description The iov address may be in the bech32 (iov....) or hex (ON3LK...) format.
// @Tags IOV token
// @Param address path string false "Bech32 or hex representation of an address"
// @Param offset query string false "Bech32 or hex representation of an address to be used as offset"
// @Success 200
// @Failure 404
// @Failure 500
// @Router /cash/balances [get]
func (h *CashBalanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if !AtMostOne(q, "address", "offset") {
		JSONErr(w, http.StatusBadRequest, "At most one filter can be used at a time.")
		return
	}

	key := q.Get("address")
	if key != "" {
		if strings.HasPrefix(key, "iov") || strings.HasPrefix(key, "tiov") {
			key = "bech32:" + key
		}
		addr, err := WeaveAddressFromQuery(key)

		if err != nil {
			log.Print(err)
			JSONErr(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}
		var set cash.Set

		res := models.KeyModel{
			Model: &set,
		}
		switch err := client.ABCIKeyQuery(r.Context(), h.Bns, "/wallets", addr, &res); {
		case err == nil:
			JSONResp(w, http.StatusOK, set)
		case errors.ErrNotFound.Is(err):
			JSONErr(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		default:
			log.Printf("account ABCI query: %s", err)
			JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	} else {
		// query all wallets
		offset := ExtractIDFromKey(q.Get("offset"))
		it := client.ABCIRangeQuery(r.Context(), h.Bns, "/wallets", fmt.Sprintf("%x:", offset))

		objects := make([]KeyValue, 0, PaginationMaxItems)
	fetchBalances:
		for {
			var set cash.Set
			switch key, err := it.Next(&set); {
			case err == nil:
				objects = append(objects, KeyValue{
					Key:   key,
					Value: &set,
				})
				if len(objects) == PaginationMaxItems {
					break fetchBalances
				}
			case errors.ErrIteratorDone.Is(err):
				break fetchBalances
			default:
				log.Printf("cash balance ABCI query: %s", err)
				JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				return
			}
		}

		JSONResp(w, http.StatusOK, MultipleObjectsResponse{
			Objects: objects,
		})
	}
}
