package gdb

import (
	"fmt"
	"github.com/freedqo/fmc-go-agents/scripts/cmd/dbgen/options"
	"github.com/freedqo/fmc-go-agents/scripts/cmd/dbgen/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"sort"
)

type DbMeta struct {
	DbName    string
	TableMeta []*QueryStructMeta
}

// QueryStructMeta struct info in generated code
type QueryStructMeta struct {
	DbName          string
	Generated       bool   // whether to generate db model
	FileName        string // generated file name
	S               string // the first letter(lower case)of simple Name (receiver)
	QueryStructName string // internal query struct name
	ModelStructName string // origin/model struct name
	TableName       string // table name in db server
	TableComment    string // table comment in db server
	Fields          []*Field
	ImportPkgPaths  []string
	interfaceMode   bool
}

// Field user input structures
type Field struct {
	Name             string
	Type             string
	ColumnName       string
	ColumnComment    string
	MultilineComment bool
	Tag              field.Tag
	GORMTag          field.GormTag
	CustomGenType    string
	Relation         *field.Relation
}

func NewGdbConn(addr string, dbname string, user string, password string) (db *gorm.DB, g *gen.Generator, tables []string, dMeta *DbMeta, err error) {
	// 数据库连接信息
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, addr, dbname)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, nil, nil, err
	}
	sqldb, err := db.DB()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	err = sqldb.Ping()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	cfg := gen.Config{
		// 模型文件包名
		ModelPkgPath: "/" + "model",
		// 生成的模型文件输出目录
		OutPath: options.ModelRootPath + "/" + dbname + "/" + "query",
		// 当字段为空时生成指针
		FieldNullable: false,
		// 生成的结构体字段使用指针类型表示可空字段
		FieldCoverable: true,
		// 生成的结构体字段添加结构体标签
		FieldWithIndexTag: true,
		// 生成的结构体字段添加 GORM 标签
		FieldWithTypeTag: true,
	}
	cfg.WithModelNameStrategy(func(tableName string) (modelName string) {
		return utils.FirstToUpper(tableName)
	})
	// 创建代码生成器实例
	g = gen.NewGenerator(cfg)
	// 设置数据库连接
	g.UseDB(db)

	tables, err = db.Migrator().GetTables()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	sort.Slice(tables, func(i, j int) bool {
		return tables[i] < tables[j]
	})
	dbMeta := &DbMeta{
		DbName:    dbname,
		TableMeta: make([]*QueryStructMeta, 0),
	}
	for _, table := range tables {
		gm := g.GenerateModel(table)
		qMeta := &QueryStructMeta{
			DbName:          dbname,
			Generated:       gm.Generated,
			FileName:        gm.FileName,
			S:               gm.S,
			QueryStructName: gm.QueryStructName,
			ModelStructName: gm.ModelStructName,
			TableName:       gm.TableName,
			TableComment:    gm.TableComment,
			ImportPkgPaths:  gm.ImportPkgPaths,
		}
		qMeta.Fields = make([]*Field, 0)
		for _, field := range gm.Fields {
			f := Field{
				Name:             field.Name,
				Type:             field.Type,
				ColumnName:       field.ColumnName,
				ColumnComment:    field.ColumnComment,
				MultilineComment: field.MultilineComment,
				Tag:              field.Tag,
				GORMTag:          field.GORMTag,
				CustomGenType:    field.CustomGenType,
				Relation:         field.Relation,
			}
			qMeta.Fields = append(qMeta.Fields, &f)
		}
		sort.Slice(qMeta.Fields, func(i, j int) bool {
			return qMeta.Fields[i].Name < qMeta.Fields[j].Name
		})
		dbMeta.TableMeta = append(dbMeta.TableMeta, qMeta)
	}
	sort.Slice(dbMeta.TableMeta, func(i, j int) bool {
		return dbMeta.TableMeta[i].ModelStructName < dbMeta.TableMeta[j].ModelStructName
	})

	return db, g, tables, dbMeta, nil
}
