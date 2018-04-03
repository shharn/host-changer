package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	//"regexp"
)

var (
	hostFilePath = os.Getenv("SystemRoot") + `\System32\drivers\etc\hosts`
	hostEnvMap   = map[string][]string{
		"local": []string{"127.0.0.1"},
		"dev":   []string{"211.218.231."},
		"test":  []string{"125.141.158."},
		"pre":   []string{"183.110.0.", "222.122.222."},
	}
	NEXT_LINE = "\r\n"
)

func main() {
	switchCommand := flag.NewFlagSet("switch", flag.ExitOnError)
	targetEnvironment := switchCommand.String("env", "live", "Target Environment you want to work. ( local | dev | test | pre | live )")

	if len(os.Args) < 2 {
		fmt.Println(`- Usage
			host-changer [command] [flag] [value] [value] ...
			
			- Available Commands
			switch
			`)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "switch":
		switchCommand.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	hosts := switchCommand.Args()
	if switchCommand.Parsed() {
		fmt.Printf("targetEnvironment : %s, hosts : %s\n", *targetEnvironment, hosts)
	}

	content, err := ioutil.ReadFile(hostFilePath)
	if err != nil {
		fmt.Println("Fail to read file")
		os.Exit(1)
	}

	splitted := strings.Split(string(content), NEXT_LINE)
	currentHostGroup := ""
	for idx := 0; idx < len(splitted); idx++ {
		singleLine := strings.TrimSpace(splitted[idx])
		if len(singleLine) > 0 && isHostGroupDeclaration(singleLine) {
			temp := strings.TrimSpace(strings.Split(singleLine, "##")[1])
			// oh, this is the target host we want to change !!
			if index, exists := contains(hosts, temp); exists == true {
				// remove this host from the list so that we won't check again
				hosts = remove(hosts, index)
				currentHostGroup = temp
			} else {
				currentHostGroup = ""
			}
		} else if len(currentHostGroup) > 0 {
			tempHost := strings.Fields(singleLine)
			if len(tempHost) == 2 && tempHost[1] == currentHostGroup {
				tempHostIP := strings.TrimSpace(tempHost[0])
				if isHostIPCommented(tempHostIP) {
					tempHostIP = tempHostIP[1:]
					if *targetEnvironment != "live" && isTargetEnvHostIP(hostEnvMap[*targetEnvironment], tempHostIP) {
						// remove leading hashbang
						singleLine = singleLine[1:]
					}
				} else { // In case of already set host ip
					if !isTargetEnvHostIP(hostEnvMap[*targetEnvironment], tempHostIP) || *targetEnvironment == "live" {
						singleLine += "#" + singleLine
					}
				}
				splitted[idx] = singleLine
			}
		}
	}
	result := strings.Join(splitted, NEXT_LINE)
	fmt.Println(result)
	ioutil.WriteFile(hostFilePath+".test", []byte(result), 0664)
}

func contains(array []string, target string) (int, bool) {
	for index, value := range array {
		if value == target {
			return index, true
		}
	}
	return -1, false
}

func remove(arr []string, index int) []string {
	arr[index] = arr[len(arr)-1]
	return arr[:len(arr)-1]
}

func isHostGroupDeclaration(str string) bool {
	return strings.HasPrefix(str, "##")
}

func isHostIPCommented(ip string) bool {
	return strings.HasPrefix(ip, "#")
}

func isTargetEnvHostIP(list []string, ip string) bool {
	for _, ipPrefix := range list {
		if strings.HasPrefix(ip, ipPrefix) {
			return true
		}
	}
	return false
}
