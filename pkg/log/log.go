package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

const (
	BASE_PATH          string = "/var/log/grp"
	BGP_PATH           string = "/var/log/grp/bgp"
	RIP_PATH           string = "/var/log/grp/rip"
	ROUTE_MANAGER_PATH string = "/var/log/grp/route"
)

type Log struct {
	Level int    `json:"level" yaml:"level"`
	Out   string `json:"out,omitempty" yaml:"out,omitempty"`
}

type Logger interface {
	Info(string, ...any)
	Warn(string, ...any)
	Err(string, ...any)
	SetProtocol(protocol string)
	Set(key, value string)
	With() Logger
	Level() Level
	Path() string
	GetLogger() *zerolog.Logger
}

type logger struct {
	lg       zerolog.Logger
	level    Level
	out      io.Writer
	path     string
	protocol string
}

type Level uint8

const (
	NoLog Level = iota
	Info  Level = iota
	Warn  Level = iota
	Error Level = iota
)

func (l Level) String() string {
	switch l {
	case NoLog:
		return "NoLog"
	case Info:
		return "Info"
	case Warn:
		return "Warn"
	case Error:
		return "Error"
	default:
		return "Unknown"
	}
}

func New(level Level, out string) (Logger, error) {
	if level > Error {
		return nil, fmt.Errorf("New logger: Invalid log level: %d", level)
	}
	l := &logger{
		level: level,
		path:  out,
	}
	switch out {
	case "stdout":
		l.out = os.Stdout
	case "stderr":
		l.out = os.Stderr
	case "":
		l.out = ioutil.Discard
	default:
		if err := prepareLogDir(); err != nil {
			return nil, err
		}
		ok, err := filepath.Match(BASE_PATH+"/*", out)
		if err != nil {
			return nil, fmt.Errorf("New logger: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("New logger: Invalid output path %s.\n  Output path must be under %s", out, BASE_PATH)
		}
		file, err := os.OpenFile(out, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
		if err != nil {
			return nil, fmt.Errorf("New logger: %w", err)
		}
		l.out = file
	}
	l.lg = zerolog.New(l.out).With().Timestamp().Logger()
	return l, nil
}

func (l *logger) Info(format string, v ...any) {
	if l.level < Warn && l.level > NoLog {
		l.lg.Info().Msgf(format, v...)
	}
}

func (l *logger) Warn(format string, v ...any) {
	if l.level < Error && l.level > NoLog {
		l.lg.Warn().Msgf(format, v...)
	}
}

func (l *logger) Err(format string, v ...any) {
	if l.level > NoLog {
		l.lg.Error().Msgf(format, v...)
	}
}

func (l *logger) SetProtocol(protocol string) {
	l.protocol = protocol
	l.lg = l.lg.With().Str("protocol", protocol).Logger()
}

func (l *logger) Set(key, value string) {
	l.lg = l.lg.With().Str(key, value).Logger()
}

func (l *logger) With() Logger {
	return &logger{
		level:    l.level,
		out:      l.out,
		protocol: l.protocol,
		lg:       zerolog.New(l.out).With().Timestamp().Logger().With().Str("protocol", l.protocol).Logger(),
	}
}

func (l *logger) GetLogger() *zerolog.Logger {
	return &l.lg
}

func (l *logger) Level() Level {
	return l.level
}

func (l *logger) Path() string {
	return l.path
}

func prepareLogDir() error {
	return os.MkdirAll(BASE_PATH, 0644)
}
