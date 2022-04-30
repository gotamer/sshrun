package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/gotamer/sshrun"
)

var (
	Info = *log.Default()

	help  = flag.Bool("h", false, "Display help")
	debug = flag.Bool("d", false, "Verbose/debug mode")

	host = flag.String("host", "", "Remote host name")
	port = flag.Int("post", 22, "Remote port address")
	user = flag.String("user", "", "Remote host user name")
	pass = flag.String("pass", "", "Remote host user password, leave blank for ssh key login")

	cmd = flag.String("cmd", sshrun.CMD_LS, "command to run on remote host")
)

func init() {
	log.SetPrefix("ERR ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	Info.SetOutput(ioutil.Discard)
}

func main() {
	if *debug {
		Info = *log.Default()
		Info.SetPrefix("INF ")
		Info.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	}
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	var session = sshrun.NewSSH(*host, *port, *user, *pass, *cmd, *debug)
	defer session.End()
	fmt.Println("SSH Session Command: ", session.Cmd)
	fmt.Println(session.BOut.String())
	fmt.Println(session.BErr.String())
}
