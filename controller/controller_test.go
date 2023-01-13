package controller

import (
	"net/http"
	"testing"
)

/*
func TestTransferWemix(t *testing.T) {
	transfer.TransferWemix()
}

func TestTransferCtxCoz(t *testing.T) {
	transfer.TransferCtxCoz()
}
*/

func TestSearchTokenSymbolByTokenNameController(t *testing.T) {

	req, err := http.NewRequest("GET", "http://localhost:8080/v1/token/symbol", nil)
	if err != nil {
		t.Errorf("request Error: %s", err)
	}

	q := req.URL.Query()
	q.Add("tokenName", "YKK Token")
	req.URL.RawQuery = q.Encode()

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Errorf("TestSearchTokenSymbolByTokenNameController Error: %s", err)
	}

	t.Log(res.Body)

}
