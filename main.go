package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/victorgeel/PsiTurtle/src/libpsiphon"
	"github.com/victorgeel/libinject"
	"github.com/victorgeel/liblog"
	"github.com/victorgeel/proxyrotator"
	"github.com/aztecrabbit/libredsocks"
	"github.com/victorgeel/libutils"
)

const (
	    appName    = "•Psiphon-Unlimited•"
        appVersionName = "Psi-pro-Mod"
        appVersionCode = "1.02.102724"

        ReleaseYear    = "2024"
        ModifiedAuthor = "••Victor⌐⁠■⁠-⁠■ Geek••"
)

var (
	InterruptHandler = new(libutils.InterruptHandler)
	Redsocks         = new(libredsocks.Redsocks)
)

type Config struct {
	ProxyRotator *libproxyrotator.Config
	Inject       *libinject.Config
	PsiphonCore  int
	Psiphon      *libpsiphon.Config
}

func init() {
	InterruptHandler.Handle = func() {
		libpsiphon.Stop()
		libredsocks.Stop(Redsocks)
		liblog.LogKeyboardInterrupt()
	}
	InterruptHandler.Start()
}

func GetConfigPath(filename string) string {
	return libutils.GetConfigPath("PsiTurtle-run", filename)
}

func main() {
	liblog.Header(
		[]string{
			fmt.Sprintf("%s [%s Version. %s]", appName, appVersionName, appVersionCode),
			fmt.Sprintf("(©) %s %s.", ReleaseYear, ModifiedAuthor),
		},
		liblog.Colors["Y1"],
	)

	config := new(Config)
	defaultConfig := new(Config)
	defaultConfig.ProxyRotator = libproxyrotator.DefaultConfig
	defaultConfig.Inject = libinject.DefaultConfig
	defaultConfig.Inject.Type = 2
	defaultConfig.Inject.Rules = map[string][]string{
		"akamai.net:80": []string{
			"www.pubgmobile.com",
                        "m.mobilelegends.com",
			"a1815.dscv.akamai.net",
                        "a1845.dscb.akamai.net",
                        "a1815.dscb.akamai.net",
                        "m.mobilelegends.com.akamaized.net",
                        "esports.pubgmobile.com",
                        "play.mobilelegends.com",
                        "www.pubgmobile.com.cdn.ettdnsv.com",
                        "www.pubgmobile.com.edgesuite.net",
                },
	}
	defaultConfig.Inject.Payload = ""
	defaultConfig.Inject.Timeout = 5
	defaultConfig.PsiphonCore = 6
	defaultConfig.Psiphon = libpsiphon.DefaultConfig

	if runtime.GOOS == "windows" {
		defaultConfig.Psiphon.CoreName += ".exe"
	}

	libutils.JsonReadWrite(GetConfigPath("config.json"), config, defaultConfig)

	var flagPro = true
	var flagRefresh = false
	var flagVerbose = false
	var flagFrontend string
	var flagWhitelist string

	flag.BoolVar(&flagPro, "pro", flagPro, "Pro Version?")
	flag.BoolVar(&flagRefresh, "refresh", flagRefresh, "Refresh Data")
	flag.BoolVar(&flagVerbose, "verbose", flagVerbose, "Verbose Log?")
	flag.StringVar(&flagFrontend, "f", flagFrontend, "-f frontend-domains (e.g. -f cdn.com,cdn.com:443)")
	flag.StringVar(&flagWhitelist, "w", flagWhitelist, "-w whitelist-request (e.g. -w akamai.net:80)")
	flag.IntVar(&config.Inject.MeekType, "mt", config.Inject.MeekType, "-mt meek type (0 and 1 for fastly)")
	flag.IntVar(&config.PsiphonCore, "c", config.PsiphonCore, "-c core (e.g. -c 4) (1 for Pro Version)")
	flag.StringVar(&config.Psiphon.Region, "r", config.Psiphon.Region, "-r region (e.g. -r sg)")
	flag.IntVar(&config.Psiphon.Tunnel, "t", config.Psiphon.Tunnel, "-t tunnel (e.g. -t 4) (1 for Reconnect Version)")
	flag.IntVar(&config.Psiphon.TunnelWorkers, "tw", config.Psiphon.TunnelWorkers, "-tw tunnel-workers (e.g. -tw 6) (8 for Pro Version)")
	flag.IntVar(&config.Psiphon.KuotaDataLimit, "l", config.Psiphon.KuotaDataLimit, "-l limit (in MB) (e.g. -l 4) (0 for Pro Version (unlimited))")
	flag.Parse()

	if !flagPro {
		config.Psiphon.Authorizations = make([]string, 0)
	}

	if flagRefresh {
		libpsiphon.RemoveData()
	}

	if flagFrontend != "" || flagWhitelist != "" {
		if flagFrontend == "" {
			flagFrontend = "*"
		}
		if flagWhitelist == "" {
			flagWhitelist = "*:*"
		}

		config.Inject.Rules = map[string][]string{
			flagWhitelist: strings.Split(flagFrontend, ","),
		}
	}

	ProxyRotator := new(libproxyrotator.ProxyRotator)
	ProxyRotator.Config = config.ProxyRotator

	Inject := new(libinject.Inject)
	Inject.Redsocks = Redsocks
	Inject.Config = config.Inject

	go ProxyRotator.Start()
	go Inject.Start()

	time.Sleep(200 * time.Millisecond)

	liblog.LogInfo("Domain Fronting running on port "+Inject.Config.Port, "INFO", liblog.Colors["B1"])
	liblog.LogInfo("Proxy Rotator running on port "+ProxyRotator.Config.Port, "INFO", liblog.Colors["B1"])

	if _, err := os.Stat(libutils.RealPath(config.Psiphon.CoreName)); os.IsNotExist(err) {
		liblog.LogInfo(
			fmt.Sprintf(
				"Exception:\n\n"+
					"|	 File '%s' not exist!\n"+
					"|	 Exiting...\n"+
					"|\n",
				config.Psiphon.CoreName,
			),
			"INFO", liblog.Colors["M1"],
		)
		return
	}

	Redsocks.Config = libredsocks.DefaultConfig
	Redsocks.Config.LogOutput = GetConfigPath("redsocks.log")
	Redsocks.Config.ConfigOutput = GetConfigPath("redsocks.conf")
	Redsocks.Start()

	for i := 1; i <= config.PsiphonCore; i++ {
		Psiphon := new(libpsiphon.Psiphon)
		Psiphon.ProxyRotator = ProxyRotator
		Psiphon.Config = config.Psiphon
		Psiphon.ProxyPort = Inject.Config.Port
		Psiphon.KuotaData = libpsiphon.DefaultKuotaData
		Psiphon.ListenPort = libutils.Atoi(ProxyRotator.Config.Port) + i
		Psiphon.Verbose = flagVerbose

		go Psiphon.Start()
	}

	InterruptHandler.Wait()
}
