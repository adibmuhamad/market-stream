package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/piquette/finance-go/equity"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type StockData struct {
	Symbol        string  `json:"symbol"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	PreviousClose float64 `json:"previousClose"`
	Timestamp     int64   `json:"timestamp"`
	Volume        int64   `json:"volume"`
	Value         float64 `json:"value"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
}

type Subscriber struct {
	conn   *websocket.Conn
	symbol string
}

type Publisher struct {
	subscribers map[*Subscriber]bool
	mu          sync.RWMutex
}

func (p *Publisher) NotifySymbols(symbols []string) {
	for _, symbol := range symbols {
		p.Notify(symbol)
	}
}

func NewPublisher() *Publisher {
	return &Publisher{
		subscribers: make(map[*Subscriber]bool),
	}
}

func (p *Publisher) Subscribe(conn *websocket.Conn, symbol string) *Subscriber {
	p.mu.Lock()
	defer p.mu.Unlock()

	subscriber := &Subscriber{conn: conn, symbol: symbol}
	p.subscribers[subscriber] = true
	return subscriber
}

func (p *Publisher) Unsubscribe(subscriber *Subscriber) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.subscribers, subscriber)
}

func (p *Publisher) Notify(symbol string) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for subscriber := range p.subscribers {
		if subscriber.symbol == symbol {
			err := sendStockData(subscriber.conn, symbol)
			if err != nil {
				log.Println("sendStockData:", err)
				p.Unsubscribe(subscriber)
				subscriber.conn.Close()
			}
		}
	}
}

func (p *Publisher) SubscribedSymbols() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	symbolsMap := make(map[string]bool)
	for subscriber := range p.subscribers {
		symbolsMap[subscriber.symbol] = true
	}

	symbols := make([]string, 0, len(symbolsMap))
	for symbol := range symbolsMap {
		symbols = append(symbols, symbol)
	}

	return symbols
}

func fetchStockData(symbol string) (*StockData, error) {
	q, err := equity.Get(symbol)
	if err != nil {
		return nil, err
	}

	stock := q
	change := stock.RegularMarketPrice - stock.RegularMarketPreviousClose
	changePercent := (change / stock.RegularMarketPreviousClose) * 100

	data := &StockData{
		Symbol:        stock.Symbol,
		Open:          stock.RegularMarketOpen,
		High:          stock.RegularMarketDayHigh,
		Low:           stock.RegularMarketDayLow,
		Close:         stock.RegularMarketPrice,
		PreviousClose: stock.RegularMarketPreviousClose,
		Timestamp:     int64(stock.RegularMarketTime),
		Volume:        int64(stock.RegularMarketVolume),
		Value:         float64(stock.MarketCap),
		Change:        change,
		ChangePercent: changePercent,
	}

	return data, nil
}

func sendStockData(c *websocket.Conn, symbol string) error {
	stockData, err := fetchStockData(symbol)
	if err != nil {
		return err
	}

	data, err := json.Marshal(stockData)
	if err != nil {
		return err
	}

	err = c.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return err
	}

	return nil
}

func stockWebSocketHandler(p *Publisher, w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "Missing symbol query parameter", http.StatusBadRequest)
		return
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	subscriber := p.Subscribe(c, symbol)
	defer p.Unsubscribe(subscriber)

	for {
		_, _, err := c.NextReader()
		if err != nil {
			log.Println("read:", err)
			break
		}
	}
}

func main() {
	addr := "localhost:8080"

	publisher := NewPublisher()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			<-ticker.C
			symbols := publisher.SubscribedSymbols()
			publisher.NotifySymbols(symbols)
		}
	}()

	http.HandleFunc("/stock", func(w http.ResponseWriter, r *http.Request) {
		stockWebSocketHandler(publisher, w, r)
	})

	log.Printf("Server listening on %s", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
