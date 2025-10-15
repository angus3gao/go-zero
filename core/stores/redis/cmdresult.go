package redis

import (
	"github.com/redis/go-redis/v9"
)

type CmdResult struct {
	Cmd *redis.Cmd
	Err error
}

func cmdResult(cmd *redis.Cmd, err error) *CmdResult {
	return &CmdResult{cmd, err}
}

func (cr *CmdResult) String() (string, error) {
	return cr.StringDefalut("")
}

func (cr *CmdResult) Result() (any, error) {
	return cr.Cmd.Result()
}

func (cr *CmdResult) StringDefalut(defVal string) (string, error) {
	res, err := cr.Cmd.Text()
	if err == redis.Nil {
		return defVal, nil
	}
	return res, err
}

func (cr *CmdResult) Strings() ([]string, error) {
	res, err := cr.Cmd.StringSlice()
	if err == redis.Nil {
		return nil, nil
	}
	return res, err
}

func (r *CmdResult) Bool() (bool, error) {
	res, err := r.Cmd.Bool()
	if err == redis.Nil {
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
	if res, err := r.Cmd.Int(); err == redis.Nil {
		return defVal, nil
	} else {
		return res, err
	}
}

func (r *CmdResult) Int64Default(defVal int64) (int64, error) {
	if res, err := r.Cmd.Int64(); err == redis.Nil {
		return defVal, nil
	} else {
		return res, err
	}
}

func (r *CmdResult) Float64() (float64, error) {
	return r.Float64Default(0)
}

func (r *CmdResult) Float64Default(defVal float64) (float64, error) {
	if res, err := r.Cmd.Float64(); err == redis.Nil {
		return defVal, nil
	} else {
		return res, err
	}
}

func (r *CmdResult) Bytes() ([]byte, error) {
	if res, err := r.Cmd.Text(); err == redis.Nil {
		return nil, nil
	} else {
		return []byte(res), err
	}
}

func (r *CmdResult) ByteSlice() ([][]byte, error) {
	strSlice, err := r.Cmd.StringSlice()
	if err == redis.Nil {
		return nil, nil
	}

	res := make([][]byte, 0)
	for _, v := range strSlice {
		res = append(res, []byte(v))
	}

	return res, err
}
