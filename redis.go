package main

import (
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// RedisClient - redis client for calc pos and store stats
type RedisClient struct {
	client *redis.Client

	SeriesTimeout time.Duration
	PosTimeout    time.Duration
}

// NewRedisClient - create new redis client for pos and stats
func NewRedisClient(addr string, seriesTimeout, posTimeout time.Duration) (*RedisClient, error) {
	if seriesTimeout >= posTimeout {
		return nil, errors.New("'seriesTimeout' must be less than 'posTimeout'")
	}

	rc := &RedisClient{
		SeriesTimeout: seriesTimeout,
		PosTimeout:    posTimeout,
	}
	rc.client = redis.NewClient(&redis.Options{
		Addr: addr,
	})
	_, err := rc.client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return rc, nil
}

// Pos - calc next pos
func (r *RedisClient) Pos(ifa string) (int, error) {
	var pos int64
	var err error

	sessKey := "pos:s:" + ifa
	posKey := "pos:r:" + ifa
	ok, err := r.client.Expire(sessKey, r.SeriesTimeout).Result() // check in current series
	if err != nil {
		return 0, errors.Wrap(err, "error on update session expire")
	}
	if ok {
		err = r.client.Expire(posKey, r.PosTimeout).Err()
		if err != nil {
			return 0, errors.Wrap(err, "error on update pos expire")
		}
		pos, err = r.client.Get(posKey).Int64()
		if err != nil {
			return 0, errors.Wrap(err, "error on get pos key")
		}
		return int(pos), nil
	}

	err = r.client.Set(sessKey, 1, r.SeriesTimeout).Err()
	if err != nil {
		return 0, errors.Wrap(err, "error on set session key")
	}

	pos, err = r.client.Incr(posKey).Result()
	if err != nil {
		return 0, errors.Wrap(err, "error on incr pos")
	}
	err = r.client.Expire(posKey, r.PosTimeout).Err()
	if err != nil {
		return 0, errors.Wrap(err, "error on set pos expire")
	}
	return int(pos), nil
}

// Add - add stat
func (r *RedisClient) Add(app, platform, country string) error {
	key := "s:" + app + ":" + platform + ":" + country // TODO quote ':' from values
	return r.client.Incr(key).Err()
}

// GetAll - get all stats
func (r *RedisClient) GetAll() ([]*Stat, error) {
	var res []*Stat
	iter := r.client.Scan(0, "s:*", 20).Iterator()
	for iter.Next() {
		key := iter.Val()
		count, err := r.client.Get(key).Int64()
		if err != nil {
			return nil, errors.Wrap(err, "error on get stat")
		}
		keys := strings.Split(key, ":")
		if len(keys) != 4 {
			continue
		}
		res = append(res, &Stat{keys[1], keys[2], keys[3], int(count)})
	}
	err := iter.Err()
	if err != nil {
		return nil, errors.Wrap(err, "error on iter stats")
	}
	return res, nil
}

func (r *RedisClient) flushAll() error {
	return r.client.FlushAll().Err()
}
