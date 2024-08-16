package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/prajjwal-w/golang-choicetech/database"
	"github.com/prajjwal-w/golang-choicetech/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var dbConn = database.InitializeConn()

func StoreData(persons []model.Person) error {
	//begin the tx for inserting the data
	tx := dbConn.P_Sql.Begin()

	//inserting the data in batches
	const batchSize = 500

	if err := tx.CreateInBatches(persons, batchSize).Error; err != nil {
		tx.Rollback()
		log.Printf("Failed to insert persons: %v", err)
		return err
	}

	// Commit the transaction if everything is successful
	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return err
	}

	log.Printf("Successfully inserted %d records", len(persons))
	return nil

}

func CacheData(persons []model.Person) error {
	redis := dbConn.R_DB
	errChan := make(chan error, len(persons)+1)
	const batchSize = 200
	var wg sync.WaitGroup
	log.Println("IN the cacheData")
	//insert cache logic using go ruoutine in batchwise
	for i := 0; i < len(persons); i += batchSize {
		end := i + batchSize
		if end > len(persons) {
			end = len(persons)
		}

		batch := persons[i:end]
		wg.Add(1)
		go func(batch []model.Person) {

			defer wg.Done()

			ctx := context.Background()
			err := cacheBatch(ctx, redis, batch) //using pipelined caching
			if err != nil {
				errChan <- err
				return
			}

		}(batch)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	log.Println("error chan closed")
	//check the first encountered error from the go routine and return it
	var e error
	for err := range errChan {
		if e == nil {
			e = err
		} else {
			log.Printf("error during caching:%v", err)
		}
	}

	if e != nil {
		return e
	}
	log.Printf("Successfully cache the data len: %d", len(persons))
	return nil

}

// helper function for inseting the cache in a pipeline
func cacheBatch(ctx context.Context, rdb *redis.Client, batch []model.Person) error {
	pipe := rdb.Pipeline()
	for _, p := range batch {
		data, err := json.Marshal(p)
		if err != nil {
			return err
		}
		pipe.Set(ctx, string(p.Email), data, 5*time.Minute)
	}
	_, err := pipe.Exec(ctx)
	return err

}

func GetTheDataByEmail(email string) (*model.Person, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var person model.Person

	//check the data into the redis
	r := dbConn.R_DB
	data, err := r.Get(ctx, email).Result()

	//check the data found in the database or not
	if err == redis.Nil {
		log.Printf("Data not found in the Redis for email: %s.", email)

		if err := dbConn.P_Sql.Where("email = ?", email).First(&person).Error; err != nil {
			return nil, fmt.Errorf("person not found in the DB: %v", err)
		}

		//cache the data fetch from the database into redis
		data, err := json.Marshal(person)
		if err == nil {
			if err := r.Set(ctx, email, data, 5*time.Minute).Err(); err != nil {
				log.Printf("failed to cache person data in redis: %v", err)

			}
		}
	} else if err != nil {
		//some other error occured
		log.Printf("Error fetching data from Redis:%v", err)
		return nil, fmt.Errorf("error fetching data from Redis:%v", err)
	} else {
		//unmarshal the data from redis
		if err := json.Unmarshal([]byte(data), &person); err != nil {
			log.Printf("Error processing data :%v", err)
			return nil, fmt.Errorf("error processing data :%v", err)
		}
		log.Println("Found data in cache")
	}

	//log.Printf("retrieved data %v", person)

	return &person, nil

}

func UpdateData(updatePer *model.Person) error {

	// Check if email is provided
	if updatePer.Email == "" {
		return fmt.Errorf("email is required for updating a person record")
	}
	var person model.Person

	//check and get the data from the databse
	if err := dbConn.P_Sql.Where("email = ?", updatePer.Email).First(&person).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("no record found with email: %v", updatePer.Email)
		} else {
			log.Printf("Error retrieving record from database: %v", err)
			return fmt.Errorf("error while retrieving record: %v", err)
		}

	}

	//update the data into database
	if err := dbConn.P_Sql.Model(&person).Where("email=?", updatePer.Email).Updates(updatePer).Error; err != nil {
		log.Printf("Error updating record in database: %v", err)
		return fmt.Errorf("error updating record in database: %v", err)
	}

	// Serialize the updated record to JSON
	data, err := json.Marshal(updatePer)
	if err != nil {
		log.Printf("Error marshalling data to JSON: %v", err)
		return fmt.Errorf("error marshalling data to JSON: %v", err)
	}
	// Update the Redis cache with the new record
	ctx := context.Background()
	if err := dbConn.R_DB.Set(ctx, updatePer.Email, data, 5*time.Minute).Err(); err != nil {
		log.Printf("Error updating Redis cache: %v", err)
	} else {
		log.Printf("Record updated in Redis cache: %v", updatePer)
	}

	return nil

}
