package client

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iov-one/bns/cmd/bnsapi/models"
	weaveapp "github.com/iov-one/weave/app"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
	"github.com/tendermint/tendermint/rpc/lib/types"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// BnsClient is implemented by any service that provides access to BNS API.
type BnsClient interface {
	Get(ctx context.Context, path string, dest interface{}) error
	Post(ctx context.Context, data []byte, dest interface{}) error
}

// HTTPBnsClient implements BnsClient interface and it is using HTTP transport
// to communicate with BNS instance.
type HTTPBnsClient struct {
	apiURL string
	cli    http.Client
}

// NewHTTPBnsClient returns an instance of a BnsClient that is using HTTP
// transport.
func NewHTTPBnsClient(apiURL string) *HTTPBnsClient {
	return &HTTPBnsClient{
		apiURL: apiURL,
	}
}

func (c *HTTPBnsClient) Get(ctx context.Context, path string, dest interface{}) error {
	req, err := http.NewRequest("GET", c.apiURL+path, nil)
	if err != nil {
		return errors.Wrap(err, "create http request")
	}
	req = req.WithContext(ctx)

	resp, err := c.cli.Do(req)
	if err != nil {
		return errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 1e5))
		return errors.Wrapf(errors.ErrDatabase, "bad response: %d %s", resp.StatusCode, string(b))
	}

	payload := jsonrpcResponse{Result: dest}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1e6)).Decode(&payload); err != nil {
		return errors.Wrap(err, "decode response")
	}
	if payload.Error != nil {
		return payload.Error
	}
	return nil
}

func (c *HTTPBnsClient) Post(ctx context.Context, data []byte, dest interface{}) error {
	req, err := http.NewRequest("POST", c.apiURL, bytes.NewReader(data))
	if err != nil {
		return errors.Wrap(err, "create http request")
	}
	req = req.WithContext(ctx)

	resp, err := c.cli.Do(req)
	if err != nil {
		return errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 1e5))
		return errors.Wrapf(errors.ErrDatabase, "bad response: %d %s", resp.StatusCode, string(b))
	}

	payload := jsonrpcResponse{Result: dest}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1e6)).Decode(&payload); err != nil {
		return errors.Wrap(err, "decode response")
	}
	if payload.Error != nil {
		return payload.Error
	}
	return nil
}

type jsonrpcResponse struct {
	Error  *jsonResponseError
	Result interface{}
}

type jsonResponseError struct {
	Code    int
	Message string
	Data    string
}

type abciQueryParams struct {
	Path string `json:"path"`
	Data string `json:"data"`
}

func (e *jsonResponseError) Error() string {
	if len(e.Data) != 0 {
		return fmt.Sprintf("code %d, %s", e.Code, e.Data)
	}
	return fmt.Sprintf("code %d, %s", e.Code, e.Message)
}

func ABCIKeyQuery(ctx context.Context, c BnsClient, path string, data []byte, destination *models.KeyModel) error {
	params := abciQueryParams{
		Path: path,
		Data: strings.ToUpper(hex.EncodeToString(data)),
	}

	p, err := json.Marshal(params)
	if err != nil {
		return errors.Wrap(err, "param")
	}

	request := rpctypes.NewRPCRequest(rpctypes.JSONRPCIntID(1), "abci_query", p)
	r, err := json.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "response")
	}

	var abciResponse models.AbciQueryResponse
	if err := c.Post(ctx, r, &abciResponse); err != nil {
		return errors.Wrap(err, "response")
	}

	if len(abciResponse.Response.Key) == 0 && len(abciResponse.Response.Value) == 0 {
		return errors.Wrap(errors.ErrNotFound, "empty response")
	}

	var keys weaveapp.ResultSet
	if err := keys.Unmarshal(abciResponse.Response.Key); err != nil {
		return errors.Wrap(err, "cannot unmarshal values")
	}

	var values weaveapp.ResultSet
	if err := values.Unmarshal(abciResponse.Response.Value); err != nil {
		return errors.Wrap(err, "cannot unmarshal values")
	}

	if err := destination.Model.Unmarshal(values.Results[0]); err != nil {
		return errors.Wrap(err, "cannot unmarshal to destination")
	}

	destination.Key = keys.Results[0]

	return nil
}

func ABCIRangeQuery(ctx context.Context, c BnsClient, path string, data string) ABCIIterator {
	params := abciQueryParams{
		Path: path + "?range",
		Data: strings.ToUpper(hex.EncodeToString([]byte(data))),
	}

	p, err := json.Marshal(params)
	if err != nil {
		return &resultIterator{err: errors.Wrap(err, "param")}
	}

	request := rpctypes.NewRPCRequest(rpctypes.JSONRPCIntID(1), "abci_query", p)
	r, err := json.Marshal(request)
	if err != nil {
		return &resultIterator{err: errors.Wrap(err, "response")}
	}

	var abciResponse models.AbciQueryResponse
	if err := c.Post(ctx, r, &abciResponse); err != nil {
		return &resultIterator{err: errors.Wrap(err, "bns client")}
	}

	var values weaveapp.ResultSet
	if err := values.Unmarshal(abciResponse.Response.Value); err != nil {
		return &resultIterator{err: errors.Wrap(err, "unmarshal values response")}
	}
	var keys weaveapp.ResultSet
	if err := keys.Unmarshal(abciResponse.Response.Key); err != nil {
		return &resultIterator{err: errors.Wrap(err, "unmarshal keys response")}
	}

	return &resultIterator{
		keys:   keys.Results,
		values: values.Results,
	}
}

func ABCIPrefixQuery(ctx context.Context, c BnsClient, path string, prefix []byte) ABCIIterator {
	params := abciQueryParams{
		Path: path + "?prefix",
		Data: strings.ToUpper(hex.EncodeToString(prefix)),
	}

	p, err := json.Marshal(params)
	if err != nil {
		return &resultIterator{err: errors.Wrap(err, "param")}
	}

	request := rpctypes.NewRPCRequest(rpctypes.JSONRPCIntID(1), "abci_query", p)
	r, err := json.Marshal(request)
	if err != nil {
		return &resultIterator{err: errors.Wrap(err, "response")}
	}

	var abciResponse models.AbciQueryResponse
	if err := c.Post(ctx, r, &abciResponse); err != nil {
		return &resultIterator{err: errors.Wrap(err, "response")}
	}

	var values weaveapp.ResultSet
	if err := values.Unmarshal(abciResponse.Response.Value); err != nil {
		return &resultIterator{err: errors.Wrap(err, "unmarshal values response")}
	}
	var keys weaveapp.ResultSet
	if err := keys.Unmarshal(abciResponse.Response.Key); err != nil {
		return &resultIterator{err: errors.Wrap(err, "unmarshal keys response")}
	}

	return &resultIterator{
		keys:   keys.Results,
		values: values.Results,
	}
}

func ABCIKeyQueryIter(ctx context.Context, c BnsClient, path string, data []byte) ABCIIterator {
	params := abciQueryParams{
		Path: path,
		Data: strings.ToUpper(hex.EncodeToString(data)),
	}

	p, err := json.Marshal(params)
	if err != nil {
		return &resultIterator{err: errors.Wrap(err, "param")}
	}

	request := rpctypes.NewRPCRequest(rpctypes.JSONRPCIntID(1), "abci_query", p)
	r, err := json.Marshal(request)
	if err != nil {
		return &resultIterator{err: errors.Wrap(err, "response")}
	}

	var abciResponse models.AbciQueryResponse
	if err := c.Post(ctx, r, &abciResponse); err != nil {
		return &resultIterator{err: errors.Wrap(err, "response")}
	}

	if len(abciResponse.Response.Key) == 0 && len(abciResponse.Response.Value) == 0 {
		return &resultIterator{err: errors.Wrap(errors.ErrNotFound, "empty response")}
	}

	var keys weaveapp.ResultSet
	if err := keys.Unmarshal(abciResponse.Response.Key); err != nil {
		return &resultIterator{err: errors.Wrap(errors.ErrNotFound, "cannot unmarshal values")}
	}

	var values weaveapp.ResultSet
	if err := values.Unmarshal(abciResponse.Response.Value); err != nil {
		return &resultIterator{err: errors.Wrap(errors.ErrNotFound, "cannot unmarshal values")}
	}

	return &resultIterator{
		keys:   keys.Results,
		values: values.Results,
	}
}

type ABCIIterator interface {
	Next(orm.Model) ([]byte, error)
}

type resultIterator struct {
	err    error
	keys   [][]byte
	values [][]byte
}

func (it *resultIterator) Next(model orm.Model) ([]byte, error) {
	if it.err != nil {
		return nil, it.err
	}
	if len(it.keys) == 0 {
		return nil, errors.ErrIteratorDone
	}
	val := it.values[0]
	if err := model.Unmarshal(val); err != nil {
		return nil, errors.Wrap(err, "unmarshal model")
	}
	it.values = it.values[1:]
	key := it.keys[0]
	it.keys = it.keys[1:]
	return key, nil
}

func ABCIFullRangeQuery(ctx context.Context, bns BnsClient, path, data string) ABCIIterator {
	return &abciFullIterator{
		ctx:  ctx,
		bns:  bns,
		path: path,
		it:   ABCIRangeQuery(ctx, bns, path, data),
	}
}

type abciFullIterator struct {
	ctx  context.Context
	bns  BnsClient
	path string

	it      ABCIIterator
	lastKey []byte
	done    bool
}

func (fi *abciFullIterator) Next(model orm.Model) ([]byte, error) {
	if fi.done {
		return nil, errors.ErrIteratorDone
	}

	if fi.it != nil {
		switch key, err := fi.it.Next(model); {
		case errors.ErrIteratorDone.Is(err):
			fi.it = nil
		case err == nil:
			fi.lastKey = key
			return key, nil
		default:
			return key, err
		}
	}

	id := fi.lastKey
	for i, c := range fi.lastKey {
		if c == ':' {
			id = fi.lastKey[i+1:]
			break
		}
	}
	fi.it = ABCIRangeQuery(fi.ctx, fi.bns, fi.path, fmt.Sprintf("%x:", id))

	key, err := fi.it.Next(model)
	if err == nil && bytes.Equal(key, fi.lastKey) {
		// Range query filter is inclusive, so ignore entry that was once removed.
		key, err = fi.it.Next(model)
	}

	// If a fresh iterator is instantly done, there are no more
	// results ever and this iterator is done.
	if errors.ErrIteratorDone.Is(err) {
		fi.done = true
		fi.it = nil
		return nil, err
	}
	fi.lastKey = key
	return key, err
}
