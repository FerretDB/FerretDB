---
sidebar_position: 6
---

# Go

To use FerretDB with the Go programming language, you need to install the MongoDB driver, as FerretDB is compatible with the MongoDB API. Run the following commands to install the necessary packages:

```
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/mongo/options
go get go.mongodb.org/mongo-driver/bson
```

These packages include the MongoDB driver and some useful options for interacting with the database.

Now, let's move on to the code. Create a file named main.go and insert the following code:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    // Connecting to the database
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    // Checking the connection
    err = client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Connected to MongoDB!")

    // Selecting the database and the collection
    collection := client.Database("test").Collection("users")

    // Creating a new user
    user := bson.D{{"name", "John"}, {"age", 30}}
    insertResult, err := collection.InsertOne(context.TODO(), user)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("User inserted with ID:", insertResult.InsertedID)

    // Listing the users
    var results []*bson.D
    findOptions := options.Find()
    cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
    if err != nil {
        log.Fatal(err)
    }

    for cur.Next(context.TODO()) {
        var elem bson.D
        err := cur.Decode(&elem)
        if err != nil {
            log.Fatal(err)
        }

        results = append(results, &elem)
    }

    if err := cur.Err(); err != nil {
        log.Fatal(err)
    }

    cur.Close(context.TODO())

    fmt.Println("Users:")
    for _, result := range results {
        fmt.Println(result)
    }
}

```
This code will connect to FerretDB, insert a new user into the users collection of the test database, and then list all the users in that collection. Make sure that FerretDB is running and accessible on the specified port (27017 in this example).
