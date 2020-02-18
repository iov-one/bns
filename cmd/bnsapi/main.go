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

	"github.com/iov-one/weave/cmd/bnsd/x/username"
	"github.com/iov-one/weave/gconf"
	"github.com/iov-one/weave/migration"
	"github.com/iov-one/weave/x/cash"
)

type Configuration struct {
	HTTP       string
	Tendermint string
	// Domain is used for swagger docs configuration
	Domain string
}

// @title BNSAPI documentation
func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LUTC | log.Lshortfile)
	log.SetPrefix(cutstr(util.BuildHash, 6) + " ")

	conf := Configuration{
		HTTP:       env("HTTP", ":8000"),
		Tendermint: env("TENDERMINT", "http://localhost:26657"),
		Domain:     env("DOMAIN", "localhost"),
	}

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
		"cash":            func() gconf.Configuration { return &cash.Configuration{} },
		"migration":       func() gconf.Configuration { return &migration.Configuration{} },
		"username":        func() gconf.Configuration { return &username.Configuration{} },
	}

	rt := http.NewServeMux()
	rt.Handle("/info", &handlers.InfoHandler{})
	rt.Handle("/blocks/", &handlers.BlocksHandler{Bns: bnscli})
	rt.Handle("/username/owner/", &handlers.UsernameOwnerHandler{Bns: bnscli})
	rt.Handle("/cash/balances", &handlers.CashBalanceHandler{Bns: bnscli})
	rt.Handle("/multisig/contracts", &handlers.MultisigContractsHandler{Bns: bnscli})
	rt.Handle("/escrow/escrows", &handlers.EscrowEscrowsHandler{Bns: bnscli})
	rt.Handle("/gov/proposals", &handlers.GovProposalsHandler{Bns: bnscli})
	rt.Handle("/gov/votes", &handlers.GovVotesHandler{Bns: bnscli})
	rt.Handle("/gconf/", &handlers.GconfHandler{Bns: bnscli, Confs: gconfConfigurations})
	rt.Handle("/", &handlers.DefaultHandler{Domain: conf.Domain})

	docs.SwaggerInfo.Version = util.BuildVersion
	docs.SwaggerInfo.Host = conf.Domain
	docsUrl := fmt.Sprintf("http://%s%s/docs/doc.json", conf.Domain, conf.HTTP)
	rt.Handle("/docs/", httpSwagger.Handler(httpSwagger.URL(docsUrl)))

	if err := http.ListenAndServe(conf.HTTP, rt); err != nil {
		return fmt.Errorf("http server: %s", err)
	}
	return nil
}
