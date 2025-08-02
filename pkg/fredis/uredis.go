package fredis

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"strings"
	"time"
)

const (
	ErrRedisIsNil = "实时库查询异常,键不存在"
)

// RedisRouter 类似于路由，用于组织 Redis 键
type RedisRouter struct {
	ctx      context.Context
	prefixes []string
	sep      string
	client   *redis.Client
}

// NewRedisRouter 创建一个新的 RedisRouter 实例，默认分隔符为 :
func NewRedisRouter(client *redis.Client, root string) *RedisRouter {
	r := RedisRouter{
		sep:    ":",
		client: client,
	}
	return r.Group(root)
}

// Group 创建一个新的子路由组，添加前缀，若前缀中未包含分隔符则自动添加
func (r *RedisRouter) Group(prefix string) *RedisRouter {
	if strings.Contains(prefix, r.sep) {
		prefix = strings.TrimPrefix(prefix, r.sep)
	}
	newPrefixes := make([]string, len(r.prefixes))
	copy(newPrefixes, r.prefixes)
	newPrefixes = append(newPrefixes, prefix)
	return &RedisRouter{
		prefixes: newPrefixes,
		sep:      r.sep,
		client:   r.client,
	}
}

// Clear 删除指定前缀的所有键
func (r *RedisRouter) Clear() error {
	ctx := context.Background()
	pattern := r.buildKey("*")
	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 0).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return nil
}

// Close 关闭 Redis 客户端连接
func (r *RedisRouter) Close() error {
	return r.client.Close()
}

// buildKey 构建完整的 Redis 键
func (r *RedisRouter) buildKey(key string) string {
	if len(r.prefixes) == 0 {
		return key
	}
	k := strings.Join(r.prefixes, r.sep) + r.sep + key
	return k
}

/*
	字符串操作
*/
// Set 设置 Redis 键值对
func (r *RedisRouter) Set(key string, value interface{}, expiration ...time.Duration) error {
	fullKey := r.buildKey(key)
	// 手动序列化数据
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if len(expiration) > 0 {
		return r.client.Set(r.client.Context(), fullKey, data, expiration[0]).Err()
	}
	return r.client.Set(r.client.Context(), fullKey, data, 0).Err()
}

// MSet 设置 多个 Redis 键值对
func (r *RedisRouter) MSet(mapData map[string]interface{}) error {
	mapList := make(map[string]interface{})
	if len(mapData) <= 0 {
		return nil
	}
	for k, v := range mapData {
		// 手动序列化数据
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		mapList[r.buildKey(k)] = data
	}
	err := r.client.MSet(r.client.Context(), mapList).Err()
	if err != nil {
		return err
	}
	return nil
}

// Get 获取 Redis 键对应的值
func (r *RedisRouter) Get(key string, value interface{}) error {
	fullKey := r.buildKey(key)
	data, err := r.client.Get(r.client.Context(), fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.New(ErrRedisIsNil)
		}
		return err
	}
	err = json.Unmarshal([]byte(data), value)
	if err != nil {
		return err
	}
	return nil
}

// MGet 获取 多个Redis 键对应的值
func (r *RedisRouter) MGet(value interface{}, keys ...string) error {
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}
	if len(fullKeys) <= 0 {
		return nil
	}
	vs := make([]interface{}, 0)
	vList, err := r.client.MGet(r.client.Context(), fullKeys...).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.New(ErrRedisIsNil)
		}
		return err
	}
	for _, v := range vList {
		var item interface{}
		if str, ok := v.(string); ok {
			err = json.Unmarshal([]byte(str), &item)
			if err != nil {
				return err
			}
			vs = append(vs, item)
		}
	}
	data, err := json.Marshal(vs)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, value)
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisRouter) GetList(values interface{}) error {
	fullKey := r.buildKey("*")
	keys, err := r.client.Keys(r.client.Context(), fullKey).Result()
	if err != nil {
		return err
	}
	if len(keys) <= 0 {
		return nil
	}
	vs := make([]interface{}, 0)
	vList, err := r.client.MGet(r.client.Context(), keys...).Result()
	if err != nil {
		return err
	}
	for _, v := range vList {
		var item interface{}
		if str, ok := v.(string); ok {
			err = json.Unmarshal([]byte(str), &item)
			if err != nil {
				return err
			}
			vs = append(vs, item)
		}
	}
	data, err := json.Marshal(vs)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, values)
	if err != nil {
		return err
	}
	return nil
}

// Del 删除 Redis 键
func (r *RedisRouter) Del(keys ...string) error {
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}
	return r.client.Del(r.client.Context(), fullKeys...).Err()
}

/*
	哈希操作
*/

// HSet 设置哈希表中的字段值
func (r *RedisRouter) HSet(key string, data map[string]interface{}) error {
	fullKey := r.buildKey(key)
	values := make(map[string]string, 0)
	for k, v := range data {
		v, err := json.Marshal(v)
		if err != nil {
			return err
		}
		values[k] = string(v)
	}
	return r.client.HSet(r.client.Context(), fullKey, values).Err()
}

// HGet 获取哈希表中指定字段的值
func (r *RedisRouter) HGet(key, field string) (string, error) {
	fullKey := r.buildKey(key)
	return r.client.HGet(r.client.Context(), fullKey, field).Result()
}

// HGetAll 获取哈希表中所有字段的值
func (r *RedisRouter) HGetAll(key string) (map[string]string, error) {
	fullKey := r.buildKey(key)
	return r.client.HGetAll(r.client.Context(), fullKey).Result()
}

// HDel 移除哈希表中的一个或多个指定字段
func (r *RedisRouter) HDel(key string, field ...string) error {
	fullKey := r.buildKey(key)
	return r.client.HDel(r.client.Context(), fullKey, field...).Err()
}

// HDelAll 获取哈希表中所有字段的值
func (r *RedisRouter) HDelAll(key string) error {
	fullKey := r.buildKey(key)
	return r.client.Del(r.client.Context(), fullKey).Err()
}

/*
	列表操作
*/
// LPush 向列表头部插入一个或多个值
func (r *RedisRouter) LPush(key string, values ...interface{}) error {
	fullKey := r.buildKey(key)
	return r.client.LPush(r.client.Context(), fullKey, values...).Err()
}

// RPush 向列表尾部插入一个或多个值
func (r *RedisRouter) RPush(key string, values ...interface{}) error {
	fullKey := r.buildKey(key)
	return r.client.RPush(r.client.Context(), fullKey, values...).Err()
}

// LPop 移除并获取列表的最后一个元素
func (r *RedisRouter) LPop(key string) (string, error) {
	fullKey := r.buildKey(key)
	return r.client.LPop(r.client.Context(), fullKey).Result()
}

// RPop 移除并获取列表的最后一个元素
func (r *RedisRouter) RPop(key string) (string, error) {
	fullKey := r.buildKey(key)
	return r.client.RPop(r.client.Context(), fullKey).Result()
}

// LRem 移除列表中与指定值相等的元素
func (r *RedisRouter) LRem(key string, count int64, value interface{}) error {
	fullKey := r.buildKey(key)
	return r.client.LRem(r.client.Context(), fullKey, count, value).Err()
}

// LRange 获取所有列表
func (r *RedisRouter) LRange(key string, start, stop int64) ([]string, error) {
	fullKey := r.buildKey(key)
	return r.client.LRange(r.client.Context(), fullKey, start, stop).Result()
}

/*
	无序列表操作
*/

// SAdd 向集合添加一个或多个成员
func (r *RedisRouter) SAdd(key string, members ...interface{}) error {
	fullKey := r.buildKey(key)
	return r.client.SAdd(r.client.Context(), fullKey, members...).Err()
}

// SMembers 获取集合中的所有成员
func (r *RedisRouter) SMembers(key string) ([]string, error) {
	fullKey := r.buildKey(key)
	return r.client.SMembers(r.client.Context(), fullKey).Result()
}

// SRem 移除集合中的一个或多个成员
func (r *RedisRouter) SRem(key string, members ...string) error {
	fullKey := r.buildKey(key)
	return r.client.SRem(r.client.Context(), fullKey, members).Err()
}

// SMembersAll 获取该路由下，所有key的[]string值
func (r *RedisRouter) SMembersAll() (map[string][]string, error) {
	fullKey := r.buildKey("*")
	result, err := r.client.Keys(r.client.Context(), fullKey).Result()
	if err != nil {
		return nil, err
	}
	vs := make(map[string][]string, 0)
	for _, key := range result {
		v, err := r.client.SMembers(r.client.Context(), key).Result()
		if err != nil {
			return nil, err
		}
		k := strings.TrimPrefix(key, strings.Join(r.prefixes, r.sep)+r.sep)
		vs[k] = v
	}
	return vs, nil
}

/*
	有序列表操作
*/

// ZAdd 向有序集合添加一个或多个成员，或者更新已存在成员的分数
func (r *RedisRouter) ZAdd(key string, members ...*redis.Z) error {
	fullKey := r.buildKey(key)
	return r.client.ZAdd(r.client.Context(), fullKey, members...).Err()
}

// ZRange 获取有序集合中指定区间内的成员
func (r *RedisRouter) ZRange(key string, start, stop int64) ([]string, error) {
	fullKey := r.buildKey(key)
	return r.client.ZRange(r.client.Context(), fullKey, start, stop).Result()
}

/*
	自增自减操作
*/

// Incr 对键的值进行自增操作
func (r *RedisRouter) Incr(key string) *redis.IntCmd {
	fullKey := r.buildKey(key)
	return r.client.Incr(r.client.Context(), fullKey)
}

// Decr 对键的值进行自减操作
func (r *RedisRouter) Decr(key string) *redis.IntCmd {
	fullKey := r.buildKey(key)
	return r.client.Decr(r.client.Context(), fullKey)
}
