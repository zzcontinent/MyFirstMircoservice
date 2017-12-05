package main

import (
	//"encoding/json"
	//"fmt"
	"git.woda.ink/woda/pb/AASMessage"
	"github.com/xiaomi-tc/log15"
	//"log"
	"golang.org/x/net/context"
	"git.woda.ink/woda/common/service_util"
	"git.woda.ink/woda/pb/DBHelper"
	"google.golang.org/grpc"
	"errors"
	"strconv"
	"github.com/hashicorp/consul/api"
)

type Server struct{}

//proto的接口实现---------------------------------------------start
func (s *Server) WD_MFMS_ShowAllMicroservicesOnConsul(ctx context.Context, in *AASMessage.InvokeServiceRequest) (*AASMessage.InvokeServiceResponse, error) {
	//1 创建调用对象
	su,_ := service_util.NewServiceUtil("stepxx", "C:/etc/VCodeManager.yaml")
	//2 调用目录查询接口
	val,err := su.GetConsulServices()
	//3 格式化返回数据
	out := AASMessage.InvokeServiceResponse{}
	out.ServiceResponseData = "\n"
	for k,v:= range val{//map[i]
		out.ServiceResponseData += k + " : "
		for _,v1 := range v{
			out.ServiceResponseData += v1
		}
		out.ServiceResponseData += "\n"
	}
	chks,err :=su.GetServicesHealthyState(api.HealthAny)
	out.ServiceResponseData = "\n"
	for _,v:= range chks{//map[i]
			out.ServiceResponseData += v.Status +": " + v.CheckID + "\n"
			//out.ServiceResponseData += v.ServiceName + "2\n"
			//out.ServiceResponseData += v.Status + "3\n"
			//out.ServiceResponseData += v.ServiceTags[0] + "\n"
			//out.ServiceResponseData += v.ServiceID + "\n"
			//out.ServiceResponseData += v.Notes + "\n"
			//out.ServiceResponseData += v.Node + "\n"
			//out.ServiceResponseData += v.Name + "\n"
	}
	if err != nil {
		return &AASMessage.InvokeServiceResponse{err.Error()}, nil
	}
	return &out, nil
}
func (s *Server) WD_MFMS_ShowDB(ctx context.Context, in *AASMessage.InvokeServiceRequest) (*AASMessage.InvokeServiceResponse,error) {
	//1 通过grpc连接consul获取DBHelper
	//1.1创建对象工具
	su,_ := service_util.NewServiceUtil("stepxx", "C:/etc/VCodeManager.yaml")
	//1.2获取“DBHelper”微服务信息列表
	sList := su.GetServiceByName("DBHelper", "")
	if len(sList) == 0 {
		log15.Error("GetServiceByName result count = 0")
		//return result,errors.New("GetServiceByName result count = 0")
	}
	connErr := errors.New("未连接")
	//1.3 通过gprc建立与微服务器的客户端连接
	var connService *grpc.ClientConn
	for _, value := range sList {
		strAddress := value.IP + ":" + strconv.Itoa(value.Port)
		log15.Debug("GetServiceByName get service", "strAddress", strAddress)
		connService, connErr = grpc.Dial(strAddress, grpc.WithInsecure())

		if connErr != nil {
			log15.Error("grpc.Dial error...", "error")
			//u.ReportFailedService(serviceName,value.ServiceID)
			continue
		}
	}
	if connErr == nil {
		defer connService.Close()
	}
	c := DBHelper.NewDBHelperClient(connService)

	req := new(DBHelper.DBHelperRequest)
	req.ExeType = DBHelper.SqlExcuteType_SELECT
	req.SqlCMD = "WD_RCUT_GetContactList"


	r, err := c.DBExecute(context.Background(), req)
	if err != nil {
		log15.Error("could not DBExecute: %v", err)
	}

	//log.Printf("ServiceResponseData: %s", r.IsSuccess)
	//log.Printf("ServiceResponseData: %s", r.EffectRows)
	//log.Printf("ServiceResponseData: %s", r.Records)

	out := AASMessage.InvokeServiceResponse{}
	for _,v := range r.Records{
		out.ServiceResponseData += v.String() + "\n"
	}

	//return
	if err != nil {
		return &AASMessage.InvokeServiceResponse{err.Error()}, nil
	}
	return &out, nil
}
//proto的接口实现---------------------------------------------end




