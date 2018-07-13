package handler

import (
	bytes2 "bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func OrderHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "order_check.html", nil)
}

// 订单申诉
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
