package controllers

import (
	"DayDayFresh/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"strconv"
)

type CartController struct {
	beego.Controller
}

func (this *CartController) HandleAddCart() {
	goodsId, err := this.GetInt("goodsId")
	count, err1 := this.GetInt("count")

	resp := make(map[string]interface{})
	defer this.ServeJSON()

	if err != nil || err1 != nil {
		resp["code"] = 1
		resp["errmsg"] = "ajax数据传输错误"
		this.Data["json"] = resp
		return
	}

	//将数据存到redis中
	userName := this.GetSession("userName")
	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")

	conn, err := redis.Dial("tcp", "172.17.237.103:6379")
	if err != nil {
		resp["code"] = 2
		resp["errmsg"] = "redis数据库连接失败"
		this.Data["json"] = resp
		return
	}

	//获取原来redis中的数据

	res, err := conn.Do("hget", "cart_"+strconv.Itoa(user.Id), goodsId)

	preCount, _ := redis.Int(res, err)

	conn.Do("hset", "cart_"+strconv.Itoa(user.Id), goodsId, preCount+count)

	re, err := conn.Do("hlen", "cart_"+strconv.Itoa(user.Id))
	cartCount, _ := redis.Int(re, err)

	resp["code"] = 5

	resp["count"] = cartCount

	this.Data["json"] = resp

	this.ServeJSON()

}

func (this *CartController) ShowCart() {

	//从redis中获取相关数据
	conn, err := redis.Dial("tcp", "172.17.237.103:6379")
	if err != nil {
		beego.Error("redis连接失败", err)
	}

	userName := this.GetSession("userName")

	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")

	resp, err := conn.Do("hgetall", "cart_"+strconv.Itoa(user.Id))
	goodsMap, _ := redis.IntMap(resp, err)

	var goods = make([]map[string]interface{}, 0)
	for goodsId, count := range goodsMap {

		temp := make(map[string]interface{})
		id, _ := strconv.Atoi(goodsId)

		var goodsSku models.GoodsSKU
		goodsSku.Id = id
		o.Read(&goodsSku)

		temp["goodsSku"] = goodsSku
		temp["count"] = count

		temp["sumPrice"] = count * goodsSku.Price
		goods = append(goods, temp)
	}

	this.Data["goods"] = goods

	this.TplName = "cart.html"

}

//更新购物车数量

func (this *CartController) UpdateCart() {

	goodsId, err1 := this.GetInt("goodsId")
	count, err2 := this.GetInt("count")
	resp := make(map[string]interface{})
	defer this.ServeJSON()

	if err1 != nil || err2 != nil {
		resp["code"] = 1
		resp["errmsg"] = "ajax数据传输错误"
		this.Data["json"] = resp
		return
	}
	//beego.Info(goodsId,count)
	//数据存储在redis的hash中
	//1.获取用户Id
	userName := this.GetSession("userName")
	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")
	//2.向数据库中存储数据
	conn, err := redis.Dial("tcp", "172.17.237.103:6379")
	if err != nil {
		resp["code"] = 2
		resp["errmsg"] = "redis数据库链接错误"
		this.Data["json"] = resp
		return
	}
	//先获取一下原来的数量

	conn.Do("hset", "cart_"+strconv.Itoa(user.Id), goodsId, count)

	re, err := conn.Do("hlen", "cart_"+strconv.Itoa(user.Id))
	cartCount, _ := redis.Int(re, err)
	resp["code"] = 5
	resp["count"] = cartCount

	this.Data["json"] = resp

	this.ServeJSON()

}

//删除购物车数量
func (this *CartController) DeleteCart() {

	goodsId, err := this.GetInt("goodsId")
	userName := this.GetSession("userName")

	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")

	resp := make(map[string]interface{})
	defer this.ServeJSON()

	if err != nil {
		resp["code"] = 1
		resp["errmsg"] = "删除失败"
		this.Data["json"] = resp
		return
	}

	conn, err := redis.Dial("tcp", "172.17.237.103:6379")

	if err != nil {
		resp["code"] = 2
		resp["errmsg"] = "连接redis失败"
		this.Data["json"] = resp
		return
	}

	conn.Do("hdel", "cart_"+strconv.Itoa(user.Id), goodsId)
	resp["code"] = 5
	resp["errmsg"] = "Delete success!!"
	this.Data["json"] = resp

}
