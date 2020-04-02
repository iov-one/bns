package main

import (
	"fmt"
	"github.com/iov-one/bns/cmd/bnsapi/client"
	"github.com/iov-one/bns/cmd/bnsapi/docs"
	"github.com/iov-one/bns/cmd/bnsapi/handlers"
	"github.com/iov-one/bns/cmd/bnsapi/util"
	httpSwagger "github.com/swaggo/http-swagger"

	"log"
	"net/http"
	"os"

	"github.com/iov-one/weave/cmd/bnsd/x/account"
	"github.com/iov-one/weave/cmd/bnsd/x/preregistration"
	"github.com/iov-one/weave/cmd/bnsd/x/qualityscore"
	"github.com/iov-one/weave/cmd/bnsd/x/termdeposit"
	"github.com/iov-one/weave/cmd/bnsd/x/username"
	"github.com/iov-one/weave/gconf"
	"github.com/iov-one/weave/migration"
	"github.com/iov-one/weave/x/cash"
	"github.com/iov-one/weave/x/msgfee"
	"github.com/iov-one/weave/x/txfee"
)

type Configuration struct {
	HTTP       string
	Tendermint string
}

// @title BNSAPI documentation
func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LUTC | log.Lshortfile)
	log.SetPrefix(cutstr(util.BuildHash, 6) + " ")

	conf := Configuration{
		HTTP:       env("HTTP", ":8000"),
		Tendermint: env("TENDERMINT", "http://167.172.104.185:31140")}

	if err := run(conf); err != nil {
		log.Fatal(err)
	}
}

func cutstr(s string, maxchar int) string {
	if len(s) <= maxchar {
		return s
	}
	return s[:maxchar]
}

func env(name, fallback string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return fallback
}

func run(conf Configuration) error {
	bnscli := client.NewHTTPBnsClient(conf.Tendermint)

	gconfConfigurations := map[string]func() gconf.Configuration{
		"account":         func() gconf.Configuration { return &account.Configuration{} },
		"cash":            func() gconf.Configuration { return &cash.Configuration{} },
		"migration":       func() gconf.Configuration { return &migration.Configuration{} },
		"msgfee":          func() gconf.Configuration { return &msgfee.Configuration{} },
		"preregistration": func() gconf.Configuration { return &preregistration.Configuration{} },
		"qualityscore":    func() gconf.Configuration { return &qualityscore.Configuration{} },
		"termdeposit":     func() gconf.Configuration { return &termdeposit.Configuration{} },
		"txfee":           func() gconf.Configuration { return &txfee.Configuration{} },
		"username":        func() gconf.Configuration { return &username.Configuration{} },
	}

	rt := http.NewServeMux()
	rt.Handle("/info", &handlers.InfoHandler{})
	rt.Handle("/blocks/", &handlers.BlocksHandler{Bns: bnscli})
	rt.Handle("/account/domains", &handlers.DomainsHandler{Bns: bnscli})
	rt.Handle("/account/accounts", &handlers.AccountsHandler{Bns: bnscli})
	rt.Handle("/account/resolve/", &handlers.DetailHandler{Bns: bnscli})
	rt.Handle("/account/nonce/address/", &handlers.NonceAddressHandler{Bns: bnscli})
	rt.Handle("/account/nonce/pubkey/", &handlers.NoncePubKeyHandler{Bns: bnscli})
	rt.Handle("/username/owner/", &handlers.OwnerHandler{Bns: bnscli})
	rt.Handle("/cash/balances", &handlers.CashBalanceHandler{Bns: bnscli})
	rt.Handle("/termdeposit/contracts", &handlers.ContractsHandler{Bns: bnscli})
	rt.Handle("/termdeposit/deposits", &handlers.DepositsHandler{Bns: bnscli})
	rt.Handle("/multisig/contracts", &handlers.MultisigContractsHandler{Bns: bnscli})
	rt.Handle("/escrow/escrows", &handlers.EscrowEscrowsHandler{Bns: bnscli})
	rt.Handle("/gov/proposals", &handlers.GovProposalsHandler{Bns: bnscli})
	rt.Handle("/gov/votes", &handlers.GovVotesHandler{Bns: bnscli})
	rt.Handle("/gconf/", &handlers.GconfHandler{Bns: bnscli, Confs: gconfConfigurations})
	rt.Handle("/msgfee/msgfees", &handlers.MsgFeeHandler{Bns: bnscli})
	rt.Handle("/tx/submit", &handlers.TxSubmitHandler{Bns: bnscli})
	rt.Handle("/", &handlers.DefaultHandler{})

	docs.SwaggerInfo.Title = "IOV Name Service Rest API"
	docs.SwaggerInfo.Version = util.BuildVersion
	docsUrl := fmt.Sprintf("doc.json")
	rt.Handle("/docs/", httpSwagger.Handler(httpSwagger.URL(docsUrl)))

	if err := http.ListenAndServe(conf.HTTP, rt); err != nil {
		return fmt.Errorf("http server: %s", err)
	}
	return nil
}
