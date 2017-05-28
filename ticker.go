package bitbank

import (
	"fmt"
	"math"
	"time"
)

type TickerService struct {
	client *Client
}

type Tick struct {
	Sell      string
	Buy       string
	High      string
	Low       string
	Last      string
	Vol       string
	Timestamp int64
}

func (t *Tick) ParseTime() (*time.Time, error) {
	time := time.Unix(t.Timestamp/1000, int64(math.Mod(float64(t.Timestamp), 1000))*1000000)
	return &time, nil
}

// https://docs.bitbank.cc/#!/Ticker/ticker
func (s *TickerService) Get(pair string) (Tick, error) {
	req, err := s.client.newRequest("GET", pair+"/ticker", nil)

	if err != nil {
		return Tick{}, err
	}

	t := &Tick{}
	v := &BaseResponseJSON{Data: t}
	res, err := s.client.do(req, v)

	if err != nil {
		fmt.Println(res)
		return Tick{}, err
	}

	return *t, nil
}
