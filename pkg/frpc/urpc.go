package frpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
	"sync"
)

// URpc 类似Gin的Engine
type URpc struct {
	Groups     []*Group
	Rpc        *rpc.Server
	Registered map[string]interface{}
}

// Group 服务组
type Group struct {
	Name     string
	Services []*ServiceObjectInfo
	Groups   []*Group
	Parent   *Group
	Engine   *URpc
}

// ServiceObjectInfo 用于注册的服务对象信息
type ServiceObjectInfo struct {
	Name          string      //用于注册的服务名称
	ServiceObject interface{} //用于发布的服务对象接口
}

// New 创建一个新的URpc实例
func New() *URpc {
	return &URpc{
		Groups:     make([]*Group, 0),
		Rpc:        rpc.NewServer(),
		Registered: make(map[string]interface{}, 0),
	}
}

// Group 方法用于创建一个新的服务组，并支持注入服务
func (e *URpc) Group(name string, services ...interface{}) (*Group, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("路由组名称不能为空")
	}
	group := Group{
		Name:     name,
		Services: make([]*ServiceObjectInfo, 0),
		Groups:   make([]*Group, 0),
		Parent:   nil,
		Engine:   e,
	}
	e.Groups = append(e.Groups, &group)
	for _, service := range services {
		fullServiceName := group.fullServiceName("")
		if _, exit := e.Registered[fullServiceName]; exit {
			return nil, errors.New("duplicate service name: " + fullServiceName)
		}
		err := group.Engine.Rpc.RegisterName(fullServiceName, service)
		if err != nil {
			return nil, err
		}
		group.Services = append(group.Services, &ServiceObjectInfo{
			Name:          fullServiceName,
			ServiceObject: service,
		})
		e.Registered[fullServiceName] = service
	}
	return e.Groups[len(e.Groups)-1], nil
}

// RegisterService 方法用于在服务组中注册服务，并将服务注册到 Rpc 服务器中
func (g *Group) RegisterService(name string, service interface{}) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("方法组名称不能为空")
	}
	fullServiceName := g.fullServiceName(name)
	if _, exit := g.Engine.Registered[fullServiceName]; exit {
		return errors.New("duplicate service name: " + fullServiceName)
	}
	g.Services = append(g.Services, &ServiceObjectInfo{
		Name:          fullServiceName,
		ServiceObject: service,
	})
	err := g.Engine.Rpc.RegisterName(fullServiceName, service)
	if err != nil {
		return err
	}
	g.Engine.Registered[fullServiceName] = service
	return nil
}

// fullServiceName 生成完整的服务名称
func (g *Group) fullServiceName(serviceName string) string {

	var parts []string
	currentGroup := g
	for currentGroup != nil {
		parts = append([]string{currentGroup.Name}, parts...)
		currentGroup = currentGroup.Parent
	}
	if serviceName != "" {
		parts = append(parts, serviceName)
	}
	return strings.Join(parts, ".")
}

// Group 方法用于在当前组下创建一个子服务组，并支持注入服务
func (g *Group) Group(name string, services ...interface{}) (*Group, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("路由组名称不能为空")
	}
	group := Group{
		Name:     name,
		Services: make([]*ServiceObjectInfo, 0),
		Groups:   make([]*Group, 0),
		Parent:   nil,
		Engine:   g.Engine,
	}
	g.Groups = append(g.Groups, &group)
	for _, service := range services {
		fullServiceName := group.fullServiceName("")
		if _, exit := g.Engine.Registered[fullServiceName]; exit {
			return nil, errors.New("duplicate service name: " + fullServiceName)
		}
		err := group.Engine.Rpc.RegisterName(fullServiceName, service)
		if err != nil {
			return nil, err
		}
		group.Services = append(group.Services, &ServiceObjectInfo{
			Name:          fullServiceName,
			ServiceObject: service,
		})
		g.Engine.Registered[fullServiceName] = service
	}
	return g.Groups[len(g.Groups)-1], nil
}

// HttpResWriterWrapper 自定义一个包装器实现 gin.ResponseWriter 并重写 Write 方法
type HttpResWriterWrapper struct {
	body []byte
	gin.ResponseWriter
}

// 重写 Write 方法，保存响应体
func (w *HttpResWriterWrapper) Write(data []byte) (int, error) {
	w.body = data
	return len(w.body), nil
}

func (e *URpc) GinJsonHandler(c *gin.Context) {
	// 读取并复制请求体
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Result: nil,
			Error:  err.Error(),
			Id:     http.StatusInternalServerError,
		})
		return
	}
	req := Request{}
	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Result: nil,
			Error:  err.Error(),
			Id:     http.StatusInternalServerError,
		})
		return
	}
	// 恢复请求体，以便后续处理
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	inWriter := HttpResWriterWrapper{
		body:           []byte{},
		ResponseWriter: c.Writer,
	}
	//转调RPC接口
	var conn io.ReadWriteCloser = struct {
		io.Writer
		io.ReadCloser
	}{
		ReadCloser: c.Request.Body,
		Writer:     &inWriter,
	}
	var res Response
	err = e.Rpc.ServeRequest(jsonrpc.NewServerCodec(conn))
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Result: nil,
			Error:  err.Error(),
			Id:     req.Id,
		})
		return
	} else {
		err1 := json.Unmarshal(inWriter.body, &res)
		if err1 != nil {
			c.JSON(http.StatusOK, Response{
				Result: nil,
				Error:  err1.Error(),
				Id:     req.Id,
			})
			return
		}
		c.JSON(http.StatusOK, res)
		return
	}
}

// TCPServerResponseWriterWrapper 自定义一个包装器实现 io.Writer 并重写 Write 方法
type TCPServerResponseWriterWrapper struct {
	body []byte
	io.Writer
}

// 重写 Write 方法，保存响应体
func (w *TCPServerResponseWriterWrapper) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return len(data), nil
}

func (e *URpc) TcpServerJsonHandler(conn net.Conn) {
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	inWriter := &TCPServerResponseWriterWrapper{
		body:   []byte{},
		Writer: conn,
	}
	// 转调 RPC 接口
	var rpcConn io.ReadWriteCloser = struct {
		io.Writer
		io.ReadCloser
	}{
		ReadCloser: conn,
		Writer:     inWriter,
	}
	var res Response
	err := e.Rpc.ServeRequest(jsonrpc.NewServerCodec(rpcConn))
	if err != nil {
		res = Response{
			Result: nil,
			Error:  err,
			Id:     500,
		}
	} else {
		err := json.Unmarshal(inWriter.body, &res)
		if err != nil {
			res = Response{
				Result: nil,
				Error:  err,
				Id:     500,
			}
		}
	}
	responseJSON, err := json.Marshal(res)
	if err != nil {
		_, err := conn.Write([]byte(
			`{
				"result":null,
				"error":"json marshal error",
				"id":500
			}`))
		if err != nil {
			return
		}
		return
	}
	_, err = conn.Write(responseJSON)
	if err != nil {
		fmt.Printf("Write response error: %v\n", err)
	}
}
func (e *URpc) TcpServerGobHandler(ctx context.Context, wg *sync.WaitGroup, conn net.Conn) {
	defer wg.Done()
	defer conn.Close()
	go func() {
		e.Rpc.ServeConn(conn)
	}()
	<-ctx.Done()
}

type Request struct {
	Method string        `json:"method"` //请求方法名
	Params []interface{} `json:"params"` //请求参数（数组内任意字段）
	Id     interface{}   `json:"id"`     //请求id
}
type Response struct {
	Result interface{} `json:"result"` //请求结果
	Error  interface{} `json:"error"`  //请求错误
	Id     interface{} `json:"id"`     //请求id
}
