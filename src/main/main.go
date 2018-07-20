package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	chromeDataDir  = os.Getenv("LOCALAPPDATA") + "\\Google\\Chrome\\User Data\\Default"
	chromeCacheDir = chromeDataDir + "\\Cache"
)

func main() {
	// get arguments and validate
	switchCommand := flag.NewFlagSet("switch", flag.ExitOnError)
	targetEnvironment := switchCommand.String("env", "live", "Target Environment you want to work. ( local | dev | test | pre | live )")
	groupOrName := switchCommand.String("target", "host", "Specify which is your target, hostname(s) or group name. ( host[default] | group )")
	if len(os.Args) < 2 {
		fmt.Println(`- Usage
			host-changer [command] [flag1] [flag2] [value] [value] ...

			- Available Commands
			* switch
				 available flags : env target

			`)
		os.Exit(0)
	}

	switch os.Args[1] {
	case "switch":
		switchCommand.Parse(os.Args[3:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	args := switchCommand.Args()
	if switchCommand.Parsed() {
		fmt.Printf("targetEnvironment : %s, groupOrName : %s, hosts : %s\n", *targetEnvironment, *groupOrName, args)
	}

	tp := NewTaskPipeline()
	tp.Add(NewHostsFileModifyingTask(*targetEnvironment, *groupOrName, args...))
	tp.Add(NewWindowCommandTask("TASKKILL", "/F", "/IM", "iexplore.exe"))
	tp.Add(NewWindowCommandTask("RunDll32.exe", "InetCpl.cpl,ClearMyTracksByProcess", "2"))
	tp.Add(NewWindowCommandTask("RunDll32.exe", "InetCpl.cpl,ClearMyTracksByProcess", "8"))
	tp.Add(NewWindowCommandTask("TASKKILL", "/F", "/IM", "chrome.exe"))
	tp.Add(NewWindowCommandTask("cmd", "/c", "DEL", "/Q", "/S", "/F", chromeDataDir+"\\*.*"))
	tp.Add(NewWindowCommandTask("cmd", "/c", "DEL", "/Q", "/F", chromeCacheDir+"\\*Cookies*.*"))
	tp.Add(NewWindowCommandTask("ipconfig", "/flushdns"))
	tp.Run()
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
