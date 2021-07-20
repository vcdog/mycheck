package InspectionItem

import (
	"DepthInspection/api/PublicDB"
	"fmt"
	"strings"
)

func DBConfigCheck(aa *PublicDB.ConfigInfo,confParameterList map[string]string) (bool,string) {
	var acd bool
	aa.Loggs.Info("Begin to check that the database configuration parameters are properly configured")
	DBdate := aa.DatabaseExecInterf.DBQueryDateMap(aa,"show global variables")
	var configVariablesName ,configValue string
	for i := range confParameterList{
		configVariablesName = i
		configValue = confParameterList[i]
		aa.Loggs.Debug("Start checking database parameters ",configVariablesName)
		a,ok := DBdate[configVariablesName]
		if !ok {
			aa.Loggs.Error("The current data configuration parameter does not exist. Please check if it is incorrectly typed")
		}
		if !strings.EqualFold(a,configValue) {
			errorStrinfo := fmt.Sprintf("检测当前数据库配置参数为 %s 不符合预定要求! 当前值为 %s 建议设置成 %s",configVariablesName,a,configValue)
			aa.Loggs.Error(errorStrinfo)
		}
	}
	aa.Loggs.Info("The check database configuration parameters are complete")
	return acd,configValue
}

//type databaseBaseLineCheckInterface interface {
//	TableDesignCompliance(aa *PublicDB.ConfigInfo)
//}
type DatabaseBaselineCheckStruct struct {
	strSql string
	ignoreTableSchema string
}
type TableDesignComplianceStruct struct {
	DatabaseName interface{} `json: "databaseName"`
	TableName interface{} `json: "tableName"`
	Engine interface{}  `json: "engine"`
	Charset interface{} `json: "charset"`
}

func (baselineCheck *DatabaseBaselineCheckStruct) BaselineCheckTablesDesign(aa *PublicDB.ConfigInfo) {
	//表字符集检查 ~ 	表引擎检查
	aa.Loggs.Info("Begin a baseline check to check database table design compliance")
	ignoreTableSchema := "'mysql','information_schema','performance_schema','sys'"
	strSql := fmt.Sprintf("select t.table_schema databaseName,t.table_name tableName,lower(engine) engine,lower(c.CHARACTER_SET_NAME) charset from information_schema.tables as t, information_schema.COLLATIONS as c where t.TABLE_COLLATION=c.COLLATION_NAME and t.table_schema not in (%s)",ignoreTableSchema)
	cc := aa.DatabaseExecInterf.DBQueryDateJson(aa,strSql)
	for i := range cc{
		//表字符集检查
		charsetSt := fmt.Sprintf("%v",cc[i]["charset"])
		if !strings.Contains(charsetSt,"utf8"){
			aa.Loggs.Error(fmt.Sprintf("The current table character set is not UTF8 or UTF8MB4 character. error info: Database is %s table is %s table charset is %s ",cc[i]["databaseName"],cc[i]["tableName"],cc[i]["charset"]))
		}
		//表引擎检查
		if cc[i]["engine"] != "innodb"{
			aa.Loggs.Error(fmt.Sprintf("The current table engine set is not innodb engine. error info: Database is %s table is %s table engine is %s ",cc[i]["databaseName"],cc[i]["tableName"],cc[i]["engine"]))
		}
	}
	//检查表是否使用外键
	strSql = fmt.Sprintf("select CONSTRAINT_SCHEMA databaseName,TABLE_NAME tableName,COLUMN_NAME columnName,CONSTRAINT_NAME, REFERENCED_TABLE_NAME,REFERENCED_COLUMN_NAME from INFORMATION_SCHEMA.KEY_COLUMN_USAGE where CONSTRAINT_SCHEMA not in (%s)",ignoreTableSchema)
	dd := aa.DatabaseExecInterf.DBQueryDateJson(aa,strSql)
	for i := range dd {
		if dd[i]["REFERENCED_TABLE_NAME"] != nil && dd[i]["REFERENCED_COLUMN_NAME"] != nil {
			aa.Loggs.Error(fmt.Sprintf("The current table uses a foreign key constraint. The information is as follows: database: %s " +
				"tableName: %s column: %s Foreign key constraint name: %s Foreign key constraints table: %s" +
				"Foreign key constraints columns: %s",dd[i]["databaseName"],dd[i]["tableName"],dd[i]["columnName"],dd[i]["CONSTRAINT_NAME"],dd[i]["REFERENCED_TABLE_NAME"],dd[i]["REFERENCED_COLUMN_NAME"]))
		}
	}
	//检查没有主键的表
	strSql = fmt.Sprintf("select table_schema databaseName, table_name tableName from information_schema.tables where table_name not in (select distinct table_name from information_schema.columns where column_key = 'PRI' ) AND table_schema not in (%s)",ignoreTableSchema)
	ee := aa.DatabaseExecInterf.DBQueryDateJson(aa,strSql)
	for i := range ee{
		if ee[i] != nil{
			aa.Loggs.Error(fmt.Sprintf("The current table has no primary key index. The information is as follows: database: %s tableName: %s" ,ee[i]["databaseName"],ee[i]["tableName"]))
		}
	}
}

//列设计合规性
func (baselineCheck *DatabaseBaselineCheckStruct) BaselineCheckColumnsDesign(aa *PublicDB.ConfigInfo){
	ignoreTableSchema := "'mysql','information_schema','performance_schema','sys'"
	aa.Loggs.Info("Begin a baseline check to check database columns design compliance")
	strSql := fmt.Sprintf("select table_Schema databaseName,table_name tableName,column_name columnName,column_type columnType,COLUMN_KEY columnKey,EXTRA extra from information_schema.columns where table_schema not in(%s)",ignoreTableSchema)
	cc := aa.DatabaseExecInterf.DBQueryDateJson(aa,strSql)
	//var columnNumMap = make([]map[interface{}]int,len(cc))
	for i := range cc {
		//主键自增列是否为bigint
		if cc[i]["extra"] == "auto_increment" && cc[i]["columnType"] != "bigint"{
			aa.Loggs.Error(fmt.Sprintf("The primary key column is not of type Bigint. The information is as follows: database: %s tableName: %s columnsName: %s columnType: %s.", cc[i]["databaseName"],cc[i]["tableName"],cc[i]["columnName"],cc[i]["columnType"]))
		}
		//表中是否存在大字段blob、text、varchar(8099)、timestamp数据类型
		ce := fmt.Sprintf("%v",cc[i]["columnType"])
		if cc[i]["columnType"] == "blob" || strings.Contains(ce,"text") || cc[i]["columnType"] == "timestamp"{
			aa.Loggs.Error(fmt.Sprintf("The column data types of the current table in the database exist BLOB, TEXT, TIMESTAMP. The information is as follows: database: %s tableName: %s columnsName: %s columnType: %s.",cc[i]["databaseName"],cc[i]["tableName"],cc[i]["columnName"],cc[i]["columnType"]))
		}
		//var dd = make(map[string]string)
		//表列数是否大于255
	}
}
//索引设计合规性
func (baselineCheck *DatabaseBaselineCheckStruct) BaselineCheckIndexColumnDesign(aa *PublicDB.ConfigInfo){
	ignoreTableSchema := "'mysql','information_schema','performance_schema','sys'"
	aa.Loggs.Info("Begin by checking that index usage is reasonable and index column creation is standard")
	strSql := fmt.Sprintf("select a.table_schema databaseName,a.table_name tableName,a.column_name columnName,a.COLUMN_TYPE columnType,a.is_nullable isNullable,b.INDEX_NAME indexName from information_schema.columns a, information_schema.STATISTICS b  where a.table_schema not in(%s) and a.COLUMN_KEY !='' and a.TABLE_NAME = b.TABLE_NAME",ignoreTableSchema)
	cc := aa.DatabaseExecInterf.DBQueryDateJson(aa,strSql)
	for i := range cc {
		//判断索引列是否允许为空
		if cc[i]["isNullable"] == "YES" {
			aa.Loggs.Error(fmt.Sprintf("An index column is empty.The information is as follows: database: \"%s\"  tablename: \"%s\" indexName: \"%s\" columnName: \"%s\" columnType: \"%s\"",cc[i]["databaseName"],cc[i]["tableName"],cc[i]["indexName"],cc[i]["columnName"],cc[i]["columnType"]))
		}
		//判断索引列是否建立在enum或set类型上面
		if strings.Contains(fmt.Sprintf("%v",cc[i]["columnType"]),"enum") || strings.Contains(fmt.Sprintf("%v",cc[i]["columnType"]),"set"){
			aa.Loggs.Error(fmt.Sprintf("An index column is enum or set type. The information is as follows: The information is as follows: database: \"%s\"  tablename: \"%s\" indexName: \"%s\" columnName: \"%s\" columnType: \"%s\"",cc[i]["databaseName"],cc[i]["tableName"],cc[i]["indexName"],cc[i]["columnName"],cc[i]["columnType"]))
		}
		//判断索引列是否建立在大字段类型上（blob、text）
		if strings.Contains(fmt.Sprintf("%v",cc[i]["columnType"]),"blob") || strings.Contains(fmt.Sprintf("%v",cc[i]["columnType"]),"text"){
			aa.Loggs.Error(fmt.Sprintf("An index column is blob or text type. The information is as follows: The information is as follows: database: \"%s\"  tablename: \"%s\" indexName: \"%s\" columnName: \"%s\" columnType: \"%s\"",cc[i]["databaseName"],cc[i]["tableName"],cc[i]["indexName"],cc[i]["columnName"],cc[i]["columnType"]))
		}
	}
	//检查唯一索引和主键索引重复
	//strSql = fmt.Sprintf("select table_schema databaseName,table_name tableName,non_unique noUnique,index_name indexName,column_name columnName from information_schema.STATISTICS where table_schema not in (%s)",ignoreTableSchema)
	strSql = fmt.Sprintf("select table_schema databaseName,table_name tableName,non_unique noUnique,index_name indexName,column_name columnName from information_schema.STATISTICS where table_schema in (\"%s\")","wlkycs")
	//strSql = fmt.Sprintf("SELECT a.TABLE_SCHEMA databaseName, a.TABLE_NAME tableName, a.COLUMN_NAME columnName,a.INDEX_NAME AS 'index1',b.INDEX_NAME as 'index2',a.SEQ_IN_INDEX seqInIndex FROM  information_schema.STATISTICS a JOIN information_schema.STATISTICS b  ON a.TABLE_SCHEMA = b.TABLE_SCHEMA  AND a.TABLE_NAME = b.TABLE_NAME  AND a.SEQ_IN_INDEX = b.SEQ_IN_INDEX  AND a.COLUMN_NAME = b.COLUMN_NAME  WHERE  a.NON_UNIQUE in (0,1) and a.SEQ_IN_INDEX = 1 AND a.INDEX_NAME <> b.INDEX_NAME")
	dd := aa.DatabaseExecInterf.DBQueryDateJson(aa,strSql)
	//dmap["databaseName"] = ""
	//dmap["tableName"] = ""
	//dmap["columnName"] = ""
	//dmap["indexName"] = ""

	var indexCloumnMerge = make([]map[string]interface{},0)
	var columnNameString string
	//对数据进行处理，索引列进行合并，同一库表下同一索引名尤其是复合索引下将列进行合并
	for i := range dd{
		fmt.Println(dd[i])
		fmt.Println(columnNameString)
		var dmap = make(map[string]interface{})
		if dd[i]["databaseName"] == dmap["databaseName"] && dmap["tableName"] == dd[i]["tableName"] && dmap["indexName"] == dd[i]["indexName"]{
			columnNameString = fmt.Sprintf("%s,%s",dmap["columnName"],dd[i]["columnName"])
			dmap["columnName"] = columnNameString
			indexCloumnMerge = indexCloumnMerge[:len(indexCloumnMerge)-1]
		}else {
			dmap["columnName"] = dd[i]["columnName"]
		}
		dmap["databaseName"] = dd[i]["databaseName"]
		dmap["tableName"] = dd[i]["tableName"]
		dmap["indexName"] = dd[i]["indexName"]
		fmt.Println("dmap-- ",dmap)
		indexCloumnMerge = append(indexCloumnMerge,dmap)
		fmt.Println("indexColumnMerge-- ",indexCloumnMerge)
	}
	for i := range indexCloumnMerge{
		fmt.Println(indexCloumnMerge[i])
	}

	//for i := range dd{
	//
	//	if dd[i]["databaseName"] == dmap["databaseName"] && dmap["tableName"] == dd[i]["tableName"] && dmap["indexName"] != dd[i]["indexName"] {
	//		ad := fmt.Sprintf("%v",dd[i]["columnName"])
	//		ac := fmt.Sprintf("%v",dmap["columnName"])
	//		if strings.Contains(ad,ac) && strings.Split(ad,",")[0] == strings.Split(ac,",")[0] {
	//			fmt.Println("出现重复索引")
	//		}
	//	}else {
	//		dmap["databaseName"] = dd[i]["databaseName"]
	//		dmap["tableName"] = dd[i]["tableName"]
	//		dmap["columnName"] = dd[i]["columnName"]
	//	}
	//}

}
//存储过程及存储函数检查限制
func (baselineCheck *DatabaseBaselineCheckStruct) BaselineCheckProcedureTriggerDesign(aa *PublicDB.ConfigInfo){
	ignoreTableSchema := "'mysql','information_schema','performance_schema','sys'"
	aa.Loggs.Info("Begin a baseline check to checking whether the database uses stored procedures, stored functions, or triggers")
	strSql := fmt.Sprintf("select ROUTINE_SCHEMA databaseName,ROUTINE_NAME routineName,ROUTINE_TYPE routineType,DEFINER definer,CREATED created from information_schema.routines where ROUTINE_SCHEMA not in(%s)",ignoreTableSchema)
	cc := aa.DatabaseExecInterf.DBQueryDateJson(aa,strSql)
	for i := range cc{
		if cc[i]["routineType"] == "FUNCTION" || cc[i]["routineType"] == "PROCEDURE" {
			aa.Loggs.Error(fmt.Sprintf("The current database uses a storage function or storage procedure. The information is as follows: database: \"%s\" routineName: \"%s\" user: \"%s\" create time: \"%s\"" ,cc[i]["databaseName"],cc[i]["routineName"],cc[i]["definer"],cc[i]["created"]))
		}
	}
	strSql = fmt.Sprintf("select TRIGGER_SCHEMA databaseName,TRIGGER_NAME triggerName,DEFINER definer,CREATED created from information_schema.TRIGGERS where TRIGGER_SCHEMA not in (%s)",ignoreTableSchema)
	dd := aa.DatabaseExecInterf.DBQueryDateJson(aa,strSql)
	for i := range dd{
		if dd[i]["triggerName"] != nil{
			aa.Loggs.Error(fmt.Sprintf("The current database uses a trigger. The information is as follows: database: \"%s\" triggerName: \"%s\"  user: \"%s\"  create time:\"%s\"" ,dd[i]["databaseName"],dd[i]["triggerName"],dd[i]["definer"],dd[i]["created"]))
		}
	}
	aa.Loggs.Info("Check whether the database is completed using stored programs, stored functions, and stored triggers")
}
