package client

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	// WechatUnifiedOrder 微信统一下订单接口
	WechatUnifiedOrder = "https://api.mch.weixin.qq.com/pay/unifiedorder"
)

// WechatClient struct
type WechatClient struct{}

// NewWechatClient creates a new WechatClient
func NewWechatClient() *WechatClient {
	return &WechatClient{}
}

// UnifyOrderRequest 用于填入我们要传入的参数。
type UnifyOrderRequest struct {
	AppID          string `xml:"appid"`            //公众账号ID
	Body           string `xml:"body"`             //商品描述
	MchID          string `xml:"mch_id"`           //商户号
	NonceStr       string `xml:"nonce_str"`        //随机字符串
	NotifyURL      string `xml:"notify_url"`       //通知地址
	OpenID         string `xml:"openid,omitempty"` //购买商品的用户wxid
	OutTradeNo     string `xml:"out_trade_no"`     //商户订单号
	Sign           string `xml:"sign"`             //签名
	SpbillCreateIP string `xml:"spbill_create_ip"` //支付提交用户端ip
	TotalFee       string `xml:"total_fee"`        //总金额
	TradeType      string `xml:"trade_type"`       //交易类型
}

// UnifyOrderResponse 统一订单接口返回
type UnifyOrderResponse struct {
	ReturnCode string `xml:"return_code"`
	ReturnMsg  string `xml:"return_msg"`
	AppID      string `xml:"appid"`
	MchID      string `xml:"mch_id"`
	NonceStr   string `xml:"nonce_str"`
	Sign       string `xml:"sign"`
	ResultCode string `xml:"result_code"`
	ErrorCode  string `xml:"err_code"`
	PrepayID   string `xml:"prepay_id"`
	TradeType  string `xml:"trade_type"`
}

// Base xml params struct
type Base struct {
	AppID    string `xml:"appid"`
	MchID    string `xml:"mch_id"`
	NonceStr string `xml:"nonce_str"`
	Sign     string `xml:"sign"`
}

// PayNotifySuccessReturn params of notify success
type PayNotifySuccessReturn struct {
	ReturnCode CdataString `xml:"return_code"`
	ReturnMsg  CdataString `xml:"return_msg"`
}

// CdataString xml output with cdata
type CdataString struct {
	Value string `xml:",cdata"`
}

// PayNotifyResult notify params struct
type PayNotifyResult struct {
	Base
	DeviceInfo    string `xml:"device_info,omitempty"`
	SignType      string `xml:"sign_type,omitempty"`
	ReturnCode    string `xml:"return_code"`
	ReturnMsg     string `xml:"return_msg"`
	ResultCode    string `xml:"result_code"`
	OpenID        string `xml:"openid"`
	IsSubscribe   string `xml:"is_subscribe,omitempty"`
	TradeType     string `xml:"trade_type"`
	BankType      string `xml:"bank_type"`
	TotalFee      int    `xml:"total_fee"`
	FeeType       string `xml:"fee_type,omitempty"`
	CashFee       int    `xml:"cash_fee"`
	CashFeeType   string `xml:"cash_fee_type,omitempty"`
	TransactionID string `xml:"transaction_id"`
	OutTradeNo    string `xml:"out_trade_no"`
	Attach        string `xml:"attach,omitempty"`
	TimeEnd       string `xml:"time_end"`
}

// Pay wechat pay
func (wechat *WechatClient) Pay(charge *Charge) (string, error) {
	var unifyOrder UnifyOrderRequest
	unifyOrder.AppID = charge.AppID
	unifyOrder.Body = charge.Body
	unifyOrder.MchID = charge.MchID
	unifyOrder.NonceStr = RandomStr()
	unifyOrder.NotifyURL = charge.NotifyURL
	unifyOrder.OpenID = charge.OpenID
	unifyOrder.OutTradeNo = charge.OutTradeNo
	unifyOrder.SpbillCreateIP = LocalIP()
	unifyOrder.TotalFee = fmt.Sprintf("%d", int(charge.TotalFee*100))
	unifyOrder.TradeType = charge.TradeType

	var m = make(map[string]string)
	m["appid"] = unifyOrder.AppID
	m["mch_id"] = unifyOrder.MchID
	m["nonce_str"] = unifyOrder.NonceStr
	m["body"] = unifyOrder.Body
	m["out_trade_no"] = unifyOrder.OutTradeNo
	m["total_fee"] = string(unifyOrder.TotalFee)
	m["spbill_create_ip"] = unifyOrder.SpbillCreateIP
	m["notify_url"] = unifyOrder.NotifyURL
	m["trade_type"] = unifyOrder.TradeType
	m["openid"] = unifyOrder.OpenID

	unifyOrder.Sign = wechat.GenerateSign(m, charge.Key)
	code, _ := xml.Marshal(unifyOrder)

	ret := strings.Replace(string(code), "UnifyOrderRequest", "xml", -1)

	// 打印请求数据，用以错误跟踪
	fmt.Println(ret)

	xmlStr := []byte(ret)
	//发送unified order请求.
	req, err := http.NewRequest("POST", WechatUnifiedOrder, bytes.NewReader(xmlStr))

	if err != nil {
		return "New Http Request发生错误，原因:", err
	}

	req.Header.Set("Accept", "application/xml")
	//这里的http header的设置是必须设置的.
	req.Header.Set("Content-Type", "application/xml;charset=utf-8")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "请求微信支付统一下单接口发送错误, 原因:", err
	}

	var xmlRe UnifyOrderResponse
	respBytes, err := ioutil.ReadAll(resp.Body)
	err = xml.Unmarshal(respBytes, &xmlRe)
	if err != nil {
		return "解析xml错误，原因：", err
	}

	if xmlRe.ReturnCode != "SUCCESS" {
		return "通信失败，原因：", errors.New(xmlRe.ReturnMsg)
	}

	if xmlRe.ResultCode != "SUCCESS" {
		return "支付失败，原因：", errors.New(xmlRe.ErrorCode)
	}

	var c = make(map[string]string)

	if charge.TradeType == "JSAPI" {
		c["appId"] = xmlRe.AppID
		c["timeStamp"] = fmt.Sprintf("%d", time.Now().Unix())
		c["nonceStr"] = RandomStr()
		c["package"] = fmt.Sprintf("prepay_id=%s", xmlRe.PrepayID)
		c["signType"] = "MD5"
		c["paySign"] = wechat.GenerateSign(c, charge.Key)
	}

	if charge.TradeType == "APP" {
		c["appid"] = xmlRe.AppID
		c["partnerid"] = xmlRe.MchID
		c["prepayid"] = xmlRe.PrepayID
		c["package"] = "Sign=WXPay"
		c["noncestr"] = RandomStr()
		c["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
		c["sign"] = wechat.GenerateSign(c, charge.Key)
	}

	jsonC, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	return string(jsonC), nil
}

// GenerateSign generate sign
func (wechat *WechatClient) GenerateSign(m map[string]string, key string) string {
	var signData []string
	for k, v := range m {
		if v != "" {
			signData = append(signData, fmt.Sprintf("%s=%s", k, v))
		}
	}

	sort.Strings(signData)
	signStr := strings.Join(signData, "&")
	signStr = signStr + "&key=" + key

	md5Ctx := md5.New()
	md5Ctx.Write([]byte(signStr))
	cipherStr := md5Ctx.Sum(nil)
	upperSign := strings.ToUpper(hex.EncodeToString(cipherStr))

	return upperSign
}
