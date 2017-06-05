package bitbank

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pubnub/go/messaging"
)

const (
	// Pairs
	BTCJPY = "btc_jpy"
	XRPJPY = "xrp_jpy"
	LTCBTC = "ltc_btc"

	// Channel prefix
	ChanTicker       = "ticker_"
	ChanDepth        = "depth_"
	ChanTransactions = "transactions_"
	ChanCandlestick  = "candlestick_"
)

type PubnubService struct {
	// api client
	client *Client
	// pubnub client
	pubnub *messaging.Pubnub
	// map internal channels to pubnub's
	chanMap        map[string]chan Tick
	subscribes     []subscribeToChannel
	successChannel chan []byte
	errorChannel   chan []byte
}

type subscribeToChannel struct {
	Channel string
	Chan    chan Tick
}

func NewPubnubService(c *Client) *PubnubService {
	return &PubnubService{
		client:         c,
		chanMap:        make(map[string]chan Tick),
		subscribes:     make([]subscribeToChannel, 0),
		successChannel: make(chan []byte),
		errorChannel:   make(chan []byte),
	}
}

func (p *PubnubService) Connect() {
	p.pubnub = messaging.NewPubnub("", "sub-c-e12e9174-dd60-11e6-806b-02ee2ddab7fe", "", "", false, "", nil)
}

func (p *PubnubService) AddSubscribe(channel string, pair string, c chan Tick) {
	s := subscribeToChannel{
		Channel: channel + pair,
		Chan:    c,
	}
	p.subscribes = append(p.subscribes, s)
}

func (p *PubnubService) Subscribe() {
	go p.pubnub.Subscribe(p.SubscribingChannels(), "", p.successChannel, false, p.errorChannel)
	for {
		select {
		case response := <-p.successChannel:
			var msg []interface{}

			err := json.Unmarshal(response, &msg)
			if err != nil {
				panic(err.Error())
			}

			switch t := msg[0].(type) {
			case float64:
				if strings.Contains(msg[1].(string), "connected") {
					for _, s := range p.subscribes {
						if msg[2].(string) == s.Channel {
							p.chanMap[s.Channel] = s.Chan
						}
					}
				}
			case []interface{}:
				message := msg[0].([]interface{})[0].(map[string]interface{})
				d := message["data"].(map[string]interface{})
				tick := Tick{
					Sell:      d["sell"].(string),
					Buy:       d["buy"].(string),
					High:      d["high"].(string),
					Low:       d["low"].(string),
					Last:      d["last"].(string),
					Vol:       d["vol"].(string),
					Timestamp: int64(d["timestamp"].(float64)),
				}
				p.chanMap[msg[2].(string)] <- tick
			default:
				panic(fmt.Sprintf("Unknown type: %T", t))
			}
		case err := <-p.errorChannel:
			fmt.Println(string(err))
			return
		case <-messaging.SubscribeTimeout():
			panic("Subscribe timeout")
		}
	}
}

func (p *PubnubService) Unsubscribe() {
	go p.pubnub.Unsubscribe(p.SubscribingChannels(), p.successChannel, p.errorChannel)
}

func (p *PubnubService) SubscribingChannels() string {
	channels := make([]string, 0)
	for _, s := range p.subscribes {
		channels = append(channels, s.Channel)
	}
	return strings.Join(channels, ",")
}
