package usercenterm

type GetMenuListReq struct {
	TokenId string `json:"token_id"`
}

type GetMenuListResp struct {
	Code      int     `json:"code"`
	Msg       string  `json:"msg"`
	Data      []*Data `json:"data"`
	Total     int     `json:"total"`
	Timestamp int64   `json:"timestamp"`
	Success   bool    `json:"success"`
}
type Data struct {
	ID             int         `json:"id"`
	Sn             string      `json:"sn"`
	Index          string      `json:"index"`
	Title          string      `json:"title"`
	Icon           string      `json:"icon"`
	ParentID       int         `json:"parentId"`
	Subs           []*Subs     `json:"subs"`
	PermissionCode string      `json:"permissionCode"`
	Props          string      `json:"props"`
	SysType        string      `json:"sysType"`
	SystemType     string      `json:"systemType"`
	Status         int         `json:"status"`
	TopStatus      int         `json:"topStatus"`
	AuthFile       interface{} `json:"authFile"`
}
type Subs struct {
	ID             int         `json:"id"`
	Sn             string      `json:"sn"`
	Index          string      `json:"index"`
	Title          string      `json:"title"`
	Icon           string      `json:"icon"`
	ParentID       int         `json:"parentId"`
	Subs           []*Subs     `json:"subs"`
	PermissionCode string      `json:"permissionCode"`
	Props          string      `json:"props"`
	SysType        string      `json:"sysType"`
	SystemType     string      `json:"systemType"`
	Status         int         `json:"status"`
	TopStatus      int         `json:"topStatus"`
	AuthFile       interface{} `json:"authFile"`
}

// ToRoutes 将菜单响应转换为路由数组
func (resp *GetMenuListResp) ToRoutes() []Route {
	var routes []Route
	for _, item := range resp.Data {
		if item != nil {
			routes = append(routes, item.buildRoutes("", "")...)
		}
	}
	return routes
}

// buildRoutes 递归构建Data类型菜单项的路由
func (item *Data) buildRoutes(parentName, parentAddr string) []Route {
	if item == nil {
		return nil
	}

	var routes []Route

	// 拼接当前菜单项的名称和地址
	currentName := item.Title
	if parentName != "" {
		currentName = parentName + ">" + currentName
	}

	currentAddr := item.Index
	if parentAddr != "" {
		if item.Index != "" {
			currentAddr = parentAddr + "/" + currentAddr
		} else {
			currentAddr = parentAddr
		}
	}

	// 创建当前菜单项的路由
	routes = append(routes, Route{
		RouteName: currentName,
		RouteAddr: currentAddr,
	})

	// 处理子菜单项
	for _, sub := range item.Subs {
		if sub != nil {
			routes = append(routes, sub.buildRoutes(currentName, currentAddr)...)
		}
	}

	return routes
}

// buildRoutes 递归构建Subs类型菜单项的路由
func (item *Subs) buildRoutes(parentName, parentAddr string) []Route {
	if item == nil {
		return nil
	}

	var routes []Route

	// 拼接当前菜单项的名称和地址
	currentName := item.Title
	if parentName != "" {
		currentName = parentName + ">" + currentName
	}

	currentAddr := item.Index
	if parentAddr != "" {
		if item.Index != "" {
			currentAddr = parentAddr + "/" + currentAddr
		} else {
			currentAddr = parentAddr
		}
	}

	// 创建当前菜单项的路由
	routes = append(routes, Route{
		RouteName: currentName,
		RouteAddr: currentAddr,
	})

	// 处理子菜单项
	for _, sub := range item.Subs {
		if sub != nil {
			routes = append(routes, sub.buildRoutes(currentName, currentAddr)...)
		}
	}

	return routes
}

type Route struct {
	RouteName string `json:"routeName"`
	RouteAddr string `json:"routeAddr"`
}

// ToRoutes1 将菜单响应转换为路由数组（仅拼接Title，Index不拼接）
func (resp *GetMenuListResp) ToRoutes1() []Route {
	var routes []Route
	for _, item := range resp.Data {
		if item != nil {
			routes = append(routes, item.buildRoutes1("")...)
		}
	}
	return routes
}

// buildRoutes1 递归构建Data类型菜单项的路由（仅拼接Title）
func (item *Data) buildRoutes1(parentName string) []Route {
	if item == nil {
		return nil
	}

	var routes []Route

	// 拼接当前菜单项的名称（仅Title）
	currentName := item.Title
	if parentName != "" {
		currentName = parentName + ">" + currentName
	}

	// RouteAddr仅保留当前项的Index，不拼接父级
	currentAddr := item.Index

	// 创建当前菜单项的路由
	routes = append(routes, Route{
		RouteName: currentName,
		RouteAddr: currentAddr,
	})

	// 处理子菜单项（递归传递当前拼接的Title）
	for _, sub := range item.Subs {
		if sub != nil {
			routes = append(routes, sub.buildRoutes1(currentName)...)
		}
	}

	return routes
}

// buildRoutes1 递归构建Subs类型菜单项的路由（仅拼接Title）
func (item *Subs) buildRoutes1(parentName string) []Route {
	if item == nil {
		return nil
	}

	var routes []Route

	// 拼接当前菜单项的名称（仅Title）
	currentName := item.Title
	if parentName != "" {
		currentName = parentName + ">" + currentName
	}

	// RouteAddr仅保留当前项的Index，不拼接父级
	currentAddr := item.Index

	// 创建当前菜单项的路由
	routes = append(routes, Route{
		RouteName: currentName,
		RouteAddr: currentAddr,
	})

	// 处理子菜单项（递归传递当前拼接的Title）
	for _, sub := range item.Subs {
		if sub != nil {
			routes = append(routes, sub.buildRoutes1(currentName)...)
		}
	}

	return routes
}
