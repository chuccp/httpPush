package core

import (
	"github.com/chuccp/httpPush/config"
	"github.com/chuccp/httpPush/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
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
	VERSION = "0.2.1"
)

func initLogger(path string, consoleLevel string) (*zap.Logger, error) {
	writeFileCore, err := getFileLogWriter(path, consoleLevel)
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

func getFileLogWriter(path string, consoleLevel string) (zapcore.Core, error) {
	var level = zapcore.InfoLevel
	if strings.EqualFold(consoleLevel, "debug") {
		level = zapcore.DebugLevel
	}
	if strings.EqualFold(consoleLevel, "warn") {
		level = zapcore.WarnLevel
	}
	logger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, zapcore.AddSync(logger), level)
	return core, nil
}
func getStdoutLogWriter() zapcore.Core {
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, os.Stdout, zapcore.DebugLevel)
	return core
}

func (httpPush *HttpPush) Start() {
	logPath := httpPush.context.GetCfgStringDefault("log", "file.path", "push.log")
	consoleLevel := httpPush.context.GetCfgStringDefault("log", "console.level", "info")
	logger, err := initLogger(logPath, consoleLevel)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	httpPush.context.log = logger
	httpPush.context.rangeServer(func(server Server) {
		if s, ok := server.(IHttpServer); ok {
			s.init(httpPush.context)
			httpPush.context.RecoverGo(func() {
				err := s.start()
				if err != nil {
					httpPush.context.GetLog().Error("s.start", zap.Error(err))
				}
			})
		}
		server.Init(httpPush.context)
		httpPush.context.RecoverGo(func() {
			err := server.Start()
			if err != nil {
				httpPush.context.GetLog().Error("server.Start", zap.Error(err))
			}
		})
	})
	httpPush.context.SetSystemInfo("VERSION", VERSION)
	httpPush.context.RecoverGo(func() {
		err = httpPush.startHttpServer()
		if err != nil {
			httpPush.context.GetLog().Error("startHttpServer", zap.Error(err))
		}
	})
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGBUS)
	<-sig

}
