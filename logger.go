package gorm

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	o_gorm "gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	Reset       = "\033[0m"
	Red         = "\033[31;1m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35;1m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"
)

// GouefGormLogger upravuje výstup tak, aby duration byl přímo v logu
type GouefGormLogger struct {
	LogLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
	Colorful                  bool
	infoLog                   *log.Logger
	warnLog                   *log.Logger
	errLog                    *log.Logger
}

func NewGouefGormLogger(config LoggerConfig) *GouefGormLogger {
	var logLevel logger.LogLevel
	switch config.LogLevel {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Info
	}

	return &GouefGormLogger{
		LogLevel:                  logLevel,
		SlowThreshold:             config.SlowThreshold,
		IgnoreRecordNotFoundError: config.IgnoreRecordNotFoundError,
		ParameterizedQueries:      config.ParameterizedQueries,
		Colorful:                  config.Colorful,
		infoLog:                   log.New(os.Stdout, "", 0),
		warnLog:                   log.New(os.Stdout, "", 0),
		errLog:                    log.New(os.Stderr, "", 0),
	}
}

func (l *GouefGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *GouefGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.infoLog.Printf(msg, data...)
	}
}

func (l *GouefGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.warnLog.Printf(msg, data...)
	}
}

func (l *GouefGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.errLog.Printf(msg, data...)
	}
}

// Trace se volá po každém SQL dotazu a tady si poskládáme duration přesně podle sebe
func (l *GouefGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)

	sql, _ := fc()
	gormPrefix := "[GORM]"
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	sqlStr := "[SQL]: "
	ErrorStr := gormPrefix + "[ERROR]: "
	WarnStr := gormPrefix + "[WARN]: "
	InfoStr := gormPrefix + "[INFO]: "

	if l.Colorful {
		gormPrefix = White + "[GORM]" + Reset
		nowStr = Cyan + nowStr + Reset
		sqlStr = White + "[SQL]: " + Reset
		ErrorStr = gormPrefix + Red + "[ERROR]: " + Reset
		WarnStr = gormPrefix + Magenta + "[WARN]: " + Reset
		InfoStr = BlueBold + "[INFO]: " + Reset
	}

	prefix := gormPrefix

	switch {
	case err != nil && l.LogLevel >= logger.Error:
		if l.IgnoreRecordNotFoundError && err == o_gorm.ErrRecordNotFound {
			return
		}

		prefix = ErrorStr
		errStr := err.Error()
		if l.Colorful {
			errStr = Red + errStr + Reset
		}

		l.errLog.SetPrefix("\r\n" + nowStr + " " + prefix)
		l.errLog.Printf(sqlStr+"%s | %s | %s", sql, errStr, l.durationFormat(elapsed))

	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= logger.Warn:
		prefix = WarnStr

		l.warnLog.SetPrefix("\r\n" + nowStr + " " + prefix)
		l.warnLog.Printf(sqlStr+"%s | SLOW > %s | %s", sql, l.SlowThreshold, l.durationFormat(elapsed))

	case l.LogLevel >= logger.Info:
		prefix = InfoStr

		l.infoLog.SetPrefix("\r\n" + nowStr + " " + prefix)
		l.infoLog.Printf(sqlStr+"%s | %s", sql, l.durationFormat(elapsed))
	}
}

func (l *GouefGormLogger) durationFormat(elapsed time.Duration) string {
	durationPrefix := "duration: "
	duration := fmt.Sprintf("%s", elapsed)

	if l.Colorful {
		durationPrefix = BlueBold + "[DURATION]: " + Reset

		duration = Green + fmt.Sprintf("%s", elapsed) + Reset

		if l.SlowThreshold > 0 {
			beforeSlow := (l.SlowThreshold * 90) / 100

			if elapsed >= beforeSlow {
				duration = Yellow + fmt.Sprintf("%s", elapsed) + Reset
			}

			if elapsed > l.SlowThreshold {
				l.infoLog.Printf("elapse: %s, slowThreshold: %s, elapsed > slowThreshold", elapsed, l.SlowThreshold)
				duration = Red + fmt.Sprintf("%s", elapsed) + Reset
			}
		}
	}

	return durationPrefix + duration
}
