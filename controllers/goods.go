package controllers

import (
	"DayDayFresh/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"math"
	"strconv"
)

type GoodsController struct {
	beego.Controller
}

//展示首页数据
func (this *GoodsController) ShowIndex() {
	//登陆的时候显示欢迎你。。。。
	userName := this.GetSession("userName")
	if userName == nil {
		this.Data["userName"] = ""
	} else {
		this.Data["userName"] = userName.(string)

	}

	//第一模块展示

	o := orm.NewOrm()
	var goodstypes []models.GoodsType
	//查询对象
	o.QueryTable("GoodsType").All(&goodstypes)
	//数据传递
	this.Data["goodstypes"] = goodstypes

	var indexGoodsBanners []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&indexGoodsBanners)
	this.Data["indexGoodsBanners"] = indexGoodsBanners

	var indexPromotionBanners []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&indexPromotionBanners)
	this.Data["indexPromotionBanners"] = indexPromotionBanners

	//第二模块展示

	var goodsSkus = make([]map[string]interface{}, len(goodstypes))

	//把类型对象放入map容器中
	for index, _ := range goodsSkus {
		temp := make(map[string]interface{})
		temp["types"] = goodstypes[index]
		goodsSkus[index] = temp
	}

	//存商品数据
	for _, goodsMap := range goodsSkus {
		var goodsImage []models.IndexTypeGoodsBanner
		var goodsText []models.IndexTypeGoodsBanner
		o.QueryTable("IndexTypeGoodsBanner").
			RelatedSel("GoodsType", "GoodsSku").
			Filter("GoodsType", goodsMap["types"]).
			Filter("DisplayType", 0).All(&goodsText)

		o.QueryTable("IndexTypeGoodsBanner").
			RelatedSel("GoodsType", "GoodsSku").
			Filter("GoodsType", goodsMap["types"]).
			Filter("DisplayType", 1).All(&goodsImage)

		goodsMap["goodsImage"] = goodsImage
		goodsMap["goodsText"] = goodsText

	}
	this.Data["goodsSkus"] = goodsSkus

	this.TplName = "index.html"
	//beego.Info(goodsSkus)
}

func ShowGoodsListAndDetaillayout(this *GoodsController, typeId int) {

	o := orm.NewOrm()
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)
	this.Data["goodsTypes"] = goodsTypes
	//获取新品数据
	//获取同一类型的新品数据
	var newGoods []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", typeId).OrderBy("Time").Limit(2, 0).All(&newGoods)
	this.Data["newGoods"] = newGoods
	this.Layout = "GoodsListAndDetaillayout.html"

}

func (this *GoodsController) ShowGoodsDetail() {

	//获取数据
	id, err := this.GetInt("id")
	if err != nil {
		beego.Error("请求路径错误", err)
	}

	//查询操作
	o := orm.NewOrm()
	var goodsSku models.GoodsSKU

	goodsSku.Id = id

	err = o.QueryTable("GoodsSKU").
		RelatedSel("Goods", "GoodsType").
		Filter("Id", id).One(&goodsSku)

	if err != nil {
		beego.Error("查询商品失败")
	}

	//添加历史浏览记录
	//1.判断是否登陆
	userName := this.GetSession("userName")
	if userName != nil {

		//查询用户信息
		var user models.User
		user.Name = userName.(string)
		o.Read(&user, "Name")

		//获取redis操作对象
		conn, err := redis.Dial("tcp", "172.17.237.103:6379")
		defer conn.Close()
		if err != nil {
			beego.Error("redis连接失败", err)
		}
		conn.Do("lrem", "history_"+strconv.Itoa(user.Id), 0, id)
		conn.Do("lpush", "history_"+strconv.Itoa(user.Id), id)
	}

	//返回数据
	ShowGoodsListAndDetaillayout(this, goodsSku.GoodsType.Id)

	this.Data["goodsSku"] = goodsSku

	this.TplName = "detail.html"

}

func PageEdior(pageCount float64, pageIndex int) []int {
	//2.判断页码位置
	var pages []int

	if pageCount <= 5 {
		pages = make([]int, int(pageCount))

		for i := 1; pageCount > 0; i++ {
			pages[i-1] = i
			pageCount -= 1
		}
	} else if pageIndex <= 3 {
		pages = make([]int, 5)
		var temp = 5
		for i := 1; temp > 0; i++ {
			pages[i-1] = i
			temp -= 1
		}
	} else if pageIndex >= int(pageCount)-2 {

		pages = make([]int, 5)
		temp := 5
		for i := 1; temp > 0; i++ {
			pages[i-1] = int(pageCount) - temp + 1
			temp -= 1
		}
	} else {
		pages = make([]int, 5)
		temp := 2
		for i := 1; temp > -3; i++ {
			pages[i-1] = pageIndex - temp
			temp -= 1
		}
		beego.Info(pages)
	}
	return pages
}

//展示商品列表页
func (this *GoodsController) ShowGoodsList() {

	typeId, err := this.GetInt("id")
	if err != nil {
		beego.Error("获取商品类型失败", err)
	}

	o := orm.NewOrm()
	var goodsSkus []models.GoodsSKU

	sort := this.GetString("sort")
	this.Data["sort"] = sort

	//实现商品分页
	//1.获取总页数
	count, _ := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").
		Filter("GoodsType__Id", typeId).Count()
	pageSize := 1
	pageCount := math.Ceil(float64(count) / float64(pageSize))

	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}

	pages := PageEdior(pageCount, pageIndex)

	this.Data["pages"] = pages
	start := (pageIndex - 1) * pageSize

	if sort == "price" {
		o.QueryTable("GoodsSKU").
			RelatedSel("GoodsType").
			Filter("GoodsType__Id", typeId).
			OrderBy("Price").Limit(pageSize, start).All(&goodsSkus)
	} else if sort == "sale" {
		o.QueryTable("GoodsSKU").
			RelatedSel("GoodsType").
			Filter("GoodsType__Id", typeId).
			OrderBy("Sales").Limit(pageSize, start).All(&goodsSkus)
	} else {
		o.QueryTable("GoodsSKU").
			RelatedSel("GoodsType").
			Filter("GoodsType__Id", typeId).Limit(pageSize, start).All(&goodsSkus)
	}

	//头尾页码处理
	prePage := pageIndex - 1
	if prePage < 1 {
		prePage = 1
	}
	nextPage := pageIndex + 1
	if nextPage > int(pageCount) {
		nextPage = int(pageCount)
	}

	this.Data["goodsSkus"] = goodsSkus

	this.Data["prePage"] = prePage

	this.Data["nextPage"] = nextPage

	this.Data["pageIndex"] = pageIndex

	this.Data["typeId"] = typeId

	ShowGoodsListAndDetaillayout(this, typeId)

	this.TplName = "list.html"

}

func (this *GoodsController) HandleSearch() {

	search := this.GetString("searchName")

	o := orm.NewOrm()
	var goodsSkus []models.GoodsSKU

	if search == "" {
		o.QueryTable("GoodsSKU").All(&goodsSkus)
	} else {
		o.QueryTable("GoodsSKU").
			Filter("Name__icontains", search).All(&goodsSkus)
	}

	this.Data["search"] = goodsSkus
	this.Layout = "GoodsListAndDetaillayout.html"
	this.TplName = "search.html"

}
