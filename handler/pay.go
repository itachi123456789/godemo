package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	bytes2 "bytes"

	"github.com/gin-gonic/gin"
)

func QrHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func PayHandler(c *gin.Context) {
	values := url.Values{}

	// 网页获取的price:支付价格，
	prok := c.PostForm("product")
	mock := ProductMock[prok]

	// 网页获取的istype:支付渠道：100-支付宝；200-微信支付
	payType := c.PostForm("pay_type")

	orderuid := "u123"
	orderid := newOrderID()

	uid := c.PostForm("uid")
	secret := c.PostForm("secret")

	return_url := ReturnUrl
	notify_url := NotifyUrl

	values.Add("uid", uid)
	values.Add("price", mock.Price)
	values.Add("pay_type", payType)
	values.Add("notify_url", notify_url)
	values.Add("return_url", return_url)
	values.Add("orderid", orderid)
	values.Add("goodsname", mock.Goodsname)
	values.Add("orderuid", orderuid)
	values.Add("user_ip", c.ClientIP())

	sv := GetKeysAndValuesBySortKeys(values)
	md5sum := md5.New()

	params := strings.Join(sv, "&")
	md5sum.Write([]byte(params))
	md5sum.Write([]byte(secret))
	key := hex.EncodeToString(md5sum.Sum(nil))
	// 注意：Token在安全上非常重要，一定不要显示在任何网页代码、网址参数中。只可以放在服务端。计算key时，先在服务端计算好，把计算出来的key传出来。严禁在客户端计算key，严禁在客户端存储Token。
	fmt.Printf("key is %s\n", key)
	values.Add("key", key)

	var bmap = make(map[string]string)
	for k, v := range values {
		bmap[k] = v[0]
	}
	b, _ := json.Marshal(bmap)

	fmt.Printf("request: %s , body:%s\n\n", PayUrl, string(b))

	resp, err := http.DefaultClient.Post(PayUrl, "application/json", bytes2.NewBuffer(b))
	if err != nil {
		fmt.Errorf("http resp err:%v", err)
		return
	}

	// 返回值
	result := NewBaseJsonBean()
	if resp == nil {
		result.Code = -1
		result.Message = "获取支付二维码失败"
		// bytes, _ := json.Marshal(result)
		// c.JSON(http.StatusOK, string(bytes))
		// fmt.Printf("请求服务器出错: resp: %+v, err: %+v", resp, err)
		c.Redirect(http.StatusMovedPermanently, "/error")
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("请求支付服务器出错\n")
	}
	buf := bytes2.Buffer{}
	io.Copy(&buf, resp.Body)

	fmt.Printf("response :%s\n\n", buf.String())

	type Result struct {
		Code int    `json:"code,string"`
		Msg  string `json:"msg"`
		Data struct {
			Result struct {
				QrUrl      string `json:"qr_url"`
				ParseUrl   string `json:"parse_url"`
				OutOrderid string `json:"out_order_id"`
				Price      string `json:"price"`
				Goodsname  string `json:"goodsname"`
				Orderid    string `json:"orderid"`
			} `json:"result"`
		} `json:"data"`
	}

	res := new(Result)
	err = json.Unmarshal(buf.Bytes(), res)
	if err != nil {
		fmt.Printf("json unmarshal err:%s\n", err)
	}

	fmt.Printf("unmarshal:  %+v\n\n", res)

	// 1为成功 -1为失败
	if res.Code != 1 {
		c.Redirect(http.StatusMovedPermanently, "/error")
		return
	}

	c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/qr?qrcodeUrl=%s&id=%s", res.Data.Result.QrUrl, res.Data.Result.OutOrderid))
	return

}

func QueryHandler(c *gin.Context) {
	outOrderid := c.PostForm("id")

	fmt.Printf("h5 params id = %s\n", outOrderid)

	values := url.Values{
		"uid":          []string{UID},
		"out_order_id": []string{outOrderid},
	}

	sv := GetKeysAndValuesBySortKeys(values)
	md5sum := md5.New()

	params := strings.Join(sv, "&")
	md5sum.Write([]byte(params))
	md5sum.Write([]byte(SECRET))

	key := hex.EncodeToString(md5sum.Sum(nil))
	fmt.Printf("key is %s\n", key)
	values.Add("key", key)

	var bmap = make(map[string]string)
	for k, v := range values {
		bmap[k] = v[0]
	}
	b, _ := json.Marshal(bmap)

	fmt.Printf("request: %s , body:%s\n\n", QueryUrl, string(b))

	resp, err := http.DefaultClient.Post(QueryUrl, "application/json", bytes2.NewBuffer(b))
	if err != nil {
		fmt.Errorf("http resp err:%v", err)
		return
	}

	// result := NewBaseJsonBean()
	if resp == nil {
		c.Redirect(http.StatusMovedPermanently, "/error")
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("请求支付服务器出错\n")
	}
	buf := bytes2.Buffer{}
	io.Copy(&buf, resp.Body)

	fmt.Printf("response :%s\n\n", buf.String())

	type Result struct {
		Code int    `json:"code,string"`
		Msg  string `json:"msg"`
		Data struct {
			Result struct {
				OutOrderid string `json:"out_order_id"`
				Status     string `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}

	res := new(Result)
	err = json.Unmarshal(buf.Bytes(), res)
	if err != nil {
		fmt.Printf("json unmarshal err:%s\n", err)
	}

	fmt.Printf("unmarshal:  %+v\n\n", res)

	c.JSON(http.StatusOK, res)
}
