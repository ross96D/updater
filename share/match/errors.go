package match

import (
	"fmt"
	"slices"

	"github.com/rs/zerolog"
)

func FmtFromInnerError(msg string, err error) ErrLevel {
	if innerErr, ok := err.(ErrLevel); ok {
		switch innerErr.Level() {
		case "error":
			return ErrError{fmt.Errorf(msg, err)}
		case "warning":
			return ErrWarning{fmt.Errorf(msg, err)}
		default:
			panic("unreachable FromInnerError error level")
		}
	}
	return ErrError{fmt.Errorf(msg, err)}
}

type ErrLevel interface {
	error
	Level() string
	Log(*zerolog.Logger)
}

// Error that indicates a fail in the update
type ErrError struct{ err error }

func (e ErrError) Error() string { return e.err.Error() }
func (e ErrError) Level() string { return "error" }
func (e ErrError) Log(logger *zerolog.Logger) {
	logger.Error().Err(e.err).Send()
}

type ErrWarning struct{ err error }

func (e ErrWarning) Error() string { return e.err.Error() }
func (e ErrWarning) Level() string { return "warning" }
func (e ErrWarning) Log(logger *zerolog.Logger) {
	logger.Warn().Err(e.err).Send()
}

type JoinErrors struct {
	errs []ErrLevel
}

func (e JoinErrors) IsEmpty() bool {
	return len(e.errs) == 0
}

func (e JoinErrors) IsNotEmpty() bool {
	return len(e.errs) != 0
}

func (e *JoinErrors) Add(err error) {
	if err == nil {
		return
	}
	switch err := err.(type) {
	case ErrLevel:
		e.errs = append(e.errs, err)
	default:
		e.errs = append(e.errs, ErrError{err})
	}
}

func (v *JoinErrors) Concat(err JoinErrors) {
	for _, er := range err.errs {
		v.Add(er)
	}
}

func (e JoinErrors) Log(logger *zerolog.Logger) {
	slices.SortFunc(e.errs, func(a ErrLevel, b ErrLevel) int {
		valA := 0
		switch a.(type) {
		case ErrError:
			valA = 1
		case ErrWarning:
			valA = 0
		default:
			panic("unreachable unhandled type for ErrLevel")
		}
		valB := 0
		switch a.(type) {
		case ErrError:
			valB = 1
		case ErrWarning:
			valB = 0
		default:
			panic("unreachable unhandled type for ErrLevel")
		}
		return valA - valB
	})
	for _, e := range e.errs {
		e.Log(logger)
	}
}
