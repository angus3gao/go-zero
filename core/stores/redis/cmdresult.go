package redis

import (
	"errors"

	red "github.com/redis/go-redis/v9"
)

type CmdResult struct {
	Cmd *red.Cmd
	Err error
}

func cmdResult(cmd *red.Cmd, err error) *CmdResult {
	return &CmdResult{cmd, err}
}

func (cr *CmdResult) commandError() error {
	if cr == nil {
		return errors.New("nil command result")
	}
	if cr.Err != nil {
		return cr.Err
	}
	if cr.Cmd == nil {
		return errors.New("nil redis command")
	}

	return nil
}

func (cr *CmdResult) String() (string, error) {
	return cr.StringDefalut("")
}

func (cr *CmdResult) Result() (any, error) {
	if err := cr.commandError(); err != nil {
		return nil, err
	}

	return cr.Cmd.Result()
}

func (cr *CmdResult) StringDefalut(defVal string) (string, error) {
	if err := cr.commandError(); err != nil {
		return defVal, err
	}

	res, err := cr.Cmd.Text()
	if err == red.Nil {
		return defVal, nil
	}
	return res, err
}

func (cr *CmdResult) Strings() ([]string, error) {
	if err := cr.commandError(); err != nil {
		return nil, err
	}

	res, err := cr.Cmd.StringSlice()
	if err == red.Nil {
		return nil, nil
	}
	return res, err
}

func (r *CmdResult) Bool() (bool, error) {
	if err := r.commandError(); err != nil {
		return false, err
	}

	res, err := r.Cmd.Bool()
	if err == red.Nil {
		return false, nil
	}

	if v, ok := r.Cmd.Val().(string); ok && v == "OK" {
		return true, nil
	}

	return res, err
}

func (r *CmdResult) Int() (int, error) {
	return r.IntDefault(0)
}

func (r *CmdResult) Int64() (int64, error) {
	return r.Int64Default(0)
}

func (r *CmdResult) IntDefault(defVal int) (int, error) {
	if err := r.commandError(); err != nil {
		return defVal, err
	}

	if res, err := r.Cmd.Int(); err == red.Nil {
		return defVal, nil
	} else {
		return res, err
	}
}

func (r *CmdResult) Int64Default(defVal int64) (int64, error) {
	if err := r.commandError(); err != nil {
		return defVal, err
	}

	if res, err := r.Cmd.Int64(); err == red.Nil {
		return defVal, nil
	} else {
		return res, err
	}
}

func (r *CmdResult) Float64() (float64, error) {
	return r.Float64Default(0)
}

func (r *CmdResult) Float64Default(defVal float64) (float64, error) {
	if err := r.commandError(); err != nil {
		return defVal, err
	}

	if res, err := r.Cmd.Float64(); err == red.Nil {
		return defVal, nil
	} else {
		return res, err
	}
}

func (r *CmdResult) Bytes() ([]byte, error) {
	if err := r.commandError(); err != nil {
		return nil, err
	}

	if res, err := r.Cmd.Text(); err == red.Nil {
		return nil, nil
	} else {
		return []byte(res), err
	}
}

func (r *CmdResult) ByteSlice() ([][]byte, error) {
	if err := r.commandError(); err != nil {
		return nil, err
	}

	strSlice, err := r.Cmd.StringSlice()
	if err == red.Nil {
		return nil, nil
	}

	res := make([][]byte, 0)
	for _, v := range strSlice {
		res = append(res, []byte(v))
	}

	return res, err
}
