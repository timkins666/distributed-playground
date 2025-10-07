package main

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type dbPostgres struct {
	db          *sql.DB
	redisClient *redis.Client
}

// get single account matching id. always uses db for source of truth.
func (db *dbPostgres) getAccountByID(accountID int32) (*cmn.Account, error) {
	// TODO: squirrel / sqlx

	acc := cmn.Account{}

	err := db.db.QueryRow(`
		SELECT id, user_id, balance from accounts.account WHERE id = $1
	`, accountID).Scan(&acc.AccountID, &acc.UserID, &acc.Balance)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

// get all accounts for the user from redis/db
func (db *dbPostgres) getUserAccounts(userID int32) ([]cmn.Account, error) {
	// TODO: squirrel / sqlx
	// BIG TODO: put redis stuff somewhere else

	var redisKey string
	if db.redisClient != nil {
		redisKey = cmn.RedisKey(cmn.RedisKeyUserAccounts, strconv.Itoa(int(userID)))
		cached, err := db.redisClient.Get(context.Background(), redisKey).Result()

		if err == nil {
			log.Printf("found cached accounts for user %d", userID)
			accs, err := cmn.FromBytes[[]cmn.Account]([]byte(cached))
			return *accs, err
		} else if err != redis.Nil {
			log.Printf("Error getting user %d accounts from cache: %v", userID, err)
		}

		log.Printf("getUserAccounts cache miss for user %d", userID)
	}

	var accounts []cmn.Account

	rows, err := db.db.Query(`
		SELECT id, name, balance FROM accounts.account WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var acc cmn.Account
		err := rows.Scan(&acc.AccountID, &acc.Name, &acc.Balance)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("user %d accounts\n%+v", userID, accounts)

	b, err := cmn.ToBytes(accounts)
	if err != nil {
		return nil, err
	}

	if db.redisClient != nil {
		log.Printf("Setting redis key %s", redisKey)
		db.redisClient.Set(context.Background(), redisKey, string(b), time.Minute)
	}
	return accounts, nil
}

func (db *dbPostgres) createAccount(a cmn.Account) (int32, error) {
	var newAccID int32
	err := db.db.QueryRow(`
		INSERT INTO accounts.account (user_id, name)
		VALUES ($1, $2)
		RETURNING id
		`, a.UserID, a.Name).Scan(&newAccID)

	if db.redisClient != nil {
		// invalidate cache TODO: separate consumer invalidation service
		db.redisClient.Del(context.Background(), cmn.RedisKey(cmn.RedisKeyUserAccounts, strconv.Itoa(int(a.UserID))))
	}
	return newAccID, err
}

func (db *dbPostgres) getUserByID(userID int32) (*cmn.User, error) {
	log.Printf("Try load user id %d from db...", userID)
	var user cmn.User
	err := db.db.QueryRow(`
		SELECT id, username, roles FROM accounts."user" WHERE id = $1
	`, userID).Scan(&user.ID, &user.Username, pq.Array(&user.Roles))
	return &user, err
}
