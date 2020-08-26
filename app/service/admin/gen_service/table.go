package gen_service

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"gfast/app/model/admin/gen_table"
	"gfast/app/model/admin/gen_table_column"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/os/gview"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//根据条件分页查询数据
func SelectDbTableList(param *gen_table.SelectPageReq) (total int, list []*gen_table.Entity, err error) {
	return gen_table.SelectDbTableList(param)
}

//根据条件分页查询数据
func SelectListByPage(param *gen_table.SelectPageReq) (total int, list []*gen_table.Entity, err error) {
	return gen_table.SelectListByPage(param)
}

//查询据库列表
func SelectDbTableListByNames(tableNames []string) ([]*gen_table.Entity, error) {
	return gen_table.SelectDbTableListByNames(tableNames)
}

//导入表结构
func ImportGenTable(tableList []*gen_table.Entity, operName string) error {
	if tableList != nil && operName != "" {
		tx, err := g.DB().Begin()
		if err != nil {
			return err
		}

		for _, table := range tableList {
			tableName := table.TableName
			InitTable(table, operName)
			result, err := tx.Table(gen_table.Table).Insert(table)
			if err != nil {
				return err
			}

			tmpid, err := result.LastInsertId()

			if err != nil || tmpid <= 0 {
				tx.Rollback()
				return gerror.New("保存数据失败")
			}

			table.TableId = tmpid

			// 保存列信息
			genTableColumns, err := gen_table_column.SelectDbTableColumnsByName(tableName)

			if err != nil || len(genTableColumns) <= 0 {
				tx.Rollback()
				return gerror.New("获取列数据失败")
			}

			for _, column := range genTableColumns {
				InitColumnField(column, table)
				_, err = tx.Table("gen_table_column").Insert(column)
				if err != nil {
					tx.Rollback()
					return gerror.New("保存列数据失败")
				}
			}
		}
		return tx.Commit()
	} else {
		return gerror.New("参数错误")
	}
}

//获取数据库类型字段
func GetDbType(columnType string) string {
	if strings.Index(columnType, "(") > 0 {
		return columnType[0:strings.Index(columnType, "(")]
	} else {
		return columnType
	}
}

//将下划线大写方式命名的字符串转换为驼峰式。如果转换前的下划线大写方式命名的字符串为空，则返回空字符串。 例如：HELLO_WORLD->HelloWorld
func ConvertToCamelCase(name string) string {
	if name == "" {
		return ""
	} else if !strings.Contains(name, "_") {
		// 不含下划线，仅将首字母大写
		return strings.ToUpper(name[0:1]) + name[1:len(name)]
	}
	var result string = ""
	camels := strings.Split(name, "_")
	for index := range camels {
		if camels[index] == "" {
			continue
		}
		camel := camels[index]
		result = result + strings.ToUpper(camel[0:1]) + strings.ToLower(camel[1:len(camel)])
	}
	return result
}

////将下划线大写方式命名的字符串转换为驼峰式,首字母小写。如果转换前的下划线大写方式命名的字符串为空，则返回空字符串。 例如：HELLO_WORLD->helloWorld
func ConvertToCamelCase1(name string) string {
	if name == "" {
		return ""
	} else if !strings.Contains(name, "_") {
		// 不含下划线，原值返回
		return name
	}
	var result string = ""
	camels := strings.Split(name, "_")
	for index := range camels {
		if camels[index] == "" {
			continue
		}
		camel := camels[index]
		if result == "" {
			result = strings.ToLower(camel[0:1]) + strings.ToLower(camel[1:len(camel)])
		} else {
			result = result + strings.ToUpper(camel[0:1]) + strings.ToLower(camel[1:len(camel)])
		}
	}
	return result
}

//获取字段长度
func GetColumnLength(columnType string) int {
	start := strings.Index(columnType, "(")
	end := strings.Index(columnType, ")")
	result := ""
	if start >= 0 && end >= 0 {
		result = columnType[start+1 : end-1]
	}
	return gconv.Int(result)
}

//初始化列属性字段
func InitColumnField(column *gen_table_column.Entity, table *gen_table.Entity) {
	dataType := GetDbType(column.ColumnType)
	columnName := column.ColumnName
	column.TableId = table.TableId
	column.CreateBy = table.CreateBy
	column.CreateTime = gtime.Now()
	column.UpdateTime = column.CreateTime
	//设置字段名
	column.GoField = ConvertToCamelCase(columnName)
	column.HtmlField = ConvertToCamelCase1(columnName)

	if gen_table_column.IsStringObject(dataType) {
		//字段为字符串类型
		column.GoType = "string"
		columnLength := GetColumnLength(column.ColumnType)
		if columnLength >= 500 {
			column.HtmlType = "textarea"
		} else {
			column.HtmlType = "input"
		}
	} else if gen_table_column.IsTimeObject(dataType) {
		//字段为时间类型
		column.GoType = "Time"
		column.HtmlType = "datatime"
	} else if gen_table_column.IsNumberObject(dataType) {
		//字段为数字类型
		column.HtmlType = "input"
		// 如果是浮点型
		tmp := column.ColumnType
		start := strings.Index(tmp, "(")
		end := strings.Index(tmp, ")")
		result := "0"
		if start > 0 && end > 0 {
			result = tmp[start+1 : end]
		}
		arr := strings.Split(result, ",")
		if len(arr) == 2 && gconv.Int(arr[1]) > 0 {
			column.GoType = "float64"
		} else if len(arr) == 1 && gconv.Int(arr[0]) <= 10 {
			column.GoType = "int"
		} else {
			column.GoType = "int64"
		}
	}
	//新增字段
	if columnName == "create_by" || columnName == "create_time" || columnName == "update_by" || columnName == "update_time" {
		column.IsRequired = "0"
		column.IsInsert = "0"
	} else {
		column.IsRequired = "0"
		column.IsInsert = "1"
		if strings.Index(columnName, "name") >= 0 || strings.Index(columnName, "status") >= 0 {
			column.IsRequired = "1"
		}
	}

	// 编辑字段
	if gen_table_column.IsNotEdit(columnName) {
		if column.IsPk == "1" {
			column.IsEdit = "0"
		} else {
			column.IsEdit = "1"
		}
	} else {
		column.IsEdit = "0"
	}
	// 列表字段
	if gen_table_column.IsNotList(columnName) {
		column.IsList = "1"
	} else {
		column.IsList = "0"
	}
	// 查询字段
	if gen_table_column.IsNotQuery(columnName) {
		column.IsQuery = "1"
	} else {
		column.IsQuery = "0"
	}

	// 查询字段类型
	if CheckNameColumn(columnName) {
		column.QueryType = "LIKE"
	} else {
		column.QueryType = "EQ"
	}

	// 状态字段设置单选框
	if CheckStatusColumn(columnName) {
		column.HtmlType = "radio"
	} else if CheckTypeColumn(columnName) || CheckSexColumn(columnName) {
		// 类型&性别字段设置下拉框
		column.HtmlType = "select"
	}
}

//检查字段名后3位是否是sex
func CheckSexColumn(columnName string) bool {
	if len(columnName) >= 3 {
		end := len(columnName)
		start := end - 3

		if start <= 0 {
			start = 0
		}

		if columnName[start:end] == "sex" {
			return true
		}
	}
	return false
}

//检查字段名后4位是否是type
func CheckTypeColumn(columnName string) bool {
	if len(columnName) >= 4 {
		end := len(columnName)
		start := end - 4

		if start <= 0 {
			start = 0
		}

		if columnName[start:end] == "type" {
			return true
		}
	}
	return false
}

//检查字段名后6位是否是status
func CheckStatusColumn(columnName string) bool {
	if len(columnName) >= 6 {
		end := len(columnName)
		start := end - 6

		if start <= 0 {
			start = 0
		}
		tmp := columnName[start:end]

		if tmp == "status" {
			return true
		}
	}

	return false
}

//检查字段名后4位是否是name
func CheckNameColumn(columnName string) bool {
	if len(columnName) >= 4 {
		end := len(columnName)
		start := end - 4

		if start <= 0 {
			start = 0
		}

		tmp := columnName[start:end]

		if tmp == "name" {
			return true
		}
	}
	return false
}

//初始化表信息
func InitTable(table *gen_table.Entity, operName string) {
	table.ClassName = ConvertClassName(table.TableName)
	table.PackageName = g.Cfg().GetString("gen.packageName")
	table.ModuleName = g.Cfg().GetString("gen.moduleName")
	//TODO 2020-8-26 取表全名
	//table.BusinessName = GetBusinessName(table.TableName)
	table.BusinessName = table.TableName
	table.FunctionName = strings.ReplaceAll(table.TableComment, "表", "")
	table.FunctionAuthor = g.Cfg().GetString("gen.author")
	table.CreateBy = operName
	table.TplCategory = "crud"
	table.CreateTime = gtime.Now()
	table.UpdateTime = table.CreateTime
}

//表名转换成类名
func ConvertClassName(tableName string) string {
	autoRemovePre := g.Cfg().GetBool("gen.autoRemovePre")
	tablePrefix := g.Cfg().GetString("gen.tablePrefix")
	if autoRemovePre && tablePrefix != "" {
		searchList := strings.Split(tablePrefix, ",")
		for _, str := range searchList {
			tableName = strings.ReplaceAll(tableName, str, "")
		}
	}
	return tableName
}

//获取业务名
func GetBusinessName(tableName string) string {
	lastIndex := strings.LastIndex(tableName, "_")
	nameLength := len(tableName)
	businessName := tableName[lastIndex+1 : nameLength]
	return businessName
}

//根据table_id查询表列数据
func SelectGenTableColumnListByTableId(tableId int64) ([]*gen_table_column.Entity, error) {
	return gen_table_column.SelectGenTableColumnListByTableId(tableId)
}

func GetTableInfoByTableId(tableId int64) (info *gen_table.Entity, err error) {
	return gen_table.GetInfoById(tableId)
}

//修改表和列信息
func SaveEdit(req *gen_table.EditReq) (err error) {
	if req == nil {
		err = gerror.New("参数错误")
		return
	}
	table, err := gen_table.FindOne("table_id=?", req.TableId)
	if err != nil || table == nil {
		err = gerror.New("数据不存在")
		return
	}
	if req.TableName != "" {
		table.TableName = req.TableName
	}
	if req.TableComment != "" {
		table.TableComment = req.TableComment
	}
	if req.BusinessName != "" {
		table.BusinessName = req.BusinessName
	}
	if req.ClassName != "" {
		table.ClassName = req.ClassName
	}
	if req.FunctionAuthor != "" {
		table.FunctionAuthor = req.FunctionAuthor
	}
	if req.FunctionName != "" {
		table.FunctionName = req.FunctionName
	}
	if req.ModuleName != "" {
		table.ModuleName = req.ModuleName
	}
	if req.PackageName != "" {
		table.PackageName = req.PackageName
	}
	if req.Remark != "" {
		table.Remark = req.Remark
	}
	if req.TplCategory != "" {
		table.TplCategory = req.TplCategory
	}
	if req.Params != "" {
		table.Options = req.Params
	}
	table.UpdateTime = gtime.Now()
	table.UpdateBy = req.UserName
	if req.TplCategory == "tree" {
		//树表设置options
		options := g.Map{
			"tree_code":        req.TreeCode,
			"tree_parent_code": req.TreeParentCode,
			"tree_name":        req.TreeName,
		}
		table.Options = gconv.String(options)
	} else {
		table.Options = ""
	}

	var tx *gdb.TX
	tx, err = g.DB().Begin()
	if err != nil {
		return
	}
	_, err = tx.Table(gen_table.Table).Save(table)
	if err != nil {
		tx.Rollback()
		return err
	}

	//保存列数据
	if req.Columns != "" {
		var j *gjson.Json
		if j, err = gjson.DecodeToJson([]byte(req.Columns)); err != nil {
			tx.Rollback()
			return
		} else {
			var columnList []gen_table_column.Entity
			err = j.ToStructs(&columnList)
			if err == nil && columnList != nil && len(columnList) > 0 {
				for _, column := range columnList {
					if column.ColumnId > 0 {
						tmp, _ := gen_table_column.FindOne("column_id=?", column.ColumnId)
						if tmp != nil {
							tmp.ColumnComment = column.ColumnComment
							tmp.GoType = column.GoType
							tmp.HtmlType = column.HtmlType
							tmp.QueryType = column.QueryType
							tmp.GoField = column.GoField
							tmp.DictType = column.DictType
							tmp.IsInsert = column.IsInsert
							tmp.IsEdit = column.IsEdit
							tmp.IsList = column.IsList
							tmp.IsQuery = column.IsQuery
							_, err = tx.Table(gen_table_column.Table).Save(tmp)
							if err != nil {
								tx.Rollback()
								return
							}
						}
					}
				}
			}
		}
	}
	tx.Commit()
	return
}

//删除表格
func Delete(ids []int) error {
	tx, err := g.DB().Begin()
	if err != nil {
		g.Log().Error(err)
		return gerror.New("开启删除事务出错")
	}
	_, err = tx.Table(gen_table.Table).Where(gen_table.Columns.TableId+" in(?)", ids).Delete()
	if err != nil {
		g.Log().Error(err)
		tx.Rollback()
		return gerror.New("删除表格数据失败")
	}
	_, err = tx.Table(gen_table_column.Table).Where(gen_table_column.Columns.TableId+" in(?)", ids).Delete()
	if err != nil {
		g.Log().Error(err)
		tx.Rollback()
		return gerror.New("删除表格字段数据失败")
	}
	tx.Commit()
	return nil
}

func SelectRecordById(tableId int64) (entity *gen_table.EntityExtend, err error) {
	entity, err = gen_table.SelectRecordById(tableId)
	return
}

//TODO 2020-8-24 根据表名获取实体
func SelectRecordByTableName(tableName string) (entity *gen_table.EntityExtend, err error) {
	entity, err = gen_table.SelectRecordByTableName(tableName)
	return
}

//TODO 2020-8-25 目录及文件名和代码结构体
type DirCode struct {
	Name, Body string
}

//TODO 2020-8-25 只下载生成一个表
func DownloadOnce(name string) ([]byte, error) {
	return generateCode(name, false)
}

//TODO 2020-8-25 下载多个生成的表到一个zip包
//TODO 2020-8-26 需要放在同一个文件夹下
func DownloadMulti(tables []string) ([]byte, error) {
	curDir := "temp"
	if gfile.Exists(curDir) {
		gfile.Remove(curDir)
	}

	for _, table := range tables {
		_, err := generateCode(table, true)
		if err != nil {
			return nil, err
		}
	}

	//开始写入zip包
	zipFile, err := os.Create("gfast.zip")
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		f, err := w.Create(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}

	err = filepath.Walk(curDir, walker)
	if err != nil {
		return nil, err
	}

	return gfile.GetBytes("gfast.zip"), err
}

//TODO 2020-8-25 构建已生成的目录及文件名和代码的结构体数组
func generateCode(tableName string, isMulti bool) ([]byte, error) {
	entity, err := SelectRecordByTableName(tableName)
	if err != nil {
		//response.FailJson(true, r, err.Error())
		return nil, err
	}
	if entity == nil {
		//response.FailJson(true, r, "表格数据不存在")
		return nil, errors.New("表格数据不存在")
	}
	SetPkColumn(entity, entity.Columns)

	controllerKey := fmt.Sprintf("temp/app/controller/%s/%s_controller.go", entity.ModuleName, entity.BusinessName)
	controllerValue := ""
	serviceKey := fmt.Sprintf("temp/app/service/%s/%s_service/%s.go", entity.ModuleName, entity.BusinessName, entity.BusinessName)
	serviceValue := ""
	modelKey := fmt.Sprintf("temp/app/model/%s/%s/%s.go", entity.ModuleName, entity.BusinessName, entity.BusinessName)
	modelValue := ""
	modelExtendKey := fmt.Sprintf("temp/app/model/%s/%s/%s_model.go", entity.ModuleName, entity.BusinessName, entity.BusinessName)
	modelExtendValue := ""
	entityKey := fmt.Sprintf("temp/app/model/%s/%s/%s_entity.go", entity.ModuleName, entity.BusinessName, entity.BusinessName)
	entityValue := ""

	apiJsKey := fmt.Sprintf("temp/vue/src/api/%s/%s/index.js", entity.ModuleName, entity.BusinessName)
	apiJsValue := ""
	vueKey := fmt.Sprintf("temp/vue/src/views/%s/%s/index.vue", entity.ModuleName, entity.BusinessName)
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
		modelValue, err = trimBreak(modelValue)
	}
	if tmpModelEx, err := view.Parse("vm/go/"+entity.TplCategory+"/model_extend.template", g.Map{"table": entity}); err == nil {
		modelExtendValue = tmpModelEx
		modelExtendValue, err = trimBreak(modelExtendValue)
	}
	if tmpEntity, err := view.Parse("vm/go/"+entity.TplCategory+"/entity.template", g.Map{"table": entity}); err == nil {
		entityValue = tmpEntity
		entityValue, err = trimBreak(entityValue)
	}

	if tmpJs, err := view.Parse("vm/html/js.template", g.Map{"table": entity}); err == nil {
		apiJsValue = tmpJs
	}
	if tmpVue, err := view.Parse("vm/html/vue_"+entity.TplCategory+".template", g.Map{"table": entity, "options": options}); err == nil {
		vueValue = tmpVue
		vueValue, err = trimBreak(vueValue)
	}

	var bytes []byte
	dirCodeArr := []DirCode{
		{controllerKey, controllerValue},
		{serviceKey, serviceValue},
		{modelKey, modelValue},
		{modelExtendKey, modelExtendValue},
		{entityKey, entityValue},
		{apiJsKey, apiJsValue},
		{vueKey, vueValue},
	}
	if isMulti {
		bytes, err = makeFolderFiles(dirCodeArr)
	} else {
		bytes, err = zipCompress(dirCodeArr)
	}

	return bytes, err
}

//TODO 2020-8-26 生成本地文件流
func makeFolderFiles(files []DirCode) ([]byte, error) {
	//获取当前运行时目录
	curDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileName := strings.Join([]string{curDir, "/", file.Name}, "")
		if !gfile.Exists(fileName) {
			f, err := gfile.Create(fileName)
			if err == nil {
				f.WriteString(file.Body)
			}
			f.Close()
		}
	}

	return gfile.GetBytes(curDir + "/temp"), nil
}

//TODO 2020-8-25 生成压缩流
func zipCompress(files []DirCode) ([]byte, error) {
	//开始写入zip包
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for _, file := range files {
		f, err := w.Create(file.Name)
		if err != nil {
			return nil, err
		}

		_, err = f.Write([]byte(file.Body))
		if err != nil {
			return nil, err
		}
	}

	err := w.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

//剔除多余的换行
func trimBreak(str string) (s string, err error) {
	var b []byte
	if b, err = gregex.Replace("(([\\s\t]*)\r?\n){2,}", []byte("$2\n"), []byte(str)); err == nil {
		s = gconv.String(b)
	}
	return
}

//设置主键列信息
func SetPkColumn(table *gen_table.EntityExtend, columns []*gen_table_column.Entity) {
	for _, column := range columns {
		if column.IsPk == "1" {
			table.PkColumn = column
			break
		}
	}
	if table.PkColumn == nil {
		table.PkColumn = columns[0]
	}
}
