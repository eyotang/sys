package sys

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func CmdOut(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

func CmdOutBytes(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.Bytes(), err
}

func CmdOutNoLn(name string, arg ...string) (out string, err error) {
	out, err = CmdOut(name, arg...)
	if err != nil {
		return
	}

	return strings.TrimSpace(string(out)), nil
}

func CmdRunWithTimeout(cmd *exec.Cmd, timeout time.Duration) (error, bool) {
	var err error

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		log.Printf("timeout, process:%s will be killed", cmd.Path)

		go func() {
			<-done // allow goroutine to exit
		}()

		//IMPORTANT: cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} is necessary before cmd.Start()
		if p, err := os.FindProcess(-cmd.Process.Pid); err != nil {
			log.Println("can't find process, error:", err)
		} else {
			if err = p.Signal(syscall.SIGKILL); err != nil {
				log.Println("kill failed, error:", err)
			}
		}

		return err, true
	case err = <-done:
		return err, false
	}
}
