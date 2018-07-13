package module

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"godemo/conf"
	"net/http"

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

func IndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func TestHandler(c *gin.Context) {
	fmt.Printf("%+v", c.PostForm("price"))
	c.JSON(http.StatusOK, "ok")
}

// 通知异步回调接收 支付成功，我们会根据您之前传入的notify_url，回调此页URL，post回参数
func NotifyHandler(c *gin.Context) {
	// w := c.Writer
	// r := c.Request
	// body, err := ioutil.ReadAll(r.Body)
	// fmt.Printf("支付回调 notify :%s,err:%s\n", string(body), err)
	// result := NewBaseJsonBean()
	// result.Code = 1
	// result.Message = "success"

	// bytes, _ := json.Marshal(result)
	// fmt.Fprint(w, string(bytes))
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
