// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag at
// 2020-02-17 13:17:18.823642 +0300 +03 m=+1.349435170

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
                "description": "At most one of the query parameters must exist(excluding offset)",
                "summary": "Returns a list of ` + "`" + `bnsd/x/account` + "`" + ` Account entitiy.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Domain name",
                        "name": "domainKey",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Admin address",
                        "name": "ownerKey",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    }
                }
            }
        },
        "/account/accounts/{accountKey}": {
            "get": {
                "summary": "Returns a list of ` + "`" + `bnsd/x/account` + "`" + ` Account entitiy.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Address of the admin",
                        "name": "accountKey",
                        "in": "path"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    }
                }
            }
        },
        "/account/domains/": {
            "get": {
                "summary": "Returns a list of ` + "`" + `bnsd/x/account` + "`" + ` Domain entities.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Address of the admin",
                        "name": "admin",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Iteration offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    }
                }
            }
        },
        "/blocks/{blockHeight}": {
            "get": {
                "description": "get block detail by blockHeight",
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
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    }
                }
            }
        },
        "/cash/balances": {
            "get": {
                "summary": "Returns a ` + "`" + `bnsd/x/cash.Set` + "`" + ` entitiy.",
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
                        "in": "path"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    }
                }
            }
        },
        "/escrow/escrows": {
            "get": {
                "description": "At most one of the query parameters must exist(excluding offset)",
                "summary": "Returns a list of x/escrow Escrow entities.",
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
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "500": {}
                }
            }
        },
        "/gconf/{extensionName}": {
            "get": {
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
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "500": {}
                }
            }
        },
        "/info/": {
            "get": {
                "summary": "Returns information about this instance of ` + "`" + `bnsapi` + "`" + `.",
                "responses": {
                    "200": {}
                }
            }
        },
        "/multisig/contracts": {
            "get": {
                "description": "At most one of the query parameters must exist(excluding offset)",
                "summary": "Returns a list of multisig Contract entities.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Iteration offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "500": {}
                }
            }
        },
        "/termdeposit/contracts": {
            "get": {
                "description": "At most one of the query parameters must exist(excluding offset)",
                "summary": "Returns a list of bnsd/x/termdeposit Contract entities.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Iteration offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "500": {}
                }
            }
        },
        "/termdeposit/deposits": {
            "get": {
                "description": "At most one of the query parameters must exist(excluding offset)",
                "summary": "Returns a list of bnsd/x/termdeposit Deposit entities.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Depositor address",
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
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/json.RawMessage"
                        }
                    },
                    "500": {}
                }
            }
        }
    },
    "definitions": {
        "json.RawMessage": {
            "type": "array",
            "items": {
                "type": "integer"
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
