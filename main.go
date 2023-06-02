package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

type Param struct {
	Name string
}

func (param *Param) build() []string {
	deleted_args := make([]int, 0, len(os.Args))
	param.Name = "Coin"
	for i, arg := range os.Args {
		switch arg {
		case "--name", "-n":
			if i+1 < len(os.Args) {
				param.Name = os.Args[i+1]
				deleted_args = append(deleted_args, i, i+1)
			} else {
				panic(fmt.Errorf("incorrect argument for name"))
			}
		}
	}
	args := make([]string, 0, len(os.Args))
	in := func(idx int, arr []int) bool {
		for _, v := range arr {
			if v == idx {
				return true
			}
		}
		return false
	}
	for i, arg := range os.Args {
		if !in(i, deleted_args) {
			args = append(args, arg)
		}
	}
	return args
}
func main() {

	param := new(Param)
	args := param.build()
	switch os.Args[1] {
	case "exec":
		pocket(os.Args[2:], param)
	case "coin":
		coin(param, args[2], args[3:])
	default:
		panic("Incorrect arguments")
	}
}

func pocket(args []string, param *Param) {
	cmd := exec.Command("/proc/self/exe", append([]string{"coin"}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	must(cmd.Run())
}

func coin(param *Param, command string, args []string) {
	syscall.Sethostname([]byte(param.Name))
	syscall.Chroot("./ubuntu-fs")
	syscall.Chdir("/")
	syscall.Mount("proc", "proc", "proc", 0, "")

	cgroups()
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(cmd.Run())
	syscall.Unmount("/proc", 0)
}

func cgroups() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	must(os.Mkdir(filepath.Join(pids, "pocket"), 0755))
	must(ioutil.WriteFile(filepath.Join(pids, "pocket/pids.max"), []byte("20"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, "pocket/notify_on_release"), []byte("1"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, "pocket/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func must(err error) {
	if err != nil {
		log.Fatal("error :", err)
	}
}
