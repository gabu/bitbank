package bitbank

import (
	"fmt"
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

func (t *Tick) ParseTime() time.Time {
	tt := time.Unix(0, int64(time.Millisecond)*t.Timestamp)
	return tt
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
