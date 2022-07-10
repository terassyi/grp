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
	BASE_PATH string = "/var/log/grp"
)

type Logger interface {
	Info(string, ...any)
	Warn(string, ...any)
	Err(string, ...any)
	SetProtocol(protocol string)
	Set(key, value string)
	With() Logger
}

type logger struct {
	zerolog.Logger
	level    Level
	out      io.Writer
	protocol string
}

type Level uint8

const (
	NoLog Level = iota
	Info  Level = iota
	Warn  Level = iota
	Error Level = iota
)

func New(level Level, out string) (Logger, error) {
	if level > Error {
		return nil, fmt.Errorf("New logger: Invalid log level.")
	}
	l := &logger{
		level: level,
	}
	switch out {
	case "stdout":
		l.out = os.Stdout
	case "stderr":
		l.out = os.Stderr
	case "":
		l.out = ioutil.Discard
	default:
		ok, err := filepath.Match(BASE_PATH + "/*", out)
		if err != nil {
			return nil, fmt.Errorf("New logger: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("New logger: Invalid output path %s.\n  Output path must be under %s", out, BASE_PATH)
		}
		file, err := os.OpenFile(out, os.O_CREATE|os.O_RDWR, 0664)
		if err != nil {
			return nil, fmt.Errorf("New logger: %w", err)
		}
		l.out = file
	}
	// output := zerolog.ConsoleWriter{Out: l.out, TimeFormat: "2006-01-02 15:04:050"}
	// l.Logger = zerolog.New(output).With().Timestamp().Logger()
	l.Logger = zerolog.New(l.out).With().Timestamp().Logger()
	return l, nil
}

func (l *logger) Info(format string, v ...any) {
	if l.level < Warn && l.level > NoLog {
		l.Logger.Info().Msgf(format, v...)
	}
}

func (l *logger) Warn(format string, v ...any) {
	if l.level < Error && l.level > NoLog {
		l.Logger.Warn().Msgf(format, v...)
	}
}

func (l *logger) Err(format string, v ...any) {
	if l.level > NoLog {
		l.Logger.Error().Msgf(format, v...)
	}
}

func (l *logger) SetProtocol(protocol string) {
	l.protocol = protocol
	l.Logger = l.Logger.With().Str("protocol", protocol).Logger()
}

func (l *logger) Set(key, value string) {
	l.Logger = l.Logger.With().Str(key, value).Logger()
}

func (l *logger) With() Logger {
	return &logger{
		level:    l.level,
		out:      l.out,
		protocol: l.protocol,
		Logger:   zerolog.New(l.out).With().Timestamp().Logger().With().Str("protocol", l.protocol).Logger(),
	}
}
