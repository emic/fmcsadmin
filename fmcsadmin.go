/**
 * fmcsadmin
 * (c) 2017-2018 Emic Corporation <https://www.emic.co.jp/>
 * This software is distributed under the Apache License, Version 2.0,
 * see LICENSE.txt and NOTICE.txt for more information.
 *
 * FileMaker is a trademark of FileMaker, Inc., registered in the U.S. and other countries.
 */
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mattn/go-scan"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh/terminal"
)

var version string

var availableDeleteCommand = false

type cli struct {
	outStream, errStream io.Writer
}

type accountInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type dbInfo struct {
	Key string `json:"key"`
}

type output struct {
	Token  string `json:"token"`
	Result int    `json:"result"`
}

type messageInfo struct {
	Message string `json:"message"`
}

type params struct {
	key       string
	message   string
	gracetime int
	retry     int
}

type commandOptions struct {
	helpFlag    bool
	versionFlag bool
	yesFlag     bool
	statsFlag   bool
	fqdn        string
	username    string
	password    string
	key         string
	message     string
	clientID    int
	graceTime   int
}

func main() {
	cli := &cli{outStream: os.Stdout, errStream: os.Stderr}
	os.Exit(cli.Run(os.Args))
}

func (c *cli) Run(args []string) int {
	var exitStatus int

	token := ""
	exitStatus = 0
	helpFlag := false
	versionFlag := false
	yesFlag := false
	statsFlag := false
	graceTime := 90
	fqdn := ""
	username := ""
	password := ""
	key := ""
	clientID := -1
	message := ""

	commandOptions := commandOptions{}
	commandOptions.helpFlag = false
	commandOptions.versionFlag = false
	commandOptions.yesFlag = false
	commandOptions.statsFlag = false
	commandOptions.fqdn = ""
	commandOptions.username = ""
	commandOptions.password = ""
	commandOptions.key = ""
	commandOptions.message = ""
	commandOptions.clientID = -1
	commandOptions.graceTime = 90

	cmdArgs, cFlags, err := getFlags(args, commandOptions)
	if err != nil {
		fmt.Fprintln(c.outStream, flag.ErrHelp)
		exitStatus = outputInvalidCommandErrorMessage(c)
		return exitStatus
	}

	helpFlag = cFlags.helpFlag
	versionFlag = cFlags.versionFlag
	yesFlag = cFlags.yesFlag
	statsFlag = cFlags.statsFlag
	graceTime = cFlags.graceTime
	key = cFlags.key
	username = cFlags.username
	password = cFlags.password
	clientID = cFlags.clientID
	message = cFlags.message

	fqdn = cFlags.fqdn
	endpoint := getHostName(fqdn)
	u, _ := url.Parse(endpoint)

	retry := 3
	if len(username) > 0 && len(password) > 0 {
		// Don't retry when specifying username and password
		retry = 0
	}

	if len(cmdArgs) > 0 {
		switch strings.ToLower(cmdArgs[0]) {
		case "close":
			res := ""
			if yesFlag == true {
				res = "y"
			} else {
				r := bufio.NewReader(os.Stdin)
				fmt.Fprint(c.outStream, "fmcsadmin: really close database(s)? (y, n) ")
				input, _ := r.ReadString('\n')
				res = strings.ToLower(strings.TrimSpace(input))
			}
			if res == "y" {
				token, exitStatus, err = login(endpoint, username, password, params{retry: retry})
				if token != "" && err == nil {
					u.Path = path.Join(getAPIBasePath(), "databases")
					args = []string{""}
					if len(cmdArgs[1:]) > 0 {
						args = cmdArgs[1:]
					}
					idList, nameList, _ := getDatabases(u.String(), token, args, "NORMAL")
					if len(idList) > 0 {
						for i := 0; i < len(idList); i++ {
							fmt.Fprintln(c.outStream, "File Closing: "+nameList[i])
						}
						connectedClients := getClients(u.String(), token, args, "")
						for i := 0; i < len(idList); i++ {
							u.Path = path.Join(getAPIBasePath(), "databases", strconv.Itoa(idList[i]), "close")
							exitStatus, err = sendRequest("PUT", u.String(), token, params{message: message})
							if exitStatus == 0 && err == nil && len(message) == 0 && len(connectedClients) == 0 {
								// Don't output this message
								//   1. when using "-m (--message)" option
								//   2. when the clients connected to the specified databases are existing
								fmt.Fprintln(c.outStream, "File Closed: "+nameList[i])
							}
						}
					} else {
						exitStatus = 10904
					}
					if exitStatus > 0 {
						outputErrorMessage(exitStatus)
					}
					logout(endpoint, token)
				}
			}
		case "delete":
			if len(cmdArgs[1:]) > 0 && availableDeleteCommand == true {
				res := ""
				if yesFlag == true {
					res = "y"
				} else {
					r := bufio.NewReader(os.Stdin)
					fmt.Fprint(c.outStream, "fmcsadmin: really delete a schedule? (y, n) ")
					input, _ := r.ReadString('\n')
					res = strings.ToLower(strings.TrimSpace(input))
				}
				if res == "y" {
					token, exitStatus, err = login(endpoint, username, password, params{retry: retry})
					if token != "" && err == nil {
						switch strings.ToLower(cmdArgs[1]) {
						case "schedule":
							id := 0
							if len(cmdArgs) >= 3 {
								sid, err := strconv.Atoi(cmdArgs[2])
								if err == nil {
									id = sid
								}
							}
							if id > 0 {
								u.Path = path.Join(getAPIBasePath(), "schedules", strconv.Itoa(id))
								scheduleName := getScheduleName(u.String(), token, id)
								exitStatus, err = sendRequest("DELETE", u.String(), token, params{})
								if exitStatus == 0 && err == nil {
									if scheduleName != "" {
										fmt.Fprintln(c.outStream, "Schedule Deleted: "+scheduleName)
									}
								} else {
									outputErrorMessage(exitStatus)
								}
							} else {
								fmt.Fprintln(c.outStream, "Error: 10600 (Schedule at specified index no longer exists)")
								exitStatus = 10600
							}
						}
						logout(endpoint, token)
					}
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "disable":
			if len(cmdArgs[1:]) > 0 {
				res := ""
				if yesFlag == true {
					res = "y"
				} else {
					r := bufio.NewReader(os.Stdin)
					fmt.Fprint(c.outStream, "fmcsadmin: really disable a schedule? (y, n) ")
					input, _ := r.ReadString('\n')
					res = strings.ToLower(strings.TrimSpace(input))
				}
				if res == "y" {
					token, exitStatus, err = login(endpoint, username, password, params{retry: retry})
					if token != "" && err == nil {
						switch strings.ToLower(cmdArgs[1]) {
						case "schedule":
							id := 0
							if len(cmdArgs) >= 3 {
								sid, err := strconv.Atoi(cmdArgs[2])
								if err == nil {
									id = sid
								}
							}
							if id > 0 {
								u.Path = path.Join(getAPIBasePath(), "schedules", strconv.Itoa(id), "disable")
								exitStatus, err = sendRequest("PUT", u.String(), token, params{})
								if exitStatus == 0 && err == nil {
									u.Path = path.Join(getAPIBasePath(), "schedules")
									listSchedules(u.String(), token, id)
								} else {
									outputErrorMessage(exitStatus)
								}
							} else {
								fmt.Fprintln(c.outStream, "Error: 10600 (Schedule at specified index no longer exists)")
								exitStatus = 10600
							}
						}
						logout(endpoint, token)
					}
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "disconnect":
			if len(cmdArgs[1:]) > 0 {
				res := ""
				if yesFlag == true {
					res = "y"
				} else {
					r := bufio.NewReader(os.Stdin)
					fmt.Fprint(c.outStream, "fmcsadmin: really disconnect client(s)? (y, n) ")
					input, _ := r.ReadString('\n')
					res = strings.ToLower(strings.TrimSpace(input))
				}
				if res == "y" {
					token, exitStatus, err = login(endpoint, username, password, params{retry: retry})
					if token != "" && err == nil {
						switch strings.ToLower(cmdArgs[1]) {
						case "client":
							id := 0
							if len(cmdArgs) >= 3 {
								cid, err := strconv.Atoi(cmdArgs[2])
								if err == nil {
									id = cid
								}
							}
							if id > -1 {
								// check the client connection
								u.Path = path.Join(getAPIBasePath(), "databases")
								idList := getClients(u.String(), token, []string{""}, "NORMAL")
								connected := false
								if len(idList) > 0 {
									if id == 0 {
										connected = true
									} else {
										for i := 0; i < len(idList); i++ {
											if id == idList[i] {
												connected = true
											}
										}
									}
								}

								if connected == true {
									u.Path = path.Join(getAPIBasePath(), "clients", strconv.Itoa(id), "disconnect")
									exitStatus, err = sendRequest("PUT", u.String(), token, params{message: message, gracetime: graceTime})
									if exitStatus == 0 {
										fmt.Fprintln(c.outStream, "Client(s) being disconnected.")
									} else {
										outputErrorMessage(exitStatus)
									}
								} else {
									fmt.Fprintln(c.outStream, "No client is connected.")
								}
							}
						}
						logout(endpoint, token)
					}
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "enable":
			if len(cmdArgs[1:]) > 0 {
				token, exitStatus, err = login(endpoint, username, password, params{retry: retry})
				if token != "" && err == nil {
					switch strings.ToLower(cmdArgs[1]) {
					case "schedule":
						id := 0
						if len(cmdArgs) >= 3 {
							sid, err := strconv.Atoi(cmdArgs[2])
							if err == nil {
								id = sid
							}
						}
						if id > 0 {
							u.Path = path.Join(getAPIBasePath(), "schedules", strconv.Itoa(id), "enable")
							exitStatus, err = sendRequest("PUT", u.String(), token, params{})
							if exitStatus == 0 && err == nil {
								u.Path = path.Join(getAPIBasePath(), "schedules")
								listSchedules(u.String(), token, id)
							} else {
								outputErrorMessage(exitStatus)
							}
						} else {
							fmt.Fprintln(c.outStream, "Error: 10600 (Schedule at specified index no longer exists)")
							exitStatus = 10600
						}
					}
					logout(endpoint, token)
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "help":
			if len(cmdArgs[1:]) > 0 {
				switch strings.ToLower(cmdArgs[1]) {
				case "commands":
					if availableDeleteCommand == true {
						fmt.Fprint(c.outStream, commandListHelpTextTemplate2)
					} else {
						fmt.Fprint(c.outStream, commandListHelpTextTemplate)
					}
				case "options":
					fmt.Fprint(c.outStream, optionListHelpTextTemplate)
				case "close":
					fmt.Fprint(c.outStream, closeHelpTextTemplate)
				case "delete":
					if availableDeleteCommand == true {
						fmt.Fprint(c.outStream, deleteHelpTextTemplate)
					} else {
						fmt.Fprint(c.outStream, helpTextTemplate)
					}
				case "disable":
					fmt.Fprint(c.outStream, disableHelpTextTemplate)
				case "disconnect":
					fmt.Fprint(c.outStream, disconnectHelpTextTemplate)
				case "enable":
					fmt.Fprint(c.outStream, enableHelpTextTemplate)
				case "help":
					fmt.Fprint(c.outStream, helpTextTemplate)
				case "list":
					fmt.Fprint(c.outStream, listHelpTextTemplate)
				case "open":
					fmt.Fprint(c.outStream, openHelpTextTemplate)
				case "pause":
					fmt.Fprint(c.outStream, pauseHelpTextTemplate)
				case "resume":
					fmt.Fprint(c.outStream, resumeHelpTextTemplate)
				case "run":
					fmt.Fprint(c.outStream, runHelpTextTemplate)
				case "send":
					fmt.Fprint(c.outStream, sendHelpTextTemplate)
				case "status":
					fmt.Fprint(c.outStream, statusHelpTextTemplate)
				default:
					fmt.Fprint(c.outStream, helpTextTemplate)
				}
			} else {
				fmt.Fprint(c.outStream, helpTextTemplate)
			}
		case "list":
			if len(cmdArgs[1:]) > 0 {
				switch strings.ToLower(cmdArgs[1]) {
				case "clients":
					token, exitStatus, err = login(endpoint, username, password, params{retry: retry})
					if token != "" && err == nil {
						id := -1
						if statsFlag == true {
							id = 0
						}
						u.Path = path.Join(getAPIBasePath(), "databases")
						exitStatus = listClients(u.String(), token, id)
						logout(endpoint, token)
					}
				case "files":
					token, exitStatus, err = login(endpoint, username, password, params{retry: retry})
					if token != "" && err == nil {
						idList := []int{-1}
						if statsFlag == true {
							idList = []int{0}
						}
						u.Path = path.Join(getAPIBasePath(), "databases")
						exitStatus = listFiles(u.String(), token, idList)
						logout(endpoint, token)
					}
				case "schedules":
					token, exitStatus, err = login(endpoint, username, password, params{retry: retry})
					if token != "" && err == nil {
						u.Path = path.Join(getAPIBasePath(), "schedules")
						listSchedules(u.String(), token, 0)
						logout(endpoint, token)
					}
				default:
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "open":
			token, exitStatus, err = login(endpoint, username, password, params{retry: retry})
			if token != "" && err == nil {
				u.Path = path.Join(getAPIBasePath(), "databases")
				args = []string{""}
				if len(cmdArgs[1:]) > 0 {
					args = cmdArgs[1:]
				}
				idList, nameList, hintList := getDatabases(u.String(), token, args, "CLOSED")
				if len(idList) > 0 {
					for i := 0; i < len(idList); i++ {
						fmt.Fprintln(c.outStream, "File Opening: "+nameList[i])
					}
					for i := 0; i < len(idList); i++ {
						u.Path = path.Join(getAPIBasePath(), "databases", strconv.Itoa(idList[i]), "open")
						exitStatus, err = sendRequest("PUT", u.String(), token, params{key: key})
						if exitStatus == 0 && err == nil {
							// Note: FileMaker Admin API (Trial) does not validate the encryption key.
							//       You receive a result code of 0 even if you enter an invalid key.
							var openedID []int
							for value := 0; ; {
								value++
								u.Path = path.Join(getAPIBasePath(), "databases")
								openedID, _, _ = getDatabases(u.String(), token, []string{strconv.Itoa(idList[i])}, "NORMAL")
								if len(openedID) > 0 || value > 3 {
									break
								}
								time.Sleep(1 * time.Second)
							}
							if len(openedID) > 0 {
								fmt.Fprintln(c.outStream, "File Opened: "+nameList[i])
							} else {
								fmt.Fprintln(c.outStream, "Fail to open encrypted database. The correct password must be supplied with the --key option. (Hint: "+hintList[i]+")")
								fmt.Fprintln(c.outStream, "File Closed: "+nameList[i])
							}
						}
					}
				} else {
					exitStatus = 10904
				}
				if exitStatus > 0 {
					outputErrorMessage(exitStatus)
				}
				logout(endpoint, token)
			}
		case "pause":
			token, exitStatus, err = login(endpoint, username, password, params{retry: retry})
			if token != "" && err == nil {
				u.Path = path.Join(getAPIBasePath(), "databases")
				args = []string{""}
				if len(cmdArgs[1:]) > 0 {
					args = cmdArgs[1:]
				}
				idList, nameList, _ := getDatabases(u.String(), token, args, "NORMAL")
				if len(idList) > 0 {
					for i := 0; i < len(idList); i++ {
						fmt.Fprintln(c.outStream, "File Pausing: "+nameList[i])
					}
					for i := 0; i < len(idList); i++ {
						u.Path = path.Join(getAPIBasePath(), "databases", strconv.Itoa(idList[i]), "pause")
						exitStatus, err = sendRequest("PUT", u.String(), token, params{})
						if exitStatus == 0 && err == nil {
							fmt.Fprintln(c.outStream, "File Paused: "+nameList[i])
						}
					}
				} else {
					exitStatus = 10904
				}
				if exitStatus > 0 {
					outputErrorMessage(exitStatus)
				}
				logout(endpoint, token)
			}
		case "resume":
			token, exitStatus, err = login(endpoint, username, password, params{retry: retry})
			if token != "" && err == nil {
				u.Path = path.Join(getAPIBasePath(), "databases")
				args = []string{""}
				if len(cmdArgs[1:]) > 0 {
					args = cmdArgs[1:]
				}
				idList, nameList, _ := getDatabases(u.String(), token, args, "PAUSED")
				if len(idList) > 0 {
					for i := 0; i < len(idList); i++ {
						fmt.Fprintln(c.outStream, "File Resuming: "+nameList[i])
					}
					for i := 0; i < len(idList); i++ {
						u.Path = path.Join(getAPIBasePath(), "databases", strconv.Itoa(idList[i]), "resume")
						exitStatus, err = sendRequest("PUT", u.String(), token, params{})
						if exitStatus == 0 && err == nil {
							fmt.Fprintln(c.outStream, "File Resumed: "+nameList[i])
						}
					}
				} else {
					exitStatus = 10904
				}
				if exitStatus > 0 {
					outputErrorMessage(exitStatus)
				}
				logout(endpoint, token)
			}
		case "run":
			if len(cmdArgs[1:]) > 0 {
				token, exitStatus, err = login(endpoint, username, password, params{retry: retry})
				if token != "" && err == nil {
					switch strings.ToLower(cmdArgs[1]) {
					case "schedule":
						id := 0
						if len(cmdArgs) >= 3 {
							sid, err := strconv.Atoi(cmdArgs[2])
							if err == nil {
								id = sid
							}
						}
						if id > 0 {
							u.Path = path.Join(getAPIBasePath(), "schedules", strconv.Itoa(id), "run")
							exitStatus, err = sendRequest("PUT", u.String(), token, params{})
							if exitStatus == 0 && err == nil {
								u.Path = path.Join(getAPIBasePath(), "schedules", strconv.Itoa(id))
								scheduleName := getScheduleName(u.String(), token, id)
								if scheduleName != "" {
									fmt.Fprintln(c.outStream, "Schedule '"+scheduleName+"' will run now.")
								}
							} else {
								outputErrorMessage(exitStatus)
							}
						} else {
							fmt.Fprintln(c.outStream, "Error: 10600 (Schedule at specified index no longer exists)")
							exitStatus = 10600
						}
					}
					logout(endpoint, token)
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "send":
			token, exitStatus, err = login(endpoint, username, password, params{retry: retry})
			if token != "" && err == nil {
				id := 0
				if clientID > -1 {
					id = clientID
				}
				u.Path = path.Join(getAPIBasePath(), "databases")
				args = []string{""}
				if len(cmdArgs[1:]) > 0 {
					args = cmdArgs[1:]
				}
				idList := getClients(u.String(), token, args, "NORMAL")
				if len(args) == 0 || (len(args) > 0 && id > 0) {
					idList = append(idList, id)
				}
				for i := 0; i < len(idList); i++ {
					if (clientID > -1 && clientID == idList[i]) || clientID == -1 {
						u.Path = path.Join(getAPIBasePath(), "clients", strconv.Itoa(idList[i]), "message")
						exitStatus = sendMessage(u.String(), token, message)
						if clientID > -1 {
							break
						}
					}
				}
				logout(endpoint, token)
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "status":
			if len(cmdArgs[1:]) > 0 {
				token, exitStatus, err = login(endpoint, username, password, params{retry: retry})
				if token != "" && err == nil {
					switch strings.ToLower(cmdArgs[1]) {
					case "client":
						id := 0
						if len(cmdArgs) >= 3 {
							sid, err := strconv.Atoi(cmdArgs[2])
							if err == nil {
								id = sid
							}
						}
						if id > 0 {
							u.Path = path.Join(getAPIBasePath(), "databases")
							exitStatus = listClients(u.String(), token, id)
						}
					case "file":
						if len(cmdArgs[2:]) > 0 {
							u.Path = path.Join(getAPIBasePath(), "databases")
							idList, _, _ := getDatabases(u.String(), token, cmdArgs[2:], "")
							exitStatus = listFiles(u.String(), token, idList)
						} else {
							fmt.Fprintln(c.outStream, "Error: 10001 (Invalid parameter)")
							exitStatus = 10001
						}
					}
					logout(endpoint, token)
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		default:
			if helpFlag == true {
				fmt.Fprint(c.outStream, helpTextTemplate)
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		}
	} else {
		if versionFlag == true {
			fmt.Fprintln(c.outStream, "fmcsadmin "+version)
		} else {
			fmt.Fprint(c.outStream, helpTextTemplate)
		}
	}

	return exitStatus
}

func getFlags(args []string, cFlags commandOptions) ([]string, commandOptions, error) {
	var cmdArgs []string
	helpFlag := false
	versionFlag := false
	yesFlag := false
	statsFlag := false
	fqdn := ""
	username := ""
	password := ""
	key := ""
	message := ""
	clientID := -1
	graceTime := 90

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.Usage = func() {}

	flags.BoolVar(&helpFlag, "h", false, "Print help pages.")
	flags.BoolVar(&helpFlag, "help", false, "Print help pages.")
	flags.BoolVar(&versionFlag, "v", false, "Print version information.")
	flags.BoolVar(&versionFlag, "version", false, "Print version information.")
	flags.BoolVar(&yesFlag, "y", false, "Automatically answer yes to all command prompts.")
	flags.BoolVar(&yesFlag, "yes", false, "Automatically answer yes to all command prompts.")
	flags.BoolVar(&statsFlag, "s", false, "Return FILE or CLIENT stats.")
	flags.BoolVar(&statsFlag, "stats", false, "Return FILE or CLIENT stats.")
	flags.StringVar(&fqdn, "fqdn", "", "Specify the Fully Qualified Domain Name of a remote server.")
	flags.StringVar(&username, "u", "", "Username to use to authenticate with the server.")
	flags.StringVar(&username, "username", "", "Username to use to authenticate with the server.")
	flags.StringVar(&password, "p", "", "Password to use to authenticate with the server.")
	flags.StringVar(&password, "password", "", "Password to use to authenticate with the server.")
	flags.StringVar(&key, "key", "", "Specify the database encryption password.")
	flags.StringVar(&message, "m", "", "Specify a text message to send to clients.")
	flags.StringVar(&message, "message", "", "Specify a text message to send to clients.")
	flags.IntVar(&clientID, "c", -1, "Specify a client number to send a message.")
	flags.IntVar(&clientID, "client", -1, "Specify a client number to send a message.")
	flags.IntVar(&graceTime, "t", 90, "Specify time in seconds before client is forced to disconnect.")
	flags.IntVar(&graceTime, "gracetime", 90, "Specify time in seconds before client is forced to disconnect.")

	buf := &bytes.Buffer{}
	flags.SetOutput(buf)
	if len(args) == 2 {
		flag.ErrHelp = errors.New("Invalid option: " + args[1])
	}

	err := flags.Parse(args[1:])
	if err != nil {
		return cmdArgs, cFlags, err
	}

	cFlags.helpFlag = cFlags.helpFlag || helpFlag
	cFlags.versionFlag = cFlags.versionFlag || versionFlag
	cFlags.yesFlag = cFlags.yesFlag || yesFlag
	cFlags.statsFlag = cFlags.statsFlag || statsFlag
	if cFlags.fqdn == "" {
		cFlags.fqdn = fqdn
	}
	if cFlags.username == "" {
		cFlags.username = username
	}
	if cFlags.password == "" {
		cFlags.password = password
	}
	if cFlags.key == "" {
		cFlags.key = key
	}
	if cFlags.message == "" {
		cFlags.message = message
	}
	if cFlags.clientID == -1 {
		cFlags.clientID = clientID
	}
	if cFlags.graceTime == 90 {
		cFlags.graceTime = graceTime
	}

	cmdArgs = flags.Args()

	var resultArgs []string
	var subCmdArgs []string

	if len(cmdArgs) > 0 {
		resultArgs = append(resultArgs, cmdArgs[0])
	}

	subCommandOptions := commandOptions{}
	if len(cmdArgs) > 1 {
		subCmdArgs, subCommandOptions, _ = getFlags(cmdArgs[0:], cFlags)
		if len(subCmdArgs) > 0 {
			resultArgs = append(resultArgs, subCmdArgs[0:]...)
		}
		cFlags.helpFlag = cFlags.helpFlag || subCommandOptions.helpFlag
		cFlags.versionFlag = cFlags.versionFlag || subCommandOptions.versionFlag
		cFlags.yesFlag = cFlags.yesFlag || subCommandOptions.yesFlag
		cFlags.statsFlag = cFlags.statsFlag || subCommandOptions.statsFlag
		if cFlags.fqdn == "" {
			cFlags.fqdn = subCommandOptions.fqdn
		}
		if cFlags.username == "" {
			cFlags.username = subCommandOptions.username
		}
		if cFlags.password == "" {
			cFlags.password = subCommandOptions.password
		}
		if cFlags.key == "" {
			cFlags.key = subCommandOptions.key
		}
		if cFlags.message == "" {
			cFlags.message = subCommandOptions.message
		}
		if cFlags.clientID == -1 {
			cFlags.clientID = subCommandOptions.clientID
		}
		if cFlags.graceTime == 90 {
			cFlags.graceTime = subCommandOptions.graceTime
		}
	}

	return resultArgs, cFlags, nil
}

func outputInvalidCommandErrorMessage(c *cli) int {
	exitStatus := 248
	fmt.Fprintln(c.outStream, "Error: 11000 (Invalid command)")
	fmt.Fprint(c.outStream, helpTextTemplate)

	return exitStatus
}

func getHostName(fqdn string) string {
	endpoint := "http://127.0.0.1:8080"
	if len(fqdn) > 0 {
		endpoint = "https://" + strings.TrimSpace(fqdn)
	}

	return endpoint
}

func getAPIBasePath() string {
	path := "/admin/api/v1"

	return path
}

func getUsernameAndPassword(username string, password string) (string, string) {
	if len(username) == 0 {
		r := bufio.NewReader(os.Stdin)
		fmt.Print("username: ")
		input, _ := r.ReadString('\n')
		username = strings.TrimSpace(input)
	}

	if len(password) == 0 {
		fmt.Print("password: ")
		bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
		password = string(bytePassword)
		fmt.Printf("\n")
	}

	return username, password
}

func login(endpoint string, user string, pass string, p params) (string, int, error) {
	token := ""
	exitStatus := 0

	username, password := getUsernameAndPassword(user, pass)
	u, _ := url.Parse(endpoint)
	u.Path = path.Join(getAPIBasePath(), "user", "login")

	d := accountInfo{
		username,
		password,
	}
	output := output{}

	jsonStr, _ := json.Marshal(d)
	body, err := callURL("POST", u.String(), "", bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		fmt.Println("Error: 10502 (Host unreachable)")
		exitStatus = 10502
		return token, exitStatus, err
	}

	err = json.Unmarshal(body, &output)
	if err != nil {
		return token, exitStatus, err
	}

	if output.Result == 0 && err == nil {
		token = output.Token
	} else {
		if p.retry > 0 {
			fmt.Println("fmcsadmin: Permission denied, please try again.")
			token, exitStatus, err = login(endpoint, user, pass, params{retry: p.retry - 1})
		} else {
			fmt.Println("fmcsadmin: Permission denied.")
			fmt.Println("Error: 9 (Access denied)")
			exitStatus = 9
		}
	}

	return token, exitStatus, err
}

func logout(endpoint string, token string) {
	u, _ := url.Parse(endpoint)
	u.Path = path.Join(getAPIBasePath(), "user", "logout")
	sendRequest("POST", u.String(), token, params{})
}

func listClients(url string, token string, id int) int {
	body, err := callURL("GET", url, token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return -1
	}

	var v interface{}
	body2 := []byte(body)
	err = json.Unmarshal(body2, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	mode := "NORMAL"
	if id > -1 {
		mode = "DETAIL"
	}

	var count int
	var c []string
	var s string
	var s1 string
	var userName string
	var computerName string
	var extPriv string
	var ipAddress string
	var macAddress string
	var connectTime string
	var connectDuration string
	var appVersion string
	var appLanguage string
	var teamLicensed string
	var fileName string
	var accountName string
	var privsetName string
	var b bool
	var data [][]string

	err = scan.ScanTree(v, "/clients/clients", &c)
	count = len(c)

	if mode == "NORMAL" {
		if count > 0 {
			for i := 0; i < count; i++ {
				err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/status", &s)
				if s == "NORMAL" {
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/id", &s1)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/userName", &userName)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/computerName", &computerName)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/extpriv", &extPriv)
					data = append(data, []string{s1, userName, computerName, extPriv})
				}
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Client ID", "User Name", "Computer Name", "Ext Privilege"})
			table.SetAutoWrapText(false)
			for _, v := range data {
				table.Append(v)
			}
			table.Render()
		}
	} else {
		if count > 0 {
			for i := 0; i < count; i++ {
				err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/status", &s)
				if s == "NORMAL" {
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/id", &s1)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/userName", &userName)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/computerName", &computerName)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/extpriv", &extPriv)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/ipaddress", &ipAddress)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/macaddress", &macAddress)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/connectTime", &connectTime)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/connectDuration", &connectDuration)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/appVersion", &appVersion)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/appLanguage", &appLanguage)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/teamLicensed", &b)
					teamLicensed = ""
					if b == true {
						teamLicensed = "Yes"
					} else {
						teamLicensed = "No"
					}
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/guestFiles[0]/filename", &fileName)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/guestFiles[0]/accountName", &accountName)
					err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(i)+"]/guestFiles[0]/privsetName", &privsetName)
					data = append(data, []string{s1, userName, computerName, extPriv, ipAddress, macAddress, connectTime, connectDuration, appVersion, appLanguage, teamLicensed, fileName, accountName, privsetName})
				}
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Client ID", "User Name", "Computer Name", "Ext Privilege", "IP Address", "MAC Address", "Connect Time", "Duration", "App Version", "App Language", "User Connections License", "File Name", "Account Name", "Privilege Set"})
			table.SetAutoWrapText(false)
			for _, v := range data {
				table.Append(v)
			}
			table.Render()
		}
	}

	return 0
}

func listFiles(url string, token string, idList []int) int {
	body, err := callURL("GET", url, token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return -1
	}

	var v interface{}
	body2 := []byte(body)
	err = json.Unmarshal(body2, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	mode := "NORMAL"
	if (len(idList) == 1 && idList[0] > -1) || len(idList) > 1 {
		mode = "DETAIL"
	}

	var totalDbCount int
	var count int
	var c []string
	var s string
	var s1 string
	var fileName string
	var status string
	var extPriv string
	var isEncrypted string
	var num1 int
	var num2 int
	var b bool
	var data [][]string

	err = scan.ScanTree(v, "/totalDBCount", &totalDbCount)

	if mode == "NORMAL" {
		for i := 0; i < totalDbCount; i++ {
			err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/status", &s)
			if s == "NORMAL" {
				err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/folder", &s)
				fmt.Print(s)
				err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/filename", &s)
				fmt.Println(s)
			}
		}
	} else {
		for i := 0; i < totalDbCount; i++ {
			err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/id", &s1)
			for j := 0; j < len(idList); j++ {
				if s1 == strconv.Itoa(idList[j]) || idList[j] == 0 {
					err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/filename", &fileName)
					err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/clients", &num1)
					err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/size", &num2)
					err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/status", &status)

					extPriv = ""
					err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/enabledExtPrivileges", &c)
					count = len(c)
					for j := 0; j < count; j++ {
						err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/enabledExtPrivileges["+strconv.Itoa(j)+"]", &s)
						if extPriv == "" {
							extPriv = s
						} else {
							extPriv = extPriv + " " + s
						}
					}

					isEncrypted = ""
					err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/isEncrypted", &b)
					if b == true {
						isEncrypted = "Yes"
					} else {
						isEncrypted = "No"
					}

					data = append(data, []string{s1, fileName, strconv.Itoa(num1), strconv.Itoa(num2), status, extPriv, isEncrypted})
				}
			}
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "File", "Clients", "Size", "Status", "Enabled Extended Privileges", "Encrypted"})
		table.SetAutoWrapText(false)
		for _, v := range data {
			table.Append(v)
		}
		table.Render()
	}

	return 0
}

func listSchedules(url string, token string, id int) int {
	body, err := callURL("GET", url, token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return -1
	}

	var v interface{}
	body2 := []byte(body)
	err = json.Unmarshal(body2, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	var count int
	var c []string
	var s1 string
	var sID int
	var name string
	var taskType string
	var lastRun string
	var nextRun string
	var enabled string
	var status int
	var data [][]string

	err = scan.ScanTree(v, "/schedules", &c)
	count = len(c)

	if count > 0 {
		for i := 0; i < count; i++ {
			err = scan.ScanTree(v, "/schedules["+strconv.Itoa(i)+"]/id", &s1)
			err = scan.ScanTree(v, "/schedules["+strconv.Itoa(i)+"]/name", &name)
			err = scan.ScanTree(v, "/schedules["+strconv.Itoa(i)+"]/taskType", &taskType)
			err = scan.ScanTree(v, "/schedules["+strconv.Itoa(i)+"]/lastRun", &lastRun)
			err = scan.ScanTree(v, "/schedules["+strconv.Itoa(i)+"]/nextRun", &nextRun)
			err = scan.ScanTree(v, "/schedules["+strconv.Itoa(i)+"]/enabled", &enabled)
			if enabled == "false" {
				nextRun = "Disabled"
			}
			err = scan.ScanTree(v, "/schedules["+strconv.Itoa(i)+"]/status", &status)
			sID, _ = strconv.Atoi(s1)
			if id == sID || id == 0 {
				data = append(data, []string{s1, name, taskType, lastRun, nextRun, strconv.Itoa(status)})
			}
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		for _, v := range data {
			table.SetHeader([]string{"ID", "Name", "Type", "Last Completed", "Next Run", "Status"})
			table.Append(v)
		}
		table.Render()
	}

	return 0
}

func getScheduleName(url string, token string, id int) string {
	body, err := callURL("GET", url, token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	var v interface{}
	body2 := []byte(body)
	err = json.Unmarshal(body2, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	var count int
	var c []string
	var s1 string
	var sID int
	var name string

	err = scan.ScanTree(v, "/schedules", &c)
	count = len(c)

	if count > 0 {
		for i := 0; i < count; i++ {
			err = scan.ScanTree(v, "/schedules["+strconv.Itoa(i)+"]/id", &s1)
			err = scan.ScanTree(v, "/schedules["+strconv.Itoa(i)+"]/name", &name)
			sID, _ = strconv.Atoi(s1)
			if id == sID || id == 0 {
				return name
			}
		}
	}

	return ""
}

func sendMessage(url string, token string, message string) int {
	d := messageInfo{
		message,
	}
	jsonStr, _ := json.Marshal(d)
	body, err := callURL("POST", url, token, bytes.NewBuffer([]byte(jsonStr)))

	output := output{}
	err = json.Unmarshal(body, &output)
	if err != nil {
		fmt.Println(err.Error())
		return -1
	}

	return output.Result
}

func getDatabases(url string, token string, arg []string, status string) ([]int, []string, []string) {
	var fileName string
	var folderName string
	var idList []int
	var nameList []string
	var hintList []string
	var id int

	body, err := callURL("GET", url, token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return idList, nameList, hintList
	}

	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		fmt.Println(err.Error())
	}
	var totalDbCount int
	var fileStatus string
	var s1 string
	var s2 string
	var fileID string
	var decryptHint string

	err = scan.ScanTree(v, "/totalDBCount", &totalDbCount)
	for i := 0; i < totalDbCount; i++ {
		for j := 0; j < len(arg)+1; j++ {
			if j == len(arg) && j > 0 {
				break
			}

			fileName = ""
			folderName = ""
			if len(arg) > 0 {
				fileName = arg[j]
				if strings.Index(fileName, string(os.PathSeparator)) > -1 {
					folderName = fileName
				}
			}

			if len(folderName) == 0 {
				err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/status", &fileStatus)
				err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/filename", &s1)
				err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/decryptHint", &decryptHint)
				if regexp.MustCompile(`[0-9]`).Match([]byte(arg[j])) {
					// ID
					err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/id", &fileID)
					if fileID == arg[j] && (status == fileStatus || status == "") {
						nameList = append(nameList, s1)
						id, _ = strconv.Atoi(fileID)
						idList = append(idList, id)
						hintList = append(hintList, decryptHint)
					}
				} else {
					// name
					if (fileName == "" || comparePath(fileName, s1)) && (status == fileStatus || status == "") {
						nameList = append(nameList, s1)
						err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/id", &fileID)
						id, _ = strconv.Atoi(fileID)
						idList = append(idList, id)
						hintList = append(hintList, decryptHint)
					}
				}
			} else {
				err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/status", &fileStatus)
				err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/folder", &s1)
				err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/filename", &s2)
				err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/decryptHint", &decryptHint)
				if (status == fileStatus || status == "") && (comparePath(s1, folderName) || comparePath(s1+s2, fileName)) {
					nameList = append(nameList, s2)
					err = scan.ScanTree(v, "/files/files["+strconv.Itoa(i)+"]/id", &fileID)
					id, _ = strconv.Atoi(fileID)
					idList = append(idList, id)
					hintList = append(hintList, decryptHint)
				}
			}
		}
	}

	return idList, nameList, hintList
}

func getClients(url string, token string, arg []string, status string) []int {
	var fileName string
	var folderName string
	var idList []int
	var id int

	body, err := callURL("GET", url, token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return idList
	}

	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		fmt.Println(err.Error())
	}
	var clients []string
	var guestFiles []string
	var files []string
	var guestFileID string
	var guestFileName string
	var clientID string
	var fileID string
	var directory string

	for i := 0; i < len(arg)+1; i++ {
		if i == len(arg) && i > 0 {
			break
		}

		fileName = ""
		folderName = ""
		if len(arg) > 0 {
			fileName = arg[i]
			if strings.Index(fileName, string(os.PathSeparator)) > -1 {
				folderName = fileName
			}
		}

		err = scan.ScanTree(v, "/clients/clients", &clients)
		for j := 0; j < len(clients); j++ {
			err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(j)+"]/guestFiles", &guestFiles)
			for k := 0; k < len(guestFiles); k++ {
				err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(j)+"]/guestFiles["+strconv.Itoa(k)+"]/filename", &guestFileID)
				err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(j)+"]/guestFiles["+strconv.Itoa(k)+"]/filename", &guestFileName)
				if len(folderName) == 0 {
					if fileName == "" || comparePath(fileName, guestFileName) {
						err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(k)+"]/id", &clientID)
						id, _ = strconv.Atoi(clientID)
						idList = append(idList, id)
					}
				} else {
					err = scan.ScanTree(v, "/files/files", &files)
					for l := 0; k < len(files); l++ {
						err = scan.ScanTree(v, "/files/files["+strconv.Itoa(k)+"]/id", &fileID)
						err = scan.ScanTree(v, "/files/files["+strconv.Itoa(k)+"]/folder", &directory)
						if fileID == guestFileID {
							if comparePath(fileName, directory+guestFileName) {
								err = scan.ScanTree(v, "/clients/clients["+strconv.Itoa(k)+"]/id", &clientID)
								id, _ = strconv.Atoi(clientID)
								idList = append(idList, id)
							}
						}
					}
				}
			}
		}
	}

	return idList
}

func comparePath(name1 string, name2 string) bool {
	extName := ".fmp12"
	pathPrefix := []string{"filelinux:"}

	if name1 == name2 {
		return true
	} else if name1 == name2+extName {
		return true
	} else if name1+extName == name2 {
		return true
	}

	if strings.Index(name1, string(os.PathSeparator)) > -1 || strings.Index(name2, string(os.PathSeparator)) > -1 {
		for i := 0; i < len(pathPrefix); i++ {
			if pathPrefix[i]+name1 == name2 {
				return true
			} else if pathPrefix[i]+name1+extName == name2 {
				return true
			} else if name1 == pathPrefix[i]+name2 {
				return true
			} else if name1 == pathPrefix[i]+name2+extName {
				return true
			}
		}
	}

	return false
}

func outputErrorMessage(code int) {
	if code > 0 {
		fmt.Println("Error: " + strconv.Itoa(code) + " (" + getErrorDescription(code) + ")")
	}
}

func sendRequest(method string, url string, token string, p params) (int, error) {
	var jsonStr []byte
	if len(p.key) > 0 {
		d := dbInfo{
			p.key,
		}
		jsonStr, _ = json.Marshal(d)
	} else if len(p.message) > 0 {
		d := messageInfo{
			p.message,
		}
		jsonStr, _ = json.Marshal(d)
	} else {
		jsonStr = nil
	}
	body, err := callURL(method, url, token, bytes.NewBuffer([]byte(jsonStr)))
	if body != nil {
		output := output{}
		if json.Unmarshal(body, &output) == nil {
			return output.Result, nil
		}
	}
	if err != nil {
		return -1, err
	}

	return 0, nil
}

func callURL(method string, url string, token string, request io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, request)
	if err != nil {
		fmt.Println(err.Error())
		return []byte(""), err
	}

	if request == nil {
		req.Header.Set("Content-Length", "0")
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+strings.Replace(strings.Replace(token, "\n", "", -1), "\r", "", -1))
	}
	client := &http.Client{Timeout: time.Duration(5) * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return []byte(""), err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err.Error())
		return []byte(""), err
	}

	return body, err
}

func getErrorDescription(errorCode int) string {
	description := ""
	switch errorCode {
	case 8:
		description = "Empty result"
	case 9:
		description = "Insufficient privileges"
	case 212:
		description = "Invalid user account and/or password; please try again"
	case 802:
		description = "Unable to open the file"
	case 958:
		description = "Parameter missing"
	case 960:
		description = "Parameter is invalid"
	case 10600:
		description = "Schedule at specified index no longer exists"
	case 10601:
		description = "Schedule is misconfigured; invalid taskType or run status"
	case 10603:
		description = "Schedule can't be created or duplicated"
	case 10604:
		description = "Cannot enable schedule"
	case 10610:
		description = "No schedules created in configuration file"
	case 10611:
		description = "Schedule name is already used"
	case 10904:
		description = "No applicable files for this operation"
	case 10906:
		description = "Script is missing"
	case 11005:
		description = "Disconnect client invalid ID"
	case 25004:
		description = "Parameters are invalid"
	case 25006:
		description = "Invalid session error"
	default:
		description = ""
	}

	return description
}

var helpTextTemplate = `Usage: fmcsadmin [options] [COMMAND]

Description: 
    fmcsadmin is the command line tool to administer the Database Server 
	component of FileMaker Cloud via FileMaker Admin API.
	(FileMaker is a trademark of FileMaker, Inc., registered in the U.S.
	and other countries.)

    You can script many tasks with fmcsadmin by using a scripting language 
    that allows execution of shell or terminal commands.

    fmcsadmin HELP COMMANDS
       Lists available commands

    fmcsadmin HELP [COMMAND]
       Displays help on the specified COMMAND

    fmcsadmin HELP OPTIONS
       Lists available options

Author: 
    Emic Corporation <https://www.emic.co.jp/>

License:
    This software is distributed under the Apache License, Version 2.0,
    please see <https://github.com/emic/fmcsadmin/NOTICE.txt> for details.
`

var commandListHelpTextTemplate = `fmcsadmin commands are:

    CLOSE           Close databases
    DISABLE         Disable a schedule
    DISCONNECT      Disconnect clients
    ENABLE          Enable a schedule
    HELP            Get help pages
    LIST            List clients, databases, or schedules
    OPEN            Open databases
    PAUSE           Temporarily stop database access
    RESUME          Make paused databases available
    RUN             Run a schedule
    SEND            Send a message
    STATUS          Get status of clients or databases

`

var commandListHelpTextTemplate2 = `fmcsadmin commands are:

    CLOSE           Close databases
    DELETE          Delete a schedule
    DISABLE         Disable a schedule
    DISCONNECT      Disconnect clients
    ENABLE          Enable a schedule
    HELP            Get help pages
    LIST            List clients, databases, or schedules
    OPEN            Open databases
    PAUSE           Temporarily stop database access
    RESUME          Make paused databases available
    RUN             Run a schedule
    SEND            Send a message
    STATUS          Get status of clients or databases

`

var optionListHelpTextTemplate = `Many fmcsadmin commands take options and parameters.

Short Options:
    Specify single-character options after a single hyphen (-). You can 
    specify multiple options together. If an option requires a parameter, 
    that option is usually the last option that you specify. For example:
         fmcsadmin close -y -m 'Closing for maintenance' myData.fmp12 
    A space is optional between the option character and the parameter. 
    For example:
         fmcsadmin close -m Goodbye
    Note: Short options are case sensitive. 

Long Options: 
    Specify long options after two hyphens (--). Long options can be used 
    in scripts to increase readability. Long options are not case sensitive. 
    A space is required between the option and any parameters. For example:
         fmcsadmin close --yes --message "Closing for maintenance" myData.fmp12

Parameters: 
    Enclose any parameters that contain spaces in single or double quotation 
    marks (' or "). Symbols that may be interpreted by the shell must be 
    escaped, that is, preceded by a backslash character (\). Refer to the 
    documentation for your shell or command interpreter.

General Options: 
	--fqdn                     Specify the Fully Qualified Domain Name (FQDN)
	                           of a remote server via HTTPS.
    -h, --help                 Print help pages.
    -p pass, --password pass   Password to use to authenticate with the server.
    -u user, --username user   Username to use to authenticate with the server.
    -v, --version              Print version information.
    -y, --yes                  Automatically answer yes to all command prompts.

Options that apply to specific commands:
    -c NUM, --client NUM       Specify a client number to send a message.
    --key encryptpass          Specify the database encryption password.
    -m msg, --message msg      Specify a text message to send to clients. 
	-s, --stats                Return FILE or CLIENT stats.
    -t sec, --gracetime sec    Specify time in seconds before client is forced
                               to disconnect.
`

var closeHelpTextTemplate = `Usage: fmcsadmin CLOSE [FILE...] [PATH...] [options]

Description:
    Closes the specified databases (FILE) or all the hosted databases in the
    specified folders (PATH). If no FILE or PATH is specified, closes all 
    hosted databases. 

    To specify a database by its ID rather than its filename, first use the 
    LIST FILES -s command to get a list of databases and their IDs.

Options:
    -m message, --message message 
        Specifies a text message to be sent to the clients that are being 
        disconnected.
`

var deleteHelpTextTemplate = `Usage: fmcsadmin DELETE [TYPE] [SCHEDULE_NUMBER]

Description:
    Delete a schedule.

    Valid TYPEs:
        SCHEDULE        Deletes a schedule with schedule ID number
                        SCHEDULE_NUMBER. Use the LIST SCHEDULES
                        command to obtain the ID number of each
                        schedule.

Options:
    No command specific options.
`

var disableHelpTextTemplate = `Usage: fmcsadmin DISABLE [TYPE] [SCHEDULE_NUMBER]

Description:
    Disables a schedule.

    Valid TYPEs:
        SCHEDULE        Disables a schedule with schedule ID number
                        SCHEDULE_NUMBER. Use the LIST SCHEDULES
                        command to obtain the ID number of each
                        schedule.

Options:
    No command specific options.
`

var disconnectHelpTextTemplate = `Usage: fmcsadmin DISCONNECT CLIENT [CLIENT_NUMBER] [options]

Description: 
    Disconnects the specified client. The CLIENT_NUMBER is the ID number of 
    the client. Use the LIST CLIENTS command to obtain a list of clients and 
    their ID numbers. If no CLIENT_NUMBER is specified, all clients are 
    disconnected.

Options:
    -m message, --message message   
        Specifies a text message to be sent to the client that is being 
        disconnected.
`

var enableHelpTextTemplate = `Usage: fmcsadmin ENABLE [TYPE] [SCHEDULE_NUMBER]

Description:
    Enables a schedule.

    Valid TYPEs:
        SCHEDULE        Enables a schedule with schedule ID number
                        SCHEDULE_NUMBER. Use the LIST SCHEDULES
                        command to obtain the ID number of each
                        schedule.
 
Options:
    No command specific options.
`

var listHelpTextTemplate = `Usage: fmcsadmin LIST [TYPE] [options]

Description: 
    Lists items of the specified TYPE. 

    Valid TYPEs:
        CLIENTS         Lists the connected clients.
        FILES           Lists the hosted databases.
        SCHEDULES       List schedules.

Options:
    -s, --stats
        Reports additional details for each item.
`

var openHelpTextTemplate = `Usage: fmcsadmin OPEN [options] [FILE...] [PATH...]

Description:
    Opens databases in the default and additional database folders. Each FILE 
    specified is opened, or all the databases in each folder (PATH) are 
    opened. If no FILE or PATH is specified, all databases in the hosting 
    area are opened.

    To specify a database by its ID rather than its filename, first use the 
    LIST FILES -s command to get a list of databases and their IDs.

Options:
    --key encryptpass
        Specifies the encryption password for database(s) being opened.
`

var pauseHelpTextTemplate = `Usage: fmcsadmin PAUSE [FILE...] [PATH...]

Description:
    Pauses the specified databases (FILE) or all the hosted databases in the 
    specified folders (PATH). If no FILE or PATH is specified, pauses all 
    hosted databases. 

    After a database is paused, it is safe to copy or back up the database 
    until a RESUME command is performed.

Options: 
    No command specific options.
`

var resumeHelpTextTemplate = `Usage: fmcsadmin RESUME [FILE...] [PATH...]

Description:
    Makes a database that has been paused available again. Resumes activity on 
    the specified databases (FILE) or all the paused databases in the 
    specified folders (PATH). If no FILE or PATH is specified, all paused 
    databases are resumed.

Options:
    No command specific options.
`

var runHelpTextTemplate = `Usage: fmcsadmin RUN SCHEDULE [SCHEDULE_NUMBER]

Description:
    Manually runs a schedule specified by its SCHEDULE_NUMBER. To obtain a 
    list of schedules and their ID numbers, use the LIST SCHEDULES command.

Options:
    No command specific options.
`

var sendHelpTextTemplate = `Usage: fmcsadmin SEND [options] [CLIENT_NUMBER] [FILE...] [PATH...]

Description:
    Sends a text message to a client specified by CLIENT_NUMBER, to the 
    clients connected to the specified databases (FILE), or to all clients 
    connected to any database in the specified folders (PATH). 

    If no CLIENT_NUMBER, FILE, or PATH is specified, the message is sent to 
    all connected clients. By default, parameters are expected to be FILEs or 
    PATHs. To specify a CLIENT_NUMBER, you must use the -c option. 
    For example: 
        fmsadmin SEND -c 2 -m "This is a test message"

Options:
    -m message, --message message
        Specifies the text message to send.
        
    -c, --client
        Specifies a CLIENT_NUMBER.
`

var statusHelpTextTemplate = `Usage: fmcsadmin STATUS [TYPE] [CLIENT_NUMBER] [FILE...]

Description: 
    Retrieves the status of the specified TYPE.

    Valid TYPEs:
        CLIENT          Retrieves the status of a client specified by 
                        CLIENT_NUMBER.
        FILE            Retrieves the status of database(s) specified by FILE.

Options:
    No command specific options.
`
