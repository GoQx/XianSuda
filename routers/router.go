package routers

import (
	"DayDayFresh/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	beego.InsertFilter("/goods/*", beego.BeforeExec, fiterFunc)

	beego.Router("/",
		&controllers.GoodsController{},
		"get:ShowIndex")

	beego.Router("/register",
		&controllers.UserContreller{},
		"get:ShowRegister;post:HandleReg")

	beego.Router("/active",
		&controllers.UserContreller{},
		"get:HandleActive")

	beego.Router("/login",
		&controllers.UserContreller{},
		"get:ShowLogin;post:HandleLogin")

	beego.Router("/goods/logout",
		&controllers.UserContreller{},
		"get:Logout")

	beego.Router("/goods/UserCenterInfo",
		&controllers.UserContreller{},
		"get:ShowCenterInfo")

	beego.Router("/goods/UserCenterOrder",
		&controllers.UserContreller{},
		"get:ShowCenterOrder")

	beego.Router("/goods/UserCenterSite",
		&controllers.UserContreller{},
		"get:ShowCenterSite;post:HandleCenterSite")

	beego.Router("/goodsDetail",
		&controllers.GoodsController{},
		"get:ShowGoodsDetail")

	beego.Router("/goodsList",
		&controllers.GoodsController{},
		"get:ShowGoodsList")

	beego.Router("/searchGoods",
		&controllers.GoodsController{},
		"post:HandleSearch")

	beego.Router("/goods/Cart",
		&controllers.CartController{},
		"get:ShowCart")

	beego.Router("/goods/addCart",
		&controllers.CartController{},
		"post:HandleAddCart")

	beego.Router("/goods/addCart",
		&controllers.CartController{},
		"post:UpdateCart")

	beego.Router("/goods/deleteCart",
		&controllers.CartController{},
		"post:DeleteCart")

	beego.Router("/goods/showOrder",
		&controllers.OrderController{},
		"post:ShowOrder")

	beego.Router("/goods/orderInfo",
		&controllers.OrderController{},
		"post:HandleOrderInfo")

	beego.Router("/goods/PayAli",
		&controllers.UserContreller{},
		"get:PayAli")

	beego.Router("/goods/sms",
		&controllers.UserContreller{},
		"get:SMS")

}

var fiterFunc = func(ctx *context.Context) {

	userName := ctx.Input.Session("userName")
	if userName == nil {
		ctx.Redirect(302, "/login")
	}
}
