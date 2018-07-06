package client

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
}
