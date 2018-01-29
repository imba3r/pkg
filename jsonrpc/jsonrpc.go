package jsonrpc

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"

	"github.com/pkg/errors"
)

// Client represents a JSON RPC client.
type Client struct {
	rpcURL string
	client *http.Client
}

// NewClient returns a new JSON RPC client.
func NewClient(rpcURL string) *Client {
	return &Client{
		rpcURL: rpcURL,
		client: &http.Client{},
	}
}

// Call executes a JSON RPC call.
func (c *Client) Call(method string, args interface{}, result interface{}) error {
	// Encode the request args.
	message, err := encodeRequest(method, args)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.rpcURL, bytes.NewBuffer(message))
	if err != nil {
		return errors.Wrap(err, "could not create request for rpc call")
	}
	req.Header.Set("Content-Type", "application/json")

	// Do the request.
	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error executing rpc request")
	}
	defer resp.Body.Close()

	// Decode the request.
	return decodeResponse(resp.Body, &result)
}

// ----------------------------------------------------------------------------
// Request and Response
// ----------------------------------------------------------------------------

type request struct {
	Method string         `json:"method"`
	Params [1]interface{} `json:"params"`
	ID     uint64         `json:"id"`
}

type response struct {
	Result *json.RawMessage `json:"result"`
	Error  interface{}      `json:"error"`
	ID     uint64           `json:"id"`
}

func encodeRequest(method string, args interface{}) ([]byte, error) {
	c := &request{
		Method: method,
		Params: [1]interface{}{args},
		ID:     uint64(rand.Int63()),
	}
	return json.Marshal(c)
}

func decodeResponse(r io.Reader, reply interface{}) error {
	var c response
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		return errors.Wrap(err, "could not unmarshal json rpc response")
	}
	if c.Error != nil {
		return errors.Errorf("%v", c.Error)
	}
	if c.Result == nil {
		return errors.New("result is nil")
	}
	return json.Unmarshal(*c.Result, reply)
}
