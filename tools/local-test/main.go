package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/golang/glog"
)

func main() {
	// set up logging
	flag.Parse()
	defer glog.Flush()

	// check go path is set
	if os.Getenv("GOPATH") == "" {
		glog.Fatal("GOPATH not set, please set GOPATH and try again")
	}

	for _, algorithm := range []string{"delegated", "simple", "fpaxos"} {
		serverFile := os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/" + algorithm + "/server.conf"
		clientFile := os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/" + algorithm + "/client.conf"
		statsFile := "latency_" + algorithm + ".csv"

		//Create temp directory
		dir, err := ioutil.TempDir("", "IosTests_"+algorithm)
		if err != nil {
			glog.Fatal(err)
		}
		defer os.RemoveAll(dir) // clean up

		// start server
		server := exec.Command(os.Getenv("GOPATH")+"/bin/ios", "-id", "0", "-algorithm", serverFile, "-disk", dir, "-stderrthreshold", "WARNING")
		server.Stderr = os.Stderr
		err = server.Start()
		if err != nil {
			glog.Fatal("Error starting server process. Error: ", err.Error())
		}

		time.Sleep(1 * time.Second)

		//start client
		client := exec.Command(os.Getenv("GOPATH")+"/bin/test", "-stat", statsFile, "-algorithm", clientFile, "-check", "true", "-stderrthreshold", "WARNING")
		client.Stderr = os.Stderr

		err = client.Start()
		if err != nil {
			glog.Fatal("Error starting server process. Error: ", err.Error())
		}

		// wait for workload to finish
		timer := time.AfterFunc(10*time.Second, func() {
			client.Process.Kill()
			server.Process.Kill()
			glog.Fatal("Client did not complete in time")
		})
		err = client.Wait()
		if err != nil {
			glog.Fatal("Client did not exit cleanly. Error: ", err.Error())
		}
		timer.Stop()

		//close server
		err = server.Process.Signal(syscall.SIGKILL)
		if err != nil {
			glog.Fatal("Error killing server process. Error: ", err.Error())
		}

		// restart server
		server = exec.Command(os.Getenv("GOPATH")+"/bin/ios", "-id", "0", "-algorithm", serverFile, "-disk", dir, "-stderrthreshold", "WARNING")
		server.Stderr = os.Stderr
		err = server.Start()
		if err != nil {
			glog.Fatal("Error starting server process. Error: ", err.Error())

		}
		time.Sleep(1 * time.Second)

		//start client
		client = exec.Command(os.Getenv("GOPATH")+"/bin/test", "-stat", statsFile, "-algorithm", clientFile, "-stderrthreshold", "WARNING")
		client.Stderr = os.Stderr

		err = client.Start()
		if err != nil {
			glog.Fatal("Error starting server process. Error: ", err.Error())
		}

		// wait for workload to finish
		timer = time.AfterFunc(10*time.Second, func() {
			client.Process.Kill()
			server.Process.Kill()
			glog.Fatal("Client did not complete in time")
		})
		err = client.Wait()
		if err != nil {
			glog.Fatal("Client did not exit cleanly. Error: ", err.Error())
		}
		timer.Stop()

		//close server
		err = server.Process.Signal(syscall.SIGKILL)
		if err != nil {
			glog.Fatal("Error killing server process. Error: ", err.Error())
		}
	}
}
