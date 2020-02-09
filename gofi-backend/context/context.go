package context

import (
	"flag"
	"fmt"
	"github.com/go-xorm/xorm"
	"gofi/env"
	"path/filepath"

	//import sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"gofi/models"
	"net"
	"os"
	"strings"
	"time"
)

//version ,will be replaced at compile time by [-ldflags="-X 'gofi/context.Version=vX.X.X'"]
var version = "UNKOWN VERSION"

const (
	//DefaultPort default port to listen Gofi监听的默认端口号
	DefaultPort = "8080"
	portUsage   = "port to expose web services"
	ipUsage     = "server side ip for web client to request,default is lan ip"
)

//Context 上下文对象
type Context struct {
	Port          string
	ServerAddress string
	ServerIP      string //ServerIP server side ip for web client to request,default is lan ip
	Orm           *xorm.Engine
}

var instance = new(Context)
var isFlagBind = false

func init() {
	bindFlags()
}

func bindFlags() {
	if !isFlagBind {
		flag.StringVar(&instance.Port, "port", DefaultPort, portUsage)
		flag.StringVar(&instance.Port, "p", DefaultPort, portUsage+" (shorthand)")
		flag.StringVar(&instance.ServerIP, "ip", "", ipUsage)
		isFlagBind = true
	}
}

//InitContext 初始化Context,只能初始化一次
func InitContext() {
	flag.Parse()

	// if ip is empty, obtain lan ip to instead.
	if instance.ServerIP == "" || !CheckIP(instance.ServerIP) {
		instance.ServerIP = instance.GetLanIP()
	}
	instance.ServerAddress = instance.ServerIP + ":" + instance.Port
	instance.Orm = instance.initDatabase()
}

//CheckIP 校验IP是否有效
func CheckIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

//Get 返回当前Context实例
func Get() *Context {
	return instance
}

func (context *Context) GetAppVersion() string {
	return version
}

func (context *Context) GetAppName() string {
	return "gofi"
}

//GetStorageDir 获取当前仓储目录
func (context *Context) GetStorageDir() string {
	configuration := context.QueryConfiguration()
	if len(configuration.CustomStoragePath) == 0 {
		return context.GetDefaultStorageDir()
	}
	return configuration.CustomStoragePath
}

func (context *Context) GetDefaultStorageDir() string {
	return filepath.Join(context.GetWorkDir(), "storage")
}

func (context *Context) GetDatabaseFilePath() string {
	return filepath.Join(context.GetWorkDir(), context.GetAppName()+".db")
}

func (context *Context) GetLogDir() string {
	return filepath.Join(context.GetWorkDir(), "log")
}

func (context *Context) initDatabase() *xorm.Engine {
	// connect to database
	engine, err := xorm.NewEngine("sqlite3", context.GetDatabaseFilePath())
	if err != nil {
		logrus.Println(err)
		panic("failed to connect database")
	}

	if context.IsTestEnvironment() {
		logrus.Info("on environment,skip database sync")
	} else {
		// migrate database
		if err := engine.Sync2(new(models.Configuration)); err != nil {
			logrus.Error(err)
		}
	}

	if env.IsDevelop() {
		engine.ShowSQL(true)
	}

	return engine
}

func (context *Context) QueryConfiguration() *models.Configuration {
	var config = new(models.Configuration)
	//obtain first record
	if has, _ := context.Orm.Get(config); !has {
		//create new record if there is no record exist
		config = &models.Configuration{
			AppPath:            context.GetWorkDir(),
			Initialized:        false,
			CustomStoragePath:  "",
			DefaultStoragePath: context.GetDefaultStorageDir(),
			DatabaseFilePath:   context.GetDatabaseFilePath(),
			LogDirectoryPath:   context.GetLogDir(),
			ThemeStyle:         "light", // light or dark
			NavMode:            "top",   // top or side
			Created:            time.Time{},
			Updated:            time.Time{},
		}

		if _, err := context.Orm.InsertOne(config); err != nil {
			logrus.Error(err)
		}
	}

	config.Version = context.GetAppVersion()
	config.AppPath = context.GetWorkDir()
	config.DefaultStoragePath = context.GetDefaultStorageDir()
	config.DatabaseFilePath = context.GetDatabaseFilePath()
	config.LogDirectoryPath = context.GetLogDir()

	return config
}

//IsTestEnvironment 当前是否测试环境
func (context *Context) IsTestEnvironment() bool {
	for _, value := range os.Args {
		if strings.Contains(value, "-test.v") {
			return true
		}
	}
	return false
}

//GetWorkDir 获取工作目录
func (context *Context) GetWorkDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

//GetLanIP 返回本地ip
func (context *Context) GetLanIP() string {
	addresses, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logrus.Infof("print all ip address: %v\n\t", addresses)

	for _, address := range addresses {
		ipNet, ok := address.(*net.IPNet)

		if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
			continue
		}

		// 当前ip属于私有地址,直接返回
		if isIpBelongToPrivateIpNet(ipNet.IP) {
			return ipNet.IP.To4().String()
		}
	}

	return "127.0.0.1"
}

// 某个ip是否属于私有网段
func isIpBelongToPrivateIpNet(ip net.IP) bool {
	for _, ipNet := range getInternalIpNetArray() {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

// 返回私有网段切片
func getInternalIpNetArray() []*net.IPNet {
	var ipNetArrays []*net.IPNet

	for _, ip := range []string{"192.168.0.0/16", "172.16.0.0/12", "10.0.0.0/8"} {
		_, ipNet, _ := net.ParseCIDR(ip)
		ipNetArrays = append(ipNetArrays, ipNet)
	}

	return ipNetArrays
}
