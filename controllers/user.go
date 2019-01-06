package controllers

import (
	"DayDayFresh/models"
	"fmt"
	"github.com/KenmyZhang/aliyun-communicate"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils"
	"github.com/gomodule/redigo/redis"
	"github.com/smartwalle/alipay"
	"regexp"
	"strconv"
)

type UserContreller struct {
	beego.Controller
}

func (this *UserContreller) ShowRegister() {
	this.TplName = "register.html"
}

func (this *UserContreller) HandleReg() {

	userName := this.GetString("user_name")
	pwd := this.GetString("pwd")
	cpwd := this.GetString("cpwd")
	email := this.GetString("email")

	if userName == "" || pwd == "" || cpwd == "" || email == "" {
		this.Data["errmsg"] = "数据不能为空"
		this.TplName = "register.html"
		return
	}
	//邮箱格式校验
	reg, _ := regexp.Compile("^[A-Za-z0-9\u4e00-\u9fa5]+@[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)+$")
	res := reg.FindString(email)
	if res == "" {
		this.Data["errmsg"] = "邮箱格式不正确，请重新填写"
		this.TplName = "register.html"
		return

	}
	//密码校验
	if pwd != cpwd {
		this.Data["errmsg"] = "两次密码输入不正确，请重新填写"
		this.TplName = "register.html"
		return
	}
	//处理数据
	o := orm.NewOrm()

	var user models.User

	user.Name = userName
	user.PassWord = pwd
	user.Email = email

	_, err := o.Insert(&user)
	if err != nil {
		this.Data["errmsg"] = "注册失败，用户名重名，请重新填写"
		this.TplName = "register.html"
		return
	}

	//邮箱激活

	enailConfig := `{"username":"422414336@qq.com","password":"lmtcuqiofugocafc","host":"smtp.qq.com","port":587}`
	ems := utils.NewEMail(enailConfig)
	ems.From = "422414336@qq.com"
	ems.To = []string{email}
	ems.Subject = "天天生鲜用户激活"
	//ems.Text = "复制该连接到浏览器中激活：127.0.0.1:8090/active?id=" + strconv.Itoa(user.Id)
	ems.HTML = "<a href=\"http://172.17.237.103:8090/active?id=" + strconv.Itoa(user.Id) + "\">点击该链接，天天账号激活</a>"

	err = ems.Send()
	if err != nil {
		this.Data["errmsg"] = "激活邮箱发送失败"
		this.TplName = "register.html"
		return
	}

	this.Ctx.WriteString("注册成功,请去邮箱激活账号")

}

func (this *UserContreller) HandleActive() {

	id, err := this.GetInt("id")
	if err != nil {
		this.Data["errmsg"] = "激活失败，请重新注册"
		this.TplName = "register.html"
		return
	}

	o := orm.NewOrm()
	var user models.User

	user.Id = id
	err = o.Read(&user)
	if err != nil {
		this.Data["errmsg"] = "激活失败，请重新注册"
		this.TplName = "register.html"
		return
	}

	user.Active = true
	_, err = o.Update(&user)
	if err != nil {
		this.Data["errmsg"] = "插入数据库失败，请重试"
		this.TplName = "register.html"
		return
	}
	this.Redirect("/login", 302)

}

//展示登陆界面
func (this *UserContreller) ShowLogin() {

	usreName := this.Ctx.GetCookie("username")
	if usreName == "" {
		this.Data["userName"] = ""
		this.Data["checked"] = ""

	} else {
		this.Data["usreName"] = usreName
		this.Data["checked"] = "checked"
	}

	this.TplName = "login.html"

}

func (this *UserContreller) HandleLogin() {

	//获取数据
	userName := this.GetString("username")
	pwd := this.GetString("pwd")

	if userName == "" || pwd == "" {
		this.Data["errmsg"] = "用户名和密码不能为空"
		this.TplName = "login.html"
		return
	}

	o := orm.NewOrm()
	var user models.User

	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		this.Data["errmsg"] = "用户名不存在"
		this.TplName = "login.html"
		return
	}
	if user.PassWord != pwd {
		this.Data["errmsg"] = "密码错误"
		this.TplName = "login.html"
		return
	}
	if !user.Active {
		this.Data["errmsg"] = "账号未激活"
		this.TplName = "login.html"
		return
	}

	//记住用户名
	check := this.GetString("check")
	if check == "on" {
		this.Ctx.SetCookie("userName", userName, 3600)
	} else {
		this.Ctx.SetCookie("userName", userName, -1)
	}

	this.SetSession("userName", userName)
	//this.Layout="layout.html"
	//this.Layout="index.html"
	this.Redirect("/", 302)

}

func (this *UserContreller) Logout() {
	this.DelSession("userName")
	this.Redirect("/login", 302)
}

func ShowLayout(this *UserContreller) {

	userName := this.GetSession("userName")

	this.Data["userName"] = userName.(string)

	this.Layout = "usercenterlayout.html"

}

func (this *UserContreller) ShowCenterInfo() {

	userName := this.GetSession("userName")

	//user := userName.(string)

	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")

	var addr models.Address
	o.QueryTable("Address").
		RelatedSel("User").Filter("User__Id", user.Id).
		Filter("Isdefault", true).One(&addr)

	this.Data["addr"] = addr.Addr
	this.Data["Phone"] = addr.Phone

	this.Data["userName"] = userName.(string)

	//获取历史浏览记录
	conn, err := redis.Dial("tcp", "172.17.237.103:6379")
	if err != nil {
		beego.Error("redis链接失败", err)
	}
	defer conn.Close()

	resp, err := conn.Do("lrange", "history_"+strconv.Itoa(user.Id), 0, 4)
	//回复助手函数
	goodsId, err := redis.Ints(resp, err)
	if err != nil {
		beego.Error("redis获取商品错误", err)
	}
	//beego.Info(goodsId)
	var goodsSku []models.GoodsSKU
	for _, id := range goodsId {
		var goods models.GoodsSKU
		goods.Id = id
		o.Read(&goods)
		goodsSku = append(goodsSku, goods)
	}

	this.Data["goodsSkus"] = goodsSku

	ShowLayout(this)
	this.TplName = "user_center_info.html"
}

func (this *UserContreller) ShowCenterOrder() {

	//获取数据
	o := orm.NewOrm()
	var order = make([]map[string]interface{}, 0)
	//获取订单信息
	var user models.User
	userName := this.GetSession("userName")
	user.Name = userName.(string)
	o.Read(&user, "Name")

	var orderInfos []models.OrderInfo
	o.QueryTable("OrderInfo").
		RelatedSel("User").Filter("User", user).All(&orderInfos)

	//获取订单商品信息
	for _, values := range orderInfos {
		temp := make(map[string]interface{})
		var orderGoods []models.OrderGoods
		o.QueryTable("OrderGoods").
			RelatedSel("OrderInfo", "GoodsSKU").Filter("OrderInfo", values).
			All(&orderGoods)

		temp["goods"] = orderGoods
		temp["order"] = values

		order = append(order, temp)
	}

	this.Data["orders"] = order

	ShowLayout(this)

	this.TplName = "user_center_order.html"

}

func (this *UserContreller) ShowCenterSite() {

	userName := this.GetSession("userName")

	o := orm.NewOrm()

	var addr models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name", userName.(string)).Filter("Isdefault", true).One(&addr)

	this.Data["address"] = addr

	ShowLayout(this)
	this.TplName = "user_center_site.html"

}

func (this *UserContreller) HandleCenterSite() {

	recever := this.GetString("recever")
	addr := this.GetString("addr")
	zipCode := this.GetString("zipCode")
	phone := this.GetString("phone")

	if recever == "" || addr == "" || zipCode == "" || phone == "" {
		beego.Error("添加地址页面，获取数据失败")
		this.Redirect("/goods/UserCenterSite", 302)
		return
	}

	o := orm.NewOrm()

	var address models.Address

	address.Receiver = recever
	address.Zipcode = zipCode
	address.Addr = addr
	address.Phone = phone

	//一对多的插入
	var user models.User
	userName := this.GetSession("userName")
	user.Name = userName.(string)
	o.Read(&user, "Name")

	address.User = &user
	//判断当前用户是否有默认地址，如果没有,则直接插入默认地址，
	// 如果有默认地址，把默认地址更新为非默认地址，把新插入的地址设置为默认地址
	//获取当前用户的默认地址
	var oldAddress models.Address
	err := o.QueryTable("Address").
		RelatedSel("User").Filter("User__Id", user.Id).
		Filter("Isdefault", true).One(&oldAddress)
	if err != nil {
		address.Isdefault = true
	} else {
		oldAddress.Isdefault = false

		o.Update(&oldAddress)

		address.Isdefault = true
	}

	o.Insert(&address)

	this.Redirect("/goods/UserCenterSite", 302)

}

//支付宝 支付功能
func (this *UserContreller) PayAli() {

	var privateKey = "MIIEpAIBAAKCAQEA4pZCfQRnX3TKT46SRI7Mfi8skpNDj6Dpcv0gnoH3RxdMmouT" +
		"STaPPPEhUwcfbXKhVtF5qDSXsIdByQ5tZtEuZJ+1wJKCO0CCx1/cqp3apBc81GSs" +
		"ycBfSw4bIBHSLepLeBXrJFmgDrKnsoFza1bghhaVhKD6a3Ie5/cpI35/Wic72rYP" +
		"SQPBdy92VDXwNDuQMwIi+d+7/QY2QGhTAZN0cg3qud8w9g2AI6NLRmWuq6FBRhXr" +
		"dXiHw7ci7x+2U+MLHMhdH1MI04DLCO+yFsiOrozo3XOkxxxA1F6ch9vDEzcoxW4n" +
		"+qzlrtzsOS2Hk2SW5q6Gxnpe0HY1lwY9qdeckwIDAQABAoIBAGb9HSNtyP6eOwaG" +
		"Kv12WoRQNNY6kU7LONDHNPhW4moxsOPd5Qg2AE0W3Kq8ZhB9NdAcTkuh/ACEueYE" +
		"5L0C/y9FWHs7HG6KF+c/LzFtpl9HIKL5T4A0LBwVQUcGUp4EDGF8tPBEvHdxxL9i" +
		"D3AOgObxhOxPrwL/UATnVo+Hg6MZrubZN2FbwFaUU5VF0cnmWaYRGGyu0dsUEvse" +
		"LNuKybfA3wJuM32Kvp3A0EyVy17jcA9ImcmoncVqaOdLx+YVSEsYBmr8MXOcWkBt" +
		"XRBr1pemOfP0b24/D1NXJNqfjprnfAeVX656AS+ayE5JgQMOjEMAY6EYJdt+tgPm" +
		"q0j5oOECgYEA/UXfWMTe8gjwHjRSwhESRwM/FnTOIpAOtv+8c3zNDatS7Gb5wuxP" +
		"x8cdcbToa46uHZveZ0wVokpnStCawnzeW4DIwl9KgYFoQiGlILeud/qfQMb1e8De" +
		"J9dqKPbsb5wDV058g6Wai/g8sAcfaNysnyLHomJ54xX0edsQ9LE6Uq8CgYEA5QbU" +
		"UmPw0ZxXHTCd2fPKIuNJGEuquUawfg5wDwMKw9hpdq5cbIcXlA5JkhaKtCLoN3oE" +
		"AuwBCDX3nWPQ5YX7ukwSJ5khnphYbJrpGS0/otMrWSQnbynM86tTDaAXXSN6b6wT" +
		"2kUbSooksAjJwGozazjGjky7J8jn0V5qGkHCXV0CgYEArY7qJLyUQqvZT/lvFMn6" +
		"CmuxGcRlVc3+J21MSJ+nLMzQgGt4kBi7+xz5kmf0NXCK5INhfsvmr1XpPp2Az/Id" +
		"tfqkmH4QYnq5ZUgFDkyQ5Gr8Irm0k19xXUAC4ZuEHl988qE4NkaPh4dOnxnibkt6" +
		"h3qf7ykoeXMcGz0Be4zPeMUCgYEAivwA/1rM+rcwmnM1Z92tLkzVv9uzaCpA0s66" +
		"LDIBZ2Y+Yhpf1jCJG30sIm5xj+2bFIeERa2o1q3BbY70Z0VOxPiDD+q63z6+cnHz" +
		"wSaXdp1FshvhnnE0gi7XAO7FHu130KsRhSTo8ewxZW5/2LfaKlhTDmn8LaGbJJBy" +
		"PSro47UCgYBFvXsLzPnJj5n5KJirL5SjN0Kr2shXjZEwjrJ4pUAsAM+FRmuTNboq" +
		"2K9q9zyAh/tTn8B6H2zy2gcNwJqDd0nSYM+2S3PbU3ml1mvcf0Q5vOdzwTyPuGOf" +
		"AALYrypVtiEtyXYkk8G98WjgB0q1bjTh344ZQ4gfa0PbbNLAYeDGyA=="

	var appId = "2016092200569649"
	var aliPublicKey = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtPLfffeuLcVVBAZmiQuA7BtFGv7GKG6mWP7P+r9/koOTsICX6PObhGZwSR1BYtJhgcdimRI3UBBxyR3P4Ay7egpcconLuyxqZYNfohfVRL48MfIyS7cHDdNkjz2r70gOLfjYwchM6ttkzftME0k4QLJf/Y+qbSCiWvZ+9YRFmHo9Iq8juKDbnYkYmhoq7LDUxwVh7k9JeYW20kTIJecfNutCWGOcAC01jFymbNglrne8cUWet+qgY2WhGwEK1+2r1lWu+0azsNPPF3i3vVPAH1F2yxz6njhU26zO7A6+sB5Ff4DiULh3UAH9yID6LKJNBVJTpKobwidhFqk3ip5UqQIDAQAB"

	var client = alipay.New(appId, aliPublicKey, privateKey, false)

	//alipay.trade.page.pay
	var p = alipay.AliPayTradePagePay{}
	p.NotifyURL = "http://172.17.237.103:8090/user/payOk"
	p.ReturnURL = "http://172.17.237.103:8090/user/payOk"
	p.Subject = "天天生鲜"
	p.OutTradeNo = "987654321"
	p.TotalAmount = "1000.00"
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	var url, err = client.TradePagePay(p)
	if err != nil {
		fmt.Println(err)
	}

	var payURL = url.String()

	this.Redirect(payURL, 302)

}

//实现短信发送业务
func (this *UserContreller) SMS() {
	var (
		gatewayUrl      = "http://dysmsapi.aliyuncs.com/"
		accessKeyId     = "LTAIQ9aVPA8IEwCg"
		accessKeySecret = "EFwkulaxYhp4gFDP9IY4rvUVvf8NE0"
		phoneNumbers    = "13774662182"                //要发送的电话号码
		signName        = "天天生鲜"                       //签名名称
		templateCode    = "SMS_149101793"              //模板号
		templateParam   = "{\"code\":\"DayDayFresh\"}" //验证码
	)

	smsClient := aliyunsmsclient.New(gatewayUrl)
	result, err := smsClient.Execute(accessKeyId, accessKeySecret, phoneNumbers, signName, templateCode, templateParam)
	fmt.Println("Got raw response from server:", string(result.RawResponse))
	if err != nil {
		beego.Info("配置有问题")
	}

	if result.IsSuccessful() {
		this.Data["result"] = "短信已经发送"
	} else {
		this.Data["result"] = "短信发送失败"
	}
	//this.TplName = "SMS.html"

}
