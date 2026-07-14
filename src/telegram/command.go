package telegram

import tb "gopkg.in/telebot.v3"

const (
	START_CMD  = "/start"
	REPAIR_CMD = "/repair"
	NOTIFY_CMD = "/notify"
)

var Cmds = []tb.Command{
	{
		Text:        START_CMD,
		Description: "开始",
	},
	{
		Text:        REPAIR_CMD,
		Description: "过期订单补单：/repair 交易号 链上交易hash",
	},
	{
		Text:        NOTIFY_CMD,
		Description: "重发回调通知：/notify 交易号",
	},
}
