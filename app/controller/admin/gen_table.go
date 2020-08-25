package admin

import (
	"gfast/app/model/admin/gen_table"
	"gfast/app/service/admin/gen_service"
	"gfast/app/service/admin/user_service"
	"gfast/library/response"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gview"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"github.com/gogf/gf/util/gvalid"
	"strconv"
	"strings"
)

type Gen struct{}

// @Summary 查询数据库列表
// @Description 查询数据库列表
// @Tags 系统工具
// @Param data body gen_table.SelectPageReq true "data"
// @Success 0 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /system/tools/gen/dataList [get]
// @Security
func (c *Gen) DataList(r *ghttp.Request) {
	var req *gen_table.SelectPageReq
	//获取参数
	if err := r.Parse(&req); err != nil {
		response.FailJson(true, r, err.(*gvalid.Error).FirstString())
	}

	total, list, err := gen_service.SelectDbTableList(req)
	if err != nil {
		response.FailJson(true, r, err.Error())
	}
	response.SusJson(true, r, "ok", g.Map{
		"total": total,
		"list":  list,
	})
}

// @Summary 表列表
// @Description 表列表
// @Tags 系统工具
// @Param data body gen_table.SelectPageReq true "data"
// @Success 0 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /system/tools/gen/tableList [post]
// @Security
func (c *Gen) TableList(r *ghttp.Request) {
	var req *gen_table.SelectPageReq
	//获取参数
	if err := r.Parse(&req); err != nil {
		response.FailJson(true, r, err.(*gvalid.Error).FirstString())
	}
	total, list, err := gen_service.SelectListByPage(req)
	if err != nil {
		response.FailJson(true, r, err.Error())
	}
	response.SusJson(true, r, "ok", g.Map{
		"total": total,
		"list":  list,
	})
}

// @Summary 导入表结构操作
// @Description 导入表结构操作
// @Tags 系统工具
// @Param tables body string  true "tables"
// @Success 0 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /system/tools/gen/importTableSave [post]
// @Security
func (c *Gen) ImportTableSave(r *ghttp.Request) {
	tables := r.GetString("tables")
	if tables == "" {
		response.FailJson(true, r, "请选择要导入的表格")
	}
	user := user_service.GetLoginAdminInfo(r)
	operName := user.UserName
	tableArr := strings.Split(tables, ",")
	tableList, err := gen_service.SelectDbTableListByNames(tableArr)
	if err != nil {
		response.FailJson(true, r, err.Error())
	}
	if tableList == nil {
		response.FailJson(true, r, "表信息不存在")
	}
	err = gen_service.ImportGenTable(tableList, operName)
	if err != nil {
		response.FailJson(true, r, err.Error())
	}
	response.SusJson(true, r, "导入数据表成功")
}

// @Summary 根据表格ID获取表格字段列表数据
// @Description 根据表格ID获取表格字段列表数据
// @Tags 系统工具
// @Param tableId body integer   true "tableId"
// @Success 0 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /system/tools/gen/columnList [post]
// @Security
func (c *Gen) ColumnList(r *ghttp.Request) {
	tableId := r.GetInt64("tableId")
	if tableId == 0 {
		response.FailJson(true, r, "参数错误")
	}
	list, err := gen_service.SelectGenTableColumnListByTableId(tableId)
	if err != nil {
		response.FailJson(true, r, err.Error())
	}
	var tableInfo *gen_table.Entity
	var tableMap g.Map
	tableInfo, err = gen_service.GetTableInfoByTableId(tableId)
	tableMap = gconv.Map(tableInfo)
	//如果是树表则设置树表配置
	if tableInfo != nil && tableInfo.TplCategory == "tree" {
		options := gjson.New(tableInfo.Options)
		tableMap["tree_code"] = options.Get("tree_code")
		tableMap["tree_parent_code"] = options.Get("tree_parent_code")
		tableMap["tree_name"] = options.Get("tree_name")
	}

	res := g.Map{
		"rows": list,
		"info": tableMap,
	}
	response.SusJson(true, r, "ok", res)
}

// @Summary 编辑表格信息
// @Description 编辑表格信息
// @Tags 系统工具
// @Param data body gen_table.EditReq  true "data"
// @Success 0 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /system/tools/gen/editSave [post]
// @Security
func (c *Gen) EditSave(r *ghttp.Request) {
	var req *gen_table.EditReq
	//获取参数
	if err := r.Parse(&req); err != nil {
		response.FailJson(true, r, err.(*gvalid.Error).FirstString())
	}
	userInfo := user_service.GetLoginAdminInfo(r)
	req.UserName = userInfo.UserName
	err := gen_service.SaveEdit(req)
	if err != nil {
		response.FailJson(true, r, err.Error())
	}
	response.SusJson(true, r, "设置成功")
}

// @Summary 删除表格数据
// @Description 删除表格数据
// @Tags 系统工具
// @Param ids body integer   true "ids[1,2,3...]"
// @Success 0 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /system/tools/gen/delete [delete]
// @Security
func (c *Gen) Delete(r *ghttp.Request) {
	ids := r.GetInts("ids")
	if len(ids) == 0 {
		response.FailJson(true, r, "参数错误")
	}
	err := gen_service.Delete(ids)
	if err != nil {
		response.FailJson(true, r, err.Error())
	}
	response.SusJson(true, r, "删除成功")
}

// @Summary 代码生成预览
// @Description 代码生成预览
// @Tags 系统工具
// @Param tableId body integer   true "tableId"
// @Success 0 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /system/tools/gen/preview [post]
// @Security
func (c *Gen) Preview(r *ghttp.Request) {
	tableId := r.GetInt64("tableId")
	if tableId == 0 {
		response.FailJson(true, r, "参数错误")
	}
	entity, err := gen_service.SelectRecordById(tableId)
	if err != nil {
		response.FailJson(true, r, err.Error())
	}
	if entity == nil {
		response.FailJson(true, r, "表格数据不存在")
	}
	gen_service.SetPkColumn(entity, entity.Columns)

	controllerKey := "vm/go/" + entity.BusinessName + "_controller.go.vm"
	controllerValue := ""
	serviceKey := "vm/go/" + entity.BusinessName + "_service.go.vm"
	serviceValue := ""
	modelKey := "vm/go/" + entity.BusinessName + "_model.go.vm"
	modelValue := ""
	apiJsKey := "vm/html/" + entity.BusinessName + "_api.js.vm"
	apiJsValue := ""
	vueKey := "vm/html/" + entity.BusinessName + "_vue.js.vm"
	vueValue := ""

	view := gview.New()
	view.BindFuncMap(g.Map{
		"UcFirst": func(str string) string {
			return gstr.UcFirst(str)
		},
		"add": func(a, b int) int {
			return a + b
		},
	})
	view.SetConfigWithMap(g.Map{
		"Paths":      []string{"template"},
		"Delimiters": []string{"{{", "}}"},
	})
	//树形菜单选项
	var options g.Map
	if entity.TplCategory == "tree" {
		options = gjson.New(entity.Options).ToMap()
	}
	if tmpController, err := view.Parse("vm/go/"+entity.TplCategory+"/controller.template", g.Map{"table": entity}); err == nil {
		controllerValue = tmpController
	}
	if tmpService, err := view.Parse("vm/go/"+entity.TplCategory+"/service.template", g.Map{"table": entity, "options": options}); err == nil {
		serviceValue = tmpService
	}
	if tmpModel, err := view.Parse("vm/go/"+entity.TplCategory+"/model.template", g.Map{"table": entity}); err == nil {
		modelValue = tmpModel
		modelValue, err = c.trimBreak(modelValue)
	}
	if tmpJs, err := view.Parse("vm/html/js.template", g.Map{"table": entity}); err == nil {
		apiJsValue = tmpJs
	}
	if tmpVue, err := view.Parse("vm/html/vue_"+entity.TplCategory+".template", g.Map{"table": entity, "options": options}); err == nil {
		vueValue = tmpVue
		vueValue, err = c.trimBreak(vueValue)
	}

	response.SusJson(true, r, "ok", g.Map{
		modelKey:      modelValue,
		serviceKey:    serviceValue,
		controllerKey: controllerValue,
		apiJsKey:      apiJsValue,
		vueKey:        vueValue,
	})
}

// @Summary 代码生成
// @Description 代码生成
// @Tags 系统工具
// @Param tableId body integer   true "tableId"
// @Success 0 {object} response.Response "{"code": 200, "data": [...]}"
// @Router /system/tools/gen/batchGenCode [post]
// @Security
func (c *Gen) BatchGenCode(r *ghttp.Request) {
	tables := strings.Split(r.GetString("tables"), ",")
	if len(tables) == 0 {
		response.FailJson(true, r, "参数错误")
	}

	var bufBytes []byte
	var err error
	if len(tables) == 1 {
		bufBytes, err = gen_service.DownloadOnce(tables[0])
	} else {
		bufBytes, err = gen_service.DownloadMulti(tables)
	}

	if err != nil {
		response.FailJson(true, r, err.Error())
	}

	c.responseZip(r, bufBytes)
}

func (c *Gen) responseZip(r *ghttp.Request, bytes []byte) {
	//直接返回数据流，前端使用blob对象接收
	r.Response.Header().Set("Access-Control-Allow-Origin", "*")
	r.Response.Header().Set("Access-Control-Allow-Origin", "*")
	r.Response.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	r.Response.Header().Set("Content-Disposition", "attachment; filename=\"gfast.zip\"")
	r.Response.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
	r.Response.Header().Set("Content-Type", "application/octet-stream; charset=UTF-8")
	r.Response.Writer.Write(bytes)
}

//剔除多余的换行
func (c *Gen) trimBreak(str string) (s string, err error) {
	var b []byte
	if b, err = gregex.Replace("(([\\s\t]*)\r?\n){2,}", []byte("$2\n"), []byte(str)); err == nil {
		s = gconv.String(b)
	}
	return
}
