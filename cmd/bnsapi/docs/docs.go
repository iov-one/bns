// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag at
// 2020-02-29 17:02:53.548438 +0300 +03 m=+2.138682271

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/account/accounts": {
            "get": {
                "description": "The list is either the list of all the starname (orkun*neuma) for a given premium starname (*neuma), or the list of all starnames for a given owner address.\nYou need to provide exactly one argument, either the premium starname (*neuma) or the owner address.\n",
                "tags": [
                    "Starname"
                ],
                "summary": "Returns a list of ` + "`" + `bnsd/x/account` + "`" + ` entities (like orkun*neuma).",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Premium Starname ex: *neuma",
                        "name": "starname",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "The owner address format is either in iov address (iov1c9eprq0gxdmwl9u25j568zj7ylqgc7ajyu8wxr) or hex (C1721181E83376EF978AA4A9A38A5E27C08C7BB2)",
                        "name": "ownerAddress",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Pagination offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.MultipleObjectsResponse"
                        }
                    },
                    "404": {},
                    "500": {}
                }
            }
        },
        "/account/domains/": {
            "get": {
                "description": "The list of all premium starnames for a given admin.\nIf no admin address is provided, you get the list of all premium starnames.",
                "tags": [
                    "Starname"
                ],
                "summary": "Returns a list of ` + "`" + `bnsd/x/domain` + "`" + ` entities (like *neuma).",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The admin address may be in the bech32 (iov1c9eprq0gxdmwl9u25j568zj7ylqgc7ajyu8wxr) or hex (C1721181E83376EF978AA4A9A38A5E27C08C7BB2) format.",
                        "name": "adminAddress",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Pagination offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.MultipleObjectsResponse"
                        }
                    },
                    "404": {}
                }
            }
        },
        "/account/resolve/{starname}": {
            "get": {
                "description": "Resolve a given starname (like orkun*neuma) and return all metadata related to this starname,\nlist of crypto-addresses (targets), expiration date and owner address of the starname.",
                "tags": [
                    "Starname"
                ],
                "summary": "Resolve a starname (orkun*neuma) and returns a ` + "`" + `bnsd/x/account` + "`" + ` entity (the associated info).",
                "parameters": [
                    {
                        "type": "string",
                        "description": "starname ex: orkun*neuma",
                        "name": "starname",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/account.Account"
                        }
                    },
                    "404": {},
                    "500": {}
                }
            }
        },
        "/blocks/{blockHeight}": {
            "get": {
                "description": "get block detail by blockHeight",
                "tags": [
                    "Status"
                ],
                "summary": "Get block details by height",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Block Height",
                        "name": "blockHeight",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {},
                    "404": {}
                }
            }
        },
        "/cash/balances": {
            "get": {
                "description": "The iov address may be in the bech32 (iov....) or hex (ON3LK...) format.",
                "tags": [
                    "IOV token"
                ],
                "summary": "returns balance in IOV Token of the given iov address",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bech32 or hex representation of an address",
                        "name": "address",
                        "in": "path"
                    },
                    {
                        "type": "string",
                        "description": "Bech32 or hex representation of an address to be used as offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {},
                    "404": {},
                    "500": {}
                }
            }
        },
        "/escrow/escrows": {
            "get": {
                "description": "At most one of the query parameters must exist(excluding offset)",
                "tags": [
                    "IOV token"
                ],
                "summary": "Returns a list of all the smart contract Escrows.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Iteration offset",
                        "name": "offset",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Source address",
                        "name": "source",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Destination address",
                        "name": "destination",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {},
                    "400": {},
                    "404": {},
                    "500": {}
                }
            }
        },
        "/gconf/{extensionName}": {
            "get": {
                "tags": [
                    "Status"
                ],
                "summary": "Get configuration with extension name",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Extension name",
                        "name": "extensionName",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {},
                    "404": {},
                    "500": {}
                }
            }
        },
        "/gov/proposals": {
            "get": {
                "description": "At most one of the query parameters must exist(excluding offset)",
                "tags": [
                    "Governance"
                ],
                "summary": "Returns a list of x/gov Votes entities.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Author address",
                        "name": "author",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Base64 encoded electorate ID",
                        "name": "electorate",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Base64 encoded Elector ID",
                        "name": "elector",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Elector ID",
                        "name": "electorID",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {},
                    "400": {},
                    "404": {},
                    "500": {}
                }
            }
        },
        "/gov/votes": {
            "get": {
                "description": "At most one of the query parameters must exist(excluding offset)",
                "tags": [
                    "Governance"
                ],
                "summary": "Returns a list of Votes made on the governance.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Base64 encoded Proposal ID",
                        "name": "proposal",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Proposal ID",
                        "name": "proposalID",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Base64 encoded Elector ID",
                        "name": "elector",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Elector ID",
                        "name": "electorID",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {},
                    "400": {},
                    "404": {},
                    "500": {}
                }
            }
        },
        "/info/": {
            "get": {
                "tags": [
                    "Status"
                ],
                "summary": "Returns information about this instance of ` + "`" + `bnsapi` + "`" + `.",
                "responses": {
                    "200": {}
                }
            }
        },
        "/multisig/contracts": {
            "get": {
                "description": "At most one of the query parameters must exist(excluding offset)",
                "tags": [
                    "IOV token"
                ],
                "summary": "Returns a list of all the multisig Contracts.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Iteration offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {},
                    "404": {},
                    "500": {}
                }
            }
        },
        "/termdeposit/contracts": {
            "get": {
                "description": "The term deposit Contract are the contract defining the dates until which one can deposit.",
                "tags": [
                    "IOV token"
                ],
                "summary": "Returns a list of bnsd/x/termdeposit entities.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Pagination offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.MultipleObjectsResponse"
                        }
                    },
                    "404": {},
                    "500": {}
                }
            }
        },
        "/termdeposit/deposits": {
            "get": {
                "description": "At most one of the query parameters must exist (excluding offset).\nThe query may be filtered by Depositor, in which case it returns all the deposits from the Depositor.\nThe query may be filtered by Deposit Contract, in which case it returns all the deposits from this Contract.\nThe query may be filtered by Contract ID, in which case it returns the deposits from the Deposit Contract with this ID.",
                "tags": [
                    "IOV token"
                ],
                "summary": "Returns a list of bnsd/x/termdeposit Deposit entities (individual deposits).",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Depositor address in bech32 (iov1c9eprq0gxdmwl9u25j568zj7ylqgc7ajyu8wxr) or hex(C1721181E83376EF978AA4A9A38A5E27C08C7BB2)",
                        "name": "depositor",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Base64 encoded ID",
                        "name": "contract",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Contract ID as integer",
                        "name": "contract_id",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.MultipleObjectsResponse"
                        }
                    },
                    "404": {},
                    "500": {}
                }
            }
        },
        "/username/owner/{address}": {
            "get": {
                "tags": [
                    "Starname"
                ],
                "summary": "Returns the username object with associated info for an owner",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Address. example: C1721181E83376EF978AA4A9A38A5E27C08C7BB2 or iov1c9eprq0gxdmwl9u25j568zj7ylqgc7ajyu8wxr",
                        "name": "owner",
                        "in": "path"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/username.Token"
                        }
                    },
                    "404": {},
                    "500": {}
                }
            }
        },
        "/username/resolve/{username}": {
            "get": {
                "tags": [
                    "Starname"
                ],
                "summary": "Returns the username object with associated info for an iov username, like thematrix*iov",
                "parameters": [
                    {
                        "type": "string",
                        "description": "username. example: thematrix*iov",
                        "name": "username",
                        "in": "path"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/username.Token"
                        }
                    },
                    "404": {},
                    "500": {}
                }
            }
        }
    },
    "definitions": {
        "account.Account": {
            "type": "object",
            "properties": {
                "certificates": {
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
                            "type": "integer"
                        }
                    }
                },
                "domain": {
                    "description": "Domain references a domain that this account belongs to.",
                    "type": "string"
                },
                "metadata": {
                    "type": "object",
                    "$ref": "#/definitions/weave.Metadata"
                },
                "name": {
                    "type": "string"
                },
                "owner": {
                    "description": "Owner is a weave.Address that controls this account. Can be empty.\n\nAn account can be administrated by the domain admin. In addition,\nownership can be assigned to an address to allow another party to manage\nselected account.",
                    "type": "object",
                    "$ref": "#/definitions/weave.Address"
                },
                "targets": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/account.BlockchainAddress"
                    }
                },
                "valid_until": {
                    "description": "Valid until defines the expiration date for the account. Expired account\ncannot be used or modified. This date is always considered in context of\nthe domain that this account belongs. Expired domain expires all belonging\naccounts as well, event if that account valid until date is not yet due.",
                    "type": "integer"
                }
            }
        },
        "account.BlockchainAddress": {
            "type": "object",
            "properties": {
                "address": {
                    "description": "An address on the specified blockchain network. Address is not a\nweave.Address as we cannot know what is the format of an address on the\nchain that this token instance links to. Because we do not know the rules\nto validate an address for any blockchain ID, this is an arbitrary bulk of\ndata.\nIt is more convenient to always use encoded representation of each address\nand store it as a string. Using bytes while compact is not as comfortable\nto use.",
                    "type": "string"
                },
                "blockchain_id": {
                    "description": "An arbitrary blockchain ID.",
                    "type": "string"
                }
            }
        },
        "handlers.KeyValue": {
            "type": "object",
            "properties": {
                "key": {
                    "type": "object",
                    "$ref": "#/definitions/handlers.hexbytes"
                },
                "value": {
                    "type": "object",
                    "$ref": "#/definitions/orm.Model"
                }
            }
        },
        "handlers.MultipleObjectsResponse": {
            "type": "object",
            "properties": {
                "objects": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/handlers.KeyValue"
                    }
                }
            }
        },
        "handlers.hexbytes": {
            "type": "array",
            "items": {
                "type": "integer"
            }
        },
        "orm.Model": {
            "type": "object"
        },
        "username.BlockchainAddress": {
            "type": "object",
            "properties": {
                "address": {
                    "description": "An address on the specified blockchain network. Address is not a\nweave.Address as we cannot know what is the format of an address on the\nchain that this token instance links to. Because we do not know the rules\nto validate an address for any blockchain ID, this is an arbitrary bulk of\ndata.\nIt is more convinient to always use encoded representation of each address\nand store it as a string. Using bytes while compact is not as comfortable\nto use.",
                    "type": "string"
                },
                "blockchain_id": {
                    "description": "An arbitrary blockchain ID.",
                    "type": "string"
                }
            }
        },
        "username.Token": {
            "type": "object",
            "properties": {
                "metadata": {
                    "type": "object",
                    "$ref": "#/definitions/weave.Metadata"
                },
                "owner": {
                    "description": "Owner is a weave.Address that controls this token. Only the owner can\nmodify a username token.",
                    "type": "object",
                    "$ref": "#/definitions/weave.Address"
                },
                "targets": {
                    "description": "Targets specifies where this username token points to. This must be at\nleast one blockchain address elemenet.",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/username.BlockchainAddress"
                    }
                }
            }
        },
        "weave.Address": {
            "type": "array",
            "items": {
                "type": "integer"
            }
        },
        "weave.Metadata": {
            "type": "object",
            "properties": {
                "schema": {
                    "type": "integer"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "",
	Host:        "",
	BasePath:    "",
	Schemes:     []string{},
	Title:       "BNSAPI documentation",
	Description: "",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
