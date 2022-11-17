package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"time"

	manager "github.com/DataDog/ebpf-manager"
	"github.com/sirupsen/logrus"
)

//go:embed ebpf/bin/probe.o
var Probe []byte

var m = &manager.Manager{
	Probes: []*manager.Probe{
		{
			ProbeIdentificationPair: manager.ProbeIdentificationPair{
				EBPFSection:  "kprobe/input_handle_event",
				EBPFFuncName: "kprobe_input_handle_event",
			},
		},
	},
}

func checkKonamicode() (uint32, error) {
	m, _, err := m.GetMap("konamicode_activation_counter")
	if err != nil {
		logrus.Printf("checkKonamicode error: %v\n", err)
		return 0, err
	}
	var key, val uint32
	err = m.Lookup(&key, &val)
	if err != nil {
		logrus.Printf("checkKonamicode error: %v\n", err)
		return 0, err
	}
	return val, nil
}

func start_konamicode_watcher() {
	go func() {
		konamicode_check := time.NewTicker(time.Second)

		for {
			select {
			case _ = <-konamicode_check.C:
				val, err := checkKonamicode()
				if err != nil {
					continue
				} else if val != 0 {
					logrus.Printf("KONAMI CODE ACTIVATED \\o/ !\n")
				}
			}
		}
	}()
}

func main() {
	// Initialize the managers
	if err := m.Init(bytes.NewReader(Probe)); err != nil {
		panic(fmt.Errorf("failed to init manager: %w", err))
	}

	// Start
	if err := m.Start(); err != nil {
		panic(fmt.Errorf("failed to start manager: %w", err))
	}
	defer m.Stop(manager.CleanAll)
	logrus.Println("manager successfully started")

	start_konamicode_watcher()

	logrus.Println("=> Cmd+C to stop")
	wait()
}

// wait - Waits until an interrupt or kill signal is sent
func wait() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig
	fmt.Println()
}
