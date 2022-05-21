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
	Infof(string, ...any)
	Infoln(...any)
	Warnf(string, ...any)
	Warnln(...any)
	Errorf(string, ...any)
	Errorln(...any)
}

type logger struct {
	zerolog.Logger
	level Level
	out   io.Writer
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
	l := &logger{level: level}
	switch out {
	case "stdout":
		l.out = os.Stdout
	case "stderr":
		l.out = os.Stderr
	case "":
		l.out = ioutil.Discard
	default:
		ok, err := filepath.Match(BASE_PATH, out)
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
	l.Logger = zerolog.New(l.out).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: l.out})
	return l, nil
}

func (l *logger) Infof(format string, v ...any) {
	if l.level < Warn && l.level > NoLog {
		l.Logger.Info().Msgf(format, v...)
	}
}

func (l *logger) Infoln(a ...any) {
	if l.level < Warn && l.level > NoLog {
		l.Logger.Info().Msg(fmt.Sprintln(a...))
	}
}

func (l *logger) Warnf(format string, v ...any) {
	if l.level < Error && l.level > NoLog {
		l.Logger.Warn().Msgf(format, v...)
	}
}

func (l *logger) Warnln(a ...any) {
	if l.level < Error && l.level > NoLog {
		l.Logger.Warn().Msg(fmt.Sprintln(a...))
	}
}

func (l *logger) Errorf(format string, v ...any) {
	if l.level > NoLog {
		l.Logger.Error().Msgf(format, v...)
	}
}

func (l *logger) Errorln(a ...any) {
	if l.level > NoLog {
		l.Logger.Error().Msg(fmt.Sprintln(a...))
	}
}
