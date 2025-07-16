/* ComLinkToDataSourceStateChanged_message
   消息定义：到数据源的连接状态发生变化（包括间接连接）
   hws 2022-11-23
*/

package socsws

import (
	"fmt"
	"github.com/freedqo/fmc-go-agent/pkg/umsg"
)

// 创建用于推送一个到数据源的连接状态发生变化消息（用于服务端）
func NewMessageForComLinkToDataSourceStateChanged(clientID string, oldState, newState LinkToDataSourceState, linkDataSources DataSourceID, desc string) umsg.Message {
	dss := []DataSourceID{linkDataSources}
	return NewMessageForComLinkToDataSourceStateChanged2(clientID, oldState, newState, dss, desc)
}

// 创建用于推送一个到数据源的连接状态发生变化消息（用于服务端）
func NewMessageForComLinkToDataSourceStateChanged2(clientID string, oldState LinkToDataSourceState, newState LinkToDataSourceState, linkDataSources []DataSourceID, desc string) umsg.Message {
	operateData := NewPushComLinkToDataSourceStateChangedOperateData(oldState, newState, linkDataSources, desc)
	return umsg.NewMessageForSend(clientID, string("MOT_COM_LinkToDataSourceStateChanged"), fmt.Sprintf("%T", operateData), operateData)
}

// 构建推送一个到数据源的连接状态发生变化消息的操作数据
func NewPushComLinkToDataSourceStateChangedOperateData(oldState, newState LinkToDataSourceState, linkDataSources []DataSourceID, desc string) TPushComLinkToDataSourceStateChanged_OperateData {
	return TPushComLinkToDataSourceStateChanged_OperateData{OldState: oldState,
		NewState:        newState,
		LinkDataSources: linkDataSources,
		Desc:            desc}
}

// 连接到数据源的状态的数据类型
type LinkToDataSourceState int

// 连接到数据源的状态的数据类型的常量定义
const (
	LDS_State_Unknow       LinkToDataSourceState = -1 //未知状态(已断开、已连接)
	LDS_State_Disconnected LinkToDataSourceState = 0  //已断开
	LDS_State_Connected    LinkToDataSourceState = 1  //已连接
)

// 数据源名称的数据类型
type DataSourceID string

// 数据源名称数据类型的常量（已知数据源定义）
const (
	DS_MainSrvWs         DataSourceID = "DS_MainSrvWs"         //GO DS_MainSrvWs wss
	DS_ApzSrvWs          DataSourceID = "DS_ApzSrvWs"          //GO DS_ApzSrvWs wss
	DS_EboardSrvWs       DataSourceID = "DS_EboardSrvWs"       //GO DS_EboardSrvWs wss
	DS_CLDService        DataSourceID = "DS_CLDService"        //车辆段车辆设备实时库服务
	DS_CLDWFService      DataSourceID = "DS_CLDWFService"      //车辆段五防设备实时库服务
	DS_SocsBusiness      DataSourceID = "DS_SocsBusiness"      //SOCs业务服务
	DS_fromDcbmService   DataSourceID = "DS_fromDcbmService"   //DCBM服务
	DS_QxdService        DataSourceID = "DS_QxdService"        //请销点系统服务服务
	DS_OcrService        DataSourceID = "DS_OcrService"        //车号识别系统服务
	DS_AcsService        DataSourceID = "DS_AcsService"        //门禁系统服务
	DS_DataCenterService DataSourceID = "DS_DataCenterService" //数据中心系统服务
	DS_AcqService        DataSourceID = "DS_AcqService"        //数采中心系统服务
	DS_MQTT              DataSourceID = "DS_MQTT"              //DS_MQTT
)

// 推送一个到数据源的连接状态发生变化消息的数据结构体（utwebsocket）
type TPushComLinkToDataSourceStateChanged_OperateData struct {
	OldState        LinkToDataSourceState `json:"OldState"`        //旧状态
	NewState        LinkToDataSourceState `json:"NewState"`        //新状态
	LinkDataSources []DataSourceID        `json:"LinkDataSources"` //连接的数据源列表
	Desc            string                `json:"Desc"`            //描述信息
}
