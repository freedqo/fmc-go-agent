package dbm

type Page struct {
	Index int `json:"index"` // 当前页码
	Size  int `json:"size"`  // 每页大小
	Total int `json:"total"` // 总条数
}
