package client

import (
	"fmt"
	"net"
	"time"
)

// Charge order params
//
//swagger:parameters pay
type Charge struct {
	AppID      string  `json:"appid" form:"appid"`
	MchID      string  `json:"mch_id" form:"mch_id"`
	Key        string  `json:"key" form:"key"`
	Body       string  `json:"body,omitempty" form:"body"`
	OpenID     string  `json:"openid,omitempty" form:"openid"`
	OutTradeNo string  `json:"out_trade_no,omitempty" form:"out_trade_no"`
	TotalFee   float32 `json:"total_fee,omitempty" form:"total_fee"`
	TradeType  string  `json:"trade_type,omitempty" form:"trade_type"`
	NotifyURL  string  `json:"notify_url,omitempty" form:"notify_url"`
}

//RandomStr 获取一个随机字符串
func RandomStr() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// LocalIP 获取机器的IP
func LocalIP() string {
	info, _ := net.InterfaceAddrs()
	for _, addr := range info {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}
	return ""
}
