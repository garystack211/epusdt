package telegram

import (
	"fmt"
	"github.com/assimon/luuu/model/data"
	"github.com/assimon/luuu/model/mdb"
	"github.com/assimon/luuu/model/service"
	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/strutil"
	tb "gopkg.in/telebot.v3"
)

const (
	ReplayAddWallet = "请发给我一个合法的钱包地址"
)

func OnTextMessageHandle(c tb.Context) error {
	if c.Message().ReplyTo.Text == ReplayAddWallet {
		defer bots.Delete(c.Message().ReplyTo)
		_, err := data.AddWalletAddress(c.Message().Text)
		if err != nil {
			return c.Send(err.Error())
		}
		c.Send(fmt.Sprintf("钱包[%s]添加成功！", c.Message().Text))
		return WalletList(c)
	}
	return nil
}

// OrderRepairHandle 过期订单手动补单：/repair <交易号> <链上交易hash>
// 只把订单标记为支付成功，不自动回调；回调请随后用 /notify 手动触发。
func OrderRepairHandle(c tb.Context) error {
	args := c.Args()
	if len(args) < 2 {
		return c.Send("用法：/repair 交易号 链上交易hash\n\n用于对已过期但链上实际已到账的订单手动补单。\n补单成功后请再执行 /notify 交易号 手动发送回调。")
	}
	order, err := service.ManualCompleteOrder(args[0], args[1])
	if err != nil {
		return c.Send("❌ 补单失败：" + err.Error())
	}
	return c.Send(fmt.Sprintf(
		"✅ 补单成功（订单已标记为支付成功）\n交易号：%s\n订单号：%s\n实际金额：%v USDT\n链上hash：%s\n\n如需通知商户，请执行：\n/notify %s",
		order.TradeId, order.OrderId, order.ActualAmount, order.BlockTransactionId, order.TradeId))
}

// OrderNotifyHandle 手动重发回调通知：/notify <交易号>
func OrderNotifyHandle(c tb.Context) error {
	args := c.Args()
	if len(args) < 1 {
		return c.Send("用法：/notify 交易号\n\n用于补单后或回调失败后，手动向商户重发回调通知。")
	}
	order, err := service.ResendOrderCallback(args[0])
	if err != nil {
		return c.Send("❌ 回调发送失败：" + err.Error())
	}
	return c.Send(fmt.Sprintf(
		"✅ 已投递回调通知（异步发送，结果以商户接收为准）\n交易号：%s\n回调地址：%s",
		order.TradeId, order.NotifyUrl))
}

func WalletList(c tb.Context) error {
	wallets, err := data.GetAllWalletAddress()
	if err != nil {
		return err
	}
	var btnList [][]tb.InlineButton
	for _, wallet := range wallets {
		status := "已启用✅"
		if wallet.Status == mdb.TokenStatusDisable {
			status = "已禁用🚫"
		}
		var temp []tb.InlineButton
		btnInfo := tb.InlineButton{
			Unique: wallet.Token,
			Text:   fmt.Sprintf("%s[%s]", wallet.Token, status),
			Data:   strutil.MustString(wallet.ID),
		}
		bots.Handle(&btnInfo, WalletInfo)
		btnList = append(btnList, append(temp, btnInfo))
	}
	addBtn := tb.InlineButton{Text: "添加钱包地址", Unique: "AddWallet"}
	bots.Handle(&addBtn, func(c tb.Context) error {
		return c.Send(ReplayAddWallet, &tb.ReplyMarkup{
			ForceReply: true,
		})
	})
	btnList = append(btnList, []tb.InlineButton{addBtn})
	return c.EditOrSend("请点击钱包继续操作", &tb.ReplyMarkup{
		InlineKeyboard: btnList,
	})
}

func WalletInfo(c tb.Context) error {
	id := mathutil.MustUint(c.Data())
	tokenInfo, err := data.GetWalletAddressById(id)
	if err != nil {
		return c.Send(err.Error())
	}
	enableBtn := tb.InlineButton{
		Text:   "启用",
		Unique: "enableBtn",
		Data:   c.Data(),
	}
	disableBtn := tb.InlineButton{
		Text:   "禁用",
		Unique: "disableBtn",
		Data:   c.Data(),
	}
	delBtn := tb.InlineButton{
		Text:   "删除",
		Unique: "delBtn",
		Data:   c.Data(),
	}
	backBtn := tb.InlineButton{
		Text:   "返回",
		Unique: "WalletList",
	}
	bots.Handle(&enableBtn, EnableWallet)
	bots.Handle(&disableBtn, DisableWallet)
	bots.Handle(&delBtn, DelWallet)
	bots.Handle(&backBtn, WalletList)
	return c.EditOrReply(tokenInfo.Token, &tb.ReplyMarkup{InlineKeyboard: [][]tb.InlineButton{
		{
			enableBtn,
			disableBtn,
			delBtn,
		},
		{
			backBtn,
		},
	}})
}

func EnableWallet(c tb.Context) error {
	id := mathutil.MustUint(c.Data())
	if id <= 0 {
		return c.Send("请求不合法！")
	}
	err := data.ChangeWalletAddressStatus(id, mdb.TokenStatusEnable)
	if err != nil {
		return c.Send(err.Error())
	}
	return WalletList(c)
}

func DisableWallet(c tb.Context) error {
	id := mathutil.MustUint(c.Data())
	if id <= 0 {
		return c.Send("请求不合法！")
	}
	err := data.ChangeWalletAddressStatus(id, mdb.TokenStatusDisable)
	if err != nil {
		return c.Send(err.Error())
	}
	return WalletList(c)
}

func DelWallet(c tb.Context) error {
	id := mathutil.MustUint(c.Data())
	if id <= 0 {
		return c.Send("请求不合法！")
	}
	err := data.DeleteWalletAddressById(id)
	if err != nil {
		return c.Send(err.Error())
	}
	return WalletList(c)
}
