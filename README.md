# traffic-cop-go
This Proof of concept project is a Go lang version of a java project by the same name. Its sole purpose is to sit on top of a 'legacy' stack and a new 'microservices' stack and route traffic to one or the other based on how a given user has been configured in the database. As the new stack becomes more stable more and more users will be migrated to use it simply by undating the database records.

It will first check for the record in Redis. If the record is not there then it will query the Cassandra database. If it is found in the Cassandra database it will then insert the record into Redis.

## Getting started
1. Install Go
2. Install Redis
3. Install Cassandra
4. Install cqlsh
5. Run the following cql commands to create the keyspace, the tables, 
and populate them with sample data

```SQL
CREATE KEYSPACE ks1 WITH REPLICATION = {'class':'SimpleStrategy', 'replication_factor':1};

CREATE TABLE IF NOT EXISTS ks1.users (users int primary key, platform varchar);

COPY ks1.users FROM 'sample-data.csv' WITH HEADER=true;
```

## Installing the app
1. cd $GOPATH/src/github.com/stevenk15/traffic-cop-go
2. go install

## Running the app
1. cd $GOPATH/bin
2. ./traffic-cop-go

The application is available on port 5000.
 
http://localhost:5000/svc/v1/traffic-cop?userid=7777777