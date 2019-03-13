package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cyrus-and/gdb"
)

var breakpointHitNotification = make(chan int)

func handleNotifications(notification map[string]interface{}) {
	if notification["class"] == "library-loaded" || notification["class"] == "library-unloaded" {
		return
	}
	jsonStr, _ := json.Marshal(notification)
	fmt.Println(string(jsonStr))

	if notification["class"] == "stopped" {
		breakpointHitNotification <- 1
	}
}

// InitGDB initializes the GDB interpreter
func InitGDB(filename string) {
	// start a new instance and pipe the target output to stdout
	gdb, _ := gdb.New(handleNotifications)
	go io.Copy(os.Stdout, gdb)
	go io.Copy(gdb, os.Stdin)

	// load and run a program
	result, err := gdb.Send("file-exec-and-symbols", filename)

	if err != nil {
		log.Fatal(err)
	}

	handleNotifications(result)

	result, err = gdb.Send("break-insert", "4")
	if err != nil {
		log.Fatal(err)
	}

	handleNotifications(result)

	gdb.Send("exec-run")
	<-breakpointHitNotification

	fmt.Println("-------------------- Starting EXEC-NEXT")
	gdb.Send("exec-next")
	fmt.Println("-------------------- Finished EXEC-NEXT")
	<-breakpointHitNotification
	fmt.Println("-------------------- Got 'stop' notification")

	fmt.Println("-------------------- Starting DATA-EVAL-EXPRESSION")
	gdb.Send("data-evaluate-expression", "init_debugger()")
	fmt.Println("-------------------- Starting EXEC-RUN")
	gdb.Send("exec-run")
	fmt.Println("-------------------- Finished EXEC-RUN")
	<-breakpointHitNotification
	fmt.Println("-------------------- Got 'stop' notification")

	gdb.Exit()
}
