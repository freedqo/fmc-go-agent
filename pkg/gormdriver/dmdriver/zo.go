/*
 * Copyright (c) 2000-2018, 达梦数据库有限公司.
 * All rights reserved.
 */

package dmdriver

const (
	ParamDataEnum_Null = 0
	/**
	 * 只有大字段才有行内数据、行外数据的概念
	 */
	ParamDataEnum_OFF_ROW = 1
)

// JDBC中的Data
type lobCtl struct {
	value []byte
}
