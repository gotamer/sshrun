package sshrun

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	CMD_WHOAMI = "whoami"
	CMD_PWD    = "pwd"
	CMD_LS     = "ls -Alh"
	CMD_DF     = "df -h"
	CMD_OFF    = "sudo shutdown +1"
	CMD_REBOOT = "sudo shutdown -r +1"
)

var (
	Info = *log.Default()
	home string
)

func init() {
	log.SetPrefix("ERR ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	Info = *log.Default()
	Info.SetPrefix("INF ")
	Info.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	home = os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
}

type sshSession struct {
	Host string
	Port int
	User string
	Pass string
	Cmd  string
	BOut bytes.Buffer
	BErr bytes.Buffer
	auth []ssh.AuthMethod
	conn *ssh.Client
	sess *ssh.Session
}

func NewSSH(host string, port int, user, pass, cmd string, debug bool) (s *sshSession) {
	s = new(sshSession)
	if !debug {
		Info.SetOutput(ioutil.Discard)
	}
	s.Host = host
	s.Port = port
	s.User = user
	s.Pass = pass
	s.Cmd = cmd
	s.authenticate()
	s.start()
	return
}

func (s *sshSession) authenticate() {
	if s.Pass == "" {
		var pathPrivateKey = filepath.Join(home, ".ssh", "id_rsa")

		// A public key may be used to authenticate against the remote
		// server by using an unencrypted PEM-encoded private key file.
		//
		// If you have an encrypted private key, the crypto/x509 package
		// can be used to decrypt it.
		key, err := ioutil.ReadFile(pathPrivateKey)
		if err != nil {
			log.Fatalf("unable to read private key: %v", err)
		}

		// Create the Signer for this private key.
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			log.Fatalf("unable to parse private key: %v", err)
		}

		s.auth = []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		}
	} else {
		s.auth = []ssh.AuthMethod{
			ssh.Password(s.Pass),
		}
	}
}

func (s *sshSession) start() {
	// ssh config
	var pathHostKey = filepath.Join(home, ".ssh", "known_hosts")

	// ssh config
	hostKeyCallback, err := knownhosts.New(pathHostKey)
	if err != nil {
		log.Fatal(err)
	}

	config := &ssh.ClientConfig{
		User:            s.User,
		Auth:            s.auth,
		HostKeyCallback: hostKeyCallback,
	}

	// connect to ssh server
	s.conn, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port), config)
	if err != nil {
		log.Fatal(err)
	}

	Info.Println("SSH Connected")

	s.sess, err = s.conn.NewSession()
	if err != nil {
		s.conn.Close()
		log.Fatal(err)
	}

	Info.Println("SSH Session Started")

	s.sess.Stdout = &s.BOut
	s.sess.Stderr = &s.BErr

	if err := s.sess.Run(s.Cmd); err != nil {
		s.sess.Close()
		s.conn.Close()
		log.Fatal(err)
	}

	Info.Println("SSH Session Command: ", s.Cmd)
}

/*
	// configure terminal mode
	modes := ssh.TerminalModes{
		ssh.ECHO: 0, // supress echo

	}
	// run terminal session
	if err := s.sess.RequestPty("xterm", 50, 80, modes); err != nil {
		log.Fatal(err)
	}

	// start remote shell
	if err := s.sess.Shell(); err != nil {
		log.Fatal(err)
	}
	Info.Println("Shell Started")

*/

func (s *sshSession) End() {
	var err error
	if err = s.sess.Close(); err != nil {
		if err != io.EOF {
			log.Println(err)
		}
	}
	if err = s.conn.Close(); err != nil {
		if err != io.EOF {
			log.Println(err)
		}
	}
}
