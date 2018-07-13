package handler

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"godemo/conf"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	NotifyUrl     string
	ReturnUrl     string
	PayUrl        string
	QueryUrl      string
	OrderCheckUrl string
	UID           string
	SECRET        string
)

type Product struct {
	Price     string
	Goodsname string
}

type BaseJsonBean struct {
	Code    int     `json:"code,string"`
	Data    PayData `json:"data"`
	Message string  `json:"msg"`
}

type PayData struct {
	Qrcode    string `json:"Qrcode"`
	IsType    string `json:"istype"`
	Realprice string `json:"realprice"`
	Url       string `json:"url"`
}

var ProductMock map[string]Product

// const (
// 	UID    = "227591459098857472"
// 	SECRET = "7hvszwk1b182tvjzjpezi4hx9gvmkir0"
// )

func Init() {
	NotifyUrl = conf.AppConfig.Host + conf.AppConfig.ListenAddr + "/notify"
	ReturnUrl = conf.AppConfig.Host + conf.AppConfig.ListenAddr + "/return"
	PayUrl = conf.AppConfig.CcpayHost + "/ccpay/ach/pay"
	QueryUrl = conf.AppConfig.CcpayHost + "/ccpay/ach/query"
	OrderCheckUrl = conf.AppConfig.CcpayHost + "/ccpay/ach/order/check"

	UID = conf.AppConfig.Uid
	SECRET = conf.AppConfig.Secret

	p := make(map[string]Product)

	p["1"] = Product{Price: "10", Goodsname: "0.1 元"}
	p["2"] = Product{"20", "0.2 元"}
	p["3"] = Product{"1000", "10 元"}
	p["4"] = Product{"100", "1.0 元"}

	ProductMock = p
}

func ErrorHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "error.html", nil)
}

func NewBaseJsonBean() *BaseJsonBean {
	return &BaseJsonBean{}
}

func GenerateRandnum(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(max - min)
	num = num + min
	return num
}

func newOrderID() string {
	t := time.Now().Format("20060102150405")
	t = fmt.Sprintf("%s%d", t, GenerateRandnum(1000, 9999))
	id, _ := strconv.ParseInt(t, 10, 64)
	return fmt.Sprintf("%d", id)
}

// GetKeysAndValuesBySortKeys 排序
func GetKeysAndValuesBySortKeys(urlValues url.Values) (values []string) {
	vLen := len(urlValues)
	// get keys
	keys := make([]string, vLen)
	i := 0
	for k := range urlValues {
		keys[i] = k
		i++
	}
	// sort keys
	sort.Sort(sort.StringSlice(keys))
	values = make([]string, vLen)
	for i, k := range keys {
		values[i] = fmt.Sprintf(`%s=%s`, k, urlValues.Get(k))
	}
	return
}

// 秘钥 md5加密
func getKey(key string) string {
	md := md5.New()
	md.Write([]byte(key))
	md5_key := hex.EncodeToString(md.Sum(nil))
	return md5_key
}
