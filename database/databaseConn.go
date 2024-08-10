package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/prajjwal-w/golang-choicetech/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBConn struct {
	R_DB  *redis.Client
	P_Sql *gorm.DB
}

func DBConnection() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", dbUser, dbPass, dbName, dbHost, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error while connecting to db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Error while getting the DB instance: %v", err)
	}

	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(20)

	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("Error While pining db:%v", err)
	}

	db.AutoMigrate(&model.Person{})
	log.Println("Databse migration successful")

	return db

}

func RedisConn() *redis.Client {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	rdbHost := os.Getenv("REDISADDR")
	rdbPass := os.Getenv("R_PASS")
	rdbPort := os.Getenv("R_PORT")

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", rdbHost, rdbPort),
		Password: rdbPass,
		DB:       0,
	})

	ping, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	log.Printf("Connected to Redis: %s\n", ping)

	return rdb
}

func InitializeConn() *DBConn {
	postgres := DBConnection()
	redis := RedisConn()

	return &DBConn{P_Sql: postgres, R_DB: redis}
}
