package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"text/tabwriter"
)

const (
	endpoint = "api.linode.com"
)

type accountInfo struct {
	Action string `json:"ACTION"`
	Data   struct {
		ActiveSince      string  `json:"ACTIVE_SINCE"`
		TransferPool     float64 `json:"TRANSFER_POOL"`
		TransferUsed     float64 `json:"TRANSFER_USED"`
		TransferBillable float64 `json:"TRANSFER_BILLABLE"`
		Managed          bool    `json:"MANAGED"`
		Balance          float64 `json:"BALANCE"`
	} `json:"DATA"`
	Errors []struct {
		ErrorMessage string `json:"ERRORMESSAGE"`
		ErrorCode    int    `json:"ERRORCODE"`
	} `json:"ERRORARRAY"`
}

func main() {
	apiKey := os.Getenv("LINODE_API_TOKEN")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "[ERROR] Env LINODE_API_TOKEN not set")
		os.Exit(1)
	}

	info, err := fetchInfo(apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to fetch account info: %v\n", err)
		os.Exit(1)
	}

	if len(info.Errors) > 0 {
		for _, err := range info.Errors {
			fmt.Fprintf(os.Stderr, "[ERROR] %s (%v)\n", err.ErrorMessage, err.ErrorCode)
			os.Exit(1)
		}
	}

	writeInfoTable(info)
}

func fetchInfo(apiKey string) (*accountInfo, error) {
	fmtStr := "https://%s/?api_key=%s&api_action=account.info"
	url := fmt.Sprintf(fmtStr, endpoint, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	info := &accountInfo{}
	err = json.Unmarshal(body, info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func writeInfoTable(info *accountInfo) {
	w := tabwriter.NewWriter(os.Stdout, 4, 4, 1, ' ', 0)
	fmt.Fprintf(w, "ActiveSince:\t%s\n", info.Data.ActiveSince)
	fmt.Fprintf(w, "Balance:\t%v\n", info.Data.Balance)
	fmt.Fprintf(w, "Managed:\t%v\n", info.Data.Managed)
	fmt.Fprintf(w, "TransferUsed:\t%v/%v GB\n",
		info.Data.TransferUsed,
		info.Data.TransferPool)
	fmt.Fprintf(w, "TransferBillable:\t%v\n", info.Data.TransferBillable)
	w.Flush()
}
