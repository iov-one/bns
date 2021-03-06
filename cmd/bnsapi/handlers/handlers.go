package handlers

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iov-one/bns/cmd/bnsapi/models"
	weavecrypto "github.com/iov-one/weave/crypto"
	"github.com/iov-one/weave/x/msgfee"
	"github.com/iov-one/weave/x/sigs"
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
// @Param offset query int false "Pagination offset"
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
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/proposals/author", fmt.Sprintf("%s:%x:%x", rawAddr, offset, end))
	} else {
		qe := fmt.Sprintf("%x:", offset)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/proposals", qe)

	}

	objects := make([]util.KeyValue, 0, util.PaginationMaxItems)
fetchProposals:
	for {
		var p gov.Proposal
		switch key, err := it.Next(&p); {
		case err == nil:
			objects = append(objects, util.KeyValue{
				Key:   key,
				Value: &p,
			})
			if len(objects) == util.PaginationMaxItems {
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
// @Param offset query int false "Pagination offset"
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

	objects := make([]util.KeyValue, 0, util.PaginationMaxItems)
fetchVotes:
	for {
		var v gov.Vote
		switch key, err := it.Next(&v); {
		case err == nil:
			objects = append(objects, util.KeyValue{
				Key:   key,
				Value: &v,
			})
			if len(objects) == util.PaginationMaxItems {
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
// @Param offset query int false "Pagination offset"
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

	offset, err := ExtractOffsetFromParam(q.Get("offset"))
	if err != nil && !errors.ErrEmpty.Is(err) {
		JSONErr(w, http.StatusBadRequest, "offset is in wrong format. send integer")
		return
	}

	var it client.ABCIIterator
	if d := q.Get("destination"); len(d) > 0 {
		rawAddr, err := WeaveAddressFromQuery(d)
		if err != nil {
			JSONErr(w, http.StatusBadRequest, "Destination address must be a valid address value..")
			return
		}
		end := NextKeyValue(rawAddr)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/escrows/destination", fmt.Sprintf("%s:%x:%x", rawAddr, offset, end))
	} else if s := q.Get("source"); len(s) > 0 {
		rawAddr, err := WeaveAddressFromQuery(s)
		if err != nil {
			JSONErr(w, http.StatusBadRequest, "Source address must be a valid address value..")
			return
		}
		end := NextKeyValue(rawAddr)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/escrows/source", fmt.Sprintf("%s:%x:%x", rawAddr, offset, end))
	} else {
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/escrows", fmt.Sprintf("%x:", offset))
	}

	objects := make([]util.KeyValue, 0, util.PaginationMaxItems)
fetchEscrows:
	for {
		var e escrow.Escrow
		switch key, err := it.Next(&e); {
		case err == nil:
			objects = append(objects, util.KeyValue{
				Key:   key,
				Value: &e,
			})
			if len(objects) == util.PaginationMaxItems {
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
// @Param prefix query string false "Return objects with keys that start with given prefix"
// @Success 200 {object} handlers.MultipleObjectsResponse
// @Failure 404
// @Failure 500
// @Router /multisig/contracts [get]
func (h *MultisigContractsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO make this offset
	prefixQuery := r.URL.Query().Get("prefix")
	var p []byte
	if prefixQuery != "" {
		var err error
		p, err = util.NumericID(prefixQuery)
		if err != nil {
			JSONErr(w, http.StatusBadRequest, "prefix must be numeric")
			return
		}
	}

	// TODO make this range query
	it := client.ABCIPrefixQuery(r.Context(), h.Bns, "/contracts", p)
	objects := make([]util.KeyValue, 0, util.PaginationMaxItems)
fetchContracts:
	for {
		var c multisig.Contract
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
	"/account/accounts?owner=_&domain=_&offset_",
	"/account/domains?admin=_&offset=_",
	"/account/resolve/{starname}",
	"/account/accounts/{accountKey}",
	"/nonce/address/{address}",
	"/nonce/pubkey/{pubKey}",
	"/cash/balances?address=_[OR]offset=_",
	"/msgfee/msgfee?msgfee=_",
	"/username/resolve/{username}",
	"/username/owner/{ownerAddress}",
	"/escrow/escrows?source=_&destination=_&offset=_",
	"/multisig/contracts?prefix=_",
	"/termdeposit/contracts?offset=_",
	"/termdeposit/deposits?depositor=_&contract=_&contract_id=?_offset=_",
	"/gconf/{extensionName}",
	"/blocks/{blockHeight}",
	"/gov/proposals?author=_&electorate=_&electorate_id=_&offset=_",
	"/gov/votes?proposal=_&proposal_id=&elector=_&elector_id=_&offset=_",
}

var withoutParamEndpoint = []string{
	"/info/",
	"/tx/submit",
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
// @Summary returns balance in IOV Token of the given iov address. If not address is not provided returns all wallets
// @Description The iov address may be in the bech32 (iov....) or hex (ON3LK...) format.
// @Tags IOV token
// @Param address query string false "Bech32 or hex representation of an address"
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
		it := client.ABCIPrefixQuery(r.Context(), h.Bns, "/wallets", []byte{})

		objects := make([]util.KeyValue, 0, util.PaginationMaxItems)
	fetchBalances:
		for {
			var set cash.Set
			switch key, err := it.Next(&set); {
			case err == nil:
				objects = append(objects, util.KeyValue{
					Key:   key,
					Value: &set,
				})
				if len(objects) == util.PaginationMaxItems {
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

type NonceAddressHandler struct {
	Bns client.BnsClient
}

// NonceAddressHandler godoc
// @Summary Returns nonce based on an address
// @Description Returns nonce and public key registered for a given address if it was ever used.
// @Param address path string true "Address to query for nonce. ex: iov1qnpaklxv4n6cam7v99hl0tg0dkmu97sh6007un"
// @Tags Nonce
// @Success 200
// @Failure 404
// @Failure 500
// @Router /nonce/address/{address} [get]
func (h *NonceAddressHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	addressStr := LastChunk(r.URL.Path)
	addr, err := WeaveAddressFromQuery(addressStr)
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "provide a weave address")
		return
	}

	var userData sigs.UserData
	res := models.KeyModel{
		Model: &userData,
	}
	switch err := client.ABCIKeyQuery(r.Context(), h.Bns, "/auth", addr, &res); {
	case err == nil:
		JSONResp(w, http.StatusOK, res)
	case errors.ErrNotFound.Is(err):
		JSONErr(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
	default:
		log.Printf("gconf ABCI query: %s", err)
		JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

type NoncePubKeyHandler struct {
	Bns client.BnsClient
}

// NonceAddressHandler godoc
// @Summary Returns nonce based on an address
// @Description Returns nonce and public key registered for a given pubkey if it was ever used.
// @Param pubKey path string true "Public key to query for nonce. ex: 12ee6f581fe55673a1e9e1382a0829e32075a0aa4763c968bc526e1852e78c95"
// @Tags Nonce
// @Success 200
// @Failure 404
// @Failure 500
// @Router /nonce/pubkey/{pubKey} [get]
func (h *NoncePubKeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pubKeyStr := LastChunk(r.URL.Path)
	hexKey, err := hex.DecodeString(pubKeyStr)
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "please provide a hex public key")
		return
	}
	pubKey := weavecrypto.PublicKey_Ed25519{Ed25519: hexKey}
	addr := pubKey.Condition().Address()

	var userData sigs.UserData
	res := models.KeyModel{
		Model: &userData,
	}
	switch err := client.ABCIKeyQuery(r.Context(), h.Bns, "/auth", addr, &res); {
	case err == nil:
		JSONResp(w, http.StatusOK, res)
	case errors.ErrNotFound.Is(err):
		JSONErr(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
	default:
		log.Printf("gconf ABCI query: %s", err)
		JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

type MsgFeeHandler struct {
	Bns client.BnsClient
}

// MsgFeeHandler godoc
// @Summary Return message fee information based on message path: username/register_token
// @Description If msgfee parameter is provided return the queried mesgfee information
// @Description otherwise returns all available msgfees
// @Param msgfee query string false "ex: username/register_token"
// @Tags Message Fee
// @Success 200 {object} msgfee.MsgFee
// @Failure 404
// @Failure 500
// @Router /msgfee/msgfees [get]
func (h *MsgFeeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	msgFee := q.Get("msgfee")
	if msgFee != "" {
		var fee msgfee.MsgFee
		res := models.KeyModel{
			Model: &fee,
		}
		switch err := client.ABCIKeyQuery(r.Context(), h.Bns, "/msgfee", []byte(msgFee), &res); {
		case err == nil:
			JSONResp(w, http.StatusOK, res)
		case errors.ErrNotFound.Is(err):
			JSONErr(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		default:
			log.Printf("gconf ABCI query: %s", err)
			JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	} else {
		it := client.ABCIPrefixQuery(r.Context(), h.Bns, "/msgfee", []byte{})

		objects := make([]util.KeyValue, 0, util.PaginationMaxItems)
	fetchMsgFees:
		for {
			var msgFee msgfee.MsgFee
			switch key, err := it.Next(&msgFee); {
			case err == nil:
				objects = append(objects, util.KeyValue{
					Key:   key,
					Value: &msgFee,
				})
				if len(objects) == util.PaginationMaxItems {
					break fetchMsgFees
				}
			case errors.ErrIteratorDone.Is(err):
				break fetchMsgFees
			default:
				log.Printf("msgfee ABCI query: %s", err)
				JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				return
			}
		}

		JSONResp(w, http.StatusOK, MultipleObjectsResponse{
			Objects: objects,
		})
	}
}
