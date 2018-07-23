package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	chromeDataDir  = os.Getenv("LOCALAPPDATA") + "\\Google\\Chrome\\User Data\\Default"
	chromeCacheDir = chromeDataDir + "\\Cache"
)

func main() {
	// get arguments and validate
	cmdSwitch := &cobra.Command{
		Use:   "switch -env [env name] -target [host or group] host1 host2 host3",
		Short: "switch the hosts",
		Long: `switch is for changing hosts to target env ip
		[Example] switch -env test -target host www.test.com
							switch -env pre -target group member`,
		Run: func(cmd *cobra.Command, args []string) {},
	}

	var env, target string
	var args []string
	cmdSwitch.Flags().StringVarP(&env, "env", "e", "test", "-env [local | dev | test | pre | live]. when 'live' flag is specified, empty host file is created")
	cmdSwitch.Flags().StringVarP(&target, "target", "t", "host", "-target [host | group].")
	cmdSwitch.Flags().StringSliceVarP(&args, "list", "l", []string{}, "-list [group name1,name2... | hostname1,hostname2,hostname3...")

	rootCmd := &cobra.Command{Use: "hc"}
	rootCmd.AddCommand(cmdSwitch)
	rootCmd.Execute()

	tp := NewTaskPipeline()
	tp.Add(NewHostsFileModifyingTask(env, target, args...))
	tp.Add(NewWindowCommandTask("TASKKILL", "/F", "/IM", "iexplore.exe"))
	tp.Add(NewWindowCommandTask("RunDll32.exe", "InetCpl.cpl,ClearMyTracksByProcess", "2"))
	tp.Add(NewWindowCommandTask("RunDll32.exe", "InetCpl.cpl,ClearMyTracksByProcess", "8"))
	tp.Add(NewWindowCommandTask("TASKKILL", "/F", "/IM", "chrome.exe"))
	tp.Add(NewWindowCommandTask("cmd", "/c", "DEL", "/Q", "/S", "/F", chromeDataDir+"\\*.*"))
	tp.Add(NewWindowCommandTask("cmd", "/c", "DEL", "/Q", "/F", chromeCacheDir+"\\*Cookies*.*"))
	tp.Add(NewWindowCommandTask("ipconfig", "/flushdns"))
	tp.Run()
}
