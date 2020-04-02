package handlers

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/iov-one/bns/cmd/bnsapi/util"
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type MultipleObjectsResponse struct {
	Objects []util.KeyValue `json:"objects"`
}

// AtMostOne returns true if at most one non empty value from given list of
// names exists in the query.
func AtMostOne(query url.Values, names ...string) bool {
	var nonempty int
	for _, name := range names {
		if query.Get(name) != "" {
			nonempty++
		}
		if nonempty > 1 {
			return false
		}
	}
	return true
}

func ExtractRefID(s string) ([]byte, error) {
	tokens := strings.Split(s, "/")

	var version uint32
	switch len(tokens) {
	case 1:
		// Allow providing just the ID value to enable prefix queries.
		// This is a special case.
	case 2:
		if n, err := strconv.ParseUint(tokens[1], 10, 32); err != nil {
			return nil, fmt.Errorf("cannot decode version: %s", err)
		} else {
			version = uint32(n)
		}
	default:
		return nil, errors.ErrInput
	}

	encID := make([]byte, 8)
	if n, err := strconv.ParseUint(tokens[0], 10, 64); err != nil {
		return nil, fmt.Errorf("cannot decode ID: %s", err)
	} else {
		binary.BigEndian.PutUint64(encID, n)
	}

	ref := orm.VersionedIDRef{ID: encID, Version: version}

	if ref.Version == 0 {
		return ref.ID, nil
	}

	return orm.MarshalVersionedID(ref), nil
}

func ExtractAddress(rawAddr string) ([]byte, error) {
	if strings.HasPrefix(rawAddr, "iov") || strings.HasPrefix(rawAddr, "tiov") {
		rawAddr = "bech32:" + rawAddr
	}
	addr, err := weave.ParseAddress(rawAddr)
	return addr, err
}

func ExtractStrID(s string) ([]byte, error) {
	return []byte(s), nil
}

func ExtractNumericID(s string) ([]byte, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot parse number: %s", err)
	}
	encID := make([]byte, 8)
	binary.BigEndian.PutUint64(encID, n)
	return encID, nil
}

func RefKey(raw []byte) (string, error) {
	// Skip the prefix, being the characters before : (including separator)
	val := raw[bytes.Index(raw, []byte(":"))+1:]

	ref, err := orm.UnmarshalVersionedID(val)
	if err != nil {
		return "", fmt.Errorf("cannot unmarshal versioned key: %s", err)
	}

	id := binary.BigEndian.Uint64(ref.ID)
	return fmt.Sprintf("%d/%d", id, ref.Version), nil
}

func SequenceKey(raw []byte) (string, error) {
	// Skip the prefix, being the characters before : (including separator)
	seq := raw[bytes.Index(raw, []byte(":"))+1:]
	if len(seq) != 8 {
		return "", fmt.Errorf("invalid sequence length: %d", len(seq))
	}
	n := binary.BigEndian.Uint64(seq)
	return fmt.Sprint(int64(n)), nil
}

// Offset is sent as int and converted to binary
func ExtractOffsetFromParam(param string) ([]byte, error) {
	offset := make([]byte, 8)
	if len(param) > 0 {
		n, err := strconv.ParseUint(param, 10, 64)
		if err != nil {
			return nil, err
		}
		binary.BigEndian.PutUint64(offset, n)
		return offset, nil
	}
	return nil, errors.Wrap(errors.ErrEmpty, "empty offset")
}

// JSONResp write content as JSON encoded response.
func JSONResp(w http.ResponseWriter, code int, content interface{}) {
	b, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		log.Printf("cannot JSON serialize response: %s", err)
		code = http.StatusInternalServerError
		b = []byte(`{"errors":["Internal Server Errror"]}`)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)

	const MB = 1 << (10 * 2)
	if len(b) > MB {
		log.Printf("response JSON body is huge: %d", len(b))
	}
	_, _ = w.Write(b)
}

// JSONErr write single error as JSON encoded response.
func JSONErr(w http.ResponseWriter, code int, errText string) {
	JSONErrs(w, code, []string{errText})
}

// JSONErrs write multiple errors as JSON encoded response.
func JSONErrs(w http.ResponseWriter, code int, errs []string) {
	resp := struct {
		Errors []string `json:"errors"`
	}{
		Errors: errs,
	}
	JSONResp(w, code, resp)
}

// JSONRedirect return redirect response, but with JSON formatted body.
func JSONRedirect(w http.ResponseWriter, code int, urlStr string) {
	w.Header().Set("Location", urlStr)
	var content = struct {
		Code     int
		Location string
	}{
		Code:     code,
		Location: urlStr,
	}
	JSONResp(w, code, content)
}

func NextKeyValue(b []byte) []byte {
	if len(b) == 0 {
		return nil
	}
	next := make([]byte, len(b))
	copy(next, b)

	// If the last value does not overflow, increment it. Otherwise this is
	// an edge case and key must be extended.
	if next[len(next)-1] < math.MaxUint8 {
		next[len(next)-1]++
	} else {
		next = append(next, 0)
	}
	return next
}

func WeaveAddressFromQuery(rawAddr string) (weave.Address, error) {
	if strings.HasPrefix(rawAddr, "iov") || strings.HasPrefix(rawAddr, "tiov") {
		rawAddr = "bech32:" + rawAddr
	}
	addr, err := weave.ParseAddress(rawAddr)
	return addr, err
}

func EncodeSequence(val uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, val)
	return bz
}

// LastChunk returns last path chunk - everything after the last `/` character.
// For example LAST in /foo/bar/LAST and empty string in /foo/bar/
func LastChunk(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
