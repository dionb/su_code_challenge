package client

// This was added to make testing cleaner. In hindsight, I should have comitted to the approach
// and passed *testing.T into the methods for better handling of response codes. In a production
// client I would use a set of structured errors in the returned values and check for those in tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"sumup_challenge/api"

	"github.com/google/uuid"
)

var StatusErr = errors.New("unexpected status code")

type ApiClient struct {
	path   *url.URL
	client *http.Client
}

func New(path string) (ApiClient, error) {
	uri, err := url.Parse(path)
	if err != nil {
		return ApiClient{}, err
	}

	return ApiClient{
		uri,
		http.DefaultClient,
	}, nil
}

const CONTENT_TYPE = "JSON"

func (c *ApiClient) CreateAccount(name string, balance uint) (api.AccountNumber, error) {
	const endpoint = "/account"

	body, err := json.Marshal(api.CreateAccountRequest{name, balance})
	if err != nil {
		return uuid.Nil, err
	}

	resp, err := c.client.Post(c.path.JoinPath(endpoint).String(), CONTENT_TYPE, bytes.NewBuffer(body))
	if err != nil {
		return uuid.Nil, err
	}

	if resp.StatusCode != 200 {
		return uuid.Nil, StatusErr
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return uuid.Nil, err
	}

	var car api.CreateAccountResponse
	err = json.Unmarshal(body, &car)
	return car.AccountNumber, err
}

func (c *ApiClient) CreateDeposit(name api.AccountNumber, amount uint32) (uint64, error) {
	const endpoint = "/deposit"
	body, err := json.Marshal(api.DepositRequest{name, amount})
	if err != nil {
		return 0, err
	}
	resp, err := c.client.Post(c.path.JoinPath(endpoint).String(), CONTENT_TYPE, bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != 200 {
		log.Println(resp.StatusCode)
		return 0, StatusErr
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var dr api.DepositResponse
	err = json.Unmarshal(body, &dr)
	return dr.NewBalance, err
}

func (c *ApiClient) RequestTransfer(src, dst api.AccountNumber, amount uint32) error {
	const endpoint = "/transfer"
	body, err := json.Marshal(api.TransferRequest{src, dst, amount})
	if err != nil {
		return err
	}
	resp, err := c.client.Post(c.path.JoinPath(endpoint).String(), CONTENT_TYPE, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return StatusErr
	}

	return nil
}
