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
	finder      EnvFinder
}

// Execute is the implementation of the interface
func (h HostsFileModifyingTask) Execute() {
	if err := h.parser.Parse(); err != nil {
		panic(err)
	}

	config := h.parser.GetParsedData().(envConfig)
	var list []string
	if h.isGroupName {
		for _, g := range h.args {
			temp := config.Group[g]
			list = append(list, temp...)
		}
	} else {
		list = h.args
	}

	var b bytes.Buffer
	for _, host := range list {
		ip := h.getTargetIPAddress(host)
		if len(ip) > 0 {
			str := fmt.Sprintf("%s %s\n", ip, host)
			b.WriteString(str)
		}
	}
	// match the rule
	hostFilePath := os.Getenv("SystemRoot") + `\System32\drivers\etc\hosts`
	ioutil.WriteFile(hostFilePath, b.Bytes(), 0644)
}

func (h HostsFileModifyingTask) getTargetIPAddress(host string) string {
	rules := h.parser.GetParsedData().(envConfig).EnvRule[h.env]
	addrCollection := h.parser.GetParsedData().(envConfig).Address
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
		finder:      NewYamlEnvFinder(parser),
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
