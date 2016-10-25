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

// The route handler function for the GET route for /svc/v1/traffic-cop.
// This function will extract the userId parameter value out of the request, then
// it will attempt to find a matching record in Redis. If the record is found
// it will pass that to the sendResponse function to return it to the user. If
// the record is not found it will attempt to find the record in the Cassandra
// database. If it is found in Cassandra it will return it, if not it will return
// an HTTP 404 error.
func getHandler(w http.ResponseWriter, r *http.Request, session *gocql.Session, redisClient *redis.Client) {

	var platform string
	//var b []byte

	// get the userId from the request object
	userId := r.URL.Query().Get("userId")

	// check for the value in the Redis cache first
	redisVal, redisErr := redisClient.Get(userId).Result()
	if redisErr != nil {

		// query Cassandra for data if not found in Redis
		if err := session.Query("SELECT platform FROM users WHERE users = " + userId).Scan(&platform); err != nil {

			// If it is not in Redis and not in Cassandra return a 404
			http.Error(w, err.Error(), http.StatusNotFound)
			return

		}
		sendResponse(w, platform, "Cassandra")

	// Return the values from Redis
	} else {
		sendResponse(w, redisVal, "Redis")
	}

}


// Sends the actual HTTP response for a non-healthcheck request.
func sendResponse(w http.ResponseWriter, platform string, xOrigin string){
	m := tc{platform,"http://www.foo.com"}
	b, err := json.Marshal(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Data-Origin", xOrigin)
		w.Write(b)
		return
	}
}

// The healthcheck function.  Checks the status of Redis and Cassandra.  Returns
// HTTP 200 if they are healthy, HTTP 500 if they are not.
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
