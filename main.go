package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/DataDog/dev/konamicode-ebpf/pkg/sound"
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
	PerfMaps: []*manager.PerfMap{
		{
			Map: manager.Map{Name: "notes"},
		},
	},
}

type KonamiCodeCtx struct {
	sp *sound.SineWavePlayer
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

func openURL(url string) {
	cmd := exec.Command("xdg-open", url)

	// get the initial user to open the browser
	sudo_user := os.Getenv("SUDO_USER")
	if sudo_user == "" {
		sudo_user = "root" // try to open with root, but should not works
	} else {
		os.Setenv("HOME", "/home/"+sudo_user)
	}
	u, err := user.Lookup(sudo_user)
	if err != nil {
		logrus.Printf("user lookup failed: %s\n", err)
		return
	}
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}

	_, err = cmd.CombinedOutput()
	if err != nil {
		logrus.Printf("exec failed: %s\n", err)
	}
}

func playSong(sp *sound.SineWavePlayer) {
	const (
		freqC = 523
		freqE = 659
		freqG = 784
	)
	sp.QueueNote(sound.Note{Freq: freqC, Duration: 500})
	sp.QueueNote(sound.Note{Freq: freqE, Duration: 500})
	sp.QueueNote(sound.Note{Freq: freqG, Duration: 500})
}

func start_konamicode_watcher(sp *sound.SineWavePlayer) {
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
					playSong(sp)
					openURL("https://en.wikipedia.org/wiki/Konami_Code")
					return
				}
			}
		}
	}()
}

func (kcCtx *KonamiCodeCtx) NotesPerfMapHandler(cpu int, data []byte, perfmap *manager.PerfMap, manager *manager.Manager) {
	dataLen := uint64(len(data))
	if dataLen < 16 {
		logrus.Println("got note with wrong size %d\n", dataLen)
		return
	}

	freq := ByteOrder.Uint64(data[0:8])
	duration := ByteOrder.Uint64(data[8:16])

	n := sound.Note{Freq: int64(freq), Duration: int64(duration)}
	kcCtx.sp.QueueNote(n)
}

func main() {
	kcCtx := &KonamiCodeCtx{}
	sp, ready, err := sound.NewSineWavePlayer(48000, 2, sound.FormatSignedInt16LE)
	if err != nil {
		panic(fmt.Errorf(err.Error()))
	}
	<-ready
	var wg sync.WaitGroup
	wg.Add(1)
	go sp.PlayLoop(&wg)
	kcCtx.sp = sp

	notesPerfMap, ok := m.GetPerfMap("notes")
	if !ok {
		panic(fmt.Errorf("failed to get notes perfmap"))
	}
	notesPerfMap.PerfMapOptions = manager.PerfMapOptions{
		DataHandler: kcCtx.NotesPerfMapHandler,
	}

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

	start_konamicode_watcher(sp)

	logrus.Println("=> Cmd+C to stop")
	wait()
	sp.Close()
	wg.Wait()
}

// wait - Waits until an interrupt or kill signal is sent
func wait() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig
	fmt.Println()
}
