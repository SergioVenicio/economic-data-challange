package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type CurrencyData struct {
	Bid float64 `json:"bid,string"`
}

type CurrencyResponse struct {
	Currency CurrencyData `json:"currency"`
}

func init() {
	_, err := os.Stat("./cotacao.txt")
	if err == nil {
		return
	}
	os.Create("./cotacao.txt")
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*300)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	serverData, _ := ioutil.ReadAll(res.Body)
	var currencyData map[string]CurrencyData
	err = json.Unmarshal(serverData, &currencyData)
	if err != nil {
		panic(err)
	}

	for _, c := range currencyData {
		SaveCurrencyPrice(&c)
	}
}

func SaveCurrencyPrice(c *CurrencyData) {
	file, err := os.OpenFile("./cotacao.txt", os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	stringData := fmt.Sprintf("Dólar: {%.2f}", c.Bid)
	file.WriteString(stringData)
}
