package core

import (
	"github.com/chuccp/httpPush/config"
	"github.com/chuccp/httpPush/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"sync"
)

type Register struct {
	servers *sync.Map
	config  *config.Config
}

func (register *Register) AddServer(server Server) {
	register.servers.LoadOrStore(server.Name(), server)
}
func (register *Register) Create() *HttpPush {
	context := register.getContext()
	return context.GetHttpPush()
}
func (register *Register) getContext() *Context {
	context := newContext(register)
	return context
}
func (register *Register) rangeServer(f func(server Server)) {
	register.servers.Range(func(key, value any) bool {
		f(value.(Server))
		return true
	})
}
func NewRegister(config *config.Config) *Register {
	return &Register{servers: new(sync.Map), config: config}
}

type HttpPush struct {
	context    *Context
	httpServer *util.HttpServer
}

func newHttpPush(context *Context) *HttpPush {
	return &HttpPush{httpServer: util.NewServer(), context: context}
}

func (httpPush *HttpPush) startHttpServer() error {
	port := httpPush.context.GetCfgInt("core", "http.port")
	certFile := httpPush.context.GetCfgString("core", "http.certFile")
	keyFile := httpPush.context.GetCfgString("core", "http.keyFile")
	httpPush.context.log.Info("startHttpServer", zap.String("name", "core"), zap.Int("port", port))
	err := httpPush.httpServer.StartAutoTLS(port, certFile, keyFile)
	if err != nil {
		httpPush.context.log.Error("服务启动失败", zap.String("name", "httpPush"), zap.Int("port", port), zap.Error(err))
		return err
	}
	return nil
}

const (
	VERSION = "0.1.10"
)

func initLogger(path string) (*zap.Logger, error) {
	writeFileCore, err := getFileLogWriter(path)
	if err != nil {
		return nil, err
	}
	core := zapcore.NewTee(writeFileCore, getStdoutLogWriter())
	return zap.New(core, zap.AddCaller()), nil
}

func getEncoder() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout(util.TimestampFormat)
	return zapcore.NewJSONEncoder(config)
}

func getFileLogWriter(path string) (zapcore.Core, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, zapcore.AddSync(file), zapcore.InfoLevel)
	return core, nil
}
func getStdoutLogWriter() zapcore.Core {
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, os.Stdout, zapcore.DebugLevel)
	return core
}

func (httpPush *HttpPush) Start() error {

	logPath := httpPush.context.GetCfgStringDefault("log", "file.path", "push.log")
	logger, err := initLogger(logPath)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	httpPush.context.log = logger
	httpPush.context.rangeServer(func(server Server) {
		if s, ok := server.(IHttpServer); ok {
			s.init(httpPush.context)
			go func() {
				err := s.start()
				if err != nil {
					log.Panic(err)
				}
			}()
		}
		server.Init(httpPush.context)
		go func() {
			err := server.Start()
			if err != nil {
				log.Panic(err)
			}
		}()
	})
	httpPush.context.SetSystemInfo("VERSION", VERSION)
	return httpPush.startHttpServer()
}
