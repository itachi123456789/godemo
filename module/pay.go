package module

import (
	bytes2 "bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"godemo/conf"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	NotifyUrl     string
	ReturnUrl     string
	PayUrl        string
	OrderCheckUrl string
)

func Init() {
	NotifyUrl = conf.AppConfig.Host + conf.AppConfig.ListenAddr + "/notify"
	ReturnUrl = conf.AppConfig.Host + conf.AppConfig.ListenAddr + "/return"
	PayUrl = conf.AppConfig.CcpayHost + "/ccpay/ach/pay"
	OrderCheckUrl = conf.AppConfig.CcpayHost + "/ccpay/ach/order/check"
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

func IndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func TestHandler(c *gin.Context) {
	fmt.Printf("%+v", c.PostForm("price"))
	c.JSON(http.StatusOK, "ok")
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

// 发起付款
func PayHandler(c *gin.Context) {
	values := url.Values{}

	// 网页获取的price:支付价格，
	price, _ := strconv.Atoi(c.PostForm("price"))
	// 网页获取的istype:支付渠道：100-支付宝；200-微信支付
	payType := c.PostForm("pay_type")

	orderuid := newOrderID()
	// 传入用户的用户名
	goodsname := c.PostForm("goodsname")
	// 订单号 我们会据此判别是同一笔订单还是新订单。我们回调时，会带上这个参数
	orderid := newOrderID()
	// 接入服务的uid
	uid := c.PostForm("uid")
	// 接入服务的token
	secret := c.PostForm("secret")
	// 用户支付成功后，我们会让用户浏览器自动跳转到这个网址
	return_url := ReturnUrl
	// 通知回调网址, 用户支付成功后，我们服务器会主动发送一个post消息到这个网址
	notify_url := NotifyUrl

	if price == 0 {
		c.JSON(http.StatusOK, "金额不正确")
		return
	}

	values.Add("uid", uid)
	values.Add("price", fmt.Sprintf("%d", price))
	values.Add("pay_type", payType)
	values.Add("notify_url", notify_url)
	values.Add("return_url", return_url)
	values.Add("orderid", orderid)
	values.Add("goodsname", goodsname)
	values.Add("orderuid", orderuid)
	values.Add("user_ip", c.ClientIP())
	sv := GetKeysAndValuesBySortKeys(values)
	md5sum := md5.New()

	params := strings.Join(sv, "&")
	fmt.Printf("parmas:%s\n", params)
	md5sum.Write([]byte(params))
	md5sum.Write([]byte(secret))
	key := hex.EncodeToString(md5sum.Sum(nil))
	// key的拼接(秘钥)按字母升序排列
	// 注意：Token在安全上非常重要，一定不要显示在任何网页代码、网址参数中。只可以放在服务端。计算key时，先在服务端计算好，把计算出来的key传出来。严禁在客户端计算key，严禁在客户端存储Token。
	fmt.Printf("key is %s\n", key)
	values.Add("key", key)

	var bmap = make(map[string]string)
	for k, v := range values {
		bmap[k] = v[0]
	}
	b, _ := json.Marshal(bmap)
	fmt.Printf("body:%s", string(b))
	resp, err := http.DefaultClient.Post(PayUrl, "application/json", bytes2.NewBuffer(b))
	if err != nil {
		fmt.Errorf("http resp err:%v", err)
	}
	// 返回值
	result := NewBaseJsonBean()
	if resp == nil {
		result.Code = -1
		result.Message = "获取支付二维码失败"
		bytes, _ := json.Marshal(result)
		c.JSON(http.StatusOK, string(bytes))
		fmt.Printf("请求服务器出错: resp: %+v, err: %+v", resp, err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("请求支付服务器出错\n")
	}
	buf := bytes2.Buffer{}
	io.Copy(&buf, resp.Body)
	fmt.Printf("buf:%s", buf.String())
	type Result struct {
		Code int    `json:"code,string"`
		Msg  string `json:"msg"`
		Data struct {
			Result struct {
				QrUrl string `json:"qr_url"`
			} `json:"result"`
		} `json:"data"`
	}
	res := new(Result)
	err = json.Unmarshal(buf.Bytes(), res)
	if err != nil {
		fmt.Printf("json unmarshal err:%s\n", err)
	}
	// 1为成功 -1为失败
	result.Code = 1
	// 提示给用户的文字信息，会根据不同场景，展示不同内容
	result.Message = "付款即时到账"
	// 判断支付成功后，要同步跳转的URL
	result.Data.Url = "判断支付成功后，要同步跳转的URL"
	// 需要展示的二维码
	result.Data.Qrcode = res.Data.Result.QrUrl
	// 支付渠道：1-支付宝；2-微信
	result.Data.IsType = "1"
	// 显示给用户的订单金额
	result.Data.Realprice = "50.05"

	bytes, _ := json.Marshal(result)
	// fmt.Fprint(w, string(bytes))
	c.JSON(http.StatusOK, string(bytes))
}

// 通知异步回调接收 支付成功，我们会根据您之前传入的notify_url，回调此页URL，post回参数
func NotifyHandler(c *gin.Context) {
	w := c.Writer
	r := c.Request
	body, err := ioutil.ReadAll(r.Body)
	fmt.Printf("支付回调 notify :%s,err:%s\n", string(body), err)
	result := NewBaseJsonBean()
	result.Code = 1
	result.Message = "success"

	bytes, _ := json.Marshal(result)
	fmt.Fprint(w, string(bytes))
}

// 付款成功自动跳转 用户付款成功后，我们会在先通过上面的回调接口，通知您服务器付款成功
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// 您可以通过此orderid在您后台查询到付款确实成功后，给用户一个付款成功的展示。
	orderid := r.PostFormValue("orderid")
	// Todo 订单信息是否被我们异步回调接口修改成支付成功状态
	// 此处在您数据库中查询：此笔订单号是否已经异步通知给您付款成功了。如成功了，就给他返回一个支付成功的展示。
	fmt.Fprint(w, "恭喜, 支付成功! 订单号: "+orderid)
}

// 秘钥 md5加密
func getKey(key string) string {
	md := md5.New()
	md.Write([]byte(key))
	md5_key := hex.EncodeToString(md.Sum(nil))
	return md5_key
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

func OrderHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "order_check.html", nil)
}

// 发起付款
func OrderCheckHandler(c *gin.Context) {
	values := url.Values{}

	// 订单号 我们会据此判别是同一笔订单还是新订单。我们回调时，会带上这个参数
	// 接入服务的uid
	uid := c.PostForm("uid")
	secret := c.PostForm("secret")
	orderid := c.PostForm("orderid")
	certimg := c.PostForm("certimg")

	values.Add("uid", uid)
	values.Add("out_order_id", orderid)
	values.Add("cert_img", certimg)
	values.Add("user_ip", c.ClientIP())
	sv := GetKeysAndValuesBySortKeys(values)
	md5sum := md5.New()

	params := strings.Join(sv, "&")
	fmt.Printf("parmas:%s\n", params)
	md5sum.Write([]byte(params))
	md5sum.Write([]byte(secret))
	key := hex.EncodeToString(md5sum.Sum(nil))
	// key的拼接(秘钥)按字母升序排列
	fmt.Printf("key is %s\n", key)
	values.Add("key", key)

	var bmap = make(map[string]string)
	for k, v := range values {
		bmap[k] = v[0]
	}
	b, _ := json.Marshal(bmap)
	fmt.Printf("支付申诉 请求 body:%s\n", string(b))
	resp, err := http.DefaultClient.Post(OrderCheckUrl, "application/json", bytes2.NewBuffer(b))
	if err != nil {
		fmt.Errorf("http resp err:%v", err)
	}
	// 返回值
	result := NewBaseJsonBean()
	if resp == nil {
		result.Code = -1
		result.Message = "失败"
		bytes, _ := json.Marshal(result)
		c.JSON(http.StatusOK, string(bytes))
		fmt.Printf("请求服务器出错: resp: %+v, err: %+v\n", resp, err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("请求支付服务器出错\n")
	}
	buf := bytes2.Buffer{}
	io.Copy(&buf, resp.Body)
	// fmt.Printf("buf:%s", buf.String())
	type Result struct {
		Code int    `json:"code,string"`
		Msg  string `json:"msg"`
		Data struct {
			Result struct {
				Status string `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	res := new(Result)
	err = json.Unmarshal(buf.Bytes(), res)
	if err != nil {
		fmt.Printf("json unmarshal err:%s\n", err)
	}
	bytes, _ := json.Marshal(res)
	fmt.Printf("支付申诉返回 :%+v\n", string(bytes))
	c.JSON(http.StatusOK, string(bytes))
}
