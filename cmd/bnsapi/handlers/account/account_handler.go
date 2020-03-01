package account

import (
	"fmt"
	"github.com/iov-one/bns/cmd/bnsapi/client"
	"github.com/iov-one/bns/cmd/bnsapi/handlers"
	"github.com/iov-one/bns/cmd/bnsapi/models"
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/cmd/bnsd/x/account"
	"github.com/iov-one/weave/errors"
	"log"
	"net/http"
)

type DomainsHandler struct {
	Bns client.BnsClient
}

// DomainsHandler godoc
// @Summary Returns a list of `bnsd/x/domain` entities (like *neuma).
// @Description The list of all premium starnames for a given admin.
// @Description If no admin address is provided, you get the list of all premium starnames.
// @Param admin query string false "The admin address may be in the bech32 (iov1c9eprq0gxdmwl9u25j568zj7ylqgc7ajyu8wxr) or hex (C1721181E83376EF978AA4A9A38A5E27C08C7BB2) format."
// @Param offset query string false "Pagination offset"
// @Tags Starname
// @Success 200 {object} handlers.MultipleObjectsResponse
// @Failure 404
// @Redirect 303
// @Router /account/domains/ [get]
func (h *DomainsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var it client.ABCIIterator
	q := r.URL.Query()
	offset := handlers.ExtractIDFromKey(q.Get("offset"))
	if admin := q.Get("admin"); len(admin) > 0 {
		rawAddr, err := weave.ParseAddress(admin)
		if err != nil {
			handlers.JSONErr(w, http.StatusBadRequest, "Admin address must be a valid address value..")
			return
		}
		end := handlers.NextKeyValue(rawAddr)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/domains/admin", fmt.Sprintf("%s:%x:%x", admin, offset, end))
	} else {
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/domains", fmt.Sprintf("%x:", offset))
	}

	objects := make([]handlers.KeyValue, 0, handlers.PaginationMaxItems)
fetchDomains:
	for {
		var model account.Domain
		switch key, err := it.Next(&model); {
		case err == nil:
			objects = append(objects, handlers.KeyValue{
				Key:   key,
				Value: &model,
			})
			if len(objects) == handlers.PaginationMaxItems {
				break fetchDomains
			}
		case errors.ErrIteratorDone.Is(err):
			break fetchDomains
		default:
			log.Printf("account domain ABCI query: %s", err)
			handlers.JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
	}
	handlers.JSONResp(w, http.StatusOK, handlers.MultipleObjectsResponse{
		Objects: objects,
	})
}

type DetailHandler struct {
	Bns client.BnsClient
}

// DetailHandler godoc
// @Summary Resolve a starname (orkun*neuma) and returns a `bnsd/x/account` entity (the associated info).
// @Description Resolve a given starname (like orkun*neuma) and return all metadata related to this starname,
// @Description list of crypto-addresses (targets), expiration date and owner address of the starname.
// @Param starname path string true "starname ex: orkun*neuma"
// @Tags Starname
// @Success 200 {object} account.Account
// @Failure 404
// @Failure 500
// @Router /account/resolve/{starname} [get]
func (h *DetailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accountKey := handlers.LastChunk(r.URL.Path)
	var acc account.Account
	res := models.KeyModel{
		Model: &acc,
	}
	switch err := client.ABCIKeyQuery(r.Context(), h.Bns, "/accounts", []byte(accountKey), &res); {
	case err == nil:
		handlers.JSONResp(w, http.StatusOK, acc)
	case errors.ErrNotFound.Is(err):
		handlers.JSONErr(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
	default:
		log.Printf("account ABCI query: %s", err)
		handlers.JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

type AccountsHandler struct {
	Bns client.BnsClient
}

// AccountsHandler godoc
// @Summary Returns a list of `bnsd/x/account` entities (like orkun*neuma).
// @Description The list is either the list of all the starname (orkun*neuma) for a given premium starname (*neuma), or the list of all starnames for a given owner address.
// @Description You need to provide exactly one argument, either the premium starname (*neuma) or the owner address.
// @Description
// @Tags Starname
// @Param starname query string false "Premium Starname ex: *neuma"
// @Param owner query string false "The owner address format is either in iov address (iov1c9eprq0gxdmwl9u25j568zj7ylqgc7ajyu8wxr) or hex (C1721181E83376EF978AA4A9A38A5E27C08C7BB2)"
// @Param domain query string false "Query by domain"
// @Param offset query string false "Pagination offset"
// @Success 200 {object} handlers.MultipleObjectsResponse
// @Failure 404
// @Failure 500
// @Router /account/accounts [get]
func (h *AccountsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if !handlers.AtMostOne(q, "domain", "owner") {
		handlers.JSONErr(w, http.StatusBadRequest, "At most one filter can be used at a time.")
		return
	}

	var it client.ABCIIterator
	offset := handlers.ExtractIDFromKey(q.Get("offset"))
	if d := q.Get("domain"); len(d) > 0 {
		end := handlers.NextKeyValue([]byte(d))
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/accounts/domain", fmt.Sprintf("%x:%x:%x", d, offset, end))
	} else if o := q.Get("owner"); len(o) > 0 {
		rawAddr, err := weave.ParseAddress(o)
		if err != nil {
			handlers.JSONErr(w, http.StatusBadRequest, "Owner address must be a valid address value..")
			return
		}
		end := handlers.NextKeyValue(rawAddr)
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/accounts/owner", fmt.Sprintf("%s:%x:%x", o, offset, end))
	} else {
		it = client.ABCIRangeQuery(r.Context(), h.Bns, "/accounts", fmt.Sprintf("%x:", offset))
	}

	objects := make([]handlers.KeyValue, 0, handlers.PaginationMaxItems)
fetchAccounts:
	for {
		var acc account.Account
		switch key, err := it.Next(&acc); {
		case err == nil:
			objects = append(objects, handlers.KeyValue{
				Key:   key,
				Value: &acc,
			})
			if len(objects) == handlers.PaginationMaxItems {
				break fetchAccounts
			}
		case errors.ErrIteratorDone.Is(err):
			break fetchAccounts
		default:
			log.Printf("account account ABCI query: %s", err)
			handlers.JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
	}

	handlers.JSONResp(w, http.StatusOK, handlers.MultipleObjectsResponse{
		Objects: objects,
	})
}
