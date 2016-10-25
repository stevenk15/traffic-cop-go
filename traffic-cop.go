package main

import (
    "net/http"
	"encoding/json"
	"github.com/gocql/gocql"
	"log"
	"gopkg.in/redis.v5"
	"fmt"
)

type healthCheckMsg struct {
	Version string
	AppName string
	HostName string
	Redis string
	Cassandra string
}

type tc struct {
	Platform string
	Location string
}

type platformLocation struct {
	Legacy string
	OEM string
}

func getHandler(w http.ResponseWriter, r *http.Request, session *gocql.Session, redisClient *redis.Client) {

	var platform string
	//var b []byte

	// get the userId from the request object
	userId := r.URL.Query().Get("userId")

	// Set some standard headers
	w.Header().Set("Content-Type", "application/json")

	// check for the value in the Redis cache first
	redisVal, redisErr := redisClient.Get(userId).Result()
	if redisErr != nil {
		//log.Fatal(err)

		log.Print(redisErr)
		log.Print("redisVal1: " + redisVal)

		// query Cassandra for data if not found in Redis
		if err := session.Query("SELECT platform FROM users WHERE users = " + userId).Scan(&platform); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		m := tc{platform,"http://www.foo.com"}
		b, err := json.Marshal(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			w.Header().Set("X-Data-Origin", "Cassandra")
			w.Write(b)
			return
		}

	// Return the values from Redis
	} else {
		log.Print(redisErr)
		log.Print("redisVal2: " + redisVal)

		m := tc{redisVal,"http://www.foo.com"}
		b, err := json.Marshal(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			w.Header().Set("X-Data-Origin", "Redis")
			w.Write(b)
			return
		}

	}

}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	m := healthCheckMsg{"0.0.1", "traffic-cop-go", "localhost", "connected", "connected"}
	b, err := json.Marshal(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	/*pong, err := redisClient.Ping().Result()
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}*/

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func main() {

	fmt.Print("Server Started!")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Initialize Cassandra cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "ks1"
	cluster.ProtoVersion = 4

	// Establish connection to Cassandra
	cassandraSession, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

    // Get handler
    http.HandleFunc("/svc/v1/traffic-cop", func(w http.ResponseWriter, r *http.Request){ getHandler(w, r, cassandraSession, redisClient)})

	// healthcheck
	http.HandleFunc("/healthcheck", healthCheckHandler)

    http.ListenAndServe(":5000", nil)

}
