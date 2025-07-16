package knife4jgo

import (
	"embed"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed knife4j

var knife4jFiles embed.FS
var K *Knife4jGo

type Knife4jGo struct {
	docs        *embed.FS
	ServiceJson interface{}
	SwaggerJson interface{}
}

// NewKnife4jGo 创建Knife4jGo对象
// 入参： 无
// 返回： *Knife4jGo Knife4jGo对象
func NewKnife4jGo() *Knife4jGo {
	k := Knife4jGo{}
	return &k
}

// SetServicesJsonFile 设置services.json文件
// 入参： efs *embed.FS services.json文件
// 返回： error
func (k *Knife4jGo) SetServicesJsonFile(efs *embed.FS) error {
	if efs == nil {
		return errors.New("embed.FS is nil")
	}
	var servicesInfo interface{}
	var data []byte
	data, err := fs.ReadFile(efs, "docs/services.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &servicesInfo)
	if err != nil {
		return err
	}
	if servicesInfo == nil {
		return errors.New("services.json is nil")
	}
	k.ServiceJson = &servicesInfo
	return nil
}

// SetSwaggerJsonFile 设置swagger.json文件
// 入参： efs *embed.FS swagger.json文件
// 返回： error
func (k *Knife4jGo) SetSwaggerJsonFile(efs *embed.FS) error {
	if efs == nil {
		return errors.New("embed.FS is nil")
	}
	var swaggerInfo interface{}
	var data []byte
	data, err := fs.ReadFile(efs, "docs/swagger.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &swaggerInfo)
	if err != nil {
		return err
	}
	if swaggerInfo == nil {
		return errors.New("swagger.json is nil")
	}
	// 处理go swag 1.16.3~2.0.3rc版本生成的多余allOf结构字段
	k.handlerAllOf(swaggerInfo)

	k.SwaggerJson = &swaggerInfo
	return nil
}

// GinServiceJsonHandler gin路由处理函数
// 入参： c *gin.Context 上下文
// 返回： 无
func (k *Knife4jGo) GinServiceJsonHandler(c *gin.Context) {
	c.JSON(http.StatusOK, k.ServiceJson)
	return
}

// GinSwaggerJsonHandler gin路由处理函数
// 入参： c *gin.Context 上下文
// 返回： 无
func (k *Knife4jGo) GinSwaggerJsonHandler(c *gin.Context) {
	c.JSON(http.StatusOK, k.SwaggerJson)
	return
}

// GinKnife4jGoHandler gin路由处理函数
// 入参： c *gin.Context 上下文
// 返回： 无
func (k *Knife4jGo) GinKnife4jGoHandler(c *gin.Context) {
	if c.Request.Method != "GET" {
		c.JSON(http.StatusNotFound, "404 page not found")
		return
	}
	if strings.HasSuffix(c.Request.URL.Path, "/services.json") {
		c.JSON(http.StatusOK, k.ServiceJson)
		return
	}
	if strings.HasSuffix(c.Request.URL.Path, "/swagger.json") {
		c.JSON(http.StatusOK, k.SwaggerJson)
		return
	}

	c.Request.URL.Path = strings.ReplaceAll(c.Request.URL.Path, "/knife4jgo", "knife4j") //strings.TrimPrefix(r.URL.Path, "/knife4jgo")
	http.FileServer(http.FS(knife4jFiles)).ServeHTTP(c.Writer, c.Request)

	return
}

// GinKnife4jGoNoRouteHandler gin路由处理函数
// 入参： c *gin.Context 上下文
// 返回： 无
func (k *Knife4jGo) GinKnife4jGoNoRouteHandler(c *gin.Context) {
	// 重定向到首页
	c.Redirect(http.StatusMovedPermanently, "/knife4jgo/doc.html")
}

// hanlderAllOf 处理allOf字段
// 入参： data interface{} 需要处理的json数据
// 返回： interface{} 处理后的json数据
func (k *Knife4jGo) handlerAllOf(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// 如果是 map，遍历所有键值对
		for key, value := range v {
			if key == "allOf" {
				// 如果键是 "allOf"，处理它
				allOfSlice, ok := value.([]interface{})
				if ok && len(allOfSlice) == 1 {
					allOfMap, ok := allOfSlice[0].(map[string]interface{})
					if ok {
						if ref, ok := allOfMap["$ref"]; ok {
							// 将 "$ref" 的值替换到当前层级
							v["$ref"] = ref
							delete(v, "allOf")
						}
					}
				}
			} else {
				// 递归处理嵌套的值
				v[key] = k.handlerAllOf(value)
			}
		}
	case []interface{}:
		// 递归处理每个元素
		for i, item := range v {
			v[i] = k.handlerAllOf(item)
		}
	}
	return data
}

