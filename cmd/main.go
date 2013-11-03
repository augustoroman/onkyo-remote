package main

import (
	"github.com/augustoroman/onkyo-remote"
	"github.com/gobs/cmd"
	"log"
	"strings"
	"time"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}

func listAll() {
	for device := range eiscp.Discover(time.After(1*time.Second), nil) {
		log.Println(device.Info())
	}
}

type shell struct{ r eiscp.Device }

func (s shell) printInfo() {
	log.Println(s.r.Info())
	must(s.r.Send("PWR", "QSTN"))
	must(s.r.Send("AMT", "QSTN"))
	must(s.r.Send("MVL", "QSTN"))
	must(s.r.Send("SLP", "QSTN"))
	must(s.r.Send("DIM", "QSTN"))
	must(s.r.Send("IFA", "QSTN"))
	must(s.r.Send("SLI", "QSTN"))
	<-time.After(1 * time.Second)
}

func (s shell) send(args ...string) {
	if len(args) < 2 {
		log.Println("Send usage:  send <CMD> <ARG> [ARG]...")
		return
	}
	for i, arg := range args {
		args[i] = strings.ToUpper(arg)
	}

	err := s.r.Send(args[0], args[1:]...)
	if err != nil {
		log.Println("Error: ", err)
	}
	<-time.After(150 * time.Millisecond)
}

func (s shell) run() {
	commander := &cmd.Cmd{EnableShell: true}
	commander.Init()
	commander.Prompt = "> "
	commander.Add(cmd.Command{
		Name: "info",
		Help: "Print information about the current device.",
		Call: func(string) bool { s.printInfo(); return false },
	})
	commander.Add(cmd.Command{
		Name: "send",
		Help: "Send a command.",
		Call: func(args string) bool { s.send(strings.Fields(args)...); return false },
	})
	commander.Add(cmd.Command{
		Name: "get",
		Help: "Request status of a parameter.",
		Call: func(args string) bool { s.send(args, "QSTN"); return false },
	})
	commander.Add(cmd.Command{
		Name: "set",
		Help: "Set a parameter.",
		Call: func(args string) bool { s.send(strings.Fields(args)...); return false },
	})
	commander.Add(cmd.Command{
		Name: "quit",
		Help: "Quit",
		Call: func(string) bool { return true },
	})
	commander.CmdLoop()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Searching for Onkyo receiver:")
	r := <-eiscp.Discover(time.After(1*time.Second), nil)

	if r == nil {
		log.Println("Could not find any receivers.  Exiting.")
		return
	}

	go func() {
		for m := range r.Messages() {
			log.Println("Resp: ", pretty(m))
		}
	}()

	shell{r}.run()

	// must(r.Send("PWR", "01"))
	// must(r.Send("PWR", "QSTN"))

	// time.Sleep(20 * time.Second)

	// must(r.Send("PWR", "00"))
	// must(r.Send("PWR", "QSTN"))

	// log.Println("Complicated")
	// errs := make(chan error)
	// devices := eiscp.Discover(10*time.Second, errs)
	// for {
	// 	log.Println("Waiting...")
	// 	select {
	// 	case d, ok := <-devices:
	// 		if !ok {
	// 			break
	// 		}
	// 		log.Println(d.Info())
	// 	case e := <-errs:
	// 		log.Println("ERROR:", e)
	// 		break
	// 	}
	// }
}
