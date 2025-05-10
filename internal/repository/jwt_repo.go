package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
)

type JWTStorage struct {
	*redis.Client
}

func NewJWTStorage() *JWTStorage {
	return &JWTStorage{redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})}
}

func (r *JWTStorage) Close() {
	r.Client.Close()
}

func (r *JWTStorage) SaveTokens(accessToken, refreshToken string) error {
	if err := r.Client.Set(context.Background(), accessToken, refreshToken, 0).Err(); err != nil {
		return err
	}
	if err := r.Client.Set(context.Background(), refreshToken, accessToken, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (r *JWTStorage) IsTokenExist(token string) error {
	_, err := r.Client.Get(context.Background(), token).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errors.New("token not found")
		}
		return err
	}
	return nil
}

func (r *JWTStorage) GetAnotherToken(token string) (string, error) {
	anotherToken, err := r.Client.Get(context.Background(), token).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", errors.New("token not found")
		}
		return "", err
	}
	return anotherToken, nil
}

func (r *JWTStorage) RemoveTokens(token string) error {
	anotherToken, err := r.Client.Get(context.Background(), token).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return err
		}
	}
	_, err = r.Client.Del(context.Background(), token).Result()
	if err != nil {
		return fmt.Errorf("remove token failed: %w", err)
	}
	_, err = r.Client.Del(context.Background(), anotherToken).Result()
	if err != nil {
		return fmt.Errorf("remove token failed: %w", err)
	}
	return nil
}
