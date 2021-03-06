package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

//struct for save server auth data in a json file
type ServerAuth struct {
	Username  string `json:"username"`
	IpAddress string `json:"ip_addr"`
	Port      string `json:"port"`
}

//define variable
var auth ServerAuth
var jsonBool bool
var jsonFilename string
var PasswordBool bool
var Password string

func main() {

	//flag for server auth data
	flag.Parse()

	//if FileExists(jsonFilename) == true, use json auth
	if FileExists(jsonFilename) {
		jsonAuth(jsonFilename)

	}

	if auth.Username == "" && auth.IpAddress == "" {
		fmt.Println("[-]Please insert username and IP Address.")
		os.Exit(1)
	}

	//check if username and IpAddress are empty
	if auth.Username == "" {
		fmt.Println("[-]Please insert username.")
		os.Exit(1)
	}

	if auth.IpAddress == "" {
		fmt.Println("[-]Please insert IP Address.")
		os.Exit(1)
	}

	// if passwordBool == True enter password
	if PasswordBool {
		fmt.Print("Enter password: ")
		tmp_passwd, err := terminal.ReadPassword(0)
		CheckErr(err)
		Password = string(tmp_passwd)

	}

	//if '-j' used, save auth to json
	if jsonBool {
		saveToJson("auth.json")
	}

	conn_config := &ssh.ClientConfig{User: auth.Username, Auth: []ssh.AuthMethod{ssh.Password(Password)}, HostKeyCallback: ssh.InsecureIgnoreHostKey()}
	conn_addr := net.JoinHostPort(auth.IpAddress, auth.Port)

	//creating new connection
	conn, err := ssh.Dial("tcp", conn_addr, conn_config)
	CheckErr(err)

	//requesting a new session
	session, err := conn.NewSession()
	CheckErr(err)
	defer session.Close()

	//redirect IO of Server at the Client
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	stdin, err := session.StdinPipe()
	CheckErr(err)

	// requesting pseudoterminal
	terminal := ssh.TerminalModes{
		ssh.ECHO: 0,
	}
	err = session.RequestPty("vt220", 40, 130, terminal)
	CheckErr(err)

	//shell
	err = session.Shell()
	CheckErr(err)

	for {
		io.Copy(stdin, os.Stdin)
	}

}

func init() {

	//username & password
	flag.StringVar(&auth.Username, "u", "", "username")
	flag.BoolVar(&PasswordBool, "p", false, "use this flag for insert password")

	//server ip address & port
	flag.StringVar(&auth.IpAddress, "A", "", "server ip address")
	flag.StringVar(&auth.Port, "P", "22", "server port")

	//json auth file
	flag.BoolVar(&jsonBool, "j", false, "save the auth config in a file")
	flag.StringVar(&jsonFilename, "c", "", "load the auth file")

}

func saveToJson(filename string) {

	file, err := json.MarshalIndent(auth, "", "  ")
	CheckErr(err)

	_ = ioutil.WriteFile(filename, file, 0644)

}

func jsonAuth(authFilename string) {

	data, err := ioutil.ReadFile(authFilename)
	CheckErr(err)

	err = json.Unmarshal(data, &auth)
	CheckErr(err)

}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
