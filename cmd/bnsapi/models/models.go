package models

import (
	"github.com/iov-one/weave"
)

type KeyModel struct {
	Key   []byte           `json:"key"`
	Model weave.Persistent `json:"model"`
}

type AbciQueryResponse struct {
	Response AbciQueryResponseResponse
}

type AbciQueryResponseResponse struct {
	Key   []byte
	Value []byte
}
