package usercenterm

type Sm2LoginReq struct {
	Account  string `json:"account"`  //账号
	Password string `json:"password"` //密码,sm3加密
}

type Sm2LoginResp struct {
	Msg     string            `json:"msg"`
	Code    int               `json:"code"`
	Data    *Sm2LoginRespData `json:"data"`
	Total   int               `json:"total"`
	Pages   int               `json:"pages"`
	Success bool              `json:"success"`
}

type Permissions struct {
	PermissionID       int     `json:"permissionId"`
	SysType            string  `json:"sysType"`
	Icon               string  `json:"icon"`
	Resources          string  `json:"resources"`
	Sort               float64 `json:"sort"`
	ParentID           int     `json:"parentId"`
	Props              string  `json:"props"`
	TopStatus          int     `json:"topStatus"`
	PermissionTypeID   int     `json:"permissionTypeId"`
	SystemType         string  `json:"systemType"`
	Sn                 string  `json:"sn"`
	AuthFile           string  `json:"authFile"`
	PermissionCode     string  `json:"permissionCode"`
	PermissionName     string  `json:"permissionName"`
	PermissionTypeCode string  `json:"permissionTypeCode"`
	SecurityNeed       int     `json:"securityNeed"`
	Status             int     `json:"status"`
}
type Groups struct {
	UserGroupID int    `json:"userGroupId"`
	GroupName   string `json:"groupName"`
}
type RoleList struct {
	RoleID   int    `json:"roleId"`
	UserID   int    `json:"userId"`
	RoleName string `json:"roleName"`
	Assign   bool   `json:"assign"`
}
type DictList struct {
	DictValue string `json:"dictValue"`
	DictID    int    `json:"dictId"`
	DictType  string `json:"dictType"`
	DictName  string `json:"dictName"`
}
type StationList struct {
	StationCode string `json:"stationCode"`
	ValidFlag   int    `json:"validFlag"`
	UpdateTime  string `json:"updateTime"`
	ParentID    int    `json:"parentId"`
	DefaultAble int    `json:"defaultAble"`
	CreateTime  string `json:"createTime"`
	UpdaterID   int    `json:"updaterId"`
	TenantID    int    `json:"tenantId"`
	SortNum     int    `json:"sortNum"`
	StationName string `json:"stationName"`
	StationID   int    `json:"stationId"`
}
type Sm2LoginRespData struct {
	ClientMac          string        `json:"clientMac"`
	FirstLogin         int           `json:"firstLogin"`
	ValidFlag          int           `json:"validFlag"`
	IDCard             string        `json:"idCard"`
	OrganParentID      int           `json:"organParentId"`
	PassExpireDays     int           `json:"passExpireDays"`
	Fingers            string        `json:"fingers"`
	TenantName         string        `json:"tenantName"`
	OrganCode          string        `json:"organCode"`
	Permissions        []Permissions `json:"permissions"`
	LoginRedirectURL   string        `json:"loginRedirectUrl"`
	Email              string        `json:"email"`
	IPList             []interface{} `json:"ipList"`
	TokenExpireTime    int           `json:"tokenExpireTime"`
	TokenID            string        `json:"tokenId"`
	Level              int           `json:"level"`
	Sex                string        `json:"sex"`
	Groups             []Groups      `json:"groups"`
	AccountValidity    int           `json:"accountValidity"`
	AreaNameList       string        `json:"areaNameList"`
	RoleList           []RoleList    `json:"roleList"`
	UserName           string        `json:"userName"`
	UserID             int           `json:"userId"`
	DictList           []DictList    `json:"dictList"`
	Props              string        `json:"props"`
	UserPermissionList []interface{} `json:"userPermissionList"`
	AreaCodeList       string        `json:"areaCodeList"`
	OrganName          string        `json:"organName"`
	AreaCode           int           `json:"areaCode"`
	AreaID             string        `json:"areaId"`
	IcCard             string        `json:"icCard"`
	Phone              string        `json:"phone"`
	StationList        []StationList `json:"stationList"`
	ClientIP           string        `json:"clientIp"`
	TenantID           int           `json:"tenantId"`
	OrganID            int           `json:"organId"`
	ShieldID           string        `json:"shieldId"`
	DeptPermissionList []interface{} `json:"deptPermissionList"`
	OrganType          int           `json:"organType"`
	UserType           int           `json:"userType"`
	Account            string        `json:"account"`
}

// Sm2EncodeType_Response SM2加密类型信息响应结构体
type Sm2EncodeType_Response struct {
	Code    int           `json:"code"`
	Msg     string        `json:"msg"`
	Total   int           `json:"total"`
	Pages   int           `json:"pages"`
	Data    Sm2EncodeType `json:"data"`
	Success bool          `json:"success"`
}

// Sm2EncodeType SM2加密类型信息
type Sm2EncodeType struct {
	EncodeType string `json:"encodeType"`
	PublicKey  string `json:"publicKey"`
}
