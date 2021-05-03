package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/koron/go-dproxy"
)

var mapOrder = []string{
		"FH_TIMESTAMP", "FH_EXPIRY_DT", "FH_OPTION_TYPE", "FH_STRIKE_PRICE", "FH_OPENING_PRICE",
		"FH_TRADE_HIGH_PRICE", "FH_TRADE_LOW_PRICE", "FH_CLOSING_PRICE", "FH_LAST_TRADED_PRICE",
		"FH_SETTLE_PRICE", "FH_TOT_TRADED_QTY", "FH_TOT_TRADED_VAL", "CALCULATED_PREMIUM_VAL",
		"FH_OPEN_INT", "FH_CHANGE_IN_OI",
}

var columnNames = []string {
	"Date", "Expiry Date", "Option Type", "Strike Price", "Open Price", "High Price", " Low Price",
	"Close Price", "Last Price", "Settled Price", "Volume", "Value", "Premium Value", "Open Interest",
	"Change in OI",
}

var futureList = []string{
	"AARTIIND","ACC","ADANIENT","ADANIPORTS","ALKEM","AMARAJABAT","AMBUJACEM","APLLTD",
	"APOLLOHOSP","APOLLOTYRE","ASHOKLEY","ASIANPAINT","AUBANK","AUROPHARMA","AXISBANK",
	"BAJAJ-AUTO","BAJAJFINSV","BAJFINANCE","BALKRISIND","BANDHANBNK","BANKBARODA","BANKNIFTY",
	"BATAINDIA","BEL","BERGEPAINT","BHARATFORG","BHARTIARTL","BHEL","BIOCON","BOSCHLTD","BPCL",
	"BRITANNIA","CADILAHC","CANBK","CHOLAFIN","CIPLA","COALINDIA","COFORGE","COLPAL",
	"CONCOR","CUB","CUMMINSIND","DABUR","DEEPAKNTR","DIVISLAB","DLF","DRREDDY","EICHERMOT",
	"ESCORTS","EXIDEIND","FEDERALBNK","GAIL","GLENMARK","GMRINFRA","GODREJCP","GODREJPROP",
	"GRANULES","GRASIM","GUJGASLTD","HAVELLS","HCLTECH","HDFC","HDFCAMC","HDFCBANK","HDFCLIFE",
	"HEROMOTOCO","HINDALCO","HINDPETRO","HINDUNILVR","IBULHSGFIN","ICICIBANK","ICICIGI","ICICIPRULI",
	"IDEA","IDFCFIRSTB","IGL","INDIGO","INDUSINDBK","INDUSTOWER","INFY","IOC","IRCTC","ITC",
	"JINDALSTEL","JSWSTEEL","JUBLFOOD","KOTAKBANK","L&TFH","LALPATHLAB","LICHSGFIN","LT","LTI",
	"LTTS","LUPIN","M&M","M&MFIN","MANAPPURAM","MARICO","MARUTI","MCDOWELL-N","MFSL","MGL","MINDTREE",
	"MOTHERSUMI","MPHASIS","MRF","MUTHOOTFIN","NAM-INDIA","NATIONALUM","NAUKRI","NAVINFLUOR","NESTLEIND",
	"NIFTY","NMDC","NTPC","ONGC","PAGEIND","PEL","PETRONET","PFC","PFIZER","PIDILITIND","PIIND","PNB",
	"POWERGRID","PVR","RAMCOCEM","RBLBANK","RECLTD","RELIANCE","SAIL","SBILIFE","SBIN","SHREECEM","SIEMENS",
	"SRF","SRTRANSFIN","SUNPHARMA","SUNTV","TATACHEM","TATACONSUM","TATAMOTORS","TATAPOWER","TATASTEEL","TCS",
	"TECHM","TITAN","TORNTPHARM","TORNTPOWER","TRENT","TVSMOTOR","UBL","ULTRACEMCO","UPL","VEDL","VOLTAS","WIPRO","ZEEL",
}

var from, to, expire = "03-04-2021", "03-05-2021", "27-May-2021"

var v interface{}
var wg sync.WaitGroup
var apiUrl = "https://www.nseindia.com/api/historical/fo/derivatives?&from="+from+"&to="+to+"&expiryDate="+expire+"&instrumentType=FUTSTK&symbol="
var	symbolUrl = "https://www.nseindia.com/get-quotes/derivatives?symbol="

func main() {
	wg.Add(len(futureList))
	for i := 0 ; i < len(futureList) ; i++ {
		// getCookie(symbolUrl+futureList[i], &futureList[i])
		go process(symbolUrl+futureList[i], &futureList[i])
		if i == 2 {
			time.Sleep(10 * time.Second)
		}
	}
	wg.Wait()
}

func getCookie(url string, symbol *string) *string {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("user-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36 Edg/88.0.705.74`)
	req.Header.Set("accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`)
	req.Header.Set("accept-language", "en-GB,en;q=0.9,en-US;q=0.8")

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != 200 {
		log.Fatalf("StatusCode error: %v\n\nTrying again....", res.StatusCode)
	} else {
		log.Println("Contact OK! CODE: ", res.StatusCode)
	}

	defer res.Body.Close()

	cookie := res.Cookies()

	body := apiFetch(apiUrl+*symbol, cookie)
	return body
}

func apiFetch(url string, cookie []*http.Cookie) *string {
	client := &http.Client{Timeout: time.Second * 10}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("user-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36 Edg/88.0.705.74`)
	req.Header.Set("accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9`)
	req.Header.Set("accept-language", "en-GB,en;q=0.9,en-US;q=0.8")
	for _, c := range cookie {
		req.AddCookie(c)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	
	if res.StatusCode != 200 {
		log.Fatalf("StatusCode error: %v\n\nTrying again....", res.StatusCode)
	} else {
		log.Println("Contact OK! CODE: ", res.StatusCode)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var b = string(body)

	return &b
}

func process(url string, symbol *string) {
	var column int
	var row int = 2
	f := excelize.NewFile()

	jsonStream := getCookie(url, symbol)
	err := json.Unmarshal([]byte(``+*jsonStream), &v)
	if err != nil {
		log.Fatal(err)
	}

	v := dproxy.New(v)
	length := v.M("data").ProxySet().Len()

	for i := 0 ; i < length ; i++ {
		column = 1
		m, err := v.M("data").A(i).Map()
		if err != nil {
			log.Fatal(err)
		}

		delete(m, "_id")
		delete(m, "TIMESTAMP")
		delete(m, "FH_INSTRUMENT")
		delete(m, "FH_MARKET_TYPE")
		for i := 0; i < len(mapOrder); i++ {	
			for k, v := range m {
				if k == mapOrder[i] {
					col, _ := excelize.ColumnNumberToName(column)
					err := f.SetCellValue("Sheet1", col+strconv.Itoa(row),v)
					if err != nil {
						log.Fatal(err)
					}
					column++
				}
			}
		}
		row++
	}

	for i := 0 ; i < len(columnNames) ; i++ {
		col, _ := excelize.ColumnNumberToName(i+1)
		err := f.SetCellValue("Sheet1", col+"1", columnNames[i])
		if err != nil {
			log.Fatal(err)
		}
	}
	
	if err := f.SaveAs("spreadsheets/"+*symbol+".xlsx"); err != nil {
		log.Fatal(err)
	}
	wg.Done()
}