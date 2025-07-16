package ugrom

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

func NewUGorm(log *zap.SugaredLogger, dialector gorm.Dialector) (db *gorm.DB, err error) {
	gormLog := NewGormLog(log)
	db, err = gorm.Open(dialector, &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// 测试连接
	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
