package utils

import (
	"github.com/gomodule/redigo/redis"
	"kafka2jenkins/consts"
)

var (
	// TODO：改为pool
	RedisCli redis.Conn
)

func RedisInit()  {
	var err error
	RedisCli, err = redis.Dial("tcp", consts.REDIS_BASE_SERVER_URL)
	if err != nil {
		panic(err)
	}

}

func RedisSetString(k, v string ) error {
	_, err := RedisCli.Do("SET", k, v)
	if err != nil {
		return err
	}
	return nil
}

func RedisGetString(k string) (string, error) {
	res, err := redis.String(RedisCli.Do("GET", k))
	if err != nil && err == redis.ErrNil{
		return "", err
	}
	return res, nil
}

func RedisDeleteString(k string) (int64, error) {
	res, err := redis.Int64(RedisCli.Do("DEL", k))
	if err != nil && err != redis.ErrNil{
		return -1, err
	}
	return res, nil
}

func RedisSetStringSet(k string, value []string) error {
	// TODO: 一次完成
	for _, v := range value {
		_, err := RedisCli.Do("SADD", k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
