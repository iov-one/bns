package username

import (
	"github.com/iov-one/bns/cmd/bnsapi/client"
	"github.com/iov-one/bns/cmd/bnsapi/handlers"
	"github.com/iov-one/bns/cmd/bnsapi/models"
	"github.com/iov-one/weave/cmd/bnsd/x/username"
	"github.com/iov-one/weave/errors"
	"log"
	"net/http"
)

type OwnerHandler struct {
	Bns client.BnsClient
}

// OwnerHandler godoc
// @Summary Returns the username object with associated info for an owner
// @Tags Starname
// @Param address path string false "Address. example: 04C3DB7CCCACF58EEFCC296FF7AD0F6DB7C2FA17 or iov1qnpaklxv4n6cam7v99hl0tg0dkmu97sh6007un"
// @Success 200 {object} username.Token
// @Failure 404
// @Failure 500
// @Router /username/owner/{address} [get]
func (h *OwnerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rawKey := handlers.LastChunk(r.URL.Path)
	log.Print(r.URL.Path)
	log.Print(rawKey)
	key, err := handlers.WeaveAddressFromQuery(rawKey)
	if err != nil {
		log.Print(err)
		handlers.JSONErr(w, http.StatusBadRequest, "wrong input, must be address")
		return
	}

	var token username.Token
	res := models.KeyModel{
		Model: &token,
	}
	switch err := client.ABCIKeyQuery(r.Context(), h.Bns, "/usernames/owner", key, &res); {
	case err == nil:
		handlers.JSONResp(w, http.StatusOK, res)
	case errors.ErrNotFound.Is(err):
		handlers.JSONErr(w, http.StatusNotFound, "Username not found by owner")
	default:
		log.Printf("account ABCI query: %s", err)
		handlers.JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

type ResolveHandler struct {
	Bns client.BnsClient
}

// ResolveHandler godoc
// @Summary Returns the username object with associated info for an iov username, like thematrix*iov
// @Tags Starname
// @Param username path string false "username. example: thematrix*iov"
// @Success 200 {object} username.Token
// @Failure 404
// @Failure 500
// @Router /username/resolve/{username} [get]
func (h *ResolveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uname := handlers.LastChunk(r.URL.Path)
	if uname != "" {
		var token username.Token
		res := models.KeyModel{
			Model: &token,
		}
		switch err := client.ABCIKeyQuery(r.Context(), h.Bns, "/usernames", []byte(uname), &res); {
		case err == nil:
			handlers.JSONResp(w, http.StatusOK, res)
		case errors.ErrNotFound.Is(err):
			handlers.JSONErr(w, http.StatusNotFound, "Username not found")
		default:
			log.Printf("account ABCI query: %s", err)
			handlers.JSONErr(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	} else {
		handlers.JSONErr(w, http.StatusBadRequest, "Bad username input")
	}
}
