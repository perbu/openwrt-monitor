package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Router struct {
	baseAddr           string
	session            string
	tx, rx             uint64
	lastUpdate         time.Time
	maxBpsTx, maxBpsRx uint64
	user               string
	pass               string
}

type GenericResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  []any  `json:"result"`
}

type TrafficStats struct {
	CurTx, CurRx uint64
	MaxTx, MaxRx uint64
	Duration     time.Duration
}

func networkTrafficUbus(sessionId string) []byte {
	return []byte(fmt.Sprintf(`[{"jsonrpc":"2.0","id":32,"method":"call","params":["%s","luci-rpc","getNetworkDevices",{}]}]`, sessionId))
}

func (r *Router) getSessionId() string {
	rawReq := []byte(`{"jsonrpc":"2.0","id":1,"method":"call","params":["00000000000000000000000000000000","admb","getSid",{}]}'`)
	buffer := bytes.NewBuffer(rawReq)
	u := urlJoin(r.baseAddr, "/ubus/")
	req, err := http.NewRequest(http.MethodPost, u, buffer)
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(r.user, r.pass)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		panic(fmt.Sprintf("%s", resp.Status))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var foo GenericResponse
	err = json.Unmarshal(body, &foo)
	if err != nil {
		panic(err)
	}
	sidMap := foo.Result[1]
	sid, ok := sidMap.(map[string]interface{})["sid"].(string)
	if !ok {
		panic("can't grok response")
	}
	return sid
}

func New(baseAddr string) *Router {
	r := &Router{
		baseAddr: baseAddr,
	}
	user, ok := os.LookupEnv("ROUTER_USER")
	if !ok {
		panic("can't find ROUTER_USER env var")
	}
	r.user = user
	pass, ok := os.LookupEnv("ROUTER_PASSWORD")
	if !ok {
		panic("can't find ROUTER_PASSWORD env var")
	}
	r.pass = pass
	r.session = r.getSessionId()

	return r
}

func (r *Router) GetTrafficStats() TrafficStats {
	u := urlJoin(r.baseAddr, "/ubus/")
	reqBody := networkTrafficUbus(r.session)
	// fmt.Printf("Request body: %s\n", reqBody)
	buffer := bytes.NewBuffer(reqBody)
	req, err := http.NewRequest(http.MethodPost, u, buffer)
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(r.user, r.pass)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		panic(fmt.Sprintf("%s", resp.Status))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var parsedResp []GenericResponse
	err = json.Unmarshal(body, &parsedResp)
	if err != nil {
		panic(err)
	}
	tx, rx, err := extractTxRx(parsedResp)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("TX: %d, RX: %d\n", tx, rx)
	deltaTx := tx - r.tx
	r.tx = tx
	deltaRx := rx - r.rx
	r.rx = rx
	duration := time.Since(r.lastUpdate)
	//fmt.Printf("Delta time: %.3f\n", time.Since(r.lastUpdate).Seconds())
	//fmt.Printf("deltaTx: %d, deltaRx: %d\n", deltaTx, deltaRx)
	curBpsTx := uint64(float64(deltaTx) / duration.Seconds())
	curBpsRx := uint64(float64(deltaRx) / duration.Seconds())

	if curBpsTx > r.maxBpsTx {
		r.maxBpsTx = curBpsTx
	}
	if curBpsRx > r.maxBpsRx {
		r.maxBpsRx = curBpsRx
	}

	relTx := float64(curBpsTx) / float64(r.maxBpsTx)
	relRx := float64(curBpsRx) / float64(r.maxBpsRx)

	// Clamp these to 1.0 if they're too high (shouldn't happen)
	if relRx > 1.0 {
		relRx = 1.0
	}
	if relTx > 1.0 {
		relTx = 1.0
	}
	r.lastUpdate = time.Now()
	return TrafficStats{
		CurTx:    curBpsTx,
		CurRx:    curBpsRx,
		MaxTx:    r.maxBpsTx,
		MaxRx:    r.maxBpsRx,
		Duration: duration,
	}
}

// NOTE: This code is pretty ugly. It extracts the TX and RX values from the
// ubus response, which is very ugly.
// The proper thing to do would be to write a proper unmarshaller for the response,
// but that is a lot of work and little gain.
func extractTxRx(response []GenericResponse) (uint64, uint64, error) {
	ifaceMap, ok := response[0].Result[1].(map[string]interface{})
	if !ok {
		return 0, 0, errors.New("can't grok response")
	}
	wanIface, ok := ifaceMap["wan"]
	if !ok {
		return 0, 0, errors.New("can't find WAN interface")
	}
	stats, ok := wanIface.(map[string]interface{})["stats"].(map[string]interface{})
	if !ok {
		return 0, 0, errors.New("can't find stats in WAN interface")
	}
	tx := uint64(stats["tx_bytes"].(float64))
	rx := uint64(stats["rx_bytes"].(float64))

	return tx, rx, nil
}

func urlJoin(baseUrl, path string) string {
	if baseUrl[len(baseUrl)-1] != '/' {
		baseUrl += "/"
	}
	if path[0] == '/' {
		path = path[1:]
	}
	return baseUrl + path
}
