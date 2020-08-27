// ============================================================================
// This is auto-generated by gf cli tool only once. Fill this file as you wish.
// ============================================================================

package gen_table_column

import (
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"strings"
)

//数据库字符串类型
var COLUMNTYPE_STR = []string{"char", "varchar", "narchar", "varchar2", "tinytext", "text", "mediumtext", "longtext", "enum"}

//数据库时间类型
var COLUMNTYPE_TIME = []string{"datetime", "time", "date", "timestamp"}

//数据库数字类型
var COLUMNTYPE_NUMBER = []string{"tinyint", "smallint", "mediumint", "int", "number", "integer", "bigint", "float", "float", "double", "decimal", "int unsigned", "bigint unsigned", "tinyint unsigned"}

//页面不需要编辑字段
var COLUMNNAME_NOT_EDIT = []string{"id", "create_by", "create_time", "del_flag", "update_by", "update_time"}

//页面不需要显示的列表字段
var COLUMNNAME_NOT_LIST = []string{"id", "create_by", "create_time", "del_flag", "update_by", "update_time"}

//页面不需要查询字段
var COLUMNNAME_NOT_QUERY = []string{"id", "create_by", "create_time", "del_flag", "update_by", "update_time", "remark"}

//根据表名称查询列信息
func SelectDbTableColumnsByName(tableName string) ([]*Entity, error) {
	db := g.DB()
	var entity []*Entity
	sql := " select column_name, (case when (is_nullable = 'no' && column_key != 'PRI') then '1' else null end) as is_required, " +
		"(case when column_key = 'PRI' then '1' else '0' end) as is_pk, ordinal_position as sort, column_comment," +
		" (case when extra = 'auto_increment' then '1' else '0' end) as is_increment, column_type from information_schema.columns" +
		" where table_schema = (select database()) "
	sql += " and " + gdb.FormatSqlWithArgs(" table_name=? ", []interface{}{tableName}) + " order by ordinal_position ASC "
	result, err := db.GetAll(sql)
	if err != nil {
		g.Log().Error(err)
		return nil, gerror.New("查询列信息失败")
	}
	result.Structs(&entity)
	return entity, nil
}

//判断是否是数据库字符串类型
func IsStringObject(value string) bool {
	return IsExistInArray(value, COLUMNTYPE_STR)
}

//判断是否是数据库时间类型
func IsTimeObject(value string) bool {
	return IsExistInArray(value, COLUMNTYPE_TIME)
}

//判断是否是数据库数字类型
func IsNumberObject(value string) bool {
	return IsExistInArray(value, COLUMNTYPE_NUMBER)
}

//页面不需要编辑字段
func IsNotEdit(value string) bool {
	return !IsExistInArray(value, COLUMNNAME_NOT_EDIT)
}

//页面不需要显示的列表字段
func IsNotList(value string) bool {
	return !IsExistInArray(value, COLUMNNAME_NOT_LIST)
}

//页面不需要查询字段
func IsNotQuery(value string) bool {
	return !IsExistInArray(value, COLUMNNAME_NOT_QUERY)
}

//判断string 是否存在在数组中
func IsExistInArray(value string, array []string) bool {
	for _, v := range array {
		if strings.Contains(value, v) {
			return true
		}
	}
	return false
}

//查询业务字段列表
func SelectGenTableColumnListByTableId(tableId int64) ([]*Entity, error) {
	list, err := Model.Where(Columns.TableId, tableId).Order(Columns.Sort + " asc, " + Columns.ColumnId + " asc").All()
	if err != nil {
		g.Log().Error(err)
		return nil, gerror.New("获取字段信息出错")
	}
	return list, nil
}
