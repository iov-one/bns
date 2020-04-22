package util

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iov-one/weave/orm"
	"strconv"
)

type KeyValue struct {
	Key   hexbytes  `json:"key"`
	Value orm.Model `json:"value"`
}

// hexbytes is a byte type that JSON serialize to hex encoded string.
type hexbytes []byte

func (b hexbytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(b))
}

func (b *hexbytes) UnmarshalJSON(enc []byte) error {
	var s string
	if err := json.Unmarshal(enc, &s); err != nil {
		return err
	}
	val, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	*b = val
	return nil
}

// paginationMaxItems defines how many items should a single result return.
// This values should not be greater than orm.queryRangeLimit so that each
// query returns enough results.
const PaginationMaxItems = 1000

func NumericID(s string) ([]byte, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot parse number: %s", err)
	}
	encID := make([]byte, 8)
	binary.BigEndian.PutUint64(encID, n)
	return encID, nil
}
