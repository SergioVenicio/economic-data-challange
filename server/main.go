package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

const baseUrl string = "https://economia.awesomeapi.com.br/json/last/"

var db *sql.DB

type CurrencyData struct {
	Code       string  `json:"code"`
	Codein     string  `json:"codein"`
	Name       string  `json:"name"`
	High       float64 `json:"high,string"`
	Low        float64 `json:"low,string"`
	VarBid     float64 `json:"varBid,string"`
	PctChange  float64 `json:"pctChange,string"`
	Bid        float64 `json:"bid,string"`
	Ask        float64 `json:"ask,string"`
	Timestamp  string  `json:"timestamp"`
	CreateDate string  `json:"create_date"`
}

type CurrencyResponse struct {
	Currency CurrencyData `json:"currency"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/cotacao", GetCurrencyData)
	r.HandleFunc("/cotacao/{currency-code}", GetCurrencyData)
	log.Fatal(http.ListenAndServe(":8080", r))
}

func GetCurrencyData(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()
	vars := mux.Vars(r)
	currency := vars["currency-code"]
	if currency == "" {
		currency = "USD-BRL"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseUrl+currency, nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	w.WriteHeader(res.StatusCode)
	w.Header().Set("Content-Type", "application/json")
	if res.StatusCode != 200 {
		return
	}

	serverData, _ := ioutil.ReadAll(res.Body)
	var currencyData map[string]CurrencyData
	err = json.Unmarshal(serverData, &currencyData)
	if err != nil {
		panic(err)
	}

	w.Write(serverData)
	for _, c := range currencyData {
		SaveCurrencyData(db, &c)
	}
}

func SaveCurrencyData(db *sql.DB, c *CurrencyData) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	defer cancel()

	insertStr := `
		INSERT INTO
			currencies
			(code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	stmt, err := db.PrepareContext(ctx, insertStr)
	if err != nil {
		log.New(os.Stdout, "ERROR:", log.Ldate|log.Ltime).Panicln(err.Error())
		return
	}
	defer stmt.Close()

	stmt.ExecContext(
		ctx,
		c.Code,
		c.Codein,
		c.Name,
		c.High,
		c.Low,
		c.VarBid,
		c.PctChange,
		c.Bid,
		c.Ask,
		c.Timestamp,
		c.CreateDate,
	)
}

func init() {
	_, err := os.Stat("./currency.db")
	if err != nil {
		createDBFile()
	}
	db, _ = sql.Open("sqlite3", "./currency.db")
	createCurrencyTable(db)
}

func createDBFile() {
	_, err := os.Create("./currency.db")
	if err != nil {
		panic(err)
	}
}

func createCurrencyTable(db *sql.DB) {
	createString := `CREATE TABLE IF NOT EXISTS currencies (
		code VARCHAR(3)
		, codein VARCHAR(3)
		, name VARCHAR(150)
		, high DECIMAL(19, 4)
		, low DECIMAL(19, 4)
		, varBid DECIMAL(5, 4)
		, pctChange DECIMAL(5, 4)
		, bid DECIMAL(19, 4)
		, ask DECIMAL(19, 4)
		, timestamp TIMESTAMP
    	, create_date DATETIME
		, CONSTRAINT PK_CURRENCY PRIMARY KEY (code, codein, timestamp)
	)`

	_, err := db.Exec(createString)
	if err != nil {
		panic(err)
	}
}
