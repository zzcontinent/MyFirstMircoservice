package main

import (
	"flag"
	su "git.woda.ink/woda/common/service_util"
	pbmyfirst "git.woda.ink/woda/pb/MyFirstMicroservice"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/xiaomi-tc/log15"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strconv"
	"syscall"
)

var logPath, logLevel, YamlPath string
var serviceName, packageName string = "MyFirstMicroservice", "MyFirstMicroservice"
var sumanager *su.ServiceUtil
var srvPort int
var listen net.Listener

func waitToExit(srvId string) {
	c := make(chan os.Signal)
	signal.Notify(c)
	for {
		s := <-c
		log.Error("waitToExit", "get signal", s)
		if s == syscall.SIGINT || s == syscall.SIGKILL || s == syscall.SIGTERM {
			sumanager.UnRegistService(srvId)
			//db.Close()
			listen.Close()
			os.Exit(1)
		}
	}

}

func checkErr(err error, errMessage string, isQuit bool) {
	if err != nil {
		log.Error(errMessage, "error", err)
		if isQuit {
			//db.Close()
			listen.Close()
			os.Exit(1)
		}
	}
}

func loadInitConfig(yamlPath string) (map[string]interface{}, error) {
	m := make(map[string]interface{})

	data, err := ioutil.ReadFile(yamlPath)
	checkErr(err, "ReadFile error", false)

	err = yaml.Unmarshal([]byte(data), &m)
	checkErr(err, "yamlConfig Unmarshal error", true)

	if runtime.GOOS == "windows" {
		logPath = "c:\\log\\woda\\" + serviceName + "\\"
	} else {
		logPath = "/var/log/woda/" + serviceName + "/"
	}

	var filenameWithSuffix = su.GetFileName(yamlPath)
	log.Debug("GetFileName", "name", filenameWithSuffix)

	logPath = logPath + filenameWithSuffix + ".log"

	return m, err

}

func registerToConsul(svName, pgName string) string {
	svId, err := sumanager.RegistService(svName, pgName)
	if err != nil {
		checkErr(err, "registerToConsul", true)
	} else {
		checkErr(err, "registerToConsul", false)
	}

	return svId
}

func init() {
	// 一个服务多个配置的情况有待处理
	var project_conf string
	if runtime.GOOS == "windows" {
		project_conf = "c:\\etc\\woda\\modules\\" + serviceName + "\\" + serviceName + ".yaml"
	} else {
		project_conf = "/etc/woda/modules/" + serviceName + "/" + serviceName + ".yaml"
	}
	flag.StringVar(&YamlPath, "c", project_conf, "config_file path,example /etc/woda/MyFirstMicroservice.yaml")
	flag.StringVar(&logLevel, "l", "debug", "log level: debug, info, error")
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Crit("main", "server crash: ", err)
			log.Crit("main", "stack: ", string(debug.Stack()))
		}
	}()

	flag.Parse()
	_, err := loadInitConfig(YamlPath)

	if err != nil {
		checkErr(err, "loadInitConfig error", true)
	}

	sumanager, srvPort = su.NewServiceUtil(serviceName, YamlPath)

	serviceId := registerToConsul(serviceName, packageName)
	//监听信号，退出时，反向注销服务
	go waitToExit(serviceId)

	// Open log file
	h, _ := log.FileHandler(logPath, log.LogfmtFormat())
	log.Root().SetHandler(h)

	switch logLevel {
	case "debug":
		log.SetOutLevel(log.LvlDebug)
	case "info":
		log.SetOutLevel(log.LvlInfo)
	case "error":
		log.SetOutLevel(log.LvlError)
	default:
		log.SetOutLevel(log.LvlDebug)
	}
	//log.SetOutLevel(log.LvlError)
	log.Debug("start...", "serviceid", serviceId, "port", srvPort)

	listen, err = net.Listen("tcp", ":"+strconv.Itoa(srvPort))
	if err != nil {
		log.Error("tcp port error", "port", srvPort)
	}

	s := grpc.NewServer(
		//grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		//数据上报
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	//  注册服务
	pbmyfirst.RegisterMyFirstMicroserviceServer(s, &Server{})

	//数据上报
	grpc_prometheus.Register(s)
	grpc_prometheus.EnableHandlingTimeHistogram()

	reflection.Register(s)
	if err := s.Serve(listen); err != nil {
		log.Error("ServiceStartFailed", "error", err)
	}
}
