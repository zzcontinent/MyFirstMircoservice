//调用pb.go接口，以客户端身份测试MyFirstMicroservice微服务
package main

import (
	"errors"
	su "git.woda.ink/woda/common/service_util" //实际引入路径为"git.woda.ink/woda/common/service_util"，请先go get安装
	pb "git.woda.ink/woda/pb/MyFirstMicroservice"
	pbdb "git.woda.ink/woda/pb/DBHelper"
	log "github.com/xiaomi-tc/log15"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	logx "log"
	"strconv"
	pbaas "git.woda.ink/woda/pb/AASMessage"
	//"fmt"
)

func main() {
	//1 查询consul获取DBHelper名称
	s, _ := su.NewServiceUtil("stepxx", "C:/etc/VCodeManager.yaml")
	sList := s.GetServiceByName("MyFirstMicroservice", "")
	if len(sList) == 0 {
		log.Error("GetServiceByName result count = 0")
		return
	}
	connErr := errors.New("未连接")
	var connService *grpc.ClientConn
	//过滤已经失效的服务列表
	for _, value := range sList {
		strAddress := value.IP + ":" + strconv.Itoa(value.Port)
		log.Debug("GetServiceByName get service", "strAddress", strAddress)
		connService, connErr = grpc.Dial(strAddress, grpc.WithInsecure())
		if connErr != nil {
			log.Error("grpc.Dial error...", "error")
			continue
		}
	}
	//获取服务成功
	if connErr == nil {
		//最后关闭服务
		defer connService.Close()
		//test1 : DB Interface
		c := pb.NewMyFirstMicroserviceClient(connService)
		req := new(pbdb.DBHelperRequest)
		req.ExeType = pbdb.SqlExcuteType_SELECT
		req.SqlCMD = "WD_RCUT_GetContactList"
		r, err := c.WD_MFMS_ShowDB(context.Background(), &pbaas.InvokeServiceRequest{})
		if err != nil{
			log.Error("could not DBExecute: %v", err)
		}else{
			logx.Printf("ServiceResponseData: %s", r.GetServiceResponseData())
		}

		//test2 : Consul interface
		//c := pb.NewMyFirstMicroserviceClient(connService)
		//r, err := c.WD_MFMS_ShowAllMicroservicesOnConsul(context.Background(), &pbaas.InvokeServiceRequest{})
		//if err != nil{
		//	log.Error("Show service on consul error: %v", err)
		//}else{
		//	logx.Printf("ServiceResponseData: %s", r.GetServiceResponseData())
		//}
	}

}
