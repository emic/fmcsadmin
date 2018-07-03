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
	"reflect"
	"regexp"
	"runtime"
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
	Token   string `json:"token"`
	Result  int    `json:"result"`
	Running bool   `json:"running"`
}

type generalConfigInfo struct {
	CacheSize         int `json:"cacheSize"`
	MaxFiles          int `json:"maxFiles"`
	MaxProConnections int `json:"maxProConnections"`
	MaxPSOS           int `json:"maxPSOS"`
}

type securityConfigInfo struct {
	RequireSecureDB bool `json:"requireSecureDB"`
}

type phpConfigInfo struct {
	Enabled              bool   `json:"enabled"`
	CharacterEncoding    string `json:"characterEncoding"`
	ErrorMessageLanguage string `json:"errorMessageLanguage"`
	DataPreValidation    bool   `json:"dataPreValidation"`
	UseFileMakerPhp      bool   `json:"useFileMakerPhp"`
}

type xmlConfigInfo struct {
	Enabled bool `json:"enabled"`
}

type disconnectMessageInfo struct {
	Message   string `json:"message"`
	Gracetime int    `json:"gracetime"`
}

type messageInfo struct {
	Message string `json:"message"`
}

type runningInfo struct {
	Running bool `json:"running"`
}

type params struct {
	key                  string
	message              string
	gracetime            int
	retry                int
	running              string
	enabled              string
	cachesize            int
	maxfiles             int
	maxproconnections    int
	maxpsos              int
	requiresecuredb      string
	characterencoding    string
	errormessagelanguage string
	dataprevalidation    bool
	usefilemakerphp      bool
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

	// detect an invalid command
	cmdArgs, cFlags, err := getFlags(args, commandOptions)
	if err != nil {
		fmt.Fprintln(c.outStream, flag.ErrHelp)
		exitStatus = outputInvalidCommandErrorMessage(c)
		return exitStatus
	}

	// detect an invalid option
	for i := 0; i < len(args); i++ {
		var invalidOption bool
		allowedOptions := []string{"-h", "-v", "-y", "-s", "-u", "-p", "-m", "-c", "-t", "--help", "--version", "--yes", "--stats", "--fqdn", "--username", "--password", "--key", "--message", "--client", "--gracetime"}
		for j := 0; j < len(allowedOptions); j++ {
			if string([]rune(args[i])[:1]) == "-" {
				invalidOption = true
				for _, v := range allowedOptions {
					if strings.ToLower(args[i]) == v {
						invalidOption = false
					}
				}
				if invalidOption == true {
					exitStatus = outputInvalidOptionErrorMessage(c, args[i])
					return exitStatus
				}
			}
		}
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
	baseURI := getHostName(fqdn)
	u, _ := url.Parse(baseURI)

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
				token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
				if token != "" && err == nil {
					u.Path = path.Join(getAPIBasePath(baseURI), "databases")
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
							u.Path = path.Join(getAPIBasePath(baseURI), "databases", strconv.Itoa(idList[i]), "close")
							exitStatus, _, err = sendRequest("PUT", u.String(), token, params{message: message})
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
					logout(baseURI, token)
				} else if exitStatus != 9 {
					exitStatus = 10502
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
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
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
								u.Path = path.Join(getAPIBasePath(baseURI), "schedules", strconv.Itoa(id))
								scheduleName := getScheduleName(u.String(), token, id)
								exitStatus, _, err = sendRequest("DELETE", u.String(), token, params{})
								if exitStatus == 0 && err == nil {
									if scheduleName != "" {
										fmt.Fprintln(c.outStream, "Schedule Deleted: "+scheduleName)
									} else {
										exitStatus = 10600
									}
								}
							} else {
								exitStatus = 10600
							}
						}
						logout(baseURI, token)
					} else if exitStatus != 9 {
						exitStatus = 10502
					}
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "disable":
			if len(cmdArgs[1:]) > 0 {
				token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
				if token != "" && err == nil {
					switch strings.ToLower(cmdArgs[1]) {
					case "schedule":
						res := ""
						if yesFlag == true {
							res = "y"
						} else {
							r := bufio.NewReader(os.Stdin)
							fmt.Fprint(c.outStream, "fmcsadmin: really disable a schedule(s)? (y, n) ")
							input, _ := r.ReadString('\n')
							res = strings.ToLower(strings.TrimSpace(input))
						}
						if res == "y" {
							id := 0
							if len(cmdArgs) >= 3 {
								sid, err := strconv.Atoi(cmdArgs[2])
								if err == nil {
									id = sid
								}
							}
							if id > 0 {
								u.Path = path.Join(getAPIBasePath(baseURI), "schedules", strconv.Itoa(id), "disable")
								exitStatus, _, err = sendRequest("PUT", u.String(), token, params{})
								if exitStatus == 0 && err == nil {
									u.Path = path.Join(getAPIBasePath(baseURI), "schedules")
									exitStatus = listSchedules(u.String(), token, id)
								}
							} else {
								exitStatus = 10600
							}
						}
					default:
						exitStatus = -1
					}
					logout(baseURI, token)
				} else if exitStatus != 9 {
					exitStatus = 10502
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "disconnect":
			if len(cmdArgs[1:]) > 0 {
				switch strings.ToLower(cmdArgs[1]) {
				case "client":
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
						token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
						if token != "" && err == nil {
							id := 0
							if len(cmdArgs) >= 3 {
								cid, err := strconv.Atoi(cmdArgs[2])
								if err == nil {
									id = cid
								}
							}
							if id > -1 {
								// check the client connection
								u.Path = path.Join(getAPIBasePath(baseURI), "databases")
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
									u.Path = path.Join(getAPIBasePath(baseURI), "clients", strconv.Itoa(id), "disconnect")
									exitStatus, _, err = sendRequest("PUT", u.String(), token, params{message: message, gracetime: graceTime})
									if exitStatus == 0 {
										fmt.Fprintln(c.outStream, "Client(s) being disconnected.")
									}
								} else {
									if id == 0 {
										fmt.Fprintln(c.outStream, "No client is connected.")
									} else {
										exitStatus = 11005
									}
								}
							}
							logout(baseURI, token)
						} else if exitStatus != 9 {
							exitStatus = 10502
						}
					}
				default:
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "enable":
			if len(cmdArgs[1:]) > 0 {
				token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
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
							u.Path = path.Join(getAPIBasePath(baseURI), "schedules", strconv.Itoa(id), "enable")
							exitStatus, _, err = sendRequest("PUT", u.String(), token, params{})
							if exitStatus == 0 && err == nil {
								u.Path = path.Join(getAPIBasePath(baseURI), "schedules")
								exitStatus = listSchedules(u.String(), token, id)
							}
						} else {
							exitStatus = 10600
						}
					default:
						exitStatus = 11002
					}
					logout(baseURI, token)
				} else if exitStatus != 9 {
					exitStatus = 10502
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "get":
			if len(cmdArgs[1:]) > 0 {
				switch strings.ToLower(cmdArgs[1]) {
				case "cwpconfig":
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
					if token != "" && err == nil {
						printFlags := []bool{true, true, true, true, true, true}
						if len(cmdArgs[2:]) > 0 {
							printFlags = []bool{false, false, false, false, false, false}
							for i := 0; i < len(cmdArgs[2:]); i++ {
								switch strings.ToLower(cmdArgs[2:][i]) {
								case "enablephp":
									printFlags[0] = true
								case "enablexml":
									printFlags[1] = true
								case "encoding":
									printFlags[2] = true
								case "locale":
									printFlags[3] = true
								case "prevalidation":
									printFlags[4] = true
								case "usefmphp":
									printFlags[5] = true
								default:
									fmt.Println("Invalid configuration name: " + cmdArgs[2:][i])
									exitStatus = 10001
								}
								if exitStatus != 0 {
									break
								}
							}
						}
						if exitStatus == 0 {
							_, exitStatus, err = getWebTechnologyConfigurations(baseURI, getAPIBasePath(baseURI), token, printFlags)
						}
						logout(baseURI, token)
					} else if exitStatus != 9 {
						exitStatus = 10502
					}
				case "serverconfig":
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
					if token != "" && err == nil {
						printFlags := []bool{true, true, true, true}
						printFlag := true
						if len(cmdArgs[2:]) > 0 {
							printFlags = []bool{false, false, false, false}
							printFlag = false
							for i := 0; i < len(cmdArgs[2:]); i++ {
								switch strings.ToLower(cmdArgs[2:][i]) {
								case "cachesize":
									printFlags[0] = true
								case "hostedfiles":
									printFlags[1] = true
								case "proconnections":
									printFlags[2] = true
								case "scriptsessions":
									printFlags[3] = true
								case "securefilesonly":
									printFlag = true
								default:
									exitStatus = 10001
								}
								if exitStatus != 0 {
									break
								}
							}
						}
						if exitStatus == 0 {
							u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "general")
							_, exitStatus = getServerGeneralConfigurations(u.String(), token, printFlags)
							if exitStatus == 0 {
								u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "security")
								exitStatus = getServerSecurityConfigurations(u.String(), token, printFlag)
							}
						}
						logout(baseURI, token)
					} else if exitStatus != 9 {
						exitStatus = 10502
					}
				default:
					exitStatus = outputInvalidCommandErrorMessage(c)
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
				case "get":
					fmt.Fprint(c.outStream, getHelpTextTemplate)
				case "help":
					fmt.Fprint(c.outStream, helpTextTemplate)
				case "list":
					fmt.Fprint(c.outStream, listHelpTextTemplate)
				case "open":
					fmt.Fprint(c.outStream, openHelpTextTemplate)
				case "pause":
					fmt.Fprint(c.outStream, pauseHelpTextTemplate)
				case "restart":
					fmt.Fprint(c.outStream, restartHelpTextTemplate)
				case "resume":
					fmt.Fprint(c.outStream, resumeHelpTextTemplate)
				case "run":
					fmt.Fprint(c.outStream, runHelpTextTemplate)
				case "send":
					fmt.Fprint(c.outStream, sendHelpTextTemplate)
				case "set":
					fmt.Fprint(c.outStream, setHelpTextTemplate)
				case "start":
					fmt.Fprint(c.outStream, startHelpTextTemplate)
				case "status":
					fmt.Fprint(c.outStream, statusHelpTextTemplate)
				case "stop":
					fmt.Fprint(c.outStream, stopHelpTextTemplate)
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
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
					if token != "" && err == nil {
						id := -1
						if statsFlag == true {
							id = 0
						}
						u.Path = path.Join(getAPIBasePath(baseURI), "databases")
						exitStatus = listClients(u.String(), token, id)
						logout(baseURI, token)
					} else if exitStatus != 9 {
						exitStatus = 10502
					}
				case "files":
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
					if token != "" && err == nil {
						idList := []int{-1}
						if statsFlag == true {
							idList = []int{0}
						}
						u.Path = path.Join(getAPIBasePath(baseURI), "databases")
						exitStatus = listFiles(u.String(), token, idList)
						logout(baseURI, token)
					} else if exitStatus != 9 {
						exitStatus = 10502
					}
				case "schedules":
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
					if token != "" && err == nil {
						u.Path = path.Join(getAPIBasePath(baseURI), "schedules")
						exitStatus = listSchedules(u.String(), token, 0)
						logout(baseURI, token)
					} else if exitStatus != 9 {
						exitStatus = 10502
					}
				default:
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "open":
			token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
			if token != "" && err == nil {
				u.Path = path.Join(getAPIBasePath(baseURI), "databases")
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
						u.Path = path.Join(getAPIBasePath(baseURI), "databases", strconv.Itoa(idList[i]), "open")
						exitStatus, _, err = sendRequest("PUT", u.String(), token, params{key: key})
						if exitStatus == 0 && err == nil {
							// Note: FileMaker Admin API (Trial) does not validate the encryption key.
							//       You receive a result code of 0 even if you enter an invalid key.
							var openedID []int
							for value := 0; ; {
								value++
								u.Path = path.Join(getAPIBasePath(baseURI), "databases")
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
				logout(baseURI, token)
			} else if exitStatus != 9 {
				exitStatus = 10502
			}
		case "pause":
			token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
			if token != "" && err == nil {
				u.Path = path.Join(getAPIBasePath(baseURI), "databases")
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
						u.Path = path.Join(getAPIBasePath(baseURI), "databases", strconv.Itoa(idList[i]), "pause")
						exitStatus, _, err = sendRequest("PUT", u.String(), token, params{})
						if exitStatus == 0 && err == nil {
							fmt.Fprintln(c.outStream, "File Paused: "+nameList[i])
						}
					}
				} else {
					exitStatus = 10904
				}
				logout(baseURI, token)
			} else if exitStatus != 9 {
				exitStatus = 10502
			}
		case "restart":
			if fqdn == "" && (runtime.GOOS == "darwin" || runtime.GOOS == "windows") {
				if len(cmdArgs[1:]) > 0 {
					res := ""
					if yesFlag == true {
						res = "y"
					} else {
						r := bufio.NewReader(os.Stdin)
						fmt.Fprint(c.outStream, "fmcsadmin: really restart server? (y, n) ")
						input, _ := r.ReadString('\n')
						res = strings.ToLower(strings.TrimSpace(input))
					}
					if res == "y" {
						switch strings.ToLower(cmdArgs[1]) {
						case "server":
							token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
							if token != "" && err == nil {
								// stop database server
								exitStatus, err = stopDatabaseServer(u, baseURI, token, message, graceTime)
								if exitStatus == 0 {
									var running bool
									for value := 0; ; {
										time.Sleep(1 * time.Second)
										value++
										u.Path = path.Join(getAPIBasePath(baseURI), "server", "status")
										exitStatus, running, err = sendRequest("GET", u.String(), token, params{})
										if running == false || value > 120 {
											break
										}
									}
									// start database server
									exitStatus, _, err = sendRequest("PUT", u.String(), token, params{running: "true"})
								}
								logout(baseURI, token)
							} else if exitStatus != 9 {
								exitStatus = 10502
							}
						default:
							exitStatus = outputInvalidCommandParameterErrorMessage(c)
						}
					}
				} else {
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			} else {
				// for FileMaker Cloud
				exitStatus = 3
			}
		case "resume":
			token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
			if token != "" && err == nil {
				u.Path = path.Join(getAPIBasePath(baseURI), "databases")
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
						u.Path = path.Join(getAPIBasePath(baseURI), "databases", strconv.Itoa(idList[i]), "resume")
						exitStatus, _, err = sendRequest("PUT", u.String(), token, params{})
						if exitStatus == 0 && err == nil {
							fmt.Fprintln(c.outStream, "File Resumed: "+nameList[i])
						}
					}
				} else {
					exitStatus = 10904
				}
				logout(baseURI, token)
			} else if exitStatus != 9 {
				exitStatus = 10502
			}
		case "run":
			if len(cmdArgs[1:]) > 0 {
				switch strings.ToLower(cmdArgs[1]) {
				case "schedule":
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
					if token != "" && err == nil {
						id := 0
						if len(cmdArgs) >= 3 {
							sid, err := strconv.Atoi(cmdArgs[2])
							if err == nil {
								id = sid
							}
						}
						if id > 0 {
							u.Path = path.Join(getAPIBasePath(baseURI), "schedules", strconv.Itoa(id), "run")
							exitStatus, _, err = sendRequest("PUT", u.String(), token, params{})
							if exitStatus == 0 && err == nil {
								u.Path = path.Join(getAPIBasePath(baseURI), "schedules", strconv.Itoa(id))
								scheduleName := getScheduleName(u.String(), token, id)
								if scheduleName != "" {
									fmt.Fprintln(c.outStream, "Schedule '"+scheduleName+"' will run now.")
								} else {
									exitStatus = 10600
								}
							}
						} else {
							exitStatus = 10600
						}
						logout(baseURI, token)
					} else if exitStatus != 9 {
						exitStatus = 10502
					}
				default:
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "send":
			token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
			if token != "" && err == nil {
				exitStatus = sendMessages(u, baseURI, token, message, cmdArgs, clientID)
				logout(baseURI, token)
			} else if exitStatus != 9 {
				exitStatus = 10502
			}
		case "set":
			if len(cmdArgs[1:]) > 0 {
				switch strings.ToLower(cmdArgs[1]) {
				case "cwpconfig":
					if len(cmdArgs[2:]) > 0 {
						token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
						if token != "" && err == nil {
							var settings []string
							printFlags := []bool{false, false, false, false, false, false}
							settings, exitStatus, err = getWebTechnologyConfigurations(baseURI, getAPIBasePath(baseURI), token, printFlags)
							if err == nil {
								var results []string
								results, exitStatus = parseWebConfigurationSettings(c, cmdArgs[2:])

								phpFlag := results[0]
								xmlFlag := results[1]
								encoding := results[2]
								locale := results[3]
								preValidationFlag := results[4]
								useFMPHPFlag := results[5]

								var preValidation bool
								var useFMPHP bool

								if len(phpFlag) > 0 || len(encoding) > 0 || len(locale) > 0 || len(preValidationFlag) > 0 || len(useFMPHPFlag) > 0 {
									if phpFlag == "" {
										phpFlag = settings[0]
									} else if strings.ToLower(phpFlag) == "true" {
										fmt.Println("EnablePHP = true")
										if settings[0] == "false" {
											fmt.Println("Restart the FileMaker Server background processes to apply the change.")
										}
									} else if strings.ToLower(phpFlag) == "false" {
										fmt.Println("EnablePHP = false")
										if settings[0] == "true" {
											fmt.Println("Restart the FileMaker Server background processes to apply the change.")
										}
									}

									if encoding == "" {
										encoding = settings[1]
									} else if strings.ToLower(encoding) == "utf-8" {
										fmt.Println("Encoding = UTF-8 [ UTF-8 ISO-8859-1 ]")
									} else if strings.ToLower(encoding) == "iso-8859-1" {
										fmt.Println("Encoding = ISO-8859-1 [ UTF-8 ISO-8859-1 ]")
									}

									if locale == "" {
										locale = settings[2]
									} else if strings.ToLower(locale) == "en" || strings.ToLower(locale) == "de" || strings.ToLower(locale) == "fr" || strings.ToLower(locale) == "it" || strings.ToLower(locale) == "ja" || strings.ToLower(locale) == "sv" {
										fmt.Println("Locale = " + strings.ToLower(locale) + " [ en de fr it ja sv ]")
									}

									if preValidationFlag == "" {
										preValidationFlag = settings[3]
									} else if strings.ToLower(preValidationFlag) == "true" {
										fmt.Println("PreValidation = true")
									} else if strings.ToLower(preValidationFlag) == "false" {
										fmt.Println("PreValidation = false")
									}
									if preValidationFlag == "true" {
										preValidation = true
									} else if preValidationFlag == "false" {
										preValidation = false
									}

									if useFMPHPFlag == "" {
										useFMPHPFlag = settings[4]
									} else if strings.ToLower(useFMPHPFlag) == "false" && phpFlag == "true" {
										fmt.Println("UseFMPHP = false")
										if settings[4] == "true" {
											fmt.Println("Restart the FileMaker Server background processes to apply the change.")
										}
									} else {
										// UseFMPHP is always true when enablePHP is false
										fmt.Println("UseFMPHP = true")
										if settings[4] == "false" {
											fmt.Println("Restart the FileMaker Server background processes to apply the change.")
										}
									}
									if useFMPHPFlag == "true" {
										useFMPHP = true
									} else if useFMPHPFlag == "false" {
										useFMPHP = false
									}

									u.Path = path.Join(getAPIBasePath(baseURI), "php", "config")
									exitStatus, _, _ = sendRequest("PATCH", u.String(), token, params{
										enabled:              phpFlag,
										characterencoding:    encoding,
										errormessagelanguage: locale,
										dataprevalidation:    preValidation,
										usefilemakerphp:      useFMPHP,
									})
								}

								if xmlFlag == "true" || xmlFlag == "false" {
									if xmlFlag == "true" {
										fmt.Println("EnableXML = true")
									} else if xmlFlag == "false" {
										fmt.Println("EnableXML = false")
									}
									u.Path = path.Join(getAPIBasePath(baseURI), "xml", "config")
									exitStatus, _, _ = sendRequest("PATCH", u.String(), token, params{enabled: xmlFlag})
								}
							}
							logout(baseURI, token)
						} else if exitStatus != 9 {
							exitStatus = 10502
						}
					} else {
						exitStatus = 10001
					}
				case "serverconfig":
					if len(cmdArgs[2:]) > 0 {
						token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
						if token != "" && err == nil {
							var settings []int
							printFlags := []bool{false, false, false, false}
							u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "general")
							settings, exitStatus = getServerGeneralConfigurations(u.String(), token, printFlags)
							if exitStatus == 0 {
								var results []string
								results, exitStatus = parseServerConfigurationSettings(c, cmdArgs[2:])

								cacheSize, _ := strconv.Atoi(results[0])
								maxFiles, _ := strconv.Atoi(results[1])
								maxProConnections, _ := strconv.Atoi(results[2])
								maxPSOS, _ := strconv.Atoi(results[3])
								secureFilesOnlyFlag := results[4]

								if results[0] != "" || results[1] != "" || results[2] != "" || results[3] != "" || secureFilesOnlyFlag != "" {
									if results[0] == "" {
										cacheSize = settings[0]
									} else {
										if cacheSize < 64 || cacheSize > 1048576 {
											exitStatus = 10001
										}
									}

									if results[1] == "" {
										maxFiles = settings[1]
									} else {
										if maxFiles < 1 || maxFiles > 125 {
											exitStatus = 10001
										}
									}

									if results[2] == "" {
										maxProConnections = settings[2]
									} else {
										if maxProConnections < 0 || maxProConnections > 2000 {
											exitStatus = 10001
										}
									}

									if results[3] == "" {
										maxPSOS = settings[3]
									} else {
										if maxPSOS < 0 || maxPSOS > 500 {
											exitStatus = 10001
										}
									}

									if exitStatus == 0 && (results[0] != "" || results[1] != "" || results[2] != "" || results[3] != "") {
										if results[0] != "" {
											fmt.Println("CacheSize = " + strconv.Itoa(cacheSize) + " [default: 512, range: 64-1048576]")
										}
										if results[1] != "" {
											fmt.Println("HostedFiles = " + strconv.Itoa(maxFiles) + " [default: 125, range: 1-125]")
										}
										if results[2] != "" {
											fmt.Println("ProConnections = " + strconv.Itoa(maxProConnections) + " [default: 250, range: 0-2000]")
										}
										if results[3] != "" {
											fmt.Println("ScriptSessions = " + strconv.Itoa(maxPSOS) + " [default: 25, range: 0-500]")
										}
										u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "general")
										exitStatus, _, _ = sendRequest("PATCH", u.String(), token, params{
											cachesize:         cacheSize,
											maxfiles:          maxFiles,
											maxproconnections: maxProConnections,
											maxpsos:           maxPSOS,
										})
									}

									if exitStatus == 0 && (secureFilesOnlyFlag == "true" || secureFilesOnlyFlag == "false") {
										if secureFilesOnlyFlag == "true" {
											fmt.Println("SecureFilesOnly = true [default: true]")
										} else if secureFilesOnlyFlag == "false" {
											fmt.Println("SecureFilesOnly = false [default: true]")
										}
										u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "security")
										exitStatus, _, _ = sendRequest("PATCH", u.String(), token, params{requiresecuredb: secureFilesOnlyFlag})
									}
								}
							}
							logout(baseURI, token)
						} else if exitStatus != 9 {
							exitStatus = 10502
						}
					} else {
						exitStatus = 10001
					}
				default:
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "start":
			if fqdn == "" && (runtime.GOOS == "darwin" || runtime.GOOS == "windows") {
				if len(cmdArgs[1:]) > 0 {
					switch strings.ToLower(cmdArgs[1]) {
					case "server":
						token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
						if token != "" && err == nil {
							var running bool
							u.Path = path.Join(getAPIBasePath(baseURI), "server", "status")
							exitStatus, running, err = sendRequest("GET", u.String(), token, params{})
							if running == true {
								// Service already running
								exitStatus = 10006
							} else {
								exitStatus, _, err = sendRequest("PUT", u.String(), token, params{running: "true"})
							}
							logout(baseURI, token)
						} else if exitStatus != 9 {
							exitStatus = 10502
						}
					default:
						exitStatus = outputInvalidCommandParameterErrorMessage(c)
					}
				} else {
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			} else {
				// for FileMaker Cloud
				exitStatus = 3
			}
		case "status":
			if len(cmdArgs[1:]) > 0 {
				switch strings.ToLower(cmdArgs[1]) {
				case "client":
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
					if token != "" && err == nil {
						id := 0
						if len(cmdArgs) >= 3 {
							sid, err := strconv.Atoi(cmdArgs[2])
							if err == nil {
								id = sid
							}
						}
						if id > 0 {
							u.Path = path.Join(getAPIBasePath(baseURI), "databases")
							exitStatus = listClients(u.String(), token, id)
						}
						logout(baseURI, token)
					} else if exitStatus != 9 {
						exitStatus = 10502
					}
				case "file":
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
					if token != "" && err == nil {
						if len(cmdArgs[2:]) > 0 {
							u.Path = path.Join(getAPIBasePath(baseURI), "databases")
							idList, _, _ := getDatabases(u.String(), token, cmdArgs[2:], "")
							if len(idList) > 0 {
								exitStatus = listFiles(u.String(), token, idList)
							}
						} else {
							exitStatus = 10001
						}
						logout(baseURI, token)
					} else if exitStatus != 9 {
						exitStatus = 10502
					}
				default:
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "stop":
			if fqdn == "" && (runtime.GOOS == "darwin" || runtime.GOOS == "windows") {
				if len(cmdArgs[1:]) > 0 {
					res := ""
					if yesFlag == true {
						res = "y"
					} else {
						r := bufio.NewReader(os.Stdin)
						fmt.Fprint(c.outStream, "fmcsadmin: really stop server? (y, n) ")
						input, _ := r.ReadString('\n')
						res = strings.ToLower(strings.TrimSpace(input))
					}
					if res == "y" {
						switch strings.ToLower(cmdArgs[1]) {
						case "server":
							token, exitStatus, err = login(baseURI, username, password, params{retry: retry})
							if token != "" && err == nil {
								exitStatus, err = stopDatabaseServer(u, baseURI, token, message, graceTime)
								logout(baseURI, token)
							} else if exitStatus != 9 {
								exitStatus = 10502
							}
						default:
							exitStatus = outputInvalidCommandParameterErrorMessage(c)
						}
					}
				} else {
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			} else {
				// for FileMaker Cloud
				exitStatus = 3
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

	if exitStatus != 0 && exitStatus != 23 && exitStatus != 248 {
		outputErrorMessage(exitStatus, c)
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

func parseServerConfigurationSettings(c *cli, str []string) ([]string, int) {
	exitStatus := 0
	var results []string
	cacheSize := ""
	maxFiles := ""
	maxProConnections := ""
	maxPSOS := ""
	secureFilesOnlyFlag := ""

	for i := 0; i < len(str); i++ {
		val := strings.ToLower(str[i])
		if regexp.MustCompile(`cachesize=(\d+)`).Match([]byte(val)) == true {
			rep := regexp.MustCompile(`cachesize=(\d+)`)
			cacheSize = rep.ReplaceAllString(val, "$1")
		} else if regexp.MustCompile(`hostedfiles=(\d+)`).Match([]byte(val)) == true {
			rep := regexp.MustCompile(`hostedfiles=(\d+)`)
			maxFiles = rep.ReplaceAllString(val, "$1")
		} else if regexp.MustCompile(`proconnections=(\d+)`).Match([]byte(val)) == true {
			rep := regexp.MustCompile(`proconnections=(\d+)`)
			maxProConnections = rep.ReplaceAllString(val, "$1")
		} else if regexp.MustCompile(`scriptsessions=(\d+)`).Match([]byte(val)) == true {
			rep := regexp.MustCompile(`scriptsessions=(\d+)`)
			maxPSOS = rep.ReplaceAllString(val, "$1")
		} else if regexp.MustCompile(`securefilesonly=(.*)`).Match([]byte(str[i])) == true {
			if strings.ToLower(str[i]) == "securefilesonly=" {
				exitStatus = 10001
			} else if strings.ToLower(str[i]) == "securefilesonly=true" || (regexp.MustCompile(`securefilesonly=([+|-])?(\d)+`).Match([]byte(str[i])) == true && str[i] != "securefilesonly=0" && str[i] != "securefilesonly=+0" && str[i] != "securefilesonly=-0") {
				secureFilesOnlyFlag = "true"
			} else {
				secureFilesOnlyFlag = "false"
			}
		} else {
			exitStatus = 10001
		}
	}

	results = append(results, cacheSize)
	results = append(results, maxFiles)
	results = append(results, maxProConnections)
	results = append(results, maxPSOS)
	results = append(results, secureFilesOnlyFlag)

	return results, exitStatus
}

func parseWebConfigurationSettings(c *cli, str []string) ([]string, int) {
	exitStatus := 0
	var results []string
	phpFlag := ""
	xmlFlag := ""
	encoding := ""
	locale := ""
	preValidationFlag := ""
	useFMPHPFlag := ""

	for i := 0; i < len(str); i++ {
		if regexp.MustCompile(`enablephp=(.*)`).Match([]byte(str[i])) == true {
			if strings.ToLower(str[i]) == "enablephp=" {
				exitStatus = 10001
			} else if strings.ToLower(str[i]) == "enablephp=true" || (regexp.MustCompile(`enablephp=([+|-])?(\d)+`).Match([]byte(str[i])) == true && str[i] != "enablephp=0" && str[i] != "enablephp=+0" && str[i] != "enablephp=-0") {
				phpFlag = "true"
			} else {
				phpFlag = "false"
			}
		} else if regexp.MustCompile(`enablexml=(.*)`).Match([]byte(str[i])) == true {
			if strings.ToLower(str[i]) == "enablexml=" {
				exitStatus = 10001
			} else if strings.ToLower(str[i]) == "enablexml=true" || (regexp.MustCompile(`enablexml=([+|-])?(\d)+`).Match([]byte(str[i])) == true && str[i] != "enablexml=0" && str[i] != "enablexml=+0" && str[i] != "enablexml=-0") {
				xmlFlag = "true"
			} else {
				xmlFlag = "false"
			}
		} else if strings.ToLower(str[i]) == "encoding=utf-8" {
			encoding = "UTF-8"
		} else if strings.ToLower(str[i]) == "encoding=iso-8859-1" {
			encoding = "ISO-8859-1"
		} else if strings.ToLower(str[i]) == "locale=en" {
			locale = "en"
		} else if strings.ToLower(str[i]) == "locale=de" {
			locale = "de"
		} else if strings.ToLower(str[i]) == "locale=fr" {
			locale = "fr"
		} else if strings.ToLower(str[i]) == "locale=it" {
			locale = "it"
		} else if strings.ToLower(str[i]) == "locale=ja" {
			locale = "ja"
		} else if strings.ToLower(str[i]) == "locale=sv" {
			locale = "sv"
		} else if regexp.MustCompile(`prevalidation=(.*)`).Match([]byte(str[i])) == true {
			if strings.ToLower(str[i]) == "prevalidation=" {
				exitStatus = 10001
			} else if strings.ToLower(str[i]) == "prevalidation=true" || (regexp.MustCompile(`prevalidation=([+|-])?(\d)+`).Match([]byte(str[i])) == true && str[i] != "prevalidation=0" && str[i] != "prevalidation=+0" && str[i] != "prevalidation=-0") {
				preValidationFlag = "true"
			} else {
				preValidationFlag = "false"
			}
		} else if regexp.MustCompile(`usefmphp=(.*)`).Match([]byte(str[i])) == true {
			if strings.ToLower(str[i]) == "usefmphp=" {
				exitStatus = 10001
			} else if strings.ToLower(str[i]) == "usefmphp=true" || (regexp.MustCompile(`usefmphp=([+|-])?(\d)+`).Match([]byte(str[i])) == true && str[i] != "usefmphp=0" && str[i] != "usefmphp=+0" && str[i] != "usefmphp=-0") {
				useFMPHPFlag = "true"
			} else {
				useFMPHPFlag = "false"
			}
		} else {
			exitStatus = 10001
		}
	}

	results = append(results, phpFlag)
	results = append(results, xmlFlag)
	results = append(results, encoding)
	results = append(results, locale)
	results = append(results, preValidationFlag)
	results = append(results, useFMPHPFlag)

	return results, exitStatus
}

func outputInvalidCommandParameterErrorMessage(c *cli) int {
	exitStatus := 23
	fmt.Fprintln(c.outStream, "Error: 10007 (Requested object does not exist)")

	return exitStatus
}

func outputInvalidCommandErrorMessage(c *cli) int {
	exitStatus := 248
	fmt.Fprintln(c.outStream, "Error: 11000 (Invalid command)")
	fmt.Fprint(c.outStream, helpTextTemplate)

	return exitStatus
}

func outputInvalidOptionErrorMessage(c *cli, option string) int {
	exitStatus := 249
	fmt.Fprintln(c.outStream, "Invalid option: "+option)
	fmt.Fprintln(c.outStream, "Error: 11001 (Invalid option)")
	fmt.Fprint(c.outStream, helpTextTemplate)

	return exitStatus
}

func getHostName(fqdn string) string {
	baseURI := "http://127.0.0.1:8080"
	if len(fqdn) > 0 {
		baseURI = "https://" + strings.TrimSpace(fqdn)
	} else if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		baseURI = "http://127.0.0.1:16001"
	}

	return baseURI
}

func getAPIBasePath(baseURI string) string {
	path := "/admin/api/v1"
	if baseURI == "http://127.0.0.1:16001" {
		path = "/fmi/admin/api/v1"
	}

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

func login(baseURI string, user string, pass string, p params) (string, int, error) {
	token := ""
	exitStatus := 0

	username, password := getUsernameAndPassword(user, pass)
	u, _ := url.Parse(baseURI)
	u.Path = path.Join(getAPIBasePath(baseURI), "user", "login")

	d := accountInfo{
		username,
		password,
	}
	output := output{}

	jsonStr, _ := json.Marshal(d)
	body, err := callURL("POST", u.String(), "", bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
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
			token, exitStatus, err = login(baseURI, user, pass, params{retry: p.retry - 1})
		} else {
			fmt.Println("fmcsadmin: Permission denied.")
			exitStatus = 9
		}
	}

	return token, exitStatus, err
}

func logout(baseURI string, token string) {
	u, _ := url.Parse(baseURI)
	u.Path = path.Join(getAPIBasePath(baseURI), "user", "logout")
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
	var sID int

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

					connectTime = getDateTimeStringOfCurrentTimeZone(connectTime)
					if regexp.MustCompile(`(.*)\.fmp12`).Match([]byte(fileName)) == true {
						rep := regexp.MustCompile(`(.*)\.fmp12`)
						fileName = rep.ReplaceAllString(fileName, "$1")
					}

					data = append(data, []string{s1, userName, computerName, extPriv, ipAddress, macAddress, connectTime, connectDuration, appVersion, appLanguage, teamLicensed, fileName, accountName, privsetName})
				}
			}

			sID, _ = strconv.Atoi(s1)
			if id == sID || id == 0 {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Client ID", "User Name", "Computer Name", "Ext Privilege", "IP Address", "MAC Address", "Connect Time", "Duration", "App Version", "App Language", "User Connections License", "File Name", "Account Name", "Privilege Set"})
				table.SetAutoWrapText(false)
				for _, v := range data {
					table.Append(v)
				}
				table.Render()
			}
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
	var statusStr string
	var data [][]string

	err = scan.ScanTree(v, "/schedules", &c)
	count = len(c)

	if count > 0 {
		for i := 0; i < count; i++ {
			statusStr = ""
			lastRun = ""
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
				lastRun = getDateTimeStringOfCurrentTimeZone(lastRun)
				nextRun = getDateTimeStringOfCurrentTimeZone(nextRun)
				statusStr = strconv.Itoa(status)
				if status == 1 || status == 2 {
					// 2 : running
					if lastRun == "" {
						statusStr = ""
					} else {
						statusStr = "OK"
					}
				}

				data = append(data, []string{s1, name, taskType, lastRun, nextRun, statusStr})
			}
		}

		if len(data) > 0 {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoWrapText(false)
			for _, v := range data {
				table.SetHeader([]string{"ID", "Name", "Type", "Last Completed", "Next Run", "Status"})
				table.Append(v)
			}
			table.Render()
		} else {
			// Schedule at specified index no longer exists
			return 10600
		}
	}

	return 0
}

func getScheduleName(url string, token string, id int) string {
	body, err := callURL("GET", url, token, nil)
	if err != nil {
		//fmt.Println(err.Error())
		return ""
	}

	var v interface{}
	body2 := []byte(body)
	err = json.Unmarshal(body2, &v)
	if err != nil {
		//fmt.Println(err.Error())
		return ""
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

func sendMessages(u *url.URL, baseURI string, token string, message string, cmdArgs []string, clientID int) int {
	var exitStatus int

	args := []string{""}
	if len(cmdArgs[1:]) > 0 {
		args = cmdArgs[1:]
	}
	u.Path = path.Join(getAPIBasePath(baseURI), "databases")
	idList := getClients(u.String(), token, args, "NORMAL")
	id := 0
	if clientID > -1 {
		idList = append(idList, id)
	}
	if len(idList) > 0 {
		for i := 0; i < len(idList); i++ {
			u.Path = path.Join(getAPIBasePath(baseURI), "clients", strconv.Itoa(idList[i]), "message")
			exitStatus = sendMessage(u.String(), token, message)
			if clientID > -1 {
				break
			}
		}
	} else {
		exitStatus = 10904
	}

	return exitStatus
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

func getServerGeneralConfigurations(url string, token string, printFlags []bool) ([]int, int) {
	var settings []int
	var result int
	var cacheSize int
	var maxFiles int
	var maxProConnections int
	var maxPSOS int

	body, err := callURL("GET", url, token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return settings, 10502
	}

	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = scan.ScanTree(v, "/result", &result)
	if err != nil {
		return settings, 3
	}
	err = scan.ScanTree(v, "/cacheSize", &cacheSize)
	err = scan.ScanTree(v, "/maxFiles", &maxFiles)
	err = scan.ScanTree(v, "/maxProConnections", &maxProConnections)
	err = scan.ScanTree(v, "/maxPSOS", &maxPSOS)

	settings = append(settings, cacheSize)
	settings = append(settings, maxFiles)
	settings = append(settings, maxProConnections)
	settings = append(settings, maxPSOS)

	// output
	if printFlags[0] == true {
		fmt.Println("CacheSize = " + strconv.Itoa(cacheSize) + " [default: 512, range: 64-1048576] ")
	}
	if printFlags[1] == true {
		fmt.Println("HostedFiles = " + strconv.Itoa(maxFiles) + " [default: 125, range: 1-125] ")
	}
	if printFlags[2] == true {
		fmt.Println("ProConnections = " + strconv.Itoa(maxProConnections) + " [default: 250, range: 0-2000] ")
	}
	if printFlags[3] == true {
		fmt.Println("ScriptSessions = " + strconv.Itoa(maxPSOS) + " [default: 25, range: 0-500] ")
	}

	return settings, result
}

func getServerSecurityConfigurations(url string, token string, print bool) int {
	var result int
	var requireSecureDB bool
	var requireSecureDBStr string

	body, err := callURL("GET", url, token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return 10502
	}

	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = scan.ScanTree(v, "/result", &result)
	if err != nil {
		return 3
	}
	err = scan.ScanTree(v, "/requireSecureDB", &requireSecureDB)

	requireSecureDBStr = "true"
	if requireSecureDB == false {
		requireSecureDBStr = "false"
	}

	// output
	if print == true {
		fmt.Println("SecureFilesOnly = " + requireSecureDBStr + " [default: true] ")
	}

	return result
}

func getWebTechnologyConfigurations(baseURI string, basePath string, token string, printFlags []bool) ([]string, int, error) {
	var settings []string
	var result int
	var enabledPhp bool
	var enabledPhpStr string
	var enabledXML bool
	var enabledXMLStr string
	var characterEncoding string
	var dataPreValidation bool
	var dataPreValidationStr string
	var errorMessageLanguage string
	var useFileMakerPhp bool
	var useFileMakerPhpStr string

	// get PHP Technology Configuration
	u, _ := url.Parse(baseURI)
	u.Path = path.Join(basePath, "php", "config")

	body, err := callURL("GET", u.String(), token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return settings, 10502, err
	}

	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = scan.ScanTree(v, "/result", &result)
	if err != nil {
		return settings, 3, err
	}
	err = scan.ScanTree(v, "/enabled", &enabledPhp)
	err = scan.ScanTree(v, "/characterEncoding", &characterEncoding)
	err = scan.ScanTree(v, "/errorMessageLanguage", &errorMessageLanguage)
	err = scan.ScanTree(v, "/dataPreValidation", &dataPreValidation)
	err = scan.ScanTree(v, "/useFileMakerPhp", &useFileMakerPhp)

	enabledPhpStr = "true"
	if enabledPhp == false {
		enabledPhpStr = "false"
	}
	settings = append(settings, enabledPhpStr)
	settings = append(settings, characterEncoding)
	settings = append(settings, errorMessageLanguage)

	dataPreValidationStr = "true"
	if dataPreValidation == false {
		dataPreValidationStr = "false"
	}
	settings = append(settings, dataPreValidationStr)

	useFileMakerPhpStr = "true"
	if useFileMakerPhp == false {
		useFileMakerPhpStr = "false"
	}
	settings = append(settings, useFileMakerPhpStr)

	// get XML Technology Configuration
	u.Path = path.Join(basePath, "xml", "config")

	body, err = callURL("GET", u.String(), token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return settings, -1, err
	}

	err = json.Unmarshal(body, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = scan.ScanTree(v, "/result", &result)
	err = scan.ScanTree(v, "/enabled", &enabledXML)

	enabledXMLStr = "true"
	if enabledXML == false {
		enabledXMLStr = "false"
	}

	// output
	if printFlags[0] == true {
		fmt.Println("EnablePHP = " + enabledPhpStr)
	}
	if printFlags[1] == true {
		fmt.Println("EnableXML = " + enabledXMLStr)
	}
	if printFlags[2] == true {
		fmt.Println("Encoding = " + characterEncoding + " [ UTF-8 ISO-8859-1 ]")
	}
	if printFlags[3] == true {
		fmt.Println("Locale = " + errorMessageLanguage + " [ en de fr it ja sv ]")
	}
	if printFlags[4] == true {
		fmt.Println("PreValidation = " + dataPreValidationStr)
	}
	if printFlags[5] == true {
		fmt.Println("UseFMPHP = " + useFileMakerPhpStr)
	}

	return settings, result, err
}

func disconnectAllClient(u *url.URL, baseURI string, token string, message string, graceTime int) (int, error) {
	u.Path = path.Join(getAPIBasePath(baseURI), "clients", "0", "disconnect")
	exitStatus, _, err := sendRequest("PUT", u.String(), token, params{message: message, gracetime: graceTime})

	return exitStatus, err
}

func stopDatabaseServer(u *url.URL, baseURI string, token string, message string, graceTime int) (int, error) {
	exitStatus := -1
	var err error

	// disconnect clients
	exitStatus, err = disconnectAllClient(u, baseURI, token, message, graceTime)

	// close databases
	u.Path = path.Join(getAPIBasePath(baseURI), "databases")
	idList, _, _ := getDatabases(u.String(), token, []string{""}, "NORMAL")
	if len(idList) > 0 {
		for i := 0; i < len(idList); i++ {
			u.Path = path.Join(getAPIBasePath(baseURI), "databases", strconv.Itoa(idList[i]), "close")
			exitStatus, _, err = sendRequest("PUT", u.String(), token, params{message: message})
		}
	}

	var openedID []int
	for value := 0; ; {
		time.Sleep(1 * time.Second)
		value++
		u.Path = path.Join(getAPIBasePath(baseURI), "databases")
		openedID, _, _ = getDatabases(u.String(), token, []string{""}, "CLOSING")
		if len(openedID) == 0 || value > 120 {
			break
		}
	}

	// stop database server
	u.Path = path.Join(getAPIBasePath(baseURI), "server", "status")
	exitStatus, _, err = sendRequest("PUT", u.String(), token, params{running: "false"})

	return exitStatus, err
}

func comparePath(name1 string, name2 string) bool {
	extName := ".fmp12"
	pathPrefix := []string{"filelinux:", "filemac:", "filewin:"}

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

func outputErrorMessage(code int, c *cli) {
	if code >= -1 {
		fmt.Fprintln(c.outStream, "Error: "+strconv.Itoa(code)+" ("+getErrorDescription(code)+")")
	}
}

func sendRequest(method string, url string, token string, p params) (int, bool, error) {
	var jsonStr []byte

	if len(p.key) > 0 {
		d := dbInfo{
			p.key,
		}
		jsonStr, _ = json.Marshal(d)
	} else if (params{}) != p && reflect.ValueOf(p.gracetime).IsValid() == true && p.gracetime >= 0 {
		d := disconnectMessageInfo{
			p.message,
			p.gracetime,
		}
		jsonStr, _ = json.Marshal(d)
	} else if (params{}) != p && reflect.ValueOf(p.message).IsValid() == true && len(p.message) > 0 {
		d := messageInfo{
			p.message,
		}
		jsonStr, _ = json.Marshal(d)
	} else if (params{}) != p && reflect.ValueOf(p.characterencoding).IsValid() == true && len(p.characterencoding) > 0 {
		enabled := true
		if p.enabled == "false" {
			enabled = false
		}
		d := phpConfigInfo{
			enabled,
			p.characterencoding,
			p.errormessagelanguage,
			p.dataprevalidation,
			p.usefilemakerphp,
		}
		jsonStr, _ = json.Marshal(d)
	} else if (params{}) != p && reflect.ValueOf(p.cachesize).IsValid() == true && p.cachesize > 0 {
		d := generalConfigInfo{
			p.cachesize,
			p.maxfiles,
			p.maxproconnections,
			p.maxpsos,
		}
		jsonStr, _ = json.Marshal(d)
	} else if (params{}) != p && reflect.ValueOf(p.requiresecuredb).IsValid() == true && len(p.requiresecuredb) > 0 {
		requiresecuredb := true
		if p.requiresecuredb == "false" {
			requiresecuredb = false
		}
		d := securityConfigInfo{
			requiresecuredb,
		}
		jsonStr, _ = json.Marshal(d)
	} else if (params{}) != p && reflect.ValueOf(p.enabled).IsValid() == true && len(p.enabled) > 0 {
		enabled := true
		if p.enabled == "false" {
			enabled = false
		}
		d := xmlConfigInfo{
			enabled,
		}
		jsonStr, _ = json.Marshal(d)
	} else if (params{}) != p && reflect.ValueOf(p.running).IsValid() == true && len(p.running) > 0 {
		running := true
		if p.running == "false" {
			running = false
		}
		d := runningInfo{
			running,
		}
		jsonStr, _ = json.Marshal(d)
	} else {
		jsonStr = []byte("")
	}

	// for debugging
	/*
		fmt.Println(method)
		fmt.Println(url)
		fmt.Println(string(jsonStr))
	*/

	body, err := callURL(method, url, token, bytes.NewBuffer([]byte(jsonStr)))
	if body != nil {
		output := output{}
		if json.Unmarshal(body, &output) == nil {
			return output.Result, output.Running, nil
		}
	}
	if err != nil {
		return -1, false, err
	}

	return 0, false, nil
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
	case -1:
		description = "Internal error"
	case 3:
		description = "Unavailable command"
	case 4:
		description = "Command is unknown"
	case 8:
		description = "Empty result"
	case 9:
		description = "Access denied"
	case 212:
		description = "Invalid user account and/or password; please try again"
	case 802:
		description = "Unable to open the file"
	case 958:
		description = "Parameter missing"
	case 960:
		description = "Parameter is invalid"
	case 10001:
		description = "Invalid parameter"
	case 10006:
		// When a script runs and a service is already executing (for example, during a long loop), the FileMaker error 10006, "kServiceAlreadyRunning," is returned.
		description = "Service already running"
	case 10007:
		description = "Requested object does not exist"
	case 10502:
		description = "Host unreachable"
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
	case 10908:
		// When a script schedule stops executing, the FileMaker error code 10908, "System script aborted," is returned.
		description = "System script aborted"
	case 11000:
		description = "Invalid command"
	case 11002:
		description = "Unable to create command"
	case 11005:
		description = "Disconnect Client invalid ID"
	case 25004:
		description = "Parameters are invalid"
	case 25006:
		description = "Invalid session error"
	default:
		description = ""
	}

	return description
}

func getDateTimeStringOfCurrentTimeZone(dateTime string) string {
	var t time.Time
	_, offset := time.Now().Zone()

	if len(dateTime) > 0 {
		reg := `(\d+[-/]\d+[-/]\d+)`
		if regexp.MustCompile(reg).Match([]byte(dateTime)) == true {
			t, _ = time.Parse("2006-01-02", dateTime[:10])
			if t.Format("2006-01-02") == "0001-01-01" {
				t, _ = time.Parse("01/02/2006", dateTime[:10])
				if t.Format("2006-01-02") == "0001-01-01" {
					dateTime = ""
				} else {
					t, _ = time.Parse("01/02/2006 03:04:05 PM", dateTime)
					if t.Format("2006-01-02") == "0001-01-01" {
						dateTime = ""
					} else {
						// for clients (FileMaker Cloud)
						dateTime = t.Add(time.Second * time.Duration(offset)).Format("2006/01/02 15:04:05")
					}
				}
			} else {
				t, _ = time.Parse("2006-01-02 15:04:05 MST", dateTime)
				if t.Format("2006-01-02") == "0001-01-01" {
					t, _ = time.Parse("2006-01-02 15:04:05", dateTime)
				}
				if t.Format("2006-01-02") == "0001-01-01" {
					t, _ = time.Parse("2006-01-02T15:04:05.000Z", dateTime)
					if t.Format("2006-01-02") == "0001-01-01" {
						dateTime = ""
					} else {
						// for clients (FileMaker Server)
						dateTime = t.Add(time.Second * time.Duration(offset)).Format("2006/01/02 15:04:05")
					}
				} else {
					// for schedules
					dateTime = t.Add(time.Second * time.Duration(offset)).Format("2006/01/02 15:04")
				}
			}
		}
	}

	return dateTime
}

var helpTextTemplate = `Usage: fmcsadmin [options] [COMMAND]

Description: 
    fmcsadmin is the command line tool to administer the Database Server 
    component of FileMaker Cloud and FileMaker Server via FileMaker Admin 
    API. (FileMaker is a trademark of FileMaker, Inc., registered in the 
    U.S. and other countries.)

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
    DISABLE         Disable schedules
    DISCONNECT      Disconnect clients
    ENABLE          Enable schedules
    GET             Retrieve server or CWP configuration settings
                    (for FileMaker Server 17 only)
    HELP            Get help pages
    LIST            List clients, databases, or schedules
    OPEN            Open databases
    PAUSE           Temporarily stop database access
    RESTART         Restart a server process
                    (for FileMaker Server 17 only)
    RESUME          Make paused databases available
    RUN             Run a schedule
    SEND            Send a message
    SET             Change server or CWP configuration settings
                    (for FileMaker Server 17 only)
    START           Start a server process
                    (for FileMaker Server 17 only)
    STATUS          Get status of clients or databases
    STOP            Stop a server process
                    (for FileMaker Server 17 only)

`

var commandListHelpTextTemplate2 = `fmcsadmin commands are:

    CLOSE           Close databases
    DELETE          Delete a schedule
    DISABLE         Disable schedules
    DISCONNECT      Disconnect clients
    ENABLE          Enable schedules
    GET             Retrieve server or CWP configuration settings
                    (for FileMaker Server 17 only)
    HELP            Get help pages
    LIST            List clients, databases, or schedules
    OPEN            Open databases
    PAUSE           Temporarily stop database access
    RESTART         Restart a server process
                    (for FileMaker Server 17 only)
    RESUME          Make paused databases available
    RUN             Run a schedule
    SEND            Send a message
    SET             Change server or CWP configuration settings
                    (for FileMaker Server 17 only)
    START           Start a server process
                    (for FileMaker Server 17 only)
    STATUS          Get status of clients or databases
    STOP            Stop a server process
                    (for FileMaker Server 17 only)

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

var getHelpTextTemplate = `Usage: fmcsadmin GET [CONFIG_TYPE] [NAME1 NAME2 ...]

Description:
    Retrieve the server or Custom Web Publishing configurations. 
	(This command is not supported for FileMaker Cloud.)

    Valid configuration types of CONFIG_TYPE:
      SERVERCONFIG     Retrieve the server configuration settings.            
      CWPCONFIG        Retrieve the Custom Web Publishing configuration 
                       settings.
   
    Valid configuration names of SERVERCONFIG:
      CACHESIZE        Cache memory allocated by the server, in megabytes.
      HOSTEDFILES      Maximum number of databases that can be hosted.
      PROCONNECTIONS   Maximum number of FileMaker Pro Advanced client
                       connections.
      SCRIPTSESSIONS   Maximum number of script sessions that can run on the
                       server simultaneously.
      SECUREFILESONLY  Whether only databases with password-protected accounts
                       assigned the Full Access privilege set can be opened for
                       hosting.

    Valid configuration names of CWPCONFIG:
      ENABLEPHP        Whether Custom Web Publishing with PHP is enabled.
      ENABLEXML        Whether Custom Web Publishing with XML is enabled.
      ENCODING         The default character encoding for PHP files.  
      LOCALE           Language locale for error messages returned by the
                       FileMaker API for PHP.
      PREVALIDATION    Whether FileMaker API for PHP should validate record data
                       before committing changes to the Database Server.
      USEFMPHP         Whether to use the FileMaker version of the PHP engine
                       rather than your own version of PHP.

    If no configuration name is specified, all supported configurations of the
    corresponding CONFIG_TYPE are listed.

    Examples:
      fmcsadmin GET SERVERCONFIG HOSTEDFILES SCRIPTSESSIONS
      fmcsadmin GET SERVERCONFIG
      fmcsadmin GET CWPCONFIG ENABLEPHP USEFMPHP
      fmcsadmin GET CWPCONFIG

    Note: Input configuration names are not case sensitive.
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

var restartHelpTextTemplate = `Usage: fmcsadmin RESTART [TYPE]

Description:
    Restarts the server of specified TYPE. This command stops the server 
	TYPE and then starts it after a short delay.
	(This command is not supported for FileMaker Cloud.)

    Valid server TYPEs:
        SERVER          Stops then starts the Database Server.

Options: (applicable to SERVER only)
    -m message, --message message 
        Specifies a text message to send to the connected clients.
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
        fmcsadmin SEND -c 2 -m "This is a test message"

Options:
    -m message, --message message
        Specifies the text message to send.
        
    -c, --client
        Specifies a CLIENT_NUMBER.
`

var setHelpTextTemplate = `Usage: fmcsadmin SET [CONFIG_TYPE] [NAME1=VALUE1 NAME2=VALUE2 ...]

Description:
    Change the server or Custom Web Publishing configuration settings.
	(This command is not supported for FileMaker Cloud.)

    Valid configuration types of CONFIG_TYPE:
      SERVERCONFIG     Change the server configuration settings.             
      CWPCONFIG        Change the Custom Web Publishing configuration 
                       settings.
   
    Valid configuration names of SERVERCONFIG:
      CACHESIZE        Cache memory allocated by the server, in megabytes.
      HOSTEDFILES      Maximum number of databases that can be hosted.
      PROCONNECTIONS   Maximum number of FileMaker Pro Advanced client
                       connections.
      SCRIPTSESSIONS   Maximum number of script sessions that can run on the
                       server simultaneously.
      SECUREFILESONLY  Whether only databases with password-protected accounts
                       assigned the Full Access privilege set can be opened for
                       hosting. 

    Valid configuration names of CWPCONFIG:
      ENABLEPHP        Whether Custom Web Publishing with PHP is enabled.
      ENABLEXML        Whether Custom Web Publishing with XML is enabled.
      ENCODING         The default character encoding for PHP files. 
      LOCALE           Language locale for error messages returned by the
                       FileMaker API for PHP.
      PREVALIDATION    Whether FileMaker API for PHP should validate record data
                       before committing changes to the Database Server. 
      USEFMPHP         Whether to use the FileMaker version of the PHP engine
                       rather than your own version of PHP.

    Examples:
      fmsadmin SET SERVERCONFIG CACHESIZE=1024 SECUREFILESONLY=true
      fmsadmin SET CWPCONFIG ENABLEPHP=true ENCODING=ISO-8859-1 LOCALE=de
   
    Note: Input configuration names are not case sensitive.
`

var startHelpTextTemplate = `Usage: fmcsadmin START [TYPE]

Description:
    Starts the server of specified TYPE.
	(This command is not supported for FileMaker Cloud.)

    Valid server TYPE:
        SERVER          Starts the Database Server.

Options:
    No command specific options.
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

var stopHelpTextTemplate = `Usage: fmcsadmin STOP [TYPE] [options]

Description:
    Stops the server of specified TYPE.
	(This command is not supported for FileMaker Cloud.)

    Valid server TYPE:
        SERVER          Stops the Database Server. By default, all clients
                        are disconnected after 90 seconds. 

Options: (applicable to SERVER only)
    -m message, --message message 
        Specifies a text message to send to the connected clients.
`
