package ugrom

import (
	"fmt"
	"github.com/freedqo/fmc-go-agent/pkg/gormdriver/dmdriver"
	"github.com/freedqo/fmc-go-agent/pkg/gormdriver/sqlserverdriver"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDefaultOption() *Option {
	return &Option{
		DbType:        "mysql",
		Host:          "localhost",
		Port:          3306,
		Account:       "root",
		Password:      "123456",
		IsAutoMigrate: false,
	}
}

type Option struct {
	DbType        string `comment:"数据库类型,目前支持:mysql、mssql、dm"` //数据库类型
	Host          string `comment:"数据库服务的主机或IP"`               //数据库服务的主机或IP
	Port          int    `comment:"数据库服务侦听的端口号"`               //数据库服务侦听的端口号
	Account       string `comment:"登录用户的帐号"`                   //登录用户的帐号
	Password      string `comment:"登录用户的密码"`                   //登录用户的密码
	IsAutoMigrate bool   `comment:"是否自动迁移表结构"`                 //是否自动迁移表结构
}

func (m *Option) Open(db string) gorm.Dialector {
	dns := m.Dns(db)
	switch m.DbType {
	case "mysql":
		return mysql.Open(dns)
	case "mssql":
		return sqlserverdriver.Open(dns)
	case "dm":
		return dmdriver.Open(dns)
	default:
		panic("not support db type:" + m.DbType)
	}
}

func (m *Option) Dns(db string) string {
	switch m.DbType {
	case "mysql":
		return m.mySqlDns(db)
	case "mssql":
		return m.sqlServerDns(db)
	case "dm":
		return m.dmDns(db)
	default:
		panic("not support db type:" + m.DbType)
	}
}
func (m *Option) mySqlDns(db string) string {
	if m.DbType != "mysql" {
		panic("not mysql")
	}
	if db == "" {
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/", m.Account, m.Password, m.Host, m.Port)
	} else {
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", m.Account, m.Password, m.Host, m.Port, db)
	}
}

func (m *Option) sqlServerDns(db string) string {
	if m.DbType != "mssql" {
		panic("not mssql")
	}
	if db == "" {
		return fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;encrypt=disable;", m.Host, m.Account, m.Password, m.Port)
	} else {
		return fmt.Sprintf("Data Source=%s;user id=%s;password=%s;port=%d;Initial Catalog=%s", m.Host, m.Account, m.Password, m.Port, db)
	}
}
func (m *Option) dmDns(db string) string {
	if m.DbType != "dm" {
		panic("not dm")
	}
	if db == "" {
		return fmt.Sprintf("dm://%s:%s@%s:%d", m.Host, m.Account, m.Password, m.Port)
	} else {
		return fmt.Sprintf("dm://%s:%s@%s:%d?schema=%s&ignoreCase=true", m.Host, m.Account, m.Password, m.Port, db)
	}
}
