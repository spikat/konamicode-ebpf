package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/signal"

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

func main() {
	// Initialize the managers
	if err := m.Init(bytes.NewReader(Probe)); err != nil {
		panic(fmt.Errorf("failed to init manager: %w", err))
	}

	// Start
	if err := m.Start(); err != nil {
		panic(fmt.Errorf("failed to start manager: %w", err))
	}

	logrus.Println("manager successfully started")

	logrus.Println("=> Cmd+C to stop")
	wait()

	if err := m.Stop(manager.CleanAll); err != nil {
		logrus.Fatal(err)
	}
}

// wait - Waits until an interrupt or kill signal is sent
func wait() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig
	fmt.Println()
}
