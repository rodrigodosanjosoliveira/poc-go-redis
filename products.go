package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
)

type Products struct {
	ProductId   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	RetailPrice float64 `json:"retail_price"`
}

type JsonResponse struct {
	Data   []Products `json:"data"`
	Source string     `json:"source"`
}

func getProducts() (*JsonResponse, error) {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	cachedProducts, err := redisClient.Get("products").Bytes()

	response := JsonResponse{}

	if err != nil {
		dbProducts, err := fetchFromDb()

		if err != nil {
			return nil, err
		}

		cachedProducts, err = json.Marshal(dbProducts)

		if err != nil {
			return nil, err
		}

		err = redisClient.Set(
			"products", cachedProducts, 10*time.Second).Err()

		if err != nil {
			return nil, err
		}

		response = JsonResponse{Data: dbProducts, Source: "PostgreSQL"}

		return &response, err
	}

	products := []Products{}

	err = json.Unmarshal(cachedProducts, &products)

	if err != nil {
		return nil, err
	}

	response = JsonResponse{Data: products, Source: "Redis Cache"}

	return &response, nil

}

func fetchFromDb() ([]Products, error) {
	dbUser := "postgres"
	dbPassword := "Postgres2018!"
	dbName := "sample_company"

	conString := fmt.Sprintf(
		"user=%s password=%s dbname=%s sslmode=disable",
		dbUser,
		dbPassword,
		dbName,
	)

	db, err := sql.Open("postgres", conString)

	if err != nil {
		return nil, err
	}

	queryString := `select product_id, product_name, retail_price from products`

	rows, err := db.Query(queryString)

	if err != nil {
		return nil, err
	}

	var records []Products

	for rows.Next() {
		var p Products
		err = rows.Scan(&p.ProductId, &p.ProductName, &p.RetailPrice)

		records = append(records, p)

		if err != nil {
			return nil, err
		}
	}
	return records, nil

}
