package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type OrderBook struct {
	Success bool `json:"success"`
	Result  struct {
		Asks [][]float64 `json:"asks"`
		Bids [][]float64 `json:"bids"`
	} `json:"result"`
}

func Kappa(arg1 int, arg2 string) (float64, float64, float64, float64, float64) {

	// arg1 = order book depth, arg2 = ticker symbol

	depth := strconv.Itoa(arg1)

	res, err := http.Get("https://ftx.com/api/markets/" + arg2 + "/orderbook?depth=" + depth)

	if err != nil {

		fmt.Println(err)

		fmt.Println("Order Book Data Cannot Be Retrived")
		fmt.Println("Please Wait Until Order Book Data Can Be Retrived")
		time.Sleep(time.Second * 15)

		Kappa(arg1, arg2)

	}

	var response OrderBook
	json.NewDecoder(res.Body).Decode(&response)

	var bid_price []float64
	var bid_size []float64
	var ask_price []float64
	var ask_size []float64

	if len(response.Result.Bids) != arg1 || len(response.Result.Asks) != arg1 {

		fmt.Println("Order Book Data Is Incomplete")
		fmt.Println("Please Wait Until Order Book Data Is Full")
		time.Sleep(time.Second * 15)
		Kappa(arg1, arg2)

	}

	i := 0

	for i = 0; i < arg1; i++ {

		bid_price = append(bid_price, response.Result.Bids[i][0])
		bid_size = append(bid_size, response.Result.Bids[i][1])
		ask_price = append(ask_price, response.Result.Asks[i][0])
		ask_size = append(ask_size, response.Result.Asks[i][1])

	}

	var kappa float64
	var total_bid_size float64
	var total_ask_size float64

	var weighted_midpoint float64

	j := 0

	for j = 0; j < arg1; j++ {

		kappa = kappa + (bid_price[j] * bid_size[j]) + (ask_price[j] * ask_size[j])

		total_bid_size = total_bid_size + bid_size[j]
		total_ask_size = total_ask_size + ask_size[j]

	}

	midpoint := (bid_price[0] + ask_price[0]) / 2

	imbalance := total_bid_size / (total_bid_size + total_ask_size)

	weighted_midpoint = (imbalance * ask_price[0]) + ((1 - imbalance) * bid_price[0])

	return midpoint, weighted_midpoint, kappa, bid_price[0], ask_price[0]

}

type HistoricalPrices struct {
	Success bool `json:"success"`
	Result  []struct {
		Close     float64   `json:"close"`
		High      float64   `json:"high"`
		Low       float64   `json:"low"`
		Open      float64   `json:"open"`
		Starttime time.Time `json:"startTime"`
		Volume    float64   `json:"volume"`
	} `json:"result"`
}

func (client *FtxClient) GetHistoricalPrices(market string, resolution int64,
	limit int64, startTime int64, endTime int64) (HistoricalPrices, error) {

	var historicalPrices HistoricalPrices

	resp, err := client._get(
		"markets/"+market+
			"/candles?resolution="+strconv.FormatInt(resolution, 10)+
			"&limit="+strconv.FormatInt(limit, 10)+
			"&start_time="+strconv.FormatInt(startTime, 10)+
			"&end_time="+strconv.FormatInt(endTime, 10),
		[]byte(""))

	if err != nil {

		fmt.Println("Error GetHistoricalPrices", err)

		return historicalPrices, err
	}

	err = _processResponse(resp, &historicalPrices)
	return historicalPrices, err

}

func Sigma(arg1 string, arg2 string, arg3 string, arg4 string, arg5 string, arg6 float64) float64 {

	// arg1 = ticker symbol, arg2 = volatility interval

	client := New(arg3, arg4, arg5)

	interval, err := strconv.ParseInt(arg2, 0, 64)

	if err != nil {

		fmt.Println(err)

	}

	t := time.Now()

	past_time := t.Add(time.Duration(-interval) * time.Second).Unix()

	current_time := t.Unix()

	// market, candle interval, start time, end time

	candles, _ := client.GetHistoricalPrices(arg1, interval, 7, past_time, current_time)

	var price_data []float64

	if len(candles.Result) == 0 {

		if arg6 == 0 {

			fmt.Println("Vol Cannot Be Measured...Shutting Down")
			os.Exit(0)

		}

		return arg6

	}

	// fmt.Println(len(candles.Result))

	price_data = append(price_data, candles.Result[0].Open)
	price_data = append(price_data, candles.Result[0].High)
	price_data = append(price_data, candles.Result[0].Low)
	price_data = append(price_data, candles.Result[0].Close)

	// if arg2 == "15" {

	// 	price_data = append(price_data, candles.Result[0].Open)
	// 	price_data = append(price_data, candles.Result[0].High)
	// 	price_data = append(price_data, candles.Result[0].Low)
	// 	price_data = append(price_data, candles.Result[0].Close)

	// }

	// if arg2 == "60" || arg2 == "300" {

	// index := 0

	// for index = 0; index < len(candles.Result); index++ {

	// 	fmt.Println(candles.Result[index].Close)

	// }

	// }

	i := 0

	sum := 0.0

	for i = 0; i < len(price_data); i++ {

		sum = sum + price_data[i]

	}

	mean_price := sum / float64(len(price_data))

	var average_distance float64

	j := 0

	for j = 0; j < len(price_data); j++ {

		average_distance = average_distance + math.Pow((price_data[j]-mean_price), 2)

	}

	variance := average_distance / (float64(len(price_data)) - 1)

	// fmt.Printf("%f", variance)

	sigma := math.Sqrt(variance)

	// fmt.Printf("%f", sigma)

	if sigma == 0 {

		return arg6

	}

	// fmt.Printf("Returning calculated sigma")

	return sigma

}

func Reservation_Price(arg1 float64, arg2 float64, arg3 float64, arg4 float64, arg5 float64, arg6 float64) (float64, float64) {

	// arg1 = mid price, arg2 = target distance, arg3 = risk aversion parameter, arg4 = volatility, arg5 = price aggressor, arg6 = position size

	reservation_price := arg1 - (arg2 * arg3 * math.Pow(arg4, 2))
	aggressive_reservation_price := reservation_price - (arg2 / arg6 * arg5)

	return reservation_price, aggressive_reservation_price

}

func Optimal_Spread(arg1 float64, arg2 float64, arg3 float64) float64 {

	//arg1 = risk aversion parameter, arg2 = volatility, arg3 = kappa

	optimal_spread := (arg1 * math.Pow(arg2, 2)) + ((2 / arg1) * (math.Log(1 + (arg1 / arg3))))

	return optimal_spread
}

type FtxClient struct {
	Client     *http.Client
	Api        string
	Secret     []byte
	Subaccount string
}

func New(api string, secret string, subaccount string) *FtxClient {

	return &FtxClient{Client: &http.Client{}, Api: api, Secret: []byte(secret), Subaccount: url.PathEscape(subaccount)}

}

func _processResponse(resp *http.Response, result interface{}) error {

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {

		fmt.Print("Error processing response:", err)
		return err

	}

	err = json.Unmarshal(body, result)

	if err != nil {

		fmt.Println("Error processing response:", err)
		return err

	}

	return nil

}

func (client *FtxClient) sign(signaturePayload string) string {

	mac := hmac.New(sha256.New, client.Secret)
	mac.Write([]byte(signaturePayload))
	return hex.EncodeToString(mac.Sum(nil))

}

const URL = "https://ftx.com/api/"

func (client *FtxClient) signRequest(method string, path string, body []byte) *http.Request {

	ts := strconv.FormatInt(time.Now().UTC().Unix()*1000, 10)
	signaturePayload := ts + method + "/api/" + path + string(body)
	signature := client.sign(signaturePayload)
	req, _ := http.NewRequest(method, URL+path, bytes.NewBuffer(body))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("FTX-KEY", client.Api)
	req.Header.Set("FTX-SIGN", signature)
	req.Header.Set("FTX-TS", ts)

	if client.Subaccount != "" {

		req.Header.Set("FTX-SUBACCOUNT", client.Subaccount)

	}

	return req

}

func (client *FtxClient) _post(path string, body []byte) (*http.Response, error) {

	preparedRequest := client.signRequest("POST", path, body)
	resp, err := client.Client.Do(preparedRequest)
	return resp, err

}

type Order struct {
	CreatedAt     time.Time `json:"createdAt"`
	FilledSize    float64   `json:"filledSize"`
	Future        string    `json:"future"`
	ID            int64     `json:"id"`
	Market        string    `json:"market"`
	Price         float64   `json:"price"`
	AvgFillPrice  float64   `json:"avgFillPrice"`
	RemainingSize float64   `json:"remainingSize"`
	Side          string    `json:"side"`
	Size          float64   `json:"size"`
	Status        string    `json:"status"`
	Type          string    `json:"type"`
	ReduceOnly    bool      `json:"reduceOnly"`
	Ioc           bool      `json:"ioc"`
	PostOnly      bool      `json:"postOnly"`
	ClientID      string    `json:"clientId"`
}

type NewOrderResponse struct {
	Success bool  `json:"success"`
	Result  Order `json:"result"`
}

type NewOrder struct {
	Market                  string  `json:"market"`
	Side                    string  `json:"side"`
	Price                   float64 `json:"price"`
	Type                    string  `json:"type"`
	Size                    float64 `json:"size"`
	ReduceOnly              bool    `json:"reduceOnly"`
	Ioc                     bool    `json:"ioc"`
	PostOnly                bool    `json:"postOnly"`
	ExternalReferralProgram string  `json:"externalReferralProgram"`
}

func (client *FtxClient) PlaceOrder(market string, side string, price float64, _type string, size float64, reduceOnly bool, ioc bool, postOnly bool) (NewOrderResponse, error) {

	var newOrderResponse NewOrderResponse
	requestBody, err := json.Marshal(NewOrder{
		Market:     market,
		Side:       side,
		Price:      price,
		Type:       _type,
		Size:       size,
		ReduceOnly: reduceOnly,
		Ioc:        ioc,
		PostOnly:   postOnly})

	if err != nil {

		fmt.Println("Error PlaceOrder", err)
		return newOrderResponse, err

	}

	resp, err := client._post("orders", requestBody)

	if err != nil {

		fmt.Println("Error PlaceOrder", err)
		return newOrderResponse, err

	}

	err = _processResponse(resp, &newOrderResponse)
	return newOrderResponse, err

}

type OpenOrders struct {
	Success bool    `json:"success"`
	Result  []Order `json:"result"`
}

func (client *FtxClient) GetOpenOrders(market string) (OpenOrders, error) {

	var openOrders OpenOrders
	resp, err := client._get("orders?market="+market, []byte(""))

	if err != nil {

		fmt.Println("Error GetOpenOrders", err)
		return openOrders, err

	}

	err = _processResponse(resp, &openOrders)
	return openOrders, err

}

func Place_Order(arg1 string, arg2 string, arg3 string, arg4 float64, arg5 float64, arg6 float64, arg7 int64, arg8 bool, arg9 string, arg10 float64, arg11 float64, arg12 float64, arg13 float64) {

	// arg1 = api key, arg2 = api secret, arg3 = subaccount, arg4 = trade amount, arg5 = reservation price, arg6 = optimal spread,
	// arg7 = order time, arg8 = post only, arg9 = ticker symbol, arg10 = best bid, arg11 = best ask, arg12 = inventory cutoff, arg13 = target distance

	client := New(arg1, arg2, arg3)

	bid_price := arg5 - arg6
	ask_price := arg5 + arg6

	if arg8 {

		if bid_price > arg10 {

			bid_price = arg10

		}

		if ask_price < arg11 {

			ask_price = arg11

		}

	}

	if arg13 < -arg12 {

		// market, side, price, type, size, reduce only, IOC, post only

		bid_order, _ := client.PlaceOrder(arg9, "buy", bid_price, "limit", arg4, false, false, arg8)
		fmt.Println("Bid Offer: ", bid_order.Result.Price)

		time.Sleep(time.Second * time.Duration(arg7))

		openOrders, _ := client.GetOpenOrders(arg9)

		if len(openOrders.Result) == 0 {

			fmt.Println("Bid Order Has Been Filled")

		} else if len(openOrders.Result) > 4 {

			i := 0

			for i = 0; i < len(openOrders.Result); i++ {

				cancel, _ := client.CancelOrder(openOrders.Result[i].ID)
				fmt.Println("Order Side Cancelled: ", openOrders.Result[i].Side)
				fmt.Println(cancel)

			}

			fmt.Println("Too Many Cancellations Have Failed")
			os.Exit(0)

		} else {

			i := 0

			for i = 0; i < len(openOrders.Result); i++ {

				cancel, _ := client.CancelOrder(openOrders.Result[i].ID)
				fmt.Println("Order Side Cancelled: ", openOrders.Result[i].Side)
				fmt.Println(cancel)

			}

		}

	} else if arg13 > arg12 {

		// market, side, price, type, size, reduce only, IOC, post only

		ask_order, _ := client.PlaceOrder(arg9, "sell", ask_price, "limit", arg4, false, false, arg8)
		fmt.Println("Ask Offer: ", ask_order.Result.Price)

		time.Sleep(time.Second * time.Duration(arg7))

		openOrders, _ := client.GetOpenOrders(arg9)

		if len(openOrders.Result) == 0 {

			fmt.Println("Ask Order Has Been Filled")

		} else if len(openOrders.Result) > 4 {

			i := 0

			for i = 0; i < len(openOrders.Result); i++ {

				cancel, _ := client.CancelOrder(openOrders.Result[i].ID)
				fmt.Println("Order Side Cancelled: ", openOrders.Result[i].Side)
				fmt.Println(cancel)

			}

			fmt.Println("Too Many Cancellations Have Failed")
			os.Exit(0)

		} else {

			i := 0

			for i = 0; i < len(openOrders.Result); i++ {

				cancel, _ := client.CancelOrder(openOrders.Result[i].ID)
				fmt.Println("Order Side Cancelled: ", openOrders.Result[i].Side)
				fmt.Println(cancel)

			}

		}

	} else {

		// market, side, price, type, size, reduce only, IOC, post only

		bid_order, _ := client.PlaceOrder(arg9, "buy", bid_price, "limit", arg4, false, false, arg8)
		fmt.Println("Bid Offer: ", bid_order.Result.Price)

		// market, side, price, type, size, reduce only, IOC, post only

		ask_order, _ := client.PlaceOrder(arg9, "sell", ask_price, "limit", arg4, false, false, arg8)
		fmt.Println("Ask Offer: ", ask_order.Result.Price)

		time.Sleep(time.Second * time.Duration(arg7))

		openOrders, _ := client.GetOpenOrders(arg9)

		if len(openOrders.Result) == 0 {

			fmt.Println("Both Orders Have Been Filled")

		} else if len(openOrders.Result) > 4 {

			i := 0

			for i = 0; i < len(openOrders.Result); i++ {

				cancel, _ := client.CancelOrder(openOrders.Result[i].ID)
				fmt.Println("Order Side Cancelled: ", openOrders.Result[i].Side)
				fmt.Println(cancel)

			}

			fmt.Println("Too Many Cancellations Have Failed")
			os.Exit(0)

		} else {

			i := 0

			for i = 0; i < len(openOrders.Result); i++ {

				cancel, _ := client.CancelOrder(openOrders.Result[i].ID)
				fmt.Println("Order Side Cancelled: ", openOrders.Result[i].Side)
				fmt.Println(cancel)

			}

		}

	}

}

func (client *FtxClient) _delete(path string, body []byte) (*http.Response, error) {

	preparedRequest := client.signRequest("DELETE", path, body)
	resp, err := client.Client.Do(preparedRequest)
	return resp, err

}

type Response struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
}

func (client *FtxClient) CancelOrder(orderId int64) (Response, error) {

	var deleteResponse Response
	id := strconv.FormatInt(orderId, 10)
	resp, err := client._delete("orders/"+id, []byte(""))

	if err != nil {

		fmt.Println("Error CancelOrder", err)
		return deleteResponse, err

	}

	err = _processResponse(resp, &deleteResponse)
	return deleteResponse, err

}

type Positions struct {
	Success bool `json:"success"`
	Result  []struct {
		Cost                         float64 `json:"cost"`
		EntryPrice                   float64 `json:"entryPrice"`
		EstimatedLiquidationPrice    float64 `json:"estimatedLiquidationPrice"`
		Future                       string  `json:"future"`
		InitialMarginRequirement     float64 `json:"initialMarginRequirement"`
		LongOrderSize                float64 `json:"longOrderSize"`
		MaintenanceMarginRequirement float64 `json:"maintenanceMarginRequirement"`
		NetSize                      float64 `json:"netSize"`
		OpenSize                     float64 `json:"openSize"`
		RealizedPnl                  float64 `json:"realizedPnl"`
		ShortOrderSize               float64 `json:"shortOrderSize"`
		Side                         string  `json:"side"`
		Size                         float64 `json:"size"`
		UnrealizedPnl                float64 `json:"unrealizedPnl"`
	} `json:"result"`
}

func (client *FtxClient) _get(path string, body []byte) (*http.Response, error) {

	preparedRequest := client.signRequest("GET", path, body)
	resp, err := client.Client.Do(preparedRequest)
	return resp, err

}

func (client *FtxClient) GetPositions(showAvgPrice bool) (Positions, error) {

	var positions Positions

	resp, err := client._get("positions", []byte(""))

	if err != nil {

		fmt.Println("Error GetPositions", err)
		return positions, err

	}

	err = _processResponse(resp, &positions)
	return positions, err

}

func Get_Positions(arg1 string, arg2 string, arg3 string, arg4 string) (float64, float64, float64, float64) {

	// arg1 = api key, arg2 = api secret, arg3 = subaccount, arg4 = ticker symbol

	client := New(arg1, arg2, arg3)

	positions, _ := client.GetPositions(true)

	i := 0

	for i = 0; i < len(positions.Result); i++ {

		if arg4 == positions.Result[i].Future {

			return positions.Result[i].RealizedPnl, positions.Result[i].UnrealizedPnl, positions.Result[i].NetSize, positions.Result[i].EntryPrice

		}

	}

	return 0.0, 0.0, 0.0, 0.0

}

type Finnhub_Response struct {
	Lowerband  []float64 `json:"lowerband"`
	Middleband []float64 `json:"middleband"`
	Upperband  []float64 `json:"upperband"`
}

func getBollinger(symbol string, resolution string, timeperiod string, nbdevdn string, nbdevup string) (float64, float64, float64) {
	//Finnhub API Key:
	var APIKEY string = "c3r13u2ad3i98m4ier5g"

	var to_unix_timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	var from_unix_timestamp = strconv.FormatInt(time.Now().Unix() - 10000, 10)

	var generatedURL string = "https://finnhub.io/api/v1/indicator?symbol=" + symbol + "&indicator=BBANDS&resolution=" + resolution + "&from=" + from_unix_timestamp + "&to=" + to_unix_timestamp + "&timeperiod=" + timeperiod + "&nbdevup=" + nbdevup + "&dbdevdn=" + nbdevdn + "&token=" + APIKEY

	client := &http.Client{}
	req, err := http.NewRequest("POST", generatedURL, nil)
	if err != nil {
		fmt.Print(err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}

	var responseObject Finnhub_Response
	if err := json.Unmarshal(bodyBytes, &responseObject); err != nil {  // Parse []byte to the go struct pointer
	  fmt.Println("Can not unmarshal JSON")
	}
	var lowerband float64 = responseObject.Lowerband[len(responseObject.Lowerband)-1]
	var middleband float64 = responseObject.Middleband[len(responseObject.Middleband)-1]
	var upperband float64 = responseObject.Upperband[len(responseObject.Upperband)-1]

	return lowerband, middleband, upperband
}


func main() {

	// Test End
	fmt.Println("Trading Terminal Has Been Started !!")

	ticker_array := []string{"ETH-PERP", "BTC-PERP", "UNI-PERP", "LINK-PERP", "MKR-PERP", "DOGE-PERP"}

	fmt.Println("Please Enter Ticker Symbol")
	fmt.Println(ticker_array)
	var ticker_symbol string
	fmt.Scanln(&ticker_symbol)

	ticker_found := 0

	index := 0

	for index = 0; index < len(ticker_array); index++ {

		if ticker_array[index] == ticker_symbol {

			ticker_found = 1

		}

	}

	if ticker_found == 0 {

		fmt.Println("Ticker Not Defined Within Data Structure")
		os.Exit(0)

	}

	fmt.Println("Please Enter Stake Price")
	var stake_price float64
	fmt.Scanln(&stake_price)

	fmt.Println("Please Enter Upper Threshold")
	var upper_threshold float64
	fmt.Scanln(&upper_threshold)

	if upper_threshold < stake_price {

		fmt.Println("Something Went Wrong")
		os.Exit(0)

	}

	fmt.Println("Please Enter Lower Threshold")
	var lower_threshold float64
	fmt.Scanln(&lower_threshold)

	if lower_threshold > stake_price {

		fmt.Println("Something Went Wrong")
		os.Exit(0)

	}

	fmt.Println("Whats is Position Size")
	var position_size float64
	fmt.Scanln(&position_size)

	fmt.Println("Please Enter Stake Multiplier")
	var multiplier float64
	fmt.Scanln(&multiplier)

	fmt.Println("Please Enter Volatility Time Interval (15, 60, 300)")
	var volatility_interval string
	fmt.Scanln(&volatility_interval)

	if (volatility_interval != "15") && (volatility_interval != "60") && (volatility_interval != "300") {

		fmt.Println("Something Went Wrong")
		os.Exit(0)

	}

	fmt.Println("Please Enter Order Book Depth")
	var order_book_depth int
	fmt.Scanln(&order_book_depth)

	if order_book_depth < 10 {

		fmt.Println("Order Book Depth Too Small")
		os.Exit(0)

	}

	fmt.Println("Please Enter Risk Aversion Parameter")
	fmt.Println("Please Note: Gamma Must Exist Within The Set (0, 1)")
	var gamma float64
	fmt.Scanln(&gamma)

	if gamma >= 1 || gamma <= 0 {

		fmt.Println("The Parameter Entered For Gamma Is Incorrect")
		fmt.Println("Trading Terminal Has Been Shut Down")
		os.Exit(0)

	}

	fmt.Println("Please Enter Trade Amount")
	var max_trade_amount float64
	var order_trade_amount float64
	fmt.Scanln(&max_trade_amount)

	if max_trade_amount < 0.001 {

		fmt.Println("The Parameter Entered For Max Trade Amount Is Incorrect")
		fmt.Println("Trading Terminal Has Been Shut Down")
		os.Exit(0)

	}

	fmt.Println("Please Enter Order Refresh Time In Seconds")
	var order_time int64
	fmt.Scanln(&order_time)

	fmt.Println("Please Enter Minimum Spread")
	var minimum_spread float64
	fmt.Scanln(&minimum_spread)

	fmt.Println("Please Enter Price Aggressor Multiplier")
	var price_aggressor float64
	fmt.Scanln(&price_aggressor)

	fmt.Println("Post Only?")
	fmt.Println("0 For False, 1 For True")
	var taker_binary int64
	fmt.Scanln(&taker_binary)

	var post_only bool

	if taker_binary == 0 {

		post_only = false

	} else if taker_binary == 1 {

		post_only = true

	} else {

		fmt.Println("Something Went Wrong")
		os.Exit(0)

	}

	fmt.Println("Please Enter Inventory Cutoff (ABSOLUTE VALUE TARGET DISTANCE)")
	var inventory_cutoff float64
	fmt.Scanln(&inventory_cutoff)
	
    fmt.Println("Enter timeperiod for BBands (in minutes):")
    var timeperiod string
    fmt.Scanln(&timeperiod)

    fmt.Println("Enter nbdevup for Upper BBands:")
    var nbdevup string
    fmt.Scanln(&nbdevup)

    fmt.Println("Enter nbdevdn for Lower BBands:")
    var nbdevdn string
    fmt.Scanln(&nbdevdn)

    lowerband, middleband, upperband := getBollinger("BINANCE:ETHUSDT", "1", timeperiod, nbdevdn, nbdevup)
    fmt.Printf("Lowerband: ", lowerband)
    fmt.Printf("Middleband: ", middleband)
    fmt.Printf("Upperband: ", upperband)
    print("\n")


	api_key := "xxxx"
	api_secret := "yyyy"
	subaccount := "Test10"

	last_vol := 0.0
	var inventory_target float64

	i := 0

	for i == 0 {

		sigma := math.Round(Sigma(ticker_symbol, volatility_interval, api_key, api_secret, subaccount, last_vol)*100) / 100
		fmt.Println("Sigma: ", sigma)

		last_vol = sigma

		midpoint, weighted_midpoint, kappa, best_bid, best_ask := Kappa(order_book_depth, ticker_symbol)
		fmt.Println("Kappa: ", kappa)

		midpoint = math.Round(midpoint*100) / 100
		fmt.Println("Midpoint Price: ", midpoint)

		weighted_midpoint = math.Round(weighted_midpoint*100) / 100
		fmt.Println("Weighted Midpoint Price: ", weighted_midpoint)

		realized_pnl, unrealized_pnl, current_inventory, entry_price := Get_Positions(api_key, api_secret, subaccount, ticker_symbol)
		fmt.Println("Current Inventory: ", current_inventory)
		fmt.Println("Realized Profit & Loss: ", realized_pnl)
		fmt.Println("Unrealized Profit & Loss: ", unrealized_pnl)
		fmt.Println("Entry Price: ", entry_price)

		if weighted_midpoint >= upper_threshold {

			inventory_target = (position_size * multiplier)

		} else if weighted_midpoint <= lower_threshold {

			inventory_target = -(position_size * multiplier)

		} else if weighted_midpoint < stake_price {

			inventory_target = (((weighted_midpoint - stake_price) / (stake_price - lower_threshold)) * position_size * multiplier)

		} else if weighted_midpoint == stake_price {

			inventory_target = 0

		} else if weighted_midpoint > stake_price {

			inventory_target = (((weighted_midpoint - stake_price) / (upper_threshold - stake_price)) * position_size * multiplier)

		}

		inventory_target = math.Round(inventory_target*10000) / 10000
		fmt.Println("Inventory Target:", inventory_target)

		target_distance := math.Round((current_inventory-inventory_target)*10000) / 10000
		fmt.Println("Target Distance: ", target_distance)
		
		order_trade_amount =  math.Min(max_trade_amount, math.Abs(target_distance/5))
		order_trade_amount = math.Round(order_trade_amount*10000)/10000
		fmt.Println("Trade Amount: ", order_trade_amount)

		reserve_price, aggressive_reserve_price := Reservation_Price(weighted_midpoint, target_distance, gamma, sigma, price_aggressor, position_size)

		reserve_price = math.Round(reserve_price*100) / 100
		aggressive_reserve_price = math.Round(aggressive_reserve_price*100) / 100

		spread := math.Round(Optimal_Spread(gamma, sigma, kappa)*100) / 100

		if spread < minimum_spread {

			spread = minimum_spread

		}

		fmt.Println("Reservation Price: ", reserve_price)
		fmt.Println("Aggressive Reservation Price: ", aggressive_reserve_price)
		fmt.Println("Optimal Spread: ", spread)
		Place_Order(api_key, api_secret, subaccount, order_trade_amount, aggressive_reserve_price, spread, order_time, post_only, ticker_symbol, best_bid, best_ask, inventory_cutoff, target_distance)
		//if (weighted_midpoint >= upper_threshold && current_inventory >= inventory_target) {
		//	fmt.Println("weighted_midpoint >= upper_threshold && current_inventory >= inventory_target, stop placing orders.")
		//} else if (weighted_midpoint <= lower_threshold && current_inventory >= inventory_target) {
		//	fmt.Println("weighted_midpoint <= lower_threshold && current_inventory >= inventory_target, stop placing orders.")
		//} else if (weighted_midpoint <= upperband && weighted_midpoint >= lowerband) {
		//	fmt.Println("price within bollinger band, stop placing orders.")
		//} else {
		//	Place_Order(api_key, api_secret, subaccount, order_trade_amount, aggressive_reserve_price, spread, order_time, post_only, ticker_symbol, best_bid, best_ask, inventory_cutoff, target_distance)
		//}
	}

}
