package services

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogEntry struct {
	UserID        string    `json:"user_id"`
	DeviceIP      string    `json:"device_ip"`
	LatLong       string    `json:"lat_long"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	RequestMethod string    `json:"request_method"`
	RequestURI    string    `json:"request_uri"`
	UserAgent     string    `json:"user_agent"`
	RequestID     string    `json:"request_id"`
	ResponseSize  int       `json:"response_size"`
	RequestHost   string    `json:"request_host"`
	Message       string    `json:"msg"`
}

type LoggerService interface {
	Debug(entry LogEntry)
	Info(entry LogEntry)
	Warn(entry LogEntry)
	Error(entry LogEntry)
	WithFields(fields map[string]interface{}) LoggerService
}

type LoggerConfig struct {
	AWSAccessKey  string
	AWSSecretKey  string
	AWSRegion     string
	BucketName    string
	UploadToS3    bool
	LogDirectory  string
	LogFilePrefix string
}

type logger struct {
	logger     *zap.Logger
	s3Uploader *s3manager.Uploader
	config     LoggerConfig
	fields     map[string]interface{}
}

// func NewLoggerService(config LoggerConfig) LoggerService {
// 	// Ensure the logs directory exists
// 	if _, err := os.Stat(config.LogDirectory); os.IsNotExist(err) {
// 		os.Mkdir(config.LogDirectory, 0755)
// 	}

// 	loggerR, _ := zap.NewProduction()

// 	var uploader *s3manager.Uploader
// 	if config.UploadToS3 {
// 		sess := session.Must(session.NewSession(&aws.Config{
// 			Region:      aws.String(config.AWSRegion),
// 			Credentials: credentials.NewStaticCredentials(config.AWSAccessKey, config.AWSSecretKey, ""),
// 		}))
// 		uploader = s3manager.NewUploader(sess)
// 	}

// 	return &logger{
// 		logger:     loggerR,
// 		s3Uploader: uploader,
// 		config:     config,
// 		fields:     make(map[string]interface{}),
// 	}
// }

// NewLoggerService creates a new instance of LoggerService with file and S3 support
func NewLoggerService(config LoggerConfig) LoggerService {
	// Ensure the logs directory exists
	if _, err := os.Stat(config.LogDirectory); os.IsNotExist(err) {
		os.Mkdir(config.LogDirectory, 0755)
	}

	// Set up file logging
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // or another encoder if desired
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)

	// Create a file core
	fileLogPath := config.LogDirectory + "/" + config.LogFilePrefix + ".log"
	file, err := os.Create(fileLogPath)
	if err != nil {
		panic(err)
	}
	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(file), zapcore.DebugLevel)

	// Create the main logger
	loggerr := zap.New(fileCore, zap.AddCaller())

	var uploader *s3manager.Uploader
	if config.UploadToS3 {
		sess := session.Must(session.NewSession(&aws.Config{
			Region:      aws.String(config.AWSRegion),
			Credentials: credentials.NewStaticCredentials(config.AWSAccessKey, config.AWSSecretKey, ""),
		}))
		uploader = s3manager.NewUploader(sess)
	}

	return &logger{
		logger:     loggerr,
		s3Uploader: uploader,
		config:     config,
		fields:     make(map[string]interface{}),
	}
}

func (l *logger) logWithLevel(entry LogEntry, level zapcore.Level) {
	fields := []zap.Field{
		zap.String("user_id", entry.UserID),
		zap.String("device_ip", entry.DeviceIP),
		zap.String("lat_long", entry.LatLong),
		zap.Time("start_time", entry.StartTime),
		zap.Time("end_time", entry.EndTime),
		zap.String("request_method", entry.RequestMethod),
		zap.String("request_uri", entry.RequestURI),
		zap.String("user_agent", entry.UserAgent),
		zap.String("request_id", entry.RequestID),
		zap.Int("response_size", entry.ResponseSize),
		zap.String("request_host", entry.RequestHost),
	}

	for k, v := range l.fields {
		fields = append(fields, zap.Any(k, v))
	}

	switch level {
	case zap.DebugLevel:
		l.logger.Debug(entry.Message, fields...)
	case zap.InfoLevel:
		l.logger.Info(entry.Message, fields...)
	case zap.WarnLevel:
		l.logger.Warn(entry.Message, fields...)
	case zap.ErrorLevel:
		l.logger.Error(entry.Message, fields...)
	}
}

func (l *logger) Debug(entry LogEntry) {
	l.logWithLevel(entry, zap.DebugLevel)
}

func (l *logger) Info(entry LogEntry) {
	l.logWithLevel(entry, zap.InfoLevel)
}

func (l *logger) Warn(entry LogEntry) {
	l.logWithLevel(entry, zap.WarnLevel)
}

func (l *logger) Error(entry LogEntry) {
	l.logWithLevel(entry, zap.ErrorLevel)
}

func (l *logger) WithFields(fields map[string]interface{}) LoggerService {
	newLogger := &logger{
		logger:     l.logger,
		s3Uploader: l.s3Uploader,
		config:     l.config,
		fields:     fields,
	}
	return newLogger
}
