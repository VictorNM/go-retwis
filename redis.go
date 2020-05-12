package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"strconv"
	"time"
)

func createUser(client *redis.Client, user *user) (*user, error) {
	id, err := client.Incr("next_user_id").Result()
	if err != nil {
		return nil, fmt.Errorf("user id increment failed: %v", err)
	}

	key := fmt.Sprintf("user:%d", id)
	tx := client.TxPipeline()
	tx.HSet(key, "username", user.username, "password", user.password)
	tx.HSet("users", user.username, id)
	_, err = tx.Exec()
	if err != nil {
		return nil, err
	}

	user.id = id

	return user, nil
}

func getUserByID(client *redis.Client, id int64) (*user, error) {
	username, err := client.HGet("user:"+i64toa(id), "username").Result()
	if err != nil {
		return nil, err
	}

	return &user{
		id:       id,
		username: username,
	}, nil
}

func getUserByUsername(client *redis.Client, username string) (*user, error) {
	idString, err := client.HGet("users", username).Result()
	if err != nil {
		return nil, err
	}

	password, err := client.HGet("user:"+idString, "password").Result()
	if err != nil {
		return nil, err
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		return nil, err
	}

	return &user{
		id:       int64(id),
		username: username,
		password: password,
	}, nil
}

func createFollow(client *redis.Client, followedID, followerID int64) error {
	_, err := client.TxPipelined(func(tx redis.Pipeliner) error {
		tx.ZAdd(followersKey(followedID), &redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: followerID,
		})
		tx.ZAdd(followingKey(followerID), &redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: followedID,
		})

		return nil
	})

	if err != nil {
		return fmt.Errorf("create follow failed: %v", err)
	}

	return nil
}

func deleteFollow(client *redis.Client, followedID, followerID int64) error {
	_, err := client.TxPipelined(func(tx redis.Pipeliner) error {
		tx.ZRem(followersKey(followedID), followerID)
		tx.ZRem(followingKey(followerID), followedID)

		return nil
	})

	if err != nil {
		return fmt.Errorf("create follow failed: %v", err)
	}

	return nil
}

func isFollowing(client *redis.Client, followedID, followerID int64) bool {
	if _, err := client.ZRank(followersKey(followedID), i64toa(followerID)).Result(); err != nil {
		return false
	}

	if _, err := client.ZRank(followingKey(followerID), i64toa(followedID)).Result(); err != nil {
		return false
	}

	return true
}

func createPost(client *redis.Client, userID int64, p *post) (*post, error) {
	postID, err := client.Incr("next_post_id").Result()
	if err != nil {
		return nil, err
	}

	err = client.HSet(
		"post:"+i64toa(postID),
		"user_id", userID,
		"created_at", time.Now().Unix(),
		"body", p.Body,
	).Err()
	if err != nil {
		return nil, err
	}

	p.ID = postID

	// update posts list of followers
	followerIDs, err := client.ZRange("followers:"+i64toa(userID), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	followerIDs = append(followerIDs, i64toa(userID))
	_, err = client.Pipelined(func(pipeliner redis.Pipeliner) error {
		for _, followerID := range followerIDs {
			pipeliner.LPush("posts:"+followerID, postID)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return p, nil
}

func getUserPosts(client *redis.Client, userID int64, offset, limit int64) ([]*post, error) {
	postIDs, err := client.LRange("posts:"+i64toa(userID), offset, offset+limit).Result()
	if err != nil {
		return nil, err
	}
	
	var postList []*post
	for _, postIDStr := range postIDs {
		body, err := client.HGet("post:"+postIDStr, "body").Result()
		if err != nil {
			continue
		}
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			continue
		}

		postList = append(postList, &post{
			ID:   int64(postID),
			Body: body,
		})
	}

	return postList, nil
}

func countFollowers(client *redis.Client, userID int64) int {
	count, err := client.ZCard(followersKey(userID)).Result()
	if err != nil {
		return 0
	}

	return int(count)
}

func countFollowing(client *redis.Client, userID int64) int {
	count, err := client.ZCard(followingKey(userID)).Result()
	if err != nil {
		return 0
	}

	return int(count)
}

func followersKey(id int64) string {
	return fmt.Sprintf("followers:%d", id)
}

func followingKey(id int64) string {
	return fmt.Sprintf("following:%d", id)
}

func i64toa(i int64) string {
	return strconv.FormatInt(i, 10)
}
