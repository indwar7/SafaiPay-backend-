package leaderboard

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Service interface {
	UpdateScore(ctx context.Context, userID, ward string, points float64) error
	GetTopN(ctx context.Context, ward string, limit int) ([]LeaderboardEntry, error)
	GetUserRank(ctx context.Context, userID, ward string) (*UserRank, error)
}

type LeaderboardEntry struct {
	UserID string  `json:"user_id"`
	Score  float64 `json:"score"`
	Rank   int     `json:"rank"`
}

type UserRank struct {
	UserID string  `json:"user_id"`
	Score  float64 `json:"score"`
	Rank   int64   `json:"rank"`
}

type service struct {
	rdb *redis.Client
}

func NewService(rdb *redis.Client) Service {
	return &service{rdb: rdb}
}

func (s *service) getKey(ward string) string {
	if ward == "" {
		return "leaderboard:global"
	}
	return fmt.Sprintf("leaderboard:ward:%s", ward)
}

func (s *service) UpdateScore(ctx context.Context, userID, ward string, points float64) error {
	pipe := s.rdb.Pipeline()

	pipe.ZIncrBy(ctx, "leaderboard:global", points, userID)

	if ward != "" {
		pipe.ZIncrBy(ctx, fmt.Sprintf("leaderboard:ward:%s", ward), points, userID)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (s *service) GetTopN(ctx context.Context, ward string, limit int) ([]LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	key := s.getKey(ward)

	results, err := s.rdb.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]LeaderboardEntry, len(results))
	for i, z := range results {
		entries[i] = LeaderboardEntry{
			UserID: z.Member.(string),
			Score:  z.Score,
			Rank:   i + 1,
		}
	}

	return entries, nil
}

func (s *service) GetUserRank(ctx context.Context, userID, ward string) (*UserRank, error) {
	key := s.getKey(ward)

	rank, err := s.rdb.ZRevRank(ctx, key, userID).Result()
	if err == redis.Nil {
		return &UserRank{
			UserID: userID,
			Score:  0,
			Rank:   -1,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	score, err := s.rdb.ZScore(ctx, key, userID).Result()
	if err != nil {
		return nil, err
	}

	return &UserRank{
		UserID: userID,
		Score:  score,
		Rank:   rank + 1,
	}, nil
}
