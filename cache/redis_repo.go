package cache

import (
	"errors"
	"time"

	"github.com/go-redis/redis"
)

type MergeHandler = func(value string) (string, error)

const DAYS = 24 * time.Hour
const EXPIRE_DAYS = 7 * int(DAYS)

const PISEC_KEY = "PISEC:"
const DENY_KEY = PISEC_KEY + "DENY:"
const FALSE_POSITIVE_KEY = PISEC_KEY + "BLOOM_FILTER:"

const DEFAULT_PORT = "6379"
const TEST_PORT = "6378"

// Redis transactions use optimistic locking.
const maxRetries = 1000

type RedisRepository struct {
	client       *redis.Client
	dataDuration int
}

func NewTestClient() *RedisRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:" + DEFAULT_PORT,
		Password: "test", // no password set
		DB:       0,      // use default DB
	})
	return &RedisRepository{
		client:       rdb,
		dataDuration: 0,
	}
}

func NewRedisClient() *RedisRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:" + DEFAULT_PORT,
		Password: "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81", // no password set
		DB:       0,                                  // use default DB
	})

	return &RedisRepository{
		client:       rdb,
		dataDuration: EXPIRE_DAYS}
}

func (repo *RedisRepository) InitRepository() error {
	return repo.client.FlushDB().Err()
}

func (repo *RedisRepository) GetRepoSize() (int, error) {
	res := repo.client.DbSize()
	return int(res.Val()), res.Err()
}

func (repo *RedisRepository) FindAllDenyList() []string {
	result := []string{}
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = repo.client.Scan(cursor, DENY_KEY+"*", 0).Result()
		if err != nil {
			panic(err)
		}

		for _, key := range keys {
			result = append(result, key)
		}

		if cursor == 0 { // no more keys
			break
		}
	}
	return result
}

// Increment transactionally increments the key using GET and SET commands.
func (repo *RedisRepository) AddDeny(source string, day int, updater MergeHandler) error {
	key := DENY_KEY + source
	// Transactional function.
	txf := func(tx *redis.Tx) error {
		// Get the current value or zero.
		value := tx.Get(key).String()

		// Actual operation (local in optimistic lock).
		value, err := updater(value)
		if err != nil {
			return err
		}

		// Operation is commited only if the watched keys remain unchanged.
		_, err = tx.TxPipelined(func(pipe redis.Pipeliner) error {
			pipe.Set(key, value, time.Duration(repo.dataDuration))
			return nil
		})
		return err
	}

	// Retry if the key has been changed.
	for i := 0; i < maxRetries; i++ {
		err := repo.client.Watch(txf, key)
		if err == nil {
			// Success.
			return nil
		}
		if err == redis.TxFailedErr {
			// Optimistic lock lost. Retry.
			continue
		}
		// Return any other error.
		return err
	}

	return errors.New("increment reached maximum number of retries")
}
