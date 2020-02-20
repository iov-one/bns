package models

import "github.com/iov-one/weave"

type KeyModel struct {
	Key   []byte `json:"key"`
	Model weave.Persistent `json:"model"`
}
