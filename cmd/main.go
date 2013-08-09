package main

import (
	"github.com/augustoroman/onkyo-remote"
	"log"
	"time"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Simple")
	r := <-eiscp.Discover(1*time.Second, nil)

	// for device := range eiscp.Discover(1*time.Second, nil) {
	// 	log.Println(device.Info())
	// 	r = device
	// }

	go func() {
		for m := range r.Messages() {
			log.Println("Resp: ", pretty(m))
		}
	}()

	log.Println(r.Info())
	must(r.Send("PWR", "QSTN"))
	must(r.Send("AMT", "QSTN"))
	must(r.Send("MVL", "QSTN"))
	must(r.Send("SLP", "QSTN"))
	must(r.Send("DIM", "QSTN"))
	must(r.Send("IFA", "QSTN"))
	must(r.Send("SLI", "QSTN"))
	// must(r.Send("PWR", "01"))
	// must(r.Send("PWR", "QSTN"))

	time.Sleep(20 * time.Second)

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
