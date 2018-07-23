package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// Task represents anything additional tasks except parsing & modifying the hosts file
type Task interface {
	Execute()
}

// HostsFileModifyingTask DO parse the hosts file & write new entries depending on the config file
type HostsFileModifyingTask struct {
	env         string
	isGroupName bool
	args        []string
	parser      EnvParser
}

// Execute is the implementation of the interface
func (h HostsFileModifyingTask) Execute() {
	var b bytes.Buffer
	hostFilePath := fmt.Sprintf("%s\\%s", os.Getenv("SystemRoot"), `\System32\drivers\etc\hosts.`)
	if strings.ToLower(h.env) == "live" {
		ioutil.WriteFile(hostFilePath, b.Bytes(), 0644)
		return
	}

	parsed, err := h.parser.Parse()
	if err != nil {
		panic(err)
	}

	var list []string
	if h.isGroupName {
		list = h.resolveGroupNames(h.args, parsed.(envConfig).Group)
	} else {
		list = h.args
	}

	for _, host := range list {
		ip := h.getTargetIPAddress(host, parsed.(envConfig))
		if len(ip) > 0 {
			str := fmt.Sprintf("%s %s\n", ip, host)
			b.WriteString(str)
		}
	}
	ioutil.WriteFile(hostFilePath, b.Bytes(), 0644)
}

func (h HostsFileModifyingTask) resolveGroupNames(name []string, groupCol map[string][]string) []string {
	var list []string
	for _, g := range name {
		list = append(list, groupCol[g]...)
	}

	list = h.resolveEmbeddedGroupName(list, groupCol)
	return list
}

func (h HostsFileModifyingTask) resolveEmbeddedGroupName(list []string, groupCol map[string][]string) []string {
	var result, tmp []string
	for _, n := range list {
		if h.isEmbeddedGroupName(n) {
			tn := n[2 : len(n)-1]
			tmp = groupCol[tn]
			tmp = h.resolveEmbeddedGroupName(tmp, groupCol)
			result = append(result, tmp...)
		} else {
			result = append(result, n)
		}
	}
	return result
}

func (h HostsFileModifyingTask) isEmbeddedGroupName(name string) bool {
	return strings.HasPrefix(name, "${") && strings.HasSuffix(name, "}")
}

func (h HostsFileModifyingTask) getTargetIPAddress(host string, srcData envConfig) string {
	if strings.ToLower(h.env) == "local" {
		return "127.0.0.1"
	}
	rules := srcData.EnvRule[h.env]
	addrCollection := srcData.Address
	addrList, exists := addrCollection[host]
	if !exists {
		return ""
	}
	for _, addr := range addrList {
		for _, rule := range rules {
			if strings.HasPrefix(addr, rule) {
				return addr
			}
		}
	}
	return ""
}

// NewHostsFileModifyingTask is a task which parses the config file & write right host information to 'hosts' file
func NewHostsFileModifyingTask(env, groupOrHost string, args ...string) HostsFileModifyingTask {
	parser := NewYamlEnvParser("hc.config.yml", os.Getenv("HostChangerPath"))
	return HostsFileModifyingTask{
		env:         env,
		isGroupName: strings.ToLower(groupOrHost) == "group",
		parser:      parser,
		args:        args,
	}
}

// WindowCommandTask takes command name / arguments and executes it
type WindowCommandTask struct {
	name string
	args []string
}

// Execute is the implementation of 'interface Task'
func (wct WindowCommandTask) Execute() {
	wct.executeCommand()
}

func (wct WindowCommandTask) executeCommand() {
	cmd := exec.Command(wct.name, wct.args...)
	cmd.Run()
}

// NewWindowCommandTask creates a new WindowCommandTask instance
func NewWindowCommandTask(name string, args ...string) WindowCommandTask {
	return WindowCommandTask{
		name: name,
		args: args,
	}
}

// TaskPipeline manage and executes tasks in parallel
type TaskPipeline struct {
	tasks  []Task
	wg     sync.WaitGroup
	numCPU int
}

// Add adds task to it
func (tp *TaskPipeline) Add(t Task) {
	tp.tasks = append(tp.tasks, t)
}

// Run executes tasks concurrently
func (tp *TaskPipeline) Run() {
	numOfTasks := len(tp.tasks)
	if numOfTasks > 0 {
		tp.wg.Add(numOfTasks)
		for _, t := range tp.tasks {
			go func(task Task) {
				defer tp.wg.Done()
				task.Execute()
			}(t)
		}
		tp.wg.Wait()
	}
}

// NewTaskPipeline creates a new TaskPipeline instance
func NewTaskPipeline() TaskPipeline {
	return TaskPipeline{
		numCPU: runtime.NumCPU(),
	}
}
