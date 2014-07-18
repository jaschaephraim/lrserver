package lrserver

func validateHello(hello *clientHello) bool {
	if hello.Command != "hello" {
		return false
	}
	for _, c := range hello.Protocols {
		for _, s := range serverHello.Protocols {
			if c == s {
				return true
			}
		}
	}
	return false
}

var serverHello = struct {
	Command    string   `json:"command"`
	Protocols  []string `json:"protocols"`
	ServerName string   `json:"serverName"`
}{
	"hello",
	[]string{
		"http://livereload.com/protocols/official-7",
		"http://livereload.com/protocols/official-8",
		"http://livereload.com/protocols/official-9",
		"http://livereload.com/protocols/2.x-origin-version-negotiation",
		"http://livereload.com/protocols/2.x-remote-control",
	},
	"collective-dev",
}

type clientHello struct {
	Command   string   `json:"command"`
	Protocols []string `json:"protocols"`
}

type serverReload struct {
	Command string `json:"command"`
	Path    string `json:"path"`
	LiveCSS bool   `json:"liveCSS"`
}

func newServerReload(file string) serverReload {
	return serverReload{
		Command: "reload",
		Path:    file,
		LiveCSS: LiveCSS,
	}
}

type serverAlert struct {
	Command string `json:"command"`
	Message string `json:"message"`
}

func newServerAlert(msg string) serverAlert {
	return serverAlert{
		Command: "alert",
		Message: msg,
	}
}
