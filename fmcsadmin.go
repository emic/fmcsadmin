/*
fmcsadmin
Copyright 2017-2024 Emic Corporation, https://www.emic.co.jp/

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"bufio"
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/mattn/go-scan"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
)

var version string

type cli struct {
	outStream, errStream io.Writer
}

type output struct {
	Response struct {
		Status string `json:"status"`
		Token  string `json:"token"`
	} `json:"response"`
	Messages []struct {
		Code string `json:"code"`
		Text string `json:"text"`
	} `json:"messages"`
}

type generalOldConfigInfo struct {
	CacheSize                 int  `json:"cacheSize"`
	MaxFiles                  int  `json:"maxFiles"`
	MaxProConnections         int  `json:"maxProConnections"`
	MaxPSOS                   int  `json:"maxPSOS"`
	StartupRestorationEnabled bool `json:"startupRestorationEnabled"`
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

type authenticatedStreamConfigInfo struct {
	AuthenticatedStream int `json:"authenticatedStream"`
}

type parallelBackupConfigInfo struct {
	ParallelBackupEnabled bool `json:"parallelBackupEnabled"`
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

type dbInfo struct {
	Status  string `json:"status"`
	Key     string `json:"key"`
	SaveKey bool   `json:"saveKey"`
}

type closeMessageInfo struct {
	Status      string `json:"status"`
	MessageText string `json:"messageText"`
	Force       bool   `json:"force"`
}

type statusInfo struct {
	Status string `json:"status"`
}

type messageInfo struct {
	MessageText string `json:"messageText"`
}

type scheduleSettingInfo struct {
	Enabled bool `json:"enabled"`
}

type creatingCsrInfo struct {
	Subject  string `json:"subject"`
	Password string `json:"password"`
}

type importingCertificateInfo struct {
	Certificate              string `json:"certificate"`
	PrivateKey               string `json:"privateKey"`
	IntermediateCertificates string `json:"intermediateCertificates"`
	Password                 string `json:"password"`
}

type params struct {
	command                   string
	key                       string
	messageText               string
	force                     bool
	retry                     int
	status                    string
	enabled                   string
	cachesize                 int
	maxfiles                  int
	maxproconnections         int
	maxpsos                   int
	startuprestorationenabled bool
	startuprestorationbuiltin bool
	requiresecuredb           string
	authenticatedstream       int
	parallelbackupenabled     bool
	// persistcacheenabled       bool
	characterencoding        string
	errormessagelanguage     string
	dataprevalidation        bool
	usefilemakerphp          bool
	saveKey                  bool
	subject                  string
	password                 string
	certificate              string
	privateKey               string
	intermediateCertificates string
	printRefreshToken        bool
	identityFile             string
}

type commandOptions struct {
	helpFlag       bool
	versionFlag    bool
	yesFlag        bool
	statsFlag      bool
	forceFlag      bool
	saveKeyFlag    bool
	fqdn           string
	hostname       string
	username       string
	password       string
	key            string
	message        string
	keyFile        string
	keyFilePass    string
	intermediateCA string
	clientID       int
	graceTime      int
	identityFile   string
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
	forceFlag := false
	saveKeyFlag := false
	graceTime := 90
	fqdn := ""
	hostname := ""
	username := ""
	password := ""
	key := ""
	clientID := -1
	message := ""
	keyFile := ""
	keyFilePassOption := false
	keyFilePass := ""
	intermediateCA := ""
	identityFile := ""

	commandOptions := commandOptions{}
	commandOptions.helpFlag = false
	commandOptions.versionFlag = false
	commandOptions.yesFlag = false
	commandOptions.statsFlag = false
	commandOptions.forceFlag = false
	commandOptions.saveKeyFlag = false
	commandOptions.fqdn = ""
	commandOptions.hostname = ""
	commandOptions.username = ""
	commandOptions.password = ""
	commandOptions.key = ""
	commandOptions.message = ""
	commandOptions.keyFile = ""
	commandOptions.keyFilePass = ""
	commandOptions.intermediateCA = ""
	commandOptions.clientID = -1
	commandOptions.graceTime = 90
	commandOptions.identityFile = ""

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
		if regexp.MustCompile(`\-(\d+)`).Match([]byte(args[i])) {
			// Allow option (ex.: "fmcsadmin get backuptime -1")
			invalidOption = false
		} else {
			allowedOptions := []string{"-h", "-v", "-y", "-s", "-u", "-p", "-m", "-f", "-c", "-t", "-i", "--help", "--version", "--yes", "--stats", "--fqdn", "--host", "--username", "--password", "--key", "--message", "--force", "--client", "--gracetime", "--savekey", "--keyfile", "--KeyFile", "--keyfilepass", "--KeyFilePass", "--intermediateca", "--intermediateCA"}
			for j := 0; j < len(allowedOptions); j++ {
				if string([]rune(args[i])[:1]) == "-" {
					invalidOption = true
					for _, v := range allowedOptions {
						if strings.ToLower(args[i]) == v {
							if v == "--keyfilepass" {
								keyFilePassOption = true
							}
							invalidOption = false
						}
					}
					if invalidOption {
						exitStatus = outputInvalidOptionErrorMessage(c, args[i])
						return exitStatus
					}
				}
			}
		}
	}

	helpFlag = cFlags.helpFlag
	versionFlag = cFlags.versionFlag
	yesFlag = cFlags.yesFlag
	statsFlag = cFlags.statsFlag
	forceFlag = cFlags.forceFlag
	saveKeyFlag = cFlags.saveKeyFlag
	graceTime = cFlags.graceTime
	key = cFlags.key
	username = cFlags.username
	password = cFlags.password
	clientID = cFlags.clientID
	message = cFlags.message
	keyFile = cFlags.keyFile
	keyFilePass = cFlags.keyFilePass
	intermediateCA = cFlags.intermediateCA
	identityFile = cFlags.identityFile

	fqdn = cFlags.fqdn
	hostname = cFlags.hostname
	if len(fqdn) == 0 && len(hostname) > 0 && !strings.Contains(hostname, ".") {
		fqdn = hostname + ".account.filemaker-cloud.com"
	}
	baseURI := getBaseURI(fqdn)
	u, _ := url.Parse(baseURI)

	usingCloud := false
	if regexp.MustCompile(`https://(.*).account.filemaker-cloud.com`).Match([]byte(baseURI)) {
		// Not Supported
		usingCloud = true
	}

	retry := 3
	if len(username) > 0 && len(password) > 0 {
		// Don't retry when specifying username and password
		retry = 0
	}

	if len(cmdArgs) > 0 {
		switch strings.ToLower(cmdArgs[0]) {
		case "cancel":
			if usingCloud {
				exitStatus = 21
			} else {
				if len(cmdArgs[1:]) > 0 {
					switch strings.ToLower(cmdArgs[1]) {
					case "backup":
						running := true
						u.Path = path.Join(getAPIBasePath(baseURI), "server", "metadata")
						_, err := http.Get(u.String())
						if err != nil {
							running = false
						}

						if running {
							token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
							if token != "" && exitStatus == 0 && err == nil {
								version := getServerVersion(u.String(), token)
								if !usingCloud && version >= 19.5 {
									u.Path = path.Join(getAPIBasePath(baseURI), "server", "cancelbackup")
									exitStatus, _, err = sendRequest("POST", u.String(), token, params{command: "cancel backup"})
									if err == nil {
										fmt.Fprintln(c.outStream, "Command finished")
									} else {
										fmt.Fprintln(c.outStream, err.Error())
									}
								} else {
									exitStatus = outputInvalidCommandErrorMessage(c)
								}
								logout(baseURI, token)
							} else if detectHostUnreachable(exitStatus) {
								exitStatus = 10502
							}
						} else {
							exitStatus = 10502
						}
					default:
						exitStatus = outputInvalidCommandErrorMessage(c)
					}
				} else {
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			}
		case "certificate":
			if usingCloud {
				exitStatus = 21
			} else {
				if len(cmdArgs[1:]) > 0 {
					switch strings.ToLower(cmdArgs[1]) {
					case "create":
						running := true
						u.Path = path.Join(getAPIBasePath(baseURI), "server", "metadata")
						_, err := http.Get(u.String())
						if err != nil {
							running = false
						}

						if running {
							token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
							if token != "" && exitStatus == 0 && err == nil {
								version := getServerVersion(u.String(), token)
								if version >= 19.2 {
									if len(cmdArgs) < 3 {
										fmt.Fprintln(c.outStream, "Certificate subject is not specified.")
										exitStatus = 10001
									}
									if exitStatus == 0 {
										if keyFilePassOption {
											fmt.Fprintln(c.outStream, "Encryption password for the private key file is not specified.")
											exitStatus = 10001
										} else if keyFilePass == "" {
											fmt.Fprintln(c.outStream, "Invalid parameter for option: --KeyFilePass")
											exitStatus = 10001
										} else {
											u.Path = path.Join(getAPIBasePath(baseURI), "server", "certificate", "csr")
											exitStatus, _, err = sendRequest("PATCH", u.String(), token, params{command: "certificate create", subject: base64.StdEncoding.EncodeToString([]byte(cmdArgs[2])), password: keyFilePass})
											if exitStatus == 1712 {
												fmt.Fprintln(c.outStream, "Private key file already exists, please remove it and run the command again.")
												exitStatus = 20406
											} else {
												if err != nil {
													fmt.Fprintln(c.outStream, err.Error())
												}
											}
										}
									}
								} else {
									exitStatus = outputInvalidCommandErrorMessage(c)
								}
								logout(baseURI, token)
							} else if detectHostUnreachable(exitStatus) {
								exitStatus = 10502
							}
						} else {
							exitStatus = 10502
						}
					case "import":
						res := ""
						if yesFlag {
							res = "y"
						} else {
							r := bufio.NewReader(os.Stdin)
							fmt.Fprint(c.outStream, "fmcsadmin: really import certificate? (y, n) (Warning: server needs to be restarted) ")
							input, _ := r.ReadString('\n')
							res = strings.ToLower(strings.TrimSpace(input))
						}
						if res == "y" {
							token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
							if token != "" && exitStatus == 0 && err == nil {
								u.Path = path.Join(getAPIBasePath(baseURI), "server", "metadata")
								version := getServerVersion(u.String(), token)
								if version >= 19.2 {
									if len(cmdArgs[2:]) > 0 {
										keyFileData := []byte("")

										// reading certificate
										var crt *x509.Certificate
										certificateData, err := os.ReadFile(cmdArgs[2])
										if err != nil {
											if os.IsPermission(err) {
												exitStatus = 20402
											} else {
												exitStatus = 20405
											}
										} else {
											block, _ := pem.Decode(certificateData)
											if block == nil {
												exitStatus = 20408
											} else {
												crt, err = x509.ParseCertificate(block.Bytes)
												if err != nil {
													exitStatus = 20408
												}
												if time.Now().UTC().After(crt.NotAfter) {
													// if expired
													exitStatus = 20630
												}
											}
										}

										switch exitStatus {
										case 20402:
											fmt.Fprintln(c.outStream, "Cannot read certificate file")
										case 20405:
											fmt.Fprintln(c.outStream, "Certificate "+filepath.Clean(cmdArgs[2])+" does not exist.")
										case 20408:
											fmt.Fprintln(c.outStream, "The certificate file is not valid.")
										case 20630:
											fmt.Fprintln(c.outStream, "The certificate has expired.")
										}

										// reading private key
										if exitStatus == 0 {
											if keyFile != "" {
												keyFileData, err = os.ReadFile(keyFile)
												if err != nil {
													if os.IsPermission(err) {
														exitStatus = 20402
													} else {
														exitStatus = 20405
													}
												} else {
													block, _ := pem.Decode(keyFileData)
													if block == nil {
														exitStatus = 20408
													} else {
														buf := block.Bytes
														if x509.IsEncryptedPEMBlock(block) {
															buf, err = x509.DecryptPEMBlock(block, []byte(keyFilePass))
															if err != nil {
																if err == x509.IncorrectPasswordError {
																	exitStatus = 20408
																}
															}
														}

														if exitStatus == 0 {
															switch block.Type {
															case "RSA PRIVATE KEY":
																_, err = x509.ParsePKCS1PrivateKey(buf)
																if err != nil {
																	exitStatus = 20408
																}
															case "PRIVATE KEY":
																_, err := x509.ParsePKCS8PrivateKey(buf)
																if err != nil {
																	exitStatus = 20408
																}
															case "EC PRIVATE KEY":
																_, err := x509.ParseECPrivateKey(buf)
																if err != nil {
																	exitStatus = 20408
																}
															default:
																exitStatus = 20408
															}
														}
													}
												}

												switch exitStatus {
												case 20402, 20405:
													fmt.Fprintln(c.outStream, "Cannot read private key file")
												case 20408:
													fmt.Fprintln(c.outStream, "Cannot decrypt the private key file with the password. Please make sure the key file and password are correct.")
												}
											} else {
												fmt.Fprintln(c.outStream, "Private key file does not exist.")
												exitStatus = 20405
											}
										}

										// reading intermediate CA
										intermediateCAData := []byte("")
										intermediateCAExpired := false
										if exitStatus == 0 {
											if intermediateCA != "" {
												intermediateCAData, err = os.ReadFile(intermediateCA)
												if err != nil {
													if os.IsPermission(err) {
														exitStatus = 20402
													} else {
														exitStatus = 20405
													}
												} else {
													var block *pem.Block
													rest := intermediateCAData
													for {
														block, rest = pem.Decode(rest)
														if block == nil {
															exitStatus = 20632
															break
														} else {
															crt, err = x509.ParseCertificate(block.Bytes)
															if err != nil {
																exitStatus = 20632
																break
															}
															if time.Now().UTC().After(crt.NotAfter) {
																// if expired
																intermediateCAExpired = true
															}
														}
														if len(rest) == 0 {
															break
														}
													}
												}
											}
										}

										switch exitStatus {
										case 20402, 20405:
											fmt.Fprintln(c.outStream, "Cannot read intermediate CA file")
										case 20632:
											fmt.Fprintln(c.outStream, "Failed to verify the intermediate CA certificate.")
										}

										// import SSL certficates
										if exitStatus == 0 {
											u.Path = path.Join(getAPIBasePath(baseURI), "server", "certificate", "import")
											exitStatus, _, err = sendRequest("PATCH", u.String(), token, params{command: "certificate import", certificate: string(certificateData), privateKey: string(keyFileData), intermediateCertificates: string(intermediateCAData), password: keyFilePass})

											if exitStatus == 1712 {
												fmt.Fprintln(c.outStream, "Private key file already exists, please remove it and run the command again.")
												exitStatus = 20406
											} else if exitStatus == -1 && intermediateCAExpired {
												fmt.Fprintln(c.outStream, "Failed to verify the intermediate CA certificate.")
												exitStatus = 20630
											} else {
												if err != nil {
													fmt.Fprintln(c.outStream, err.Error())
												}
											}
											if exitStatus == 0 && err == nil {
												fmt.Fprintln(c.outStream, "Restart the FileMaker Server background processes to apply the change.")
											}
										}
									} else {
										fmt.Fprintln(c.outStream, "Certificate file is not specified.")
										exitStatus = 10001
									}
								} else {
									exitStatus = outputInvalidCommandErrorMessage(c)
								}
								logout(baseURI, token)
							} else if detectHostUnreachable(exitStatus) {
								exitStatus = 10502
							}
						}
					case "delete":
						res := ""
						if yesFlag {
							res = "y"
						} else {
							r := bufio.NewReader(os.Stdin)
							fmt.Fprint(c.outStream, "fmcsadmin: really delete certificate? (y, n) (Warning: server needs to be restarted) ")
							input, _ := r.ReadString('\n')
							res = strings.ToLower(strings.TrimSpace(input))
						}
						if res == "y" {
							token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
							if token != "" && exitStatus == 0 && err == nil {
								u.Path = path.Join(getAPIBasePath(baseURI), "server", "metadata")
								version := getServerVersion(u.String(), token)
								if version >= 19.2 {
									u.Path = path.Join(getAPIBasePath(baseURI), "server", "certificate", "delete")
									exitStatus, _, err = sendRequest("DELETE", u.String(), token, params{})
									if err != nil {
										fmt.Fprintln(c.outStream, err.Error())
									}
									if exitStatus == 0 && err == nil {
										fmt.Fprintln(c.outStream, "Restart the FileMaker Server background processes to apply the change.")
									}
								} else {
									exitStatus = outputInvalidCommandErrorMessage(c)
								}
								logout(baseURI, token)
							} else if detectHostUnreachable(exitStatus) {
								exitStatus = 10502
							}
						}
					default:
						exitStatus = outputInvalidCommandErrorMessage(c)
					}
				} else {
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			}
		case "close":
			res := ""
			if yesFlag {
				res = "y"
			} else {
				r := bufio.NewReader(os.Stdin)
				fmt.Fprint(c.outStream, "fmcsadmin: really close database(s)? (y, n) ")
				input, _ := r.ReadString('\n')
				res = strings.ToLower(strings.TrimSpace(input))
			}
			if res == "y" {
				token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
				if token != "" && exitStatus == 0 && err == nil {
					u.Path = path.Join(getAPIBasePath(baseURI), "databases")
					args = []string{""}
					if len(cmdArgs[1:]) > 0 {
						args = cmdArgs[1:]
					}
					idList, nameList, _ := getDatabases(u.String(), token, args, "NORMAL", false)
					if len(idList) > 0 {
						for i := 0; i < len(idList); i++ {
							fmt.Fprintln(c.outStream, "File Closing: "+nameList[i])
						}
						connectedClients := getClients(u.String(), token, args, "")
						for i := 0; i < len(idList); i++ {
							u.Path = path.Join(getAPIBasePath(baseURI), "databases", strconv.Itoa(idList[i]))
							exitStatus, _, err = sendRequest("PATCH", u.String(), token, params{command: "close", messageText: message, force: forceFlag})
							if exitStatus == 0 && err == nil && len(connectedClients) == 0 {
								// Don't output this message when the clients connected to the specified databases are existing
								fmt.Fprintln(c.outStream, "File Closed: "+nameList[i])
							}
						}
					} else {
						exitStatus = 10904
					}
					logout(baseURI, token)
				} else if detectHostUnreachable(exitStatus) {
					exitStatus = 10502
				}
			}
		case "delete":
			if len(cmdArgs[1:]) > 0 {
				switch strings.ToLower(cmdArgs[1]) {
				case "schedule":
					res := ""
					if yesFlag {
						res = "y"
					} else {
						r := bufio.NewReader(os.Stdin)
						fmt.Fprint(c.outStream, "fmcsadmin: really delete a schedule? (y, n) ")
						input, _ := r.ReadString('\n')
						res = strings.ToLower(strings.TrimSpace(input))
					}
					if res == "y" {
						token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
						if token != "" && exitStatus == 0 && err == nil {
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
								if err != nil {
									fmt.Fprintln(c.outStream, err.Error())
								}
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
							logout(baseURI, token)
						} else if detectHostUnreachable(exitStatus) {
							exitStatus = 10502
						}
					}
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "disable":
			if len(cmdArgs[1:]) > 0 {
				switch strings.ToLower(cmdArgs[1]) {
				case "schedule":
					res := ""
					if yesFlag {
						res = "y"
					} else {
						r := bufio.NewReader(os.Stdin)
						fmt.Fprint(c.outStream, "fmcsadmin: really disable schedule(s)? (y, n) ")
						input, _ := r.ReadString('\n')
						res = strings.ToLower(strings.TrimSpace(input))
					}
					if res == "y" {
						token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
						if token != "" && exitStatus == 0 && err == nil {
							id := 0
							if len(cmdArgs) >= 3 {
								sid, err := strconv.Atoi(cmdArgs[2])
								if err == nil {
									id = sid
								}
							}
							if id > 0 {
								u.Path = path.Join(getAPIBasePath(baseURI), "schedules", strconv.Itoa(id))
								exitStatus, _, err = sendRequest("PATCH", u.String(), token, params{command: "disable"})
								if exitStatus == 0 && err == nil {
									u.Path = path.Join(getAPIBasePath(baseURI), "schedules")
									exitStatus = listSchedules(u.String(), token, id)
								}
							} else {
								exitStatus = 10600
							}
							logout(baseURI, token)
						} else if detectHostUnreachable(exitStatus) {
							exitStatus = 10502
						}
					}
				default:
					exitStatus = -1
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "disconnect":
			if len(cmdArgs[1:]) > 0 {
				switch strings.ToLower(cmdArgs[1]) {
				case "client":
					res := ""
					if yesFlag {
						res = "y"
					} else {
						r := bufio.NewReader(os.Stdin)
						fmt.Fprint(c.outStream, "fmcsadmin: really disconnect client(s)? (y, n) ")
						input, _ := r.ReadString('\n')
						res = strings.ToLower(strings.TrimSpace(input))
					}
					if res == "y" {
						token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
						if token != "" && exitStatus == 0 && err == nil {
							id := 0
							if len(cmdArgs) >= 3 {
								cid, err := strconv.Atoi(cmdArgs[2])
								if err == nil {
									id = cid
								}
								if cid == 0 {
									exitStatus = 11005
								}
							}
							if id > -1 && exitStatus == 0 {
								if id == 0 {
									// disconnect clients
									exitStatus, _ = disconnectAllClient(u, baseURI, token, message, graceTime)
								} else {
									// check the client connection
									u.Path = path.Join(getAPIBasePath(baseURI), "clients")
									idList := getClients(u.String(), token, []string{""}, "NORMAL")
									connected := false
									if len(idList) > 0 && id > 0 {
										for i := 0; i < len(idList); i++ {
											if id == idList[i] {
												connected = true
												break
											}
										}
									}

									if connected {
										// disconnect a client
										u.Path = path.Join(getAPIBasePath(baseURI), "clients", strconv.Itoa(id))
										u.RawQuery = "messageText=" + url.QueryEscape(message) + "&graceTime=" + url.QueryEscape(strconv.Itoa(graceTime))
										exitStatus, _, _ = sendRequest("DELETE", u.String(), token, params{command: "disconnect"})
									} else {
										exitStatus = 11005
									}
								}
								if exitStatus == 0 {
									fmt.Fprintln(c.outStream, "Client(s) being disconnected.")
								}
							}
							logout(baseURI, token)
						} else if detectHostUnreachable(exitStatus) {
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
				token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
				if token != "" && exitStatus == 0 && err == nil {
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
							exitStatus, _, err = sendRequest("PATCH", u.String(), token, params{command: "enable"})
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
				} else if detectHostUnreachable(exitStatus) {
					exitStatus = 10502
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "get":
			if len(cmdArgs[1:]) > 0 {
				switch strings.ToLower(cmdArgs[1]) {
				case "backuptime":
					if usingCloud {
						exitStatus = 21
					} else {
						token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
						if token != "" && exitStatus == 0 && err == nil {
							id := 0
							if len(cmdArgs) >= 3 {
								sid, err := strconv.Atoi(cmdArgs[2])
								if err == nil {
									id = sid
								}
							}
							u.Path = path.Join(getAPIBasePath(baseURI), "schedules")
							exitStatus = getBackupTime(u.String(), token, id)
							logout(baseURI, token)
						} else if detectHostUnreachable(exitStatus) {
							exitStatus = 10502
						}
					}
				case "cwpconfig":
					if usingCloud {
						exitStatus = 21
					} else {
						printOptions := []string{}
						if len(cmdArgs[2:]) > 0 {
							for i := 0; i < len(cmdArgs[2:]); i++ {
								switch strings.ToLower(cmdArgs[2:][i]) {
								case "enablephp":
									printOptions = append(printOptions, "enablephp")
								case "enablexml":
									printOptions = append(printOptions, "enablexml")
								case "encoding":
									printOptions = append(printOptions, "encoding")
								case "locale":
									printOptions = append(printOptions, "locale")
								case "prevalidation":
									printOptions = append(printOptions, "prevalidation")
								case "usefmphp":
									printOptions = append(printOptions, "usefmphp")
								default:
									exitStatus = 10001
								}
								if exitStatus != 0 {
									break
								}
							}
						} else {
							printOptions = append(printOptions, "enablephp")
							printOptions = append(printOptions, "enablexml")
							printOptions = append(printOptions, "encoding")
							printOptions = append(printOptions, "locale")
							printOptions = append(printOptions, "prevalidation")
							printOptions = append(printOptions, "usefmphp")
						}

						for i := 0; i < len(cmdArgs[2:]); i++ {
							if regexp.MustCompile(`(.*)`).Match([]byte(cmdArgs[2:][i])) {
								rep := regexp.MustCompile(`(.*)`)
								option := rep.ReplaceAllString(cmdArgs[2:][i], "$1")
								switch strings.ToLower(option) {
								case "enablephp", "enablexml", "encoding", "locale", "prevalidation", "usefmphp":
								default:
									exitStatus = 10001
								}

								if exitStatus == 10001 {
									fmt.Fprintln(c.outStream, "Invalid configuration name: "+option)
									break
								}
							}
						}

						if exitStatus == 0 {
							token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
							if token != "" && exitStatus == 0 && err == nil {
								u.Path = path.Join(getAPIBasePath(baseURI), "server", "metadata")
								version := getServerVersion(u.String(), token)
								if runtime.GOOS == "linux" && fqdn == "" && version < 19.6 {
									// Not Supported
									exitStatus = 21
								} else {
									if exitStatus == 0 {
										_, exitStatus, _ = getWebTechnologyConfigurations(baseURI, getAPIBasePath(baseURI), token, printOptions)
									}
								}
								logout(baseURI, token)
							} else if detectHostUnreachable(exitStatus) {
								exitStatus = 10502
							}
						}
					}
				case "refreshtoken":
					if usingCloud {
						token, exitStatus, err = login(baseURI, username, password, params{printRefreshToken: true, retry: retry, identityFile: identityFile})
						if token != "" && exitStatus == 0 && err == nil {
							logout(baseURI, token)
						} else if detectHostUnreachable(exitStatus) {
							exitStatus = 10502
						}
					} else {
						exitStatus = outputInvalidCommandErrorMessage(c)
					}
				case "serverconfig":
					if usingCloud {
						exitStatus = 21
					} else {
						if len(cmdArgs[2:]) > 0 {
							for i := 0; i < len(cmdArgs[2:]); i++ {
								if regexp.MustCompile(`(.*)`).Match([]byte(cmdArgs[2:][i])) {
									rep := regexp.MustCompile(`(.*)`)
									option := rep.ReplaceAllString(cmdArgs[2:][i], "$1")
									switch strings.ToLower(option) {
									case "cachesize", "hostedfiles", "proconnections", "scriptsessions", "securefilesonly":
									default:
										exitStatus = 10001
									}

									if exitStatus == 10001 {
										break
									}
								}
							}
						}

						if exitStatus == 0 {
							token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
							if token != "" && exitStatus == 0 && err == nil {
								printOptions := []string{}
								if len(cmdArgs[2:]) > 0 {
									for i := 0; i < len(cmdArgs[2:]); i++ {
										switch strings.ToLower(cmdArgs[2:][i]) {
										case "cachesize":
											printOptions = append(printOptions, "cachesize")
										case "hostedfiles":
											printOptions = append(printOptions, "hostedfiles")
										case "proconnections":
											printOptions = append(printOptions, "proconnections")
										case "scriptsessions":
											printOptions = append(printOptions, "scriptsessions")
										case "securefilesonly":
											printOptions = append(printOptions, "securefilesonly")
										default:
											exitStatus = 10001
										}
										if exitStatus != 0 {
											break
										}
									}
								} else {
									printOptions = append(printOptions, "cachesize")
									printOptions = append(printOptions, "hostedfiles")
									printOptions = append(printOptions, "proconnections")
									printOptions = append(printOptions, "scriptsessions")
									printOptions = append(printOptions, "securefilesonly")
								}
								if exitStatus == 0 {
									u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "general")
									_, exitStatus = getServerGeneralConfigurations(u.String(), token, printOptions)
								}
								logout(baseURI, token)
							} else if detectHostUnreachable(exitStatus) {
								exitStatus = 10502
							}
						}
					}
				case "serverprefs":
					startupRestoration := false
					if len(cmdArgs[2:]) > 0 {
						for i := 0; i < len(cmdArgs[2:]); i++ {
							if regexp.MustCompile(`(.*)`).Match([]byte(cmdArgs[2:][i])) {
								rep := regexp.MustCompile(`(.*)`)
								option := rep.ReplaceAllString(cmdArgs[2:][i], "$1")
								switch strings.ToLower(option) {
								case "maxguests", "maxfiles", "cachesize", "allowpsos", "requiresecuredb":
								case "startuprestorationenabled":
									startupRestoration = true
								case "authenticatedstream", "parallelbackupenabled":
								case "persistcacheenabled", "syncpersistcache":
								default:
									exitStatus = 3
								}

								if exitStatus != 0 {
									break
								}
							}
						}
					}

					if exitStatus == 0 {
						token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
						if token != "" && exitStatus == 0 && err == nil {
							var versionString string
							var version float64

							printOptions := []string{}
							if usingCloud {
								// for Claris FileMaker Cloud
								printOptions = append(printOptions, "authenticatedstream")
							} else {
								// for Claris FileMaker Server
								u.Path = path.Join(getAPIBasePath(baseURI), "server", "metadata")
								versionString, _ = getServerVersionString(u.String(), token)
								version, _ = getServerVersionAsFloat(versionString)

								if len(cmdArgs[2:]) > 0 {
									for i := 0; i < len(cmdArgs[2:]); i++ {
										switch strings.ToLower(cmdArgs[2:][i]) {
										case "maxguests":
											printOptions = append(printOptions, "maxguests")
										case "maxfiles":
											printOptions = append(printOptions, "maxfiles")
										case "cachesize":
											printOptions = append(printOptions, "cachesize")
										case "allowpsos":
											printOptions = append(printOptions, "allowpsos")
										case "requiresecuredb":
											printOptions = append(printOptions, "requiresecuredb")
										case "startuprestorationenabled":
											printOptions = append(printOptions, "startuprestorationenabled")
										case "authenticatedstream":
											printOptions = append(printOptions, "authenticatedstream")
										case "parallelbackupenabled":
											printOptions = append(printOptions, "parallelbackupenabled")
										case "persistcacheenabled":
											printOptions = append(printOptions, "persistcacheenabled")
										case "syncpersistcache":
											printOptions = append(printOptions, "syncpersistcache")
										default:
											exitStatus = 3
										}
										if exitStatus != 0 {
											break
										}
									}
								} else {
									printOptions = append(printOptions, "maxguests")
									printOptions = append(printOptions, "maxfiles")
									printOptions = append(printOptions, "cachesize")
									printOptions = append(printOptions, "allowpsos")
									printOptions = append(printOptions, "requiresecuredb")
									printOptions = append(printOptions, "startuprestorationenabled")
									if (version >= 19.3 && !strings.HasPrefix(versionString, "19.3.1")) || usingCloud {
										printOptions = append(printOptions, "authenticatedstream")
									}
									if !usingCloud && version >= 19.5 {
										printOptions = append(printOptions, "parallelbackupenabled")
									}
									if !usingCloud && version >= 20.1 {
										printOptions = append(printOptions, "persistcacheenabled")
										printOptions = append(printOptions, "syncpersistcache")
									}
								}

								if version >= 19.2 && startupRestoration {
									exitStatus = 3
								}
							}

							if exitStatus == 0 {
								if !usingCloud {
									u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "general")
									_, exitStatus = getServerGeneralConfigurations(u.String(), token, printOptions)
								}

								for _, option := range printOptions {
									if option == "authenticatedstream" {
										if usingCloud {
											// for Claris FileMaker Cloud
											u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "authenticatedstream")
											_, exitStatus, _ = getAuthenticatedStreamSetting(u.String(), token, printOptions)
										} else {
											// for Claris FileMaker Server
											if version < 19.3 || strings.HasPrefix(versionString, "19.3.1") {
												exitStatus = 3
											}
										}
									}

									if option == "parallelbackupenabled" {
										if usingCloud {
											// for Claris FileMaker Cloud
											exitStatus = 3
										} else {
											// for Claris FileMaker Server
											if version < 19.5 {
												exitStatus = 3
											}
										}
									}

									if option == "persistcacheenabled" || option == "syncpersistcache" {
										if usingCloud {
											// for Claris FileMaker Cloud
											exitStatus = 3
										} else {
											// for Claris FileMaker Server
											if version < 20.1 {
												exitStatus = 3
											}
										}
									}
								}
							}

							logout(baseURI, token)
						} else if detectHostUnreachable(exitStatus) {
							exitStatus = 10502
						}
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
					fmt.Fprint(c.outStream, commandListHelpTextTemplate)
				case "options":
					fmt.Fprint(c.outStream, optionListHelpTextTemplate)
				case "cancel":
					fmt.Fprint(c.outStream, cancelHelpTextTemplate)
				case "certificate":
					fmt.Fprint(c.outStream, certificateHelpTextTemplate)
				case "close":
					fmt.Fprint(c.outStream, closeHelpTextTemplate)
				case "delete":
					fmt.Fprint(c.outStream, deleteHelpTextTemplate)
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
				case "remove":
					fmt.Fprint(c.outStream, removeHelpTextTemplate)
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
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
					if token != "" && exitStatus == 0 && err == nil {
						id := -1
						if statsFlag {
							id = 0
						}
						u.Path = path.Join(getAPIBasePath(baseURI), "clients")
						exitStatus = listClients(u.String(), token, id)
						logout(baseURI, token)
					} else if detectHostUnreachable(exitStatus) {
						exitStatus = 10502
					}
				case "files":
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
					if token != "" && exitStatus == 0 && err == nil {
						idList := []int{-1}
						if statsFlag {
							idList = []int{0}
						}
						u.Path = path.Join(getAPIBasePath(baseURI), "databases")
						exitStatus = listFiles(c, u.String(), token, idList)
						logout(baseURI, token)
					} else if detectHostUnreachable(exitStatus) {
						exitStatus = 10502
					}
				case "plugins":
					if usingCloud {
						exitStatus = 21
					} else {
						token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
						if token != "" && exitStatus == 0 && err == nil {
							u.Path = path.Join(getAPIBasePath(baseURI), "server", "metadata")
							version := getServerVersion(u.String(), token)
							if version >= 19.2 {
								u.Path = path.Join(getAPIBasePath(baseURI), "plugins")
								exitStatus = listPlugins(u.String(), token)
							} else {
								var running string
								u.Path = path.Join(getAPIBasePath(baseURI), "server", "status")
								_, running, _ = sendRequest("GET", u.String(), token, params{})
								if running == "STOPPED" {
									exitStatus = 10502
								} else {
									exitStatus = outputInvalidCommandErrorMessage(c)
								}
							}
							logout(baseURI, token)
						} else if detectHostUnreachable(exitStatus) {
							exitStatus = 10502
						}
					}
				case "schedules":
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
					if token != "" && exitStatus == 0 && err == nil {
						u.Path = path.Join(getAPIBasePath(baseURI), "schedules")
						exitStatus = listSchedules(u.String(), token, 0)
						logout(baseURI, token)
					} else if detectHostUnreachable(exitStatus) {
						exitStatus = 10502
					}
				default:
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "open":
			token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
			if token != "" && exitStatus == 0 && err == nil {
				u.Path = path.Join(getAPIBasePath(baseURI), "databases")
				args = []string{""}
				if len(cmdArgs[1:]) > 0 {
					args = cmdArgs[1:]
				}
				idList, nameList, hintList := getDatabases(u.String(), token, args, "CLOSED", false)
				if len(idList) > 0 {
					if usingCloud && (len(key) > 0 || saveKeyFlag) {
						if len(key) > 0 {
							exitStatus = outputInvalidOptionErrorMessage(c, "--key")
						} else {
							exitStatus = outputInvalidOptionErrorMessage(c, "--savekey")
						}
					} else {
						for i := 0; i < len(idList); i++ {
							fmt.Fprintln(c.outStream, "File Opening: "+nameList[i])
						}
						for i := 0; i < len(idList); i++ {
							u.Path = path.Join(getAPIBasePath(baseURI), "databases", strconv.Itoa(idList[i]))
							exitStatus, _, err = sendRequest("PATCH", u.String(), token, params{command: "open", key: key, saveKey: saveKeyFlag})
							if exitStatus == 0 && err == nil {
								// Note: FileMaker Admin API does not validate the encryption key.
								//       You receive a result code of 0 even if you enter an invalid key.
								var openedID []int
								for value := 0; ; {
									value++
									u.Path = path.Join(getAPIBasePath(baseURI), "databases")
									openedID, _, _ = getDatabases(u.String(), token, []string{strconv.Itoa(idList[i])}, "NORMAL", false)
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
					}
				} else {
					exitStatus = 10904
				}
				logout(baseURI, token)
			} else if detectHostUnreachable(exitStatus) {
				exitStatus = 10502
			}
		case "pause":
			token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
			if token != "" && exitStatus == 0 && err == nil {
				u.Path = path.Join(getAPIBasePath(baseURI), "databases")
				args = []string{""}
				if len(cmdArgs[1:]) > 0 {
					args = cmdArgs[1:]
				}
				idList, nameList, _ := getDatabases(u.String(), token, args, "NORMAL", false)
				if len(idList) > 0 {
					for i := 0; i < len(idList); i++ {
						fmt.Fprintln(c.outStream, "File Pausing: "+nameList[i])
					}
					for i := 0; i < len(idList); i++ {
						u.Path = path.Join(getAPIBasePath(baseURI), "databases", strconv.Itoa(idList[i]))
						exitStatus, _, err = sendRequest("PATCH", u.String(), token, params{command: "pause"})
						if exitStatus == 0 && err == nil {
							fmt.Fprintln(c.outStream, "File Paused: "+nameList[i])
						}
					}
				} else {
					exitStatus = 10904
				}
				logout(baseURI, token)
			} else if detectHostUnreachable(exitStatus) {
				exitStatus = 10502
			}
		case "remove":
			res := ""
			if yesFlag {
				res = "y"
			} else {
				r := bufio.NewReader(os.Stdin)
				fmt.Fprint(c.outStream, "fmcsadmin: really remove database(s)? (y, n) ")
				input, _ := r.ReadString('\n')
				res = strings.ToLower(strings.TrimSpace(input))
			}
			if res == "y" {
				token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
				if token != "" && exitStatus == 0 && err == nil {
					var version float64
					if !usingCloud {
						u.Path = path.Join(getAPIBasePath(baseURI), "server", "metadata")
						version = getServerVersion(u.String(), token)
					}
					if version >= 19.3 || usingCloud {
						u.Path = path.Join(getAPIBasePath(baseURI), "databases")
						args = []string{""}
						if len(cmdArgs[1:]) > 0 {
							args = cmdArgs[1:]
						}
						idList, nameList, _ := getDatabases(u.String(), token, args, "CLOSED", true)
						if len(idList) > 0 {
							for i := 0; i < len(idList); i++ {
								u.Path = path.Join(getAPIBasePath(baseURI), "databases", strconv.Itoa(idList[i]))
								exitStatus, _, err = sendRequest("DELETE", u.String(), token, params{})
								if exitStatus == 0 && err == nil {
									fmt.Fprintln(c.outStream, "File Removed: "+nameList[i])
								}
							}
						} else {
							_, nameList, _ = getDatabases(u.String(), token, args, "", true)
							exitStatus = 10904
							for i := 0; i < len(nameList); i++ {
								if len(args) > 0 && comparePath(args[0], string(os.PathSeparator)+"Library"+string(os.PathSeparator)+"FileMaker Server"+string(os.PathSeparator)+"Data"+string(os.PathSeparator)+"Databases"+string(os.PathSeparator)) {
									// File not found or not accessible
									exitStatus = 20405
									break
								}
								if len(args) > 0 && comparePath(args[0], filepath.Dir(nameList[i])+string(os.PathSeparator)) {
									// Directory not empty
									exitStatus = 20501
									break
								}
							}
						}
					} else {
						exitStatus = outputInvalidCommandErrorMessage(c)
					}
					logout(baseURI, token)
				} else if detectHostUnreachable(exitStatus) {
					exitStatus = 10502
				}
			}
		case "restart":
			if usingCloud {
				exitStatus = 21
			} else {
				if len(cmdArgs[1:]) > 0 {
					res := ""
					if yesFlag {
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
							token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
							if token != "" && exitStatus == 0 && err == nil {
								// stop database server
								if forceFlag {
									graceTime = 0
								}
								exitStatus, _ = stopDatabaseServer(u, baseURI, token, message, graceTime)
								if exitStatus == 0 {
									_, _ = waitStoppingServer(u, baseURI, token)
									// start database server
									exitStatus, _, _ = sendRequest("PATCH", u.String(), token, params{status: "RUNNING"})
								}
								logout(baseURI, token)
							} else if detectHostUnreachable(exitStatus) {
								exitStatus = 10502
							}
						default:
							exitStatus = outputInvalidCommandParameterErrorMessage(c)
						}
					}
				} else {
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			}
		case "resume":
			token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
			if token != "" && exitStatus == 0 && err == nil {
				u.Path = path.Join(getAPIBasePath(baseURI), "databases")
				args = []string{""}
				if len(cmdArgs[1:]) > 0 {
					args = cmdArgs[1:]
				}
				idList, nameList, _ := getDatabases(u.String(), token, args, "PAUSED", false)
				if len(idList) > 0 {
					for i := 0; i < len(idList); i++ {
						fmt.Fprintln(c.outStream, "File Resuming: "+nameList[i])
					}
					for i := 0; i < len(idList); i++ {
						u.Path = path.Join(getAPIBasePath(baseURI), "databases", strconv.Itoa(idList[i]))
						exitStatus, _, err = sendRequest("PATCH", u.String(), token, params{command: "resume"})
						if exitStatus == 0 && err == nil {
							fmt.Fprintln(c.outStream, "File Resumed: "+nameList[i])
						}
					}
				} else {
					exitStatus = 10904
				}
				logout(baseURI, token)
			} else if detectHostUnreachable(exitStatus) {
				exitStatus = 10502
			}
		case "run":
			if len(cmdArgs[1:]) > 0 {
				switch strings.ToLower(cmdArgs[1]) {
				case "schedule":
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
					if token != "" && exitStatus == 0 && err == nil {
						id := 0
						if len(cmdArgs) >= 3 {
							sid, err := strconv.Atoi(cmdArgs[2])
							if err == nil {
								id = sid
							}
						}
						if id > 0 {
							u.Path = path.Join(getAPIBasePath(baseURI), "schedules", strconv.Itoa(id))
							exitStatus, _, err = sendRequest("PATCH", u.String(), token, params{status: "RUNNING"})
							if exitStatus == 0 && err == nil {
								u.Path = path.Join(getAPIBasePath(baseURI), "schedules", strconv.Itoa(id))
								scheduleName := getScheduleName(u.String(), token, id)
								if scheduleName != "" {
									fmt.Fprintln(c.outStream, "Schedule '"+scheduleName+"' will run now.")
								} else {
									exitStatus = 10600
								}
							} else {
								exitStatus = 10600
							}
						} else {
							exitStatus = 10600
						}
						logout(baseURI, token)
					} else if detectHostUnreachable(exitStatus) {
						exitStatus = 10502
					}
				default:
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "send":
			token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
			if token != "" && exitStatus == 0 && err == nil {
				exitStatus = sendMessages(u, baseURI, token, message, cmdArgs, clientID)
				logout(baseURI, token)
			} else if detectHostUnreachable(exitStatus) {
				exitStatus = 10502
			}
		case "set":
			if len(cmdArgs[1:]) > 0 {
				switch strings.ToLower(cmdArgs[1]) {
				case "cwpconfig":
					if usingCloud {
						exitStatus = 21
					} else {
						if len(cmdArgs[2:]) > 0 {
							for i := 0; i < len(cmdArgs[2:]); i++ {
								if regexp.MustCompile(`(.*)=(.*)`).Match([]byte(cmdArgs[2:][i])) {
									rep := regexp.MustCompile(`(.*)=(.*)`)
									option := rep.ReplaceAllString(cmdArgs[2:][i], "$1")
									switch strings.ToLower(option) {
									case "enablephp", "enablexml", "encoding", "locale", "prevalidation", "usefmphp":
									default:
										exitStatus = 10001
									}

									if exitStatus == 10001 {
										fmt.Fprintln(c.outStream, "Invalid configuration name: "+option)
										break
									}
								} else {
									exitStatus = 10001
									break
								}
							}

							if exitStatus == 0 {
								token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
								if token != "" && exitStatus == 0 && err == nil {
									u.Path = path.Join(getAPIBasePath(baseURI), "server", "metadata")
									version := getServerVersion(u.String(), token)
									if runtime.GOOS == "linux" && fqdn == "" && version < 19.6 {
										// Not Supported
										exitStatus = 10001
									} else {
										var settings []string
										printOptions := []string{}
										settings, exitStatus, err = getWebTechnologyConfigurations(baseURI, getAPIBasePath(baseURI), token, printOptions)
										if err == nil {
											var results []string
											results, exitStatus = parseWebConfigurationSettings(c, cmdArgs[2:])

											phpFlag := results[0]
											xmlFlag := results[1]
											encoding := results[2]
											locale := results[3]
											preValidationFlag := results[4]
											useFMPHPFlag := results[5]

											var phpEnabled string
											var xmlEnabled string
											var preValidation bool
											var useFMPHP bool

											if len(cmdArgs[2:]) > 0 {
												for i := 0; i < len(cmdArgs[2:]); i++ {
													if regexp.MustCompile(`(.*)=(.*)`).Match([]byte(cmdArgs[2:][i])) {
														rep := regexp.MustCompile(`(.*)=(.*)`)
														option := rep.ReplaceAllString(cmdArgs[2:][i], "$1")
														value := rep.ReplaceAllString(cmdArgs[2:][i], "$2")
														switch strings.ToLower(option) {
														case "enablephp":
															printOptions = append(printOptions, "enablephp")
															if !(strings.ToLower(value) == "true" || strings.ToLower(value) == "false") {
																fmt.Println("Invalid configuration value: " + value)
																exitStatus = 10001
															}
														case "enablexml":
															printOptions = append(printOptions, "enablexml")
															if !(strings.ToLower(value) == "true" || strings.ToLower(value) == "false") {
																fmt.Println("Invalid configuration value: " + value)
																exitStatus = 10001
															}
														case "encoding":
															printOptions = append(printOptions, "encoding")
															if !(strings.ToLower(value) == "utf-8" || strings.ToLower(value) == "iso-8859-1") {
																fmt.Println("Invalid configuration value: " + value)
																exitStatus = 10001
															}
														case "locale":
															printOptions = append(printOptions, "locale")
															if !(strings.ToLower(value) == "en" || strings.ToLower(value) == "de" || strings.ToLower(value) == "fr" || strings.ToLower(value) == "it" || strings.ToLower(value) == "ja") {
																fmt.Println("Invalid configuration value: " + value)
																exitStatus = 10001
															}
														case "prevalidation":
															printOptions = append(printOptions, "prevalidation")
															if !(strings.ToLower(value) == "true" || strings.ToLower(value) == "false") {
																fmt.Println("Invalid configuration value: " + value)
																exitStatus = 10001
															}
														case "usefmphp":
															printOptions = append(printOptions, "usefmphp")
															if !(strings.ToLower(value) == "true" || strings.ToLower(value) == "false") {
																fmt.Println("Invalid configuration value: " + value)
																exitStatus = 10001
															}
														default:
															fmt.Fprintln(c.outStream, "Invalid configuration name: "+option)
															exitStatus = 10001
														}
													}
													if exitStatus != 0 {
														break
													}
												}
											} else {
												printOptions = append(printOptions, "enablephp")
												printOptions = append(printOptions, "enablexml")
												printOptions = append(printOptions, "encoding")
												printOptions = append(printOptions, "locale")
												printOptions = append(printOptions, "prevalidation")
												printOptions = append(printOptions, "usefmphp")
											}

											restartMessageFlag := false
											if exitStatus == 0 && (len(phpFlag) > 0 || len(encoding) > 0 || len(locale) > 0 || len(preValidationFlag) > 0 || len(useFMPHPFlag) > 0) {
												if strings.ToLower(phpFlag) == "true" {
													phpEnabled = "true"
													if settings[0] == "false" && settings[4] != "" {
														restartMessageFlag = true
													}
												} else if strings.ToLower(phpFlag) == "false" {
													phpEnabled = "false"
													if settings[0] == "true" && settings[4] != "" {
														restartMessageFlag = true
													}
												} else if settings[0] == "true" {
													phpEnabled = "true"
												} else if settings[0] == "false" {
													phpEnabled = "false"
												}

												if encoding == "" {
													encoding = settings[2]
												}

												if locale == "" {
													locale = settings[3]
												}

												if strings.ToLower(preValidationFlag) == "true" {
													preValidation = true
												} else if strings.ToLower(preValidationFlag) == "false" {
													preValidation = false
												} else if settings[4] == "true" {
													preValidation = true
												} else if settings[4] == "false" {
													preValidation = false
												}

												if strings.ToLower(useFMPHPFlag) == "true" {
													useFMPHP = true
													if settings[5] == "false" && settings[4] != "" {
														restartMessageFlag = true
													}
												} else if strings.ToLower(useFMPHPFlag) == "false" {
													if phpEnabled == "false" {
														// UseFMPHP is always true when enablePHP is false
														useFMPHP = true
													} else {
														useFMPHP = false
														if settings[5] == "true" && settings[4] != "" {
															restartMessageFlag = true
														}
													}
												} else if settings[5] == "true" {
													useFMPHP = true
												} else if settings[5] == "false" {
													useFMPHP = false
												}

												u.Path = path.Join(getAPIBasePath(baseURI), "php", "config")
												if settings[4] != "" {
													// exclude Claris FileMaker Server for Linux
													exitStatus, _, _ = sendRequest("PATCH", u.String(), token, params{
														command:              "set",
														enabled:              phpEnabled,
														characterencoding:    encoding,
														errormessagelanguage: locale,
														dataprevalidation:    preValidation,
														usefilemakerphp:      useFMPHP,
													})
												}
											}

											if exitStatus == 0 {
												if strings.ToLower(xmlFlag) == "true" || strings.ToLower(xmlFlag) == "false" {
													if strings.ToLower(xmlFlag) == "true" {
														xmlEnabled = "true"
													} else if strings.ToLower(xmlFlag) == "false" {
														xmlEnabled = "false"
													} else if settings[1] == "true" {
														xmlEnabled = "true"
													} else if settings[1] == "false" {
														xmlEnabled = "false"
													}

													u.Path = path.Join(getAPIBasePath(baseURI), "xml", "config")
													_, _, _ = sendRequest("PATCH", u.String(), token, params{command: "set", enabled: xmlEnabled})
												}

												_, exitStatus, _ = getWebTechnologyConfigurations(baseURI, getAPIBasePath(baseURI), token, printOptions)
												if restartMessageFlag {
													fmt.Fprintln(c.outStream, "Restart the FileMaker Server background processes to apply the change.")
												}
											}
										}
									}
									logout(baseURI, token)
								} else if detectHostUnreachable(exitStatus) {
									exitStatus = 10502
								}
							}
						} else {
							exitStatus = 10001
						}
					}
				case "serverconfig":
					if usingCloud {
						exitStatus = 21
					} else {
						if len(cmdArgs[2:]) > 0 {
							for i := 0; i < len(cmdArgs[2:]); i++ {
								if regexp.MustCompile(`(.*)=(.*)`).Match([]byte(cmdArgs[2:][i])) {
									rep := regexp.MustCompile(`(.*)=(.*)`)
									option := rep.ReplaceAllString(cmdArgs[2:][i], "$1")
									switch strings.ToLower(option) {
									case "cachesize", "hostedfiles", "proconnections", "scriptsessions", "securefilesonly":
									default:
										exitStatus = 10001
									}

									if exitStatus == 10001 {
										break
									}
								} else {
									exitStatus = 10001
									break
								}
							}
						} else {
							exitStatus = 10001
						}

						if exitStatus == 0 {
							token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
							if token != "" && exitStatus == 0 && err == nil {
								var settings []int
								printOptions := []string{}
								u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "general")
								settings, exitStatus = getServerGeneralConfigurations(u.String(), token, printOptions)
								if exitStatus == 0 {
									var results []string
									results, exitStatus = parseServerConfigurationSettings(c, cmdArgs[2:])

									cacheSize, _ := strconv.Atoi(results[0])
									maxFiles, _ := strconv.Atoi(results[1])
									maxProConnections, _ := strconv.Atoi(results[2])
									maxPSOS, _ := strconv.Atoi(results[3])
									startupRestorationEnabled := results[4]
									secureFilesOnlyFlag := results[5]
									authenticatedStream := results[6]

									if results[0] != "" || results[1] != "" || results[2] != "" || results[3] != "" || startupRestorationEnabled != "" || secureFilesOnlyFlag != "" || authenticatedStream != "" {
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

										startupRestorationBuiltin := true
										if settings[4] == -1 {
											// for Claris FileMaker Server 19.1.2 or later
											startupRestorationBuiltin = false
										}

										printOptions = []string{}
										if len(cmdArgs[2:]) > 0 {
											for i := 0; i < len(cmdArgs[2:]); i++ {
												if regexp.MustCompile(`(.*)=(.*)`).Match([]byte(cmdArgs[2:][i])) {
													rep := regexp.MustCompile(`(.*)=(.*)`)
													option := rep.ReplaceAllString(cmdArgs[2:][i], "$1")
													switch strings.ToLower(option) {
													case "cachesize":
														printOptions = append(printOptions, "cachesize")
													case "hostedfiles":
														printOptions = append(printOptions, "hostedfiles")
													case "proconnections":
														printOptions = append(printOptions, "proconnections")
													case "scriptsessions":
														printOptions = append(printOptions, "scriptsessions")
													case "securefilesonly":
														printOptions = append(printOptions, "securefilesonly")
													case "authenticatedstream":
														printOptions = append(printOptions, "authenticatedstream")
													default:
														exitStatus = 10001
													}
													if exitStatus != 0 {
														break
													}
												}
											}
										} else {
											printOptions = append(printOptions, "cachesize")
											printOptions = append(printOptions, "hostedfiles")
											printOptions = append(printOptions, "proconnections")
											printOptions = append(printOptions, "scriptsessions")
											printOptions = append(printOptions, "securefilesonly")
											printOptions = append(printOptions, "authenticatedstream")
										}
										if exitStatus == 0 {
											if results[0] != "" || results[1] != "" || results[2] != "" || results[3] != "" {
												u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "general")
												exitStatus, _, _ = sendRequest("PATCH", u.String(), token, params{
													command:                   "set",
													cachesize:                 cacheSize,
													maxfiles:                  maxFiles,
													maxproconnections:         maxProConnections,
													maxpsos:                   maxPSOS,
													startuprestorationbuiltin: startupRestorationBuiltin,
												})
											}

											if exitStatus == 0 && (secureFilesOnlyFlag == "true" || secureFilesOnlyFlag == "false") {
												u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "security")
												exitStatus, _, _ = sendRequest("PATCH", u.String(), token, params{command: "set", requiresecuredb: secureFilesOnlyFlag})
											}

											if exitStatus == 0 {
												u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "general")
												_, exitStatus = getServerGeneralConfigurations(u.String(), token, printOptions)
											}
										}
									}
								}
								logout(baseURI, token)
							} else if detectHostUnreachable(exitStatus) {
								exitStatus = 10502
							}
						}
					}
				case "serverprefs":
					if len(cmdArgs[2:]) > 0 {
						for i := 0; i < len(cmdArgs[2:]); i++ {
							if regexp.MustCompile(`(.*)=(.*)`).Match([]byte(cmdArgs[2:][i])) {
								rep := regexp.MustCompile(`(.*)=(.*)`)
								option := rep.ReplaceAllString(cmdArgs[2:][i], "$1")
								switch strings.ToLower(option) {
								case "cachesize", "maxfiles", "maxguests", "allowpsos", "startuprestorationenabled", "requiresecuredb":
								case "authenticatedstream", "parallelbackupenabled":
								//case "persistcacheenabled", "syncpersistcache":
								default:
									exitStatus = 3
								}

								if exitStatus == 3 {
									break
								}
							} else {
								exitStatus = 10001
								break
							}
						}
					} else {
						exitStatus = 10001
					}

					if exitStatus == 0 {
						token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
						if token != "" && exitStatus == 0 && err == nil {
							var versionString string
							var version float64

							if !usingCloud {
								u.Path = path.Join(getAPIBasePath(baseURI), "server", "metadata")
								versionString, _ = getServerVersionString(u.String(), token)
								version, _ = getServerVersionAsFloat(versionString)
							}

							var results []string
							var settings []int
							var settingResults []int
							printOptions := []string{}

							if usingCloud {
								// for Claris FileMaker Cloud
								if len(cmdArgs[2:]) > 0 {
									for i := 0; i < len(cmdArgs[2:]); i++ {
										if regexp.MustCompile(`(.*)=(.*)`).Match([]byte(cmdArgs[2:][i])) {
											rep := regexp.MustCompile(`(.*)=(.*)`)
											option := rep.ReplaceAllString(cmdArgs[2:][i], "$1")
											switch strings.ToLower(option) {
											case "authenticatedstream":
												printOptions = append(printOptions, "authenticatedstream")
											default:
												exitStatus = 3
											}
											if exitStatus != 0 {
												break
											}
										}
									}
								} else {
									printOptions = append(printOptions, "authenticatedstream")
								}

								results, exitStatus = parseServerConfigurationSettings(c, cmdArgs[2:])

								authenticatedStream, _ := strconv.Atoi(results[6])
								if results[6] != "" {
									if authenticatedStream < 1 || authenticatedStream > 2 {
										exitStatus = 10001
									}
								}
								if exitStatus == 0 {
									u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "authenticatedstream")
									exitStatus, _, _ = sendRequest("PATCH", u.String(), token, params{command: "set", authenticatedstream: authenticatedStream})
									if exitStatus != 0 {
										exitStatus = 10001
									} else {
										_, exitStatus, _ = getAuthenticatedStreamSetting(u.String(), token, printOptions)
									}
								}
							} else {
								// for Claris FileMaker Server
								u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "general")
								settings, exitStatus = getServerGeneralConfigurations(u.String(), token, printOptions)
								if exitStatus == 0 {
									var results []string
									results, exitStatus = parseServerConfigurationSettings(c, cmdArgs[2:])

									cacheSize, _ := strconv.Atoi(results[0])
									maxFiles, _ := strconv.Atoi(results[1])
									maxProConnections, _ := strconv.Atoi(results[2])
									maxPSOS, _ := strconv.Atoi(results[3])
									startupRestorationEnabled := false
									if results[4] == "true" {
										startupRestorationEnabled = true
									}
									secureFilesOnlyFlag := results[5]
									authenticatedStream, _ := strconv.Atoi(results[6])
									parallelBackupEnabled := false
									if results[7] == "true" {
										parallelBackupEnabled = true
									}
									/*
										persistCacheEnabled := false
										if results[8] == "true" {
											persistCacheEnabled = true
										}

										syncPersistCacheEnabled := false
										if results[9] == "true" {
											syncPersistCacheEnabled = true
										}
									*/

									if results[0] != "" || results[1] != "" || results[2] != "" || results[3] != "" || results[4] != "" || secureFilesOnlyFlag != "" || results[6] != "" || results[7] != "" {
										if results[0] != "" {
											if cacheSize < 64 || cacheSize > 1048576 {
												exitStatus = 10001
											}
										}

										if results[1] != "" {
											if maxFiles < 1 || maxFiles > 125 {
												exitStatus = 10001
											}
										}

										if results[2] != "" {
											if maxProConnections < 0 || maxProConnections > 2000 {
												exitStatus = 10001
											}
										}

										if results[3] != "" {
											if maxPSOS < 0 || maxPSOS > 500 {
												exitStatus = 10001
											}
										}

										startupRestorationBuiltin := true
										if settings[4] == -1 {
											// for Claris FileMaker Server 19.1.2 or later
											startupRestorationBuiltin = false
										}

										if results[6] != "" {
											// for Claris FileMaker Server 19.3.2 or later
											if authenticatedStream < 1 || authenticatedStream > 2 {
												exitStatus = 10001
											}
										}

										printOptions = []string{}
										if len(cmdArgs[2:]) > 0 {
											for i := 0; i < len(cmdArgs[2:]); i++ {
												if regexp.MustCompile(`(.*)=(.*)`).Match([]byte(cmdArgs[2:][i])) {
													rep := regexp.MustCompile(`(.*)=(.*)`)
													option := rep.ReplaceAllString(cmdArgs[2:][i], "$1")
													switch strings.ToLower(option) {
													case "cachesize":
														printOptions = append(printOptions, "cachesize")
													case "maxfiles":
														printOptions = append(printOptions, "maxfiles")
													case "maxguests":
														printOptions = append(printOptions, "maxguests")
													case "allowpsos":
														printOptions = append(printOptions, "allowpsos")
													case "startuprestorationenabled":
														printOptions = append(printOptions, "startuprestorationenabled")
													case "requiresecuredb":
														printOptions = append(printOptions, "requiresecuredb")
													case "authenticatedstream":
														if version >= 19.3 && !strings.HasPrefix(versionString, "19.3.1") {
															printOptions = append(printOptions, "authenticatedstream")
														} else {
															exitStatus = 3
														}
													case "parallelbackupenabled":
														if version >= 19.5 {
															printOptions = append(printOptions, "parallelbackupenabled")
														} else {
															exitStatus = 3
														}
													/*
														case "persistcacheenabled":
															if version >= 20.1 {
																printOptions = append(printOptions, "persistcacheenabled")
															} else {
																exitStatus = 3
															}

														case "syncpersistcache":
															if version >= 20.1 {
																printOptions = append(printOptions, "syncpersistcache")
															} else {
																exitStatus = 3
															}
													*/
													default:
														exitStatus = 3
													}
													if exitStatus != 0 {
														break
													}
												}
											}
										} else {
											printOptions = append(printOptions, "cachesize")
											printOptions = append(printOptions, "maxfiles")
											printOptions = append(printOptions, "maxguests")
											printOptions = append(printOptions, "allowpsos")
											printOptions = append(printOptions, "startuprestorationenabled")
											printOptions = append(printOptions, "requiresecuredb")
											if version >= 19.3 && !strings.HasPrefix(versionString, "19.3.1") {
												printOptions = append(printOptions, "authenticatedstream")
											}
											if version >= 19.5 {
												printOptions = append(printOptions, "parallelbackupenabled")
											}
											/*
												if version >= 20.1 {
													printOptions = append(printOptions, "persistcacheenabled")
													printOptions = append(printOptions, "syncpersistcache")
												}
											*/
										}
										if exitStatus == 0 {
											if results[0] != "" || results[1] != "" || results[2] != "" || results[3] != "" || results[4] != "" {
												u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "general")
												exitStatus, _, _ = sendRequest("PATCH", u.String(), token, params{
													command:                   "set",
													cachesize:                 cacheSize,
													maxfiles:                  maxFiles,
													maxproconnections:         maxProConnections,
													maxpsos:                   maxPSOS,
													startuprestorationenabled: startupRestorationEnabled,
													startuprestorationbuiltin: startupRestorationBuiltin,
												})
											}

											if exitStatus == 0 && (secureFilesOnlyFlag == "true" || secureFilesOnlyFlag == "false") {
												u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "security")
												exitStatus, _, _ = sendRequest("PATCH", u.String(), token, params{command: "set", requiresecuredb: secureFilesOnlyFlag})
											}

											if exitStatus == 0 {
												if results[6] != "" {
													if version >= 19.3 && !strings.HasPrefix(versionString, "19.3.1") {
														// for Claris FileMaker Server 19.3.2 or later
														u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "authenticatedstream")
														exitStatus, _, _ = sendRequest("PATCH", u.String(), token, params{command: "set", authenticatedstream: authenticatedStream})
														if exitStatus != 0 {
															exitStatus = 10001
														}
													} else {
														exitStatus = 3
													}
												}

												if results[7] != "" {
													// for Claris FileMaker Server 19.5.1 or later
													if version >= 19.5 {
														u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "parallelbackup")
														exitStatus, _, _ = sendRequest("PATCH", u.String(), token, params{command: "set", parallelbackupenabled: parallelBackupEnabled})
														if exitStatus != 0 {
															exitStatus = 10001
														}
													} else {
														exitStatus = 3
													}
												}

												u.Path = path.Join(getAPIBasePath(baseURI), "server", "config", "general")
												settingResults, exitStatus = getServerGeneralConfigurations(u.String(), token, printOptions)
												if startupRestorationBuiltin && settings[4] != settingResults[4] {
													// check setting of startupRestorationEnabled
													fmt.Println("Restart the FileMaker Server background processes to apply the change.")
												}
											}
										}
									}
								}
							}

							logout(baseURI, token)
						} else if detectHostUnreachable(exitStatus) {
							exitStatus = 10502
						}
					}
				default:
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "start":
			if usingCloud {
				exitStatus = 21
			} else {
				if len(cmdArgs[1:]) > 0 {
					switch strings.ToLower(cmdArgs[1]) {
					case "server":
						token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
						if token != "" && exitStatus == 0 && err == nil {
							var running string
							u.Path = path.Join(getAPIBasePath(baseURI), "server", "status")
							_, running, _ = sendRequest("GET", u.String(), token, params{})
							if running == "RUNNING" {
								// Service already running
								exitStatus = 10006
							} else {
								// start database server
								exitStatus, _, _ = sendRequest("PATCH", u.String(), token, params{status: "RUNNING"})
							}
							logout(baseURI, token)
						} else if detectHostUnreachable(exitStatus) {
							exitStatus = 10502
						}
					default:
						exitStatus = outputInvalidCommandParameterErrorMessage(c)
					}
				} else {
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			}
		case "status":
			if len(cmdArgs[1:]) > 0 {
				switch strings.ToLower(cmdArgs[1]) {
				case "client":
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
					if token != "" && exitStatus == 0 && err == nil {
						id := 0
						if len(cmdArgs) >= 3 {
							sid, err := strconv.Atoi(cmdArgs[2])
							if err == nil {
								id = sid
							}
						}
						if id > 0 {
							u.Path = path.Join(getAPIBasePath(baseURI), "clients")
							exitStatus = listClients(u.String(), token, id)
						}
						logout(baseURI, token)
					} else if detectHostUnreachable(exitStatus) {
						exitStatus = 10502
					}
				case "file":
					token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
					if token != "" && exitStatus == 0 && err == nil {
						if len(cmdArgs[2:]) > 0 {
							u.Path = path.Join(getAPIBasePath(baseURI), "databases")
							idList, _, _ := getDatabases(u.String(), token, cmdArgs[2:], "", false)
							if len(idList) > 0 {
								exitStatus = listFiles(c, u.String(), token, idList)
							}
						} else {
							exitStatus = 10001
						}
						logout(baseURI, token)
					} else if detectHostUnreachable(exitStatus) {
						exitStatus = 10502
					}
				default:
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		case "stop":
			if usingCloud {
				exitStatus = 21
			} else {
				if len(cmdArgs[1:]) > 0 {
					res := ""
					if yesFlag {
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
							token, exitStatus, err = login(baseURI, username, password, params{retry: retry, identityFile: identityFile})
							if token != "" && exitStatus == 0 && err == nil {
								message = "Stopping FileMaker Database Engine..."
								// message = "FileMaker データベースエンジンの停止中..."
								if forceFlag {
									graceTime = 0
								}
								exitStatus, _ = stopDatabaseServer(u, baseURI, token, message, graceTime)
								if exitStatus == 0 {
									exitStatus, _ = waitStoppingServer(u, baseURI, token)
								}
								logout(baseURI, token)
							} else if detectHostUnreachable(exitStatus) {
								exitStatus = 10502
							}
						default:
							exitStatus = outputInvalidCommandParameterErrorMessage(c)
						}
					}
				} else {
					exitStatus = outputInvalidCommandErrorMessage(c)
				}
			}
		default:
			if helpFlag {
				fmt.Fprint(c.outStream, helpTextTemplate)
			} else {
				exitStatus = outputInvalidCommandErrorMessage(c)
			}
		}
	} else {
		if versionFlag {
			fmt.Fprintln(c.outStream, "fmcsadmin "+version)
		} else {
			fmt.Fprint(c.outStream, helpTextTemplate)
		}
	}

	if exitStatus != 0 && exitStatus != 23 && exitStatus != 248 && exitStatus != 249 {
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
	forceFlag := false
	saveKeyFlag := false
	fqdn := ""
	hostname := ""
	username := ""
	password := ""
	key := ""
	message := ""
	keyFile := ""
	keyFilePass := ""
	intermediateCA := ""
	clientID := -1
	graceTime := 90
	identityFile := ""

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
	flags.StringVar(&hostname, "host", "", "specify your host name of FileMaker Cloud")
	flags.StringVar(&username, "u", "", "Username to use to authenticate with the server.")
	flags.StringVar(&username, "username", "", "Username to use to authenticate with the server.")
	flags.StringVar(&password, "p", "", "Password to use to authenticate with the server.")
	flags.StringVar(&password, "password", "", "Password to use to authenticate with the server.")
	flags.StringVar(&key, "key", "", "Specify the database encryption password.")
	flags.StringVar(&message, "m", "", "Specify a text message to send to clients.")
	flags.StringVar(&message, "message", "", "Specify a text message to send to clients.")
	flags.StringVar(&keyFile, "keyfile", "", "Specify private key file for certificate import.")
	flags.StringVar(&keyFile, "KeyFile", "", "Specify private key file for certificate import.")
	flags.StringVar(&keyFilePass, "keyfilepass", "", "Specify password needed to read KEYFILE.")
	flags.StringVar(&keyFilePass, "KeyFilePass", "", "Specify password needed to read KEYFILE.")
	flags.StringVar(&intermediateCA, "intermediateca", "", "Specify the file that contains the intermediate CA certificate(s) for certificate import.")
	flags.StringVar(&intermediateCA, "intermediateCA", "", "Specify the file that contains the intermediate CA certificate(s) for certificate import.")
	flags.IntVar(&clientID, "c", -1, "Specify a client number to send a message.")
	flags.IntVar(&clientID, "client", -1, "Specify a client number to send a message.")
	flags.BoolVar(&forceFlag, "f", false, "Force database to close or Database Server to stop, immediately disconnecting clients.")
	flags.BoolVar(&forceFlag, "force", false, "Force database to close or Database Server to stop, immediately disconnecting clients.")
	flags.BoolVar(&saveKeyFlag, "savekey", false, "Save the database encryption password.")
	flags.IntVar(&graceTime, "t", 90, "Specify time in seconds before client is forced to disconnect.")
	flags.IntVar(&graceTime, "gracetime", 90, "Specify time in seconds before client is forced to disconnect.")
	flags.StringVar(&identityFile, "i", "", "Specify a private key file for FileMaker Admin API PKI Authentication.")

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
	cFlags.forceFlag = cFlags.forceFlag || forceFlag
	cFlags.saveKeyFlag = cFlags.saveKeyFlag || saveKeyFlag
	if cFlags.fqdn == "" {
		cFlags.fqdn = fqdn
	}
	if cFlags.hostname == "" {
		cFlags.hostname = hostname
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
	if cFlags.keyFile == "" {
		cFlags.keyFile = keyFile
	}
	if cFlags.keyFilePass == "" {
		cFlags.keyFilePass = keyFilePass
	}
	if cFlags.intermediateCA == "" {
		cFlags.intermediateCA = intermediateCA
	}
	if cFlags.clientID == -1 {
		cFlags.clientID = clientID
	}
	if cFlags.graceTime == 90 {
		cFlags.graceTime = graceTime
	}
	if cFlags.identityFile == "" {
		cFlags.identityFile = identityFile
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
		cFlags.forceFlag = cFlags.forceFlag || subCommandOptions.forceFlag
		cFlags.saveKeyFlag = cFlags.saveKeyFlag || subCommandOptions.saveKeyFlag
		if cFlags.fqdn == "" {
			cFlags.fqdn = subCommandOptions.fqdn
		}
		if cFlags.hostname == "" {
			cFlags.hostname = subCommandOptions.hostname
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
		if cFlags.keyFile == "" {
			cFlags.keyFile = subCommandOptions.keyFile
		}
		if cFlags.keyFilePass == "" {
			cFlags.keyFilePass = subCommandOptions.keyFilePass
		}
		if cFlags.intermediateCA == "" {
			cFlags.intermediateCA = subCommandOptions.intermediateCA
		}
		if cFlags.clientID == -1 {
			cFlags.clientID = subCommandOptions.clientID
		}
		if cFlags.graceTime == 90 {
			cFlags.graceTime = subCommandOptions.graceTime
		}
		if cFlags.identityFile == "" {
			cFlags.identityFile = subCommandOptions.identityFile
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
	startupRestorationEnabled := ""
	secureFilesOnlyFlag := ""
	authenticatedStream := ""
	parallelBackupEnabled := ""

	for i := 0; i < len(str); i++ {
		val := strings.ToLower(str[i])
		if regexp.MustCompile(`cachesize=(\d+)`).Match([]byte(val)) {
			rep := regexp.MustCompile(`cachesize=(\d+)`)
			cacheSize = rep.ReplaceAllString(val, "$1")
		} else if regexp.MustCompile(`hostedfiles=(\d+)`).Match([]byte(val)) {
			rep := regexp.MustCompile(`hostedfiles=(\d+)`)
			maxFiles = rep.ReplaceAllString(val, "$1")
		} else if regexp.MustCompile(`maxfiles=(\d+)`).Match([]byte(val)) {
			rep := regexp.MustCompile(`maxfiles=(\d+)`)
			maxFiles = rep.ReplaceAllString(val, "$1")
		} else if regexp.MustCompile(`maxguests=(\d+)`).Match([]byte(val)) {
			rep := regexp.MustCompile(`maxguests=(\d+)`)
			maxProConnections = rep.ReplaceAllString(val, "$1")
		} else if regexp.MustCompile(`proconnections=(\d+)`).Match([]byte(val)) {
			rep := regexp.MustCompile(`proconnections=(\d+)`)
			maxProConnections = rep.ReplaceAllString(val, "$1")
		} else if regexp.MustCompile(`allowpsos=(\d+)`).Match([]byte(val)) {
			rep := regexp.MustCompile(`allowpsos=(\d+)`)
			maxPSOS = rep.ReplaceAllString(val, "$1")
		} else if regexp.MustCompile(`scriptsessions=(\d+)`).Match([]byte(val)) {
			rep := regexp.MustCompile(`scriptsessions=(\d+)`)
			maxPSOS = rep.ReplaceAllString(val, "$1")
		} else if regexp.MustCompile(`startuprestorationenabled=(.*)`).Match([]byte(val)) {
			if strings.ToLower(str[i]) == "startuprestorationenabled=" {
				exitStatus = 10001
			} else if strings.ToLower(str[i]) == "startuprestorationenabled=true" || (regexp.MustCompile(`startuprestorationenabled=([+|-])?(\d)+`).Match([]byte(str[i])) && str[i] != "startuprestorationenabled=0" && str[i] != "startuprestorationenabled=+0" && str[i] != "startuprestorationenabled=-0") {
				startupRestorationEnabled = "true"
			} else {
				startupRestorationEnabled = "false"
			}
		} else if regexp.MustCompile(`requiresecuredb=(.*)`).Match([]byte(val)) {
			if strings.ToLower(str[i]) == "requiresecuredb=" {
				exitStatus = 10001
			} else if strings.ToLower(str[i]) == "requiresecuredb=true" || (regexp.MustCompile(`requiresecuredb=([+|-])?(\d)+`).Match([]byte(val)) && val != "requiresecuredb=0" && val != "requiresecuredb=+0" && val != "requiresecuredb=-0") {
				secureFilesOnlyFlag = "true"
			} else {
				secureFilesOnlyFlag = "false"
			}
		} else if regexp.MustCompile(`securefilesonly=(.*)`).Match([]byte(val)) {
			if strings.ToLower(str[i]) == "securefilesonly=" {
				exitStatus = 10001
			} else if strings.ToLower(str[i]) == "securefilesonly=true" || (regexp.MustCompile(`securefilesonly=([+|-])?(\d)+`).Match([]byte(val)) && val != "securefilesonly=0" && val != "securefilesonly=+0" && val != "securefilesonly=-0") {
				secureFilesOnlyFlag = "true"
			} else {
				secureFilesOnlyFlag = "false"
			}
		} else if regexp.MustCompile(`authenticatedstream=(\d+)`).Match([]byte(val)) {
			rep := regexp.MustCompile(`authenticatedstream=(\d+)`)
			authenticatedStream = rep.ReplaceAllString(val, "$1")
		} else if regexp.MustCompile(`parallelbackupenabled=(.*)`).Match([]byte(val)) {
			if strings.ToLower(str[i]) == "parallelbackupenabled=" {
				exitStatus = 10001
			} else if strings.ToLower(str[i]) == "parallelbackupenabled=true" || (regexp.MustCompile(`parallelbackupenabled=([+|-])?(\d)+`).Match([]byte(str[i])) && str[i] != "parallelbackupenabled=0" && str[i] != "parallelbackupenabled=+0" && str[i] != "parallelbackupenabled=-0") {
				parallelBackupEnabled = "true"
			} else {
				parallelBackupEnabled = "false"
			}
			/*
				} else if regexp.MustCompile(`persistcacheenabled=(.*)`).Match([]byte(val)) {
					if strings.ToLower(str[i]) == "persistcacheenabled=" {
						exitStatus = 10001
					} else if strings.ToLower(str[i]) == "persistcacheenabled=true" || (regexp.MustCompile(`persistcacheenabled=([+|-])?(\d)+`).Match([]byte(str[i])) && str[i] != "persistcacheenabled=0" && str[i] != "persistcacheenabled=+0" && str[i] != "persistcacheenabled=-0") {
						persistCacheEnabled = "true"
					} else {
						persistCacheEnabled = "false"
					}
			*/
		} else {
			exitStatus = 10001
		}
	}

	results = append(results, cacheSize)
	results = append(results, maxFiles)
	results = append(results, maxProConnections)
	results = append(results, maxPSOS)
	results = append(results, startupRestorationEnabled)
	results = append(results, secureFilesOnlyFlag)
	results = append(results, authenticatedStream)
	results = append(results, parallelBackupEnabled)

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
		val := strings.ToLower(str[i])
		if regexp.MustCompile(`enablephp=(.*)`).Match([]byte(val)) {
			if strings.ToLower(str[i]) == "enablephp=" {
				exitStatus = 10001
			} else if strings.ToLower(str[i]) == "enablephp=true" || (regexp.MustCompile(`enablephp=([+|-])?(\d)+`).Match([]byte(val)) && val != "enablephp=0" && val != "enablephp=+0" && val != "enablephp=-0") {
				phpFlag = "true"
			} else {
				phpFlag = "false"
			}
		} else if regexp.MustCompile(`enablexml=(.*)`).Match([]byte(val)) {
			if strings.ToLower(str[i]) == "enablexml=" {
				exitStatus = 10001
			} else if strings.ToLower(str[i]) == "enablexml=true" || (regexp.MustCompile(`enablexml=([+|-])?(\d)+`).Match([]byte(val)) && val != "enablexml=0" && val != "enablexml=+0" && val != "enablexml=-0") {
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
		} else if regexp.MustCompile(`prevalidation=(.*)`).Match([]byte(val)) {
			if strings.ToLower(str[i]) == "prevalidation=" {
				exitStatus = 10001
			} else if strings.ToLower(str[i]) == "prevalidation=true" || (regexp.MustCompile(`prevalidation=([+|-])?(\d)+`).Match([]byte(val)) && val != "prevalidation=0" && val != "prevalidation=+0" && val != "prevalidation=-0") {
				preValidationFlag = "true"
			} else {
				preValidationFlag = "false"
			}
		} else if regexp.MustCompile(`usefmphp=(.*)`).Match([]byte(val)) {
			if strings.ToLower(str[i]) == "usefmphp=" {
				exitStatus = 10001
			} else if strings.ToLower(str[i]) == "usefmphp=true" || (regexp.MustCompile(`usefmphp=([+|-])?(\d)+`).Match([]byte(val)) && val != "usefmphp=0" && val != "usefmphp=+0" && val != "usefmphp=-0") {
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

func getBaseURI(fqdn string) string {
	baseURI := "http://127.0.0.1:16001"
	if len(fqdn) > 0 {
		baseURI = "https://" + strings.TrimSpace(fqdn)
	}

	return baseURI
}

func getAPIBasePath(baseURI string) string {
	path := "/fmi/admin/api/v2"

	return path
}

func getUsernameAndPassword(username string, password string, product int) (string, string) {
	if len(username) == 0 {
		if product == 1 {
			username = os.Getenv("FMS_USERNAME")
		} else {
			username = os.Getenv("FMC_USERNAME")
		}
		if len(username) == 0 {
			r := bufio.NewReader(os.Stdin)
			fmt.Print("username: ")
			input, _ := r.ReadString('\n')
			username = strings.TrimSpace(input)
		}
	}

	if len(password) == 0 {
		if product == 1 {
			password = os.Getenv("FMS_PASSWORD")
		} else {
			password = os.Getenv("FMC_PASSWORD")
		}
		if len(password) == 0 {
			fmt.Print("password: ")
			bytePassword, _ := term.ReadPassword(int(syscall.Stdin))
			password = string(bytePassword)
			fmt.Printf("\n")
		}
	}

	return username, password
}

func login(baseURI string, user string, pass string, p params) (string, int, error) {
	var body []byte
	var err error
	token := ""
	exitStatus := 0

	if regexp.MustCompile(`https://(.*).account.filemaker-cloud.com`).Match([]byte(baseURI)) {
		// for Claris FileMaker Cloud
		exitStatus = 21
		err = fmt.Errorf("%s", "Not Supported")
	} else {
		// for Claris FileMaker Server
		username := user
		password := pass
		if p.identityFile == "" {
			username, password = getUsernameAndPassword(user, pass, 1)
		}

		u, _ := url.Parse(baseURI)
		u.Path = path.Join(getAPIBasePath(baseURI), "user", "auth")

		output := output{}
		if p.identityFile == "" {
			body, _, err = callURL("POST", u.String(), "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)), nil)
		} else {
			var jwtToken string
			jwtToken, exitStatus, err = getJWTToken(p.identityFile)
			if err != nil || exitStatus > 0 {
				return token, exitStatus, err
			}
			body, _, _ = callURL("POST", u.String(), "PKI "+jwtToken, nil)
		}

		/* for debugging */
		//fmt.Println(bytes.NewBuffer([]byte(body)))

		if err != nil {
			exitStatus = 10502
			return token, exitStatus, err
		}

		err = json.Unmarshal(body, &output)
		if err != nil {
			return token, exitStatus, err
		}

		code := ""
		for i := 0; i < len(output.Messages); i++ {
			if reflect.ValueOf(output.Messages[i].Code).IsValid() {
				code = output.Messages[i].Code
				break
			}
		}
		if code == "0" {
			token = output.Response.Token
		} else {
			if p.retry > 0 {
				fmt.Println("fmcsadmin: Permission denied, please try again.")
				token, exitStatus, err = login(baseURI, user, pass, params{retry: p.retry - 1, identityFile: p.identityFile})
				if err != nil {
					exitStatus = 10502
					return token, exitStatus, err
				}
			} else {
				fmt.Println("fmcsadmin: Permission denied.")
				exitStatus = 9
			}
		}
	}

	return token, exitStatus, err
}

func getJWTToken(filePath string) (string, int, error) {
	// for public key infrastructure (PKI) authentication
	var err error
	var pkey *rsa.PrivateKey
	passphrase := ""

	keyData, keyFormat, exitStatus := detectPrivateKeyFormat(filePath, "")
	if exitStatus != 0 && exitStatus != 212 {
		return "", exitStatus, nil
	}

	if exitStatus == 212 {
		fmt.Print("Enter passphrase: ")
		bytePassphrase, _ := term.ReadPassword(int(syscall.Stdin))
		passphrase = string(bytePassphrase)
		fmt.Printf("\n")
		keyData, _, exitStatus = detectPrivateKeyFormat(filePath, passphrase)
	} else if keyFormat == "PRIVATE KEY" || keyFormat == "EC PRIVATE KEY" || keyFormat == "EC PARAMETERS" {
		exitStatus = 21
		return "", exitStatus, nil
	}

	// Name of public key on FileMaker Server Admin Console
	keyName := strings.Replace(filepath.Base(filePath), filepath.Ext(filePath), "", 1)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": strings.Replace(keyName, "_", " ", -1),
		"aud": "fmsadminapi",
		"exp": time.Now().Add(time.Minute * 15).Unix(),
	})

	if passphrase == "" {
		pkey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(keyData))
		if err != nil {
			exitStatus = 20408
			return "", exitStatus, err
		}
	} else {
		pkey, err = jwt.ParseRSAPrivateKeyFromPEMWithPassword([]byte(keyData), passphrase)
		if err != nil {
			exitStatus = 20408
			return "", exitStatus, err
		}
	}
	tokenString, _ := jwtToken.SignedString(pkey)

	return tokenString, exitStatus, err
}

func detectPrivateKeyFormat(filePath string, keyFilePass string) ([]byte, string, int) {
	keyType := ""
	exitStatus := 0

	_, err := os.Stat(filePath)
	if err != nil {
		exitStatus = 20405
	}

	keyData, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsPermission(err) {
			exitStatus = 20402
		} else {
			exitStatus = 20405
		}
	} else {
		block, _ := pem.Decode(keyData)
		if block == nil {
			exitStatus = 20408
		} else {
			buf := block.Bytes
			if x509.IsEncryptedPEMBlock(block) {
				buf, err = x509.DecryptPEMBlock(block, []byte(keyFilePass))
				if err != nil {
					if err == x509.IncorrectPasswordError {
						exitStatus = 212
					}
				}
			}

			if exitStatus == 0 {
				switch block.Type {
				case "RSA PRIVATE KEY":
					_, err = x509.ParsePKCS1PrivateKey(buf)
					if err != nil {
						exitStatus = 20408
					}
				case "PRIVATE KEY":
					_, err := x509.ParsePKCS8PrivateKey(buf)
					if err != nil {
						exitStatus = 20408
					}
				case "EC PRIVATE KEY":
					_, err := x509.ParseECPrivateKey(buf)
					if err != nil {
						exitStatus = 20408
					}
				default:
					exitStatus = 20408
				}
				keyType = block.Type
			}
		}
	}

	return keyData, keyType, exitStatus
}

func logout(baseURI string, token string) {
	u, _ := url.Parse(baseURI)
	u.Path = path.Join(getAPIBasePath(baseURI), "user", "auth", token)
	sendRequest("DELETE", u.String(), token, params{})
}

func getResultCode(v interface{}) int {
	var resultCode string

	err := scan.ScanTree(v, "/messages[0]/code", &resultCode)
	if err != nil {
		return -1
	}
	code, _ := strconv.Atoi(resultCode)

	return code
}

func listClients(urlString string, token string, id int) int {
	usingCloud := false
	if regexp.MustCompile(`https://(.*).account.filemaker-cloud.com`).Match([]byte(urlString)) {
		usingCloud = true
	}

	body, _, err := callURL("GET", urlString, token, nil)
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

	result := getResultCode(v)
	if result == 1701 {
		// when fmserverd is stopping
		return 10502
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
	//var teamLicensed string
	var fileName string
	var accountName string
	var privsetName string
	//var b bool
	var data [][]string
	var sID int

	_ = scan.ScanTree(v, "/response/clients", &c)
	count = len(c)

	if mode == "NORMAL" {
		if count > 0 {
			for i := 0; i < count; i++ {
				_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/status", &s)
				if s == "NORMAL" {
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/id", &s1)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/userName", &userName)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/computerName", &computerName)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/extpriv", &extPriv)
					data = append(data, []string{s1, userName, computerName, extPriv})
				}
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Client ID", "User Name", "Computer Name", "Ext Privilege"})
			table.SetAutoWrapText(false)
			table.SetAutoFormatHeaders(false)
			for _, v := range data {
				table.Append(v)
			}
			table.Render()
		}
	} else {
		if count > 0 {
			for i := 0; i < count; i++ {
				_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/status", &s)
				if s == "NORMAL" {
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/id", &s1)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/userName", &userName)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/computerName", &computerName)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/extpriv", &extPriv)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/ipaddress", &ipAddress)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/macaddress", &macAddress)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/connectTime", &connectTime)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/connectDuration", &connectDuration)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/appVersion", &appVersion)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/appLanguage", &appLanguage)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/guestFiles[0]/filename", &fileName)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/guestFiles[0]/accountName", &accountName)
					_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(i)+"]/guestFiles[0]/privsetName", &privsetName)

					connectTime = getDateTimeStringOfCurrentTimeZone(connectTime, "2006/01/02 15:04:05", usingCloud)
					if regexp.MustCompile(`(.*)\.fmp12`).Match([]byte(fileName)) {
						rep := regexp.MustCompile(`(.*)\.fmp12`)
						fileName = rep.ReplaceAllString(fileName, "$1")
					}

					data = append(data, []string{s1, userName, computerName, extPriv, ipAddress, macAddress, connectTime, connectDuration, appVersion, appLanguage, fileName, accountName, privsetName})
				}
			}

			sID, _ = strconv.Atoi(s1)
			if id == sID || id == 0 {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Client ID", "User Name", "Computer Name", "Ext Privilege", "IP Address", "MAC Address", "Connect Time", "Duration", "App Version", "App Language", "File Name", "Account Name", "Privilege Set"})
				table.SetAutoWrapText(false)
				table.SetAutoFormatHeaders(false)
				for _, v := range data {
					table.Append(v)
				}
				table.Render()
			}
		}
	}

	return 0
}

func listFiles(c *cli, url string, token string, idList []int) int {
	body, _, err := callURL("GET", url, token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return -1
	}

	/* for debugging */
	//fmt.Println(bytes.NewBuffer([]byte(body)))

	var v interface{}
	body2 := []byte(body)
	err = json.Unmarshal(body2, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	result := getResultCode(v)
	if result == 1701 {
		// when fmserverd is stopping
		return 10502
	}

	mode := "NORMAL"
	if (len(idList) == 1 && idList[0] > -1) || len(idList) > 1 {
		mode = "DETAIL"
	}

	var totalDbCount int
	var count []string
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

	_ = scan.ScanTree(v, "/response/totalDBCount", &totalDbCount)

	if mode == "NORMAL" {
		for i := 0; i < totalDbCount; i++ {
			_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/status", &s)
			if s == "NORMAL" {
				_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/folder", &s)
				fmt.Fprint(c.outStream, s)
				_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/filename", &s)
				fmt.Fprintln(c.outStream, s)
			}
		}
	} else {
		for i := 0; i < totalDbCount; i++ {
			_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/id", &s1)
			for j := 0; j < len(idList); j++ {
				if s1 == strconv.Itoa(idList[j]) || idList[j] == 0 {
					_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/filename", &fileName)
					_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/clients", &num1)
					_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/size", &num2)
					_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/status", &status)

					if status == "CLOSED" {
						extPriv = "-"
					} else {
						extPriv = ""
						_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/enabledExtPrivileges", &count)
						for j := 0; j < len(count); j++ {
							_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/enabledExtPrivileges["+strconv.Itoa(j)+"]", &s)
							if extPriv == "" {
								extPriv = s
							} else {
								extPriv = extPriv + " " + s
							}
						}
					}

					isEncrypted = ""
					_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/isEncrypted", &b)
					if b {
						isEncrypted = "Yes"
					} else {
						isEncrypted = "No"
					}

					data = append(data, []string{s1, fileName, strconv.Itoa(num1), strconv.Itoa(num2), status[:1] + strings.ToLower(status[1:]), extPriv, isEncrypted})
				}
			}
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "File", "Clients", "Size", "Status", "Enabled Extended Privileges", "Encrypted"})
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(false)
		for _, v := range data {
			table.Append(v)
		}
		table.Render()
	}

	return 0
}

func getServerVersion(url string, token string) float64 {
	versionString, err := getServerVersionString(url, token)
	if err != nil {
		return 0.0
	}

	version, err := getServerVersionAsFloat(versionString)
	if err != nil {
		return 0.0
	}

	return version
}

func getServerVersionString(urlString string, token string) (string, error) {
	body, _, err := callURL("GET", urlString, token, nil)
	if err != nil {
		return "0.0.0", err
	}

	/* for debugging */
	//fmt.Println(bytes.NewBuffer([]byte(body)))

	var v interface{}
	body2 := []byte(body)
	err = json.Unmarshal(body2, &v)
	if err != nil {
		return "0.0.0", err
	}

	var versionString string
	err = scan.ScanTree(v, "/response/ServerVersion", &versionString)
	if err != nil {
		return "0.0.0", err
	}

	return versionString, err
}

func getServerVersionAsFloat(versionString string) (float64, error) {
	version, err := strconv.ParseFloat(strings.Join(strings.Split(versionString, ".")[0:2], "."), 64)
	if err != nil {
		return 0.0, err
	}

	return version, err
}

func listPlugins(url string, token string) int {
	body, _, err := callURL("GET", url, token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return -1
	}

	/* for debugging */
	//fmt.Println(bytes.NewBuffer([]byte(body)))

	var v interface{}
	body2 := []byte(body)
	err = json.Unmarshal(body2, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	result := getResultCode(v)
	if result == 1701 {
		// when fmserverd is stopping
		return 10502
	}

	var count int
	var c []string
	var s1 string
	var pluginName string
	var fileName string
	var enabled bool
	var status string
	var data [][]string

	_ = scan.ScanTree(v, "/response/plugins", &c)
	count = len(c)

	if count > 0 {
		for i := 0; i < count; i++ {
			_ = scan.ScanTree(v, "/response/plugins["+strconv.Itoa(i)+"]/id", &s1)
			_ = scan.ScanTree(v, "/response/plugins["+strconv.Itoa(i)+"]/pluginName", &pluginName)
			_ = scan.ScanTree(v, "/response/plugins["+strconv.Itoa(i)+"]/filename", &fileName)
			_ = scan.ScanTree(v, "/response/plugins["+strconv.Itoa(i)+"]/enabled", &enabled)
			status = "Disabled"
			if enabled {
				status = "Enabled"
			}
			data = append(data, []string{s1, pluginName, fileName, status})
		}

		if len(data) > 0 {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoWrapText(false)
			table.SetAutoFormatHeaders(false)
			for _, v := range data {
				table.SetHeader([]string{"ID", "Name", "File", "Status"})
				table.Append(v)
			}
			table.Render()
		}
	}

	return 0
}

func listSchedules(urlString string, token string, id int) int {
	usingCloud := false
	if regexp.MustCompile(`https://(.*).account.filemaker-cloud.com`).Match([]byte(urlString)) {
		usingCloud = true
	}

	body, _, err := callURL("GET", urlString, token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return -1
	}

	/* for debugging */
	//fmt.Println(bytes.NewBuffer([]byte(body)))

	var v interface{}
	body2 := []byte(body)
	err = json.Unmarshal(body2, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	result := getResultCode(v)
	if result == 1701 {
		// when fmserverd is stopping
		return 10502
	}

	var count int
	var c []string
	var s1 string
	var sID int
	var name string
	var taskType string
	var backupType string
	var filemakerScriptType string
	var messageType string
	var scriptSequenceType string
	var systemScriptType string
	var verifyType string
	var lastRun string
	var nextRun string
	var enabled bool
	var status string
	var data [][]string

	_ = scan.ScanTree(v, "/response/schedules", &c)
	count = len(c)

	if count > 0 {
		for i := 0; i < count; i++ {
			_ = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/id", &s1)
			_ = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/name", &name)
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/backupType/resourceType", &backupType)
			if err != nil {
				backupType = ""
			}
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/filemakerScriptType/resource", &filemakerScriptType)
			if err != nil {
				filemakerScriptType = ""
			}
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/messageType/resourceType", &messageType)
			if err != nil {
				messageType = ""
			}
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/scriptSequenceType/resource", &scriptSequenceType)
			if err != nil {
				scriptSequenceType = ""
			}
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/systemScriptType/osScript", &systemScriptType)
			if err != nil {
				systemScriptType = ""
			}
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/verifyType/resourceType", &verifyType)
			if err != nil {
				verifyType = ""
			}
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/lastRun", &lastRun)
			if err != nil {
				lastRun = ""
			}
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/nextRun", &nextRun)
			if err != nil {
				nextRun = ""
			}
			_ = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/enabled", &enabled)
			if !enabled {
				nextRun = "Disabled"
			}
			_ = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/status", &status)

			sID, _ = strconv.Atoi(s1)
			if id == sID || id == 0 {
				if status == "IDLE" || status == "RUNNING" {
					if lastRun == "" || lastRun == "0000-00-00T00:00:00" {
						status = ""
					} else {
						status = "OK"
					}
				}
				if backupType != "" {
					taskType = "Backup"
				} else if filemakerScriptType != "" {
					taskType = "FileMaker Script"
				} else if messageType != "" {
					taskType = "Message"
				} else if scriptSequenceType != "" {
					taskType = "Script Sequence"
				} else if systemScriptType != "" {
					taskType = "System Script"
				} else if verifyType != "" {
					taskType = "Verify"
				}
				lastRun = getDateTimeStringOfCurrentTimeZone(lastRun, "2006/01/02 15:04", usingCloud)
				nextRun = getDateTimeStringOfCurrentTimeZone(nextRun, "2006/01/02 15:04", usingCloud)
				data = append(data, []string{s1, name, taskType, lastRun, nextRun, status})
			}
		}

		if len(data) > 0 {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoWrapText(false)
			table.SetAutoFormatHeaders(false)
			for _, v := range data {
				table.SetHeader([]string{"ID", "Name", "Type", "Last Completed", "Next Run", "Status"})
				table.Append(v)
			}
			table.Render()
		} else {
			return 10600
		}
	}

	return 0
}

func getScheduleName(url string, token string, id int) string {
	body, _, err := callURL("GET", url, token, nil)
	if err != nil {
		//fmt.Println(err.Error())
		return ""
	}

	/* for debugging */
	//fmt.Println(bytes.NewBuffer([]byte(body)))

	var v interface{}
	body2 := []byte(body)
	err = json.Unmarshal(body2, &v)
	if err != nil {
		//fmt.Println(err.Error())
		return ""
	}

	var s1 string
	var sID int
	var name string

	_ = scan.ScanTree(v, "/response/schedule/id", &s1)
	_ = scan.ScanTree(v, "/response/schedule/name", &name)
	sID, _ = strconv.Atoi(s1)
	if id == sID || id == 0 {
		return name
	}

	return ""
}

func sendMessages(u *url.URL, baseURI string, token string, message string, cmdArgs []string, clientID int) int {
	var exitStatus int

	args := []string{""}
	if len(cmdArgs[1:]) > 0 {
		args = cmdArgs[1:]
	}
	u.Path = path.Join(getAPIBasePath(baseURI), "clients")
	idList := getClients(u.String(), token, args, "NORMAL")
	if len(idList) > 0 {
		for i := 0; i < len(idList); i++ {
			if clientID == -1 || clientID == idList[i] {
				u.Path = path.Join(getAPIBasePath(baseURI), "clients", strconv.Itoa(idList[i]), "message")
				exitStatus = sendMessage(u.String(), token, message)
			}
			if clientID > 0 {
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
	body, _, err := callURL("POST", url, token, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		fmt.Println(err.Error())
	}

	/* for debugging */
	//fmt.Println(bytes.NewBuffer([]byte(body)))

	code := -1
	output := output{}
	err = json.Unmarshal(body, &output)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for i := 0; i < len(output.Messages); i++ {
			if reflect.ValueOf(output.Messages[i].Code).IsValid() {
				code, _ = strconv.Atoi(output.Messages[i].Code)
				break
			}
		}
	}

	return code
}

func getDatabases(url string, token string, arg []string, status string, fullPath bool) ([]int, []string, []string) {
	var fileName string
	var folderName string
	var idList []int
	var nameList []string
	var hintList []string
	var id int

	body, _, err := callURL("GET", url, token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return idList, nameList, hintList
	}

	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	// for debugging
	/*
		fmt.Println(url)
		fmt.Println(bytes.NewBuffer([]byte(body)))
	*/

	var totalDbCount int
	var fileStatus string
	var s1 string
	var s2 string
	var fileID string
	var decryptHint string

	_ = scan.ScanTree(v, "/response/totalDBCount", &totalDbCount)
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
				_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/status", &fileStatus)
				_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/filename", &s1)
				_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/folder", &s2)
				_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/decryptHint", &decryptHint)
				if regexp.MustCompile(`^[0-9]+$`).Match([]byte(arg[j])) {
					// ID
					_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/id", &fileID)
					if fileID == arg[j] && (status == fileStatus || status == "") {
						if fullPath {
							// for "remove" command
							nameList = append(nameList, s2+s1)
						} else {
							nameList = append(nameList, s1)
						}
						id, _ = strconv.Atoi(fileID)
						idList = append(idList, id)
						hintList = append(hintList, decryptHint)
					}
				} else {
					// name
					if (fileName == "" || comparePath(fileName, s1)) && (status == fileStatus || status == "") {
						if fullPath {
							// for "remove" command
							nameList = append(nameList, s2+s1)
						} else {
							nameList = append(nameList, s1)
						}
						_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/id", &fileID)
						id, _ = strconv.Atoi(fileID)
						idList = append(idList, id)
						hintList = append(hintList, decryptHint)
					}
				}
			} else {
				_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/status", &fileStatus)
				_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/folder", &s1)
				_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/filename", &s2)
				_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/decryptHint", &decryptHint)
				if (status == fileStatus || status == "") && (comparePath(s1, folderName) || comparePath(s1+s2, fileName)) {
					if fullPath {
						// for "remove" command
						nameList = append(nameList, s1+s2)
					} else {
						nameList = append(nameList, s2)
					}
					_ = scan.ScanTree(v, "/response/databases["+strconv.Itoa(i)+"]/id", &fileID)
					id, _ = strconv.Atoi(fileID)
					idList = append(idList, id)
					hintList = append(hintList, decryptHint)
				}
			}
		}
	}

	if fullPath {
		// for "remove" command
		for i := 0; i < len(nameList); i++ {
			if strings.Index(nameList[i], "filelinux:/") == 0 {
				nameList[i] = strings.Replace(nameList[i], "filelinux:/", "/", 1)
			} else if strings.Index(nameList[i], "filewin:/") == 0 {
				nameList[i] = strings.Replace(strings.Replace(nameList[i], "filewin:/", "", 1), "/", "¥", -1)
			} else if strings.Index(nameList[i], "filemac:/") == 0 {
				nameList[i] = strings.Replace(nameList[i], "filemac:/", "/Volumes/", 1)
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

	body, _, err := callURL("GET", url, token, nil)
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

		_ = scan.ScanTree(v, "/response/clients", &clients)
		for j := 0; j < len(clients); j++ {
			_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(j)+"]/guestFiles", &guestFiles)
			for k := 0; k < len(guestFiles); k++ {
				_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(j)+"]/guestFiles["+strconv.Itoa(k)+"]/filename", &guestFileID)
				_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(j)+"]/guestFiles["+strconv.Itoa(k)+"]/filename", &guestFileName)
				if len(folderName) == 0 {
					if fileName == "" || comparePath(fileName, guestFileName) {
						_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(k)+"]/id", &clientID)
						id, _ = strconv.Atoi(clientID)
						idList = append(idList, id)
					}
				} else {
					_ = scan.ScanTree(v, "/files/files", &files)
					for l := 0; k < len(files); l++ {
						_ = scan.ScanTree(v, "/files/files["+strconv.Itoa(k)+"]/id", &fileID)
						_ = scan.ScanTree(v, "/files/files["+strconv.Itoa(k)+"]/folder", &directory)
						if fileID == guestFileID {
							if comparePath(fileName, directory+guestFileName) {
								_ = scan.ScanTree(v, "/response/clients["+strconv.Itoa(k)+"]/id", &clientID)
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

func getServerGeneralConfigurations(urlString string, token string, printOptions []string) ([]int, int) {
	var settings []int
	var resultCode string
	var result int
	var cacheSize int
	var maxFiles int
	var maxProConnections int
	var maxPSOS int
	var startupRestorationEnabled bool

	body, _, err := callURL("GET", urlString, token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return settings, 10502
	}

	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = scan.ScanTree(v, "/messages[0]/code", &resultCode)
	if err != nil {
		return settings, 3
	}
	result, _ = strconv.Atoi(resultCode)
	_ = scan.ScanTree(v, "/response/cacheSize", &cacheSize)
	_ = scan.ScanTree(v, "/response/maxFiles", &maxFiles)
	_ = scan.ScanTree(v, "/response/maxProConnections", &maxProConnections)
	_ = scan.ScanTree(v, "/response/maxPSOS", &maxPSOS)
	startupRestorationBuiltin := true
	err = scan.ScanTree(v, "/response/startupRestorationEnabled", &startupRestorationEnabled)
	if err != nil {
		// for Claris FileMaker Server 19.1.2 or later
		startupRestorationBuiltin = false
	}

	settings = append(settings, cacheSize)
	settings = append(settings, maxFiles)
	settings = append(settings, maxProConnections)
	settings = append(settings, maxPSOS)
	if startupRestorationBuiltin {
		if startupRestorationEnabled {
			settings = append(settings, 1)
		} else {
			settings = append(settings, 0)
		}
	} else {
		// for Claris FileMaker Server 19.1.2 or later
		settings = append(settings, -1)
	}

	versionString, _ := getServerVersionString(strings.Replace(urlString, "/config/general", "/metadata", 1), token)
	version, _ := getServerVersionAsFloat(versionString)

	// output
	if result == 0 {
		for _, option := range printOptions {
			if option == "maxguests" {
				fmt.Println("MaxGuests = " + strconv.Itoa(maxProConnections) + " [default: 250, range: 0-2000] ")
			}
			if option == "maxfiles" {
				fmt.Println("MaxFiles = " + strconv.Itoa(maxFiles) + " [default: 125, range: 1-125] ")
			}
			if option == "cachesize" {
				fmt.Println("CacheSize = " + strconv.Itoa(cacheSize) + " [default: 512, range: 64-1048576] ")
			}
			if option == "hostedfiles" {
				fmt.Println("HostedFiles = " + strconv.Itoa(maxFiles) + " [default: 125, range: 1-125] ")
			}
			if option == "proconnections" {
				fmt.Println("ProConnections = " + strconv.Itoa(maxProConnections) + " [default: 250, range: 0-2000] ")
			}
			if option == "scriptsessions" {
				fmt.Println("ScriptSessions = " + strconv.Itoa(maxPSOS) + " [default: 100, range: 0-500] ")
			} else if option == "allowpsos" {
				fmt.Println("AllowPSOS = " + strconv.Itoa(maxPSOS) + " [default: 100, range: 0-500] ")
			}

			if option == "securefilesonly" || option == "requiresecuredb" {
				getServerSettingAsBool(strings.Replace(urlString, "/general", "/security", 1), token, []string{option})
			}

			if startupRestorationBuiltin && option == "startuprestorationenabled" {
				if startupRestorationEnabled {
					fmt.Println("StartupRestorationEnabled = true [default: true] ")
				} else {
					fmt.Println("StartupRestorationEnabled = false [default: true] ")
				}
			}

			if option == "authenticatedstream" {
				if version >= 19.3 && !strings.HasPrefix(versionString, "19.3.1") {
					getAuthenticatedStreamSetting(strings.Replace(urlString, "/general", "/authenticatedstream", 1), token, []string{option})
				}
			}

			if option == "parallelbackupenabled" {
				if version >= 19.5 {
					getServerSettingAsBool(strings.Replace(urlString, "/general", "/parallelbackup", 1), token, []string{option})
				}
			}

			if option == "persistcacheenabled" || option == "syncpersistcache" {
				if version >= 20.1 {
					getServerSettingAsBool(strings.Replace(urlString, "/general", "/persistentcache", 1), token, []string{option})
				}
			}
		}
	}

	return settings, result
}

func getAuthenticatedStreamSetting(urlString string, token string, printOptions []string) (int, int, error) {
	var resultCode string
	var result int
	var authenticatedStream int

	body, _, err := callURL("GET", urlString, token, nil)
	if err != nil {
		return 0, 10502, err
	}

	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = scan.ScanTree(v, "/messages[0]/code", &resultCode)
	if err != nil {
		return 0, 3, err
	}
	result, _ = strconv.Atoi(resultCode)
	err = scan.ScanTree(v, "/response/authenticatedStream", &authenticatedStream)

	// output
	if result == 0 {
		for _, option := range printOptions {
			if option == "authenticatedstream" {
				fmt.Println("AuthenticatedStream = " + strconv.Itoa(authenticatedStream) + " [default: 1, range: 1-2] ")
			}
		}
	}

	return authenticatedStream, result, err
}

func getServerSettingAsBool(urlString string, token string, printOptions []string) (bool, int, error) {
	var resultCode string
	var result int
	var enabled bool
	var enabledStr string
	var syncPersistCacheEnabled bool
	var syncPersistCacheEnabledStr string

	body, _, err := callURL("GET", urlString, token, nil)
	if err != nil {
		return false, 10502, err
	}

	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = scan.ScanTree(v, "/messages[0]/code", &resultCode)
	if err != nil {
		return false, 3, err
	}
	result, _ = strconv.Atoi(resultCode)

	u, err := url.Parse(urlString)
	if err != nil {
		return false, 3, err
	}

	if u.Path == path.Join(getAPIBasePath(urlString), "server", "config", "security") {
		err = scan.ScanTree(v, "/response/requireSecureDB", &enabled)
	} else if u.Path == path.Join(getAPIBasePath(urlString), "server", "config", "parallelbackup") {
		err = scan.ScanTree(v, "/response/parallelBackupEnabled", &enabled)
	} else if u.Path == path.Join(getAPIBasePath(urlString), "server", "config", "persistentcache") {
		err = scan.ScanTree(v, "/response/persistentCache", &enabled)
		_ = scan.ScanTree(v, "/response/persistentCacheSync", &syncPersistCacheEnabled)
		syncPersistCacheEnabledStr = "false"
		if syncPersistCacheEnabled {
			syncPersistCacheEnabledStr = "true"
		}
	}

	enabledStr = "false"
	if enabled {
		enabledStr = "true"
	}

	// output
	if result == 0 {
		for _, option := range printOptions {
			switch option {
			case "securefilesonly":
				fmt.Println("SecureFilesOnly = " + enabledStr + " [default: true] ")
			case "requiresecuredb":
				fmt.Println("RequireSecureDB = " + enabledStr + " [default: true] ")
			case "parallelbackupenabled":
				fmt.Println("ParallelBackupEnabled = " + enabledStr + " [default: false] ")
			case "persistcacheenabled":
				fmt.Println("PersistCacheEnabled = " + enabledStr + " [default: false] ")
			case "syncpersistcache":
				fmt.Println("SyncPersistCache = " + syncPersistCacheEnabledStr + " [default: false] ")
			default:
			}
		}
	}

	return enabled, result, err
}

func getWebTechnologyConfigurations(baseURI string, basePath string, token string, printOptions []string) ([]string, int, error) {
	var settings []string
	var resultCode string
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

	body, exitStatus, err := callURL("GET", u.String(), token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return settings, 10502, err
	}

	//fmt.Println(bytes.NewBuffer([]byte(body)))

	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return settings, 3, err
	}

	err = scan.ScanTree(v, "/messages[0]/code", &resultCode)
	if err != nil {
		return settings, 3, err
	}
	_, _ = strconv.Atoi(resultCode)
	_ = scan.ScanTree(v, "/response/enabled", &enabledPhp)
	_ = scan.ScanTree(v, "/response/characterEncoding", &characterEncoding)
	_ = scan.ScanTree(v, "/response/errorMessageLanguage", &errorMessageLanguage)
	_ = scan.ScanTree(v, "/response/dataPreValidation", &dataPreValidation)
	_ = scan.ScanTree(v, "/response/useFileMakerPhp", &useFileMakerPhp)

	enabledPhpStr = "true"
	if !enabledPhp {
		enabledPhpStr = "false"
	}

	if exitStatus == 500 {
		// for Claris FileMaker Server for Linux
		dataPreValidationStr = ""
		useFileMakerPhpStr = "true"
	} else {
		dataPreValidationStr = "true"
		if !dataPreValidation {
			dataPreValidationStr = "false"
		}

		useFileMakerPhpStr = "true"
		if !useFileMakerPhp {
			useFileMakerPhpStr = "false"
		}
	}

	// get XML Technology Configuration
	u.Path = path.Join(basePath, "xml", "config")

	body, _, err = callURL("GET", u.String(), token, nil)
	if err != nil {
		fmt.Println(err.Error())
		return settings, -1, err
	}

	err = json.Unmarshal(body, &v)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = scan.ScanTree(v, "/messages[0]/code", &resultCode)
	if err != nil {
		return settings, 3, err
	}
	result, _ = strconv.Atoi(resultCode)
	err = scan.ScanTree(v, "/response/enabled", &enabledXML)

	enabledXMLStr = "true"
	if !enabledXML {
		enabledXMLStr = "false"
	}

	settings = append(settings, enabledPhpStr)
	settings = append(settings, enabledXMLStr)
	settings = append(settings, characterEncoding)
	settings = append(settings, errorMessageLanguage)
	settings = append(settings, dataPreValidationStr)
	settings = append(settings, useFileMakerPhpStr)

	// output
	if result == 0 {
		for _, option := range printOptions {
			if option == "enablephp" {
				fmt.Println("EnablePHP = " + enabledPhpStr)
			}
			if option == "enablexml" {
				fmt.Println("EnableXML = " + enabledXMLStr)
			}
			if option == "encoding" {
				fmt.Println("Encoding = " + characterEncoding + " [ UTF-8 ISO-8859-1 ]")
			}
			if option == "locale" {
				fmt.Println("Locale = " + errorMessageLanguage + " [ en de fr it ja ]")
			}
			if option == "prevalidation" {
				fmt.Println("PreValidation = " + dataPreValidationStr)
			}
			if option == "usefmphp" {
				fmt.Println("UseFMPHP = " + useFileMakerPhpStr)
			}
		}
	}

	return settings, result, err
}

func disconnectAllClient(u *url.URL, baseURI string, token string, message string, graceTime int) (int, error) {
	exitStatus := 0
	var err error

	// check the client connection
	u.Path = path.Join(getAPIBasePath(baseURI), "clients")
	idList := getClients(u.String(), token, []string{""}, "NORMAL")

	// disconnect clients
	if len(idList) > 0 {
		for i := 0; i < len(idList); i++ {
			u.Path = path.Join(getAPIBasePath(baseURI), "clients", strconv.Itoa(idList[i]))
			u.RawQuery = "messageText=" + url.QueryEscape(message) + "&graceTime=" + url.QueryEscape(strconv.Itoa(graceTime))
			exitStatus, _, err = sendRequest("DELETE", u.String(), token, params{command: "disconnect"})
			if err != nil {
				break
			}
		}
	}

	return exitStatus, err
}

func stopDatabaseServer(u *url.URL, baseURI string, token string, message string, graceTime int) (int, error) {
	exitStatus := -1
	forceFlag := false
	var err error

	// disconnect clients
	_, _ = disconnectAllClient(u, baseURI, token, message, graceTime)

	// close databases
	u.Path = path.Join(getAPIBasePath(baseURI), "databases")
	idList, _, _ := getDatabases(u.String(), token, []string{""}, "NORMAL", false)
	if len(idList) > 0 {
		for i := 0; i < len(idList); i++ {
			u.Path = path.Join(getAPIBasePath(baseURI), "databases", strconv.Itoa(idList[i]))
			if graceTime == 0 {
				forceFlag = true
			}
			_, _, _ = sendRequest("PATCH", u.String(), token, params{command: "close", messageText: message, force: forceFlag})
		}
	}

	var openedID []int
	for value := 0; ; {
		time.Sleep(1 * time.Second)
		value++
		u.Path = path.Join(getAPIBasePath(baseURI), "databases")
		openedID, _, _ = getDatabases(u.String(), token, []string{""}, "CLOSING", false)
		if len(openedID) == 0 || value > 120 {
			break
		}
	}

	// stop database server
	u.Path = path.Join(getAPIBasePath(baseURI), "server", "status")
	exitStatus, _, err = sendRequest("PATCH", u.String(), token, params{status: "STOPPED"})

	return exitStatus, err
}

func waitStoppingServer(u *url.URL, baseURI string, token string) (int, error) {
	exitStatus := 0
	var err error
	var running string

	for value := 0; ; {
		time.Sleep(1 * time.Second)
		value++
		u.Path = path.Join(getAPIBasePath(baseURI), "server", "status")
		exitStatus, running, err = sendRequest("GET", u.String(), token, params{})
		if running == "STOPPED" || value > 120 {
			break
		}
	}

	return exitStatus, err
}

func getBackupTime(urlString string, token string, id int) int {
	body, _, err := callURL("GET", urlString, token, nil)
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
	var backupType string
	var filemakerScriptType string
	var messageType string
	var scriptSequenceType string
	var systemScriptType string
	var verifyType string
	var nextRun string
	var enabled bool
	var status string
	var data [][]string

	_ = scan.ScanTree(v, "/response/schedules", &c)
	count = len(c)

	if count > 0 {
		for i := 0; i < count; i++ {
			_ = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/id", &s1)
			_ = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/name", &name)
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/backupType/resourceType", &backupType)
			if err != nil {
				backupType = ""
			}
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/filemakerScriptType/resource", &filemakerScriptType)
			if err != nil {
				filemakerScriptType = ""
			}
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/messageType/resourceType", &messageType)
			if err != nil {
				messageType = ""
			}
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/scriptSequenceType/resource", &scriptSequenceType)
			if err != nil {
				scriptSequenceType = ""
			}
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/systemScriptType/osScript", &systemScriptType)
			if err != nil {
				systemScriptType = ""
			}
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/verifyType/resourceType", &verifyType)
			if err != nil {
				verifyType = ""
			}
			err = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/nextRun", &nextRun)
			if err != nil {
				nextRun = ""
			}
			_ = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/enabled", &enabled)
			if !enabled {
				nextRun = "Disabled"
			}
			_ = scan.ScanTree(v, "/response/schedules["+strconv.Itoa(i)+"]/status", &status)

			sID, _ = strconv.Atoi(s1)
			if id == sID || id == 0 {
				if backupType != "" {
					taskType = "Backup"
				} else if filemakerScriptType != "" {
					taskType = "FileMaker Script"
				} else if messageType != "" {
					taskType = "Message"
				} else if scriptSequenceType != "" {
					taskType = "Script Sequence"
				} else if systemScriptType != "" {
					taskType = "System Script"
				} else if verifyType != "" {
					taskType = "Verify"
				}
				nextRun = getDateTimeStringOfCurrentTimeZone(nextRun, "15:04", false)
				if taskType == "Backup" {
					data = append(data, []string{s1, name, nextRun})
				}
			}
		}

		if len(data) > 0 {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoWrapText(false)
			table.SetAutoFormatHeaders(false)
			for _, v := range data {
				table.SetHeader([]string{"ID", "Name", "Start time"})
				table.Append(v)
			}
			table.Render()
		} else {
			return 10600
		}
	}

	return 0
}

func getVolumeName() string {
	if runtime.GOOS == "darwin" {
		files, err := os.ReadDir("/Volumes/")
		if err != nil {
			return ""
		}
		for _, file := range files {
			if !file.IsDir() {
				path, err := filepath.EvalSymlinks("/Volumes/" + file.Name())
				if err != nil {
					return ""
				}
				if path == "/" {
					return file.Name()
				}
			}
		}
	}

	return ""
}

func comparePath(name1 string, name2 string) bool {
	extName := ".fmp12"
	pathPrefix := []string{"filelinux:", "filemac:", "filewin:"}
	volumeName := getVolumeName()

	if name1 == name2 {
		return true
	} else if name1 == name2+extName {
		return true
	} else if name1+extName == name2 {
		return true
	}

	if strings.Contains(name1, filepath.ToSlash(string(os.PathSeparator))) || strings.Contains(name2, filepath.ToSlash(string(os.PathSeparator))) {
		for i := 0; i < len(pathPrefix); i++ {
			if pathPrefix[i] == "filemac:" && (strings.Contains(name1, pathPrefix[i]) || strings.Contains(name2, pathPrefix[i])) {
				name1 = strings.Replace(name1, "/Volumes", pathPrefix[i], 1)
				name2 = strings.Replace(name2, "/Volumes", pathPrefix[i], 1)
			}

			if name1 == name2 {
				return true
			} else if name1 == name2+extName {
				return true
			} else if name1+extName == name2 {
				return true
			}

			if pathPrefix[i]+name1 == name2 {
				return true
			} else if pathPrefix[i]+name1+extName == name2 {
				return true
			} else if pathPrefix[i]+string(os.PathSeparator)+volumeName+name1 == name2 {
				return true
			} else if pathPrefix[i]+string(os.PathSeparator)+volumeName+name1+extName == name2 {
				return true
			} else if name1 == pathPrefix[i]+name2 {
				return true
			} else if name1 == pathPrefix[i]+name2+extName {
				return true
			} else if name1 == pathPrefix[i]+string(os.PathSeparator)+volumeName+name2 {
				return true
			} else if name1 == pathPrefix[i]+string(os.PathSeparator)+volumeName+name2+extName {
				return true
			}
		}
	}

	return false
}

func outputErrorMessage(code int, c *cli) {
	if code >= -1 {
		if code == 1701 {
			// when fmserverd is stopping
			code = 10502
		}
		fmt.Fprintln(c.outStream, "Error: "+strconv.Itoa(code)+" ("+getErrorDescription(code)+")")
	}
}

func sendRequest(method string, urlString string, token string, p params) (int, string, error) {
	var jsonStr []byte
	code := -1

	if ((params{}) != p && reflect.ValueOf(p.command).IsValid() && p.command == "open") || len(p.key) > 0 {
		d := dbInfo{
			"OPENED",
			p.key,
			p.saveKey,
		}
		jsonStr, _ = json.Marshal(d)
	} else if (params{}) != p && reflect.ValueOf(p.command).IsValid() && p.command == "close" {
		d := closeMessageInfo{
			"CLOSED",
			p.messageText,
			p.force,
		}
		jsonStr, _ = json.Marshal(d)
	} else if (params{}) != p && reflect.ValueOf(p.command).IsValid() && (p.command == "resume" || p.command == "pause") {
		if p.command == "resume" {
			d := statusInfo{
				"RESUMED",
			}
			jsonStr, _ = json.Marshal(d)
		} else if p.command == "pause" {
			d := statusInfo{
				"PAUSED",
			}
			jsonStr, _ = json.Marshal(d)
		}
	} else if (params{}) != p && reflect.ValueOf(p.command).IsValid() && p.command == "enable" {
		d := scheduleSettingInfo{
			true,
		}
		jsonStr, _ = json.Marshal(d)
	} else if (params{}) != p && reflect.ValueOf(p.command).IsValid() && p.command == "disable" {
		d := scheduleSettingInfo{
			false,
		}
		jsonStr, _ = json.Marshal(d)
	} else if (params{}) != p && reflect.ValueOf(p.command).IsValid() && p.command == "set" {
		if strings.HasSuffix(urlString, "/server/config/authenticatedstream") {
			// for Claris FileMaker Server 19.3.2 or later
			d := authenticatedStreamConfigInfo{
				p.authenticatedstream,
			}
			jsonStr, _ = json.Marshal(d)
		} else if strings.HasSuffix(urlString, "/server/config/parallelbackup") {
			// for Claris FileMaker Server 19.5.1 or later
			d := parallelBackupConfigInfo{
				p.parallelbackupenabled,
			}
			jsonStr, _ = json.Marshal(d)
		} else if strings.HasSuffix(urlString, "/server/config/general") && p.startuprestorationbuiltin {
			d := generalOldConfigInfo{
				p.cachesize,
				p.maxfiles,
				p.maxproconnections,
				p.maxpsos,
				p.startuprestorationenabled,
			}
			jsonStr, _ = json.Marshal(d)
		} else if strings.HasSuffix(urlString, "/server/config/general") && !p.startuprestorationbuiltin {
			// for Claris FileMaker Server 19.1.2 or later
			d := generalConfigInfo{
				p.cachesize,
				p.maxfiles,
				p.maxproconnections,
				p.maxpsos,
			}
			jsonStr, _ = json.Marshal(d)
		} else if strings.HasSuffix(urlString, "/server/config/security") {
			requiresecuredb := true
			if p.requiresecuredb == "false" {
				requiresecuredb = false
			}
			d := securityConfigInfo{
				requiresecuredb,
			}
			jsonStr, _ = json.Marshal(d)
		} else if strings.HasSuffix(urlString, "/php/config") {
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
		} else if strings.HasSuffix(urlString, "/xml/config") {
			enabled := true
			if p.enabled == "false" {
				enabled = false
			}
			d := xmlConfigInfo{
				enabled,
			}
			jsonStr, _ = json.Marshal(d)
		}
	} else if (params{}) != p && reflect.ValueOf(p.command).IsValid() && p.command == "certificate create" {
		d := creatingCsrInfo{
			p.subject,
			p.password,
		}
		jsonStr, _ = json.Marshal(d)
	} else if (params{}) != p && reflect.ValueOf(p.command).IsValid() && p.command == "certificate import" {
		d := importingCertificateInfo{
			p.certificate,
			p.privateKey,
			p.intermediateCertificates,
			p.password,
		}
		jsonStr, _ = json.Marshal(d)
	} else if (params{}) != p && reflect.ValueOf(p.status).IsValid() && (p.status == "RUNNING" || p.status == "STOPPED") {
		d := statusInfo{
			p.status,
		}
		jsonStr, _ = json.Marshal(d)
	} else {
		jsonStr = []byte("")
	}

	body, statusCode, err := callURL(method, urlString, token, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return -1, "", err
	}

	// for debugging
	/*
		fmt.Println(method)
		fmt.Println(urlString)
		fmt.Println(statusCode)
		fmt.Println(string(jsonStr))
		fmt.Println(bytes.NewBuffer([]byte(body)))
	*/

	if statusCode >= 400 {
		return 10001, "", err
	}

	if body != nil {
		output := output{}
		if json.Unmarshal(body, &output) == nil {
			for i := 0; i < len(output.Messages); i++ {
				if reflect.ValueOf(output.Messages[i].Code).IsValid() {
					code, _ = strconv.Atoi(output.Messages[i].Code)
					break
				}
			}
			return code, output.Response.Status, nil
		} else {
			// In case of detecting a server-side error
			return 3, "", err
		}
	}

	return 0, "", nil
}

func callURL(method string, urlString string, token string, request io.Reader) ([]byte, int, error) {
	req, err := http.NewRequest(method, urlString, request)
	if err != nil {
		fmt.Println(err.Error())
		return []byte(""), 500, err
	}

	if request == nil {
		req.Header.Set("Content-Length", "0")
	}
	req.Header.Set("Content-Type", "application/json")
	if len(token) >= 5 && (token[:5] == "FMID " || token[:6] == "Basic " || token[:4] == "PKI ") {
		req.Header.Set("Authorization", token)
	} else {
		req.Header.Set("Authorization", "Bearer "+strings.Replace(strings.Replace(token, "\n", "", -1), "\r", "", -1))
	}
	client := &http.Client{Timeout: time.Duration(5) * time.Second}
	res, err := client.Do(req)
	if err != nil {
		// for debugging
		//fmt.Println(err.Error())

		if res == nil {
			return []byte(""), 404, err
		}

		return []byte(""), res.StatusCode, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	// for debugging
	//fmt.Println(bytes.NewBuffer([]byte(body)))

	if err != nil {
		fmt.Println(err.Error())
		return []byte(""), res.StatusCode, err
	}

	return body, res.StatusCode, err
}

func detectHostUnreachable(exitStatus int) bool {
	switch exitStatus {
	case 9:
		return false
	case 21:
		return false
	case 956:
		return false
	case 20402:
		return false
	case 20405:
		return false
	case 20408:
		return false
	default:
		return true
	}
}

func getErrorDescription(errorCode int) string {
	description := ""

	switch errorCode {
	case -1:
		description = "Unknown error"
	case 3:
		description = "Unavailable command"
	case 4:
		description = "Command is unknown"
	case 8:
		description = "Empty result"
	case 9:
		description = "Access denied"
	case 21:
		description = "Not Supported"
	case 212:
		description = "Invalid user account and/or password; please try again"
	case 214:
		description = "Too many login attempts, account locked out"
	case 802:
		description = "Unable to open the file"
	case 956:
		description = "Maximum number of Admin API sessions exceeded"
	case 958:
		description = "Parameter missing"
	case 960:
		description = "Parameter is invalid"
	case 1700:
		description = "Resource doesn't exist"
	case 1702:
		description = "Authentication information wasn't provided in the correct format; verify the value of the Authorization header"
	case 1708:
		description = "Parameter value is invalid"
	case 1713:
		description = "The API request is not supported for this operating system"
	case 1717:
		description = "PHP config file does not exist; PHP may not be installed on the server"
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
		description = "Schedule at specified index does not exist"
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
	case 20402:
		description = "File permission error"
	case 20405:
		description = "File not found or not accessible."
	case 20406:
		description = "File already exists"
	case 20408:
		description = "File read error"
	case 20501:
		description = "Directory not empty"
	case 20630:
		description = "SSL certificate expired"
	case 20632:
		description = "SSL certificate verification error"
	case 25004:
		description = "Parameters are invalid"
	case 25006:
		description = "Invalid session error"
	default:
		description = ""
	}

	return description
}

func getDateTimeStringOfCurrentTimeZone(dateTime string, outputFormat string, usingCloud bool) string {
	var t time.Time
	_, offset := time.Now().Zone()

	if len(dateTime) > 0 {
		reg := `(\d+[-/]\d+[-/]\d+)`
		if regexp.MustCompile(reg).Match([]byte(dateTime)) {
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
						// for clients (FileMaker Cloud for AWS)
						dateTime = t.Add(time.Second * time.Duration(offset)).Format(outputFormat)
					}
				}
			} else {
				t, _ = time.Parse("2006-01-02 15:04:05 MST", dateTime)
				if t.Format("2006-01-02") == "0001-01-01" {
					t, _ = time.Parse("2006-01-02T15:04:05", dateTime)
				}
				if t.Format("2006-01-02") == "0001-01-01" {
					t, _ = time.Parse("2006-01-02T15:04:05.000Z", dateTime)
					if t.Format("2006-01-02") == "0001-01-01" {
						dateTime = ""
					} else {
						// for clients (FileMaker Server)
						dateTime = t.Add(time.Second * time.Duration(offset)).Format(outputFormat)
					}
				} else {
					// for clients and schedules
					if usingCloud {
						// for Claris FileMaker Cloud
						dateTime = t.Add(time.Second * time.Duration(offset)).Format(outputFormat)
					} else {
						// for Claris FileMaker Server
						dateTime = t.Format(outputFormat)
					}
				}
			}
		}
	}

	return dateTime
}

var helpTextTemplate = `Usage: fmcsadmin [options] [COMMAND]

Description: 
    fmcsadmin is a command line tool to administer the Database Server 
    component of Claris FileMaker Server via Claris FileMaker Admin API.

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
`

var commandListHelpTextTemplate = `fmcsadmin commands are:

    CANCEL          Cancel the currently running operation
                    (for FileMaker Server 19.5.1 or later)
    CERTIFICATE     Manage SSL certificates
                    (for FileMaker Server 19.2.1 or later)
    CLOSE           Close databases
    DELETE          Delete a schedule
    DISABLE         Disable schedules
    DISCONNECT      Disconnect clients
    ENABLE          Enable schedules
    GET             Retrieve server or CWP configuration settings, or retrieve 
                    the start time of a backup schedule or schedules
    HELP            Get help pages
    LIST            List clients, databases, plug-ins, or schedules
    OPEN            Open databases
    PAUSE           Temporarily stop database access
    REMOVE          Move databases out of hosted folder
                    (for FileMaker Server 19.3.1 or later)
    RESTART         Restart a server process (for FileMaker Server)
    RESUME          Make paused databases available
    RUN             Run a schedule
    SEND            Send a message
    SET             Change server or CWP configuration settings, or change the 
                    start time of a backup schedule
    START           Start a server process (for FileMaker Server)
    STATUS          Get status of clients or databases
    STOP            Stop a server process (for FileMaker Server)
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
    -h, --help                 Print this page.
    -i IDENTITYFILE            Specify a private key file for PKI Authentication.
    -p pass, --password pass   Password to use to authenticate with the server.
    -u user, --username user   Username to use to authenticate with the server.
    -v, --version              Print version information.
    -y, --yes                  Automatically answer yes to all command prompts.

Options that apply to specific commands:
    -c NUM, --client NUM       Specify a client number to send a message.
    -f, --force                Force database to close or Database Server 
                               to stop, immediately disconnecting clients.
    --intermediateCA IMCAFILE  Specify the file that contains the intermediate
                               CA certificate(s) for certificate import.
    --key encryptpass          Specify the database encryption password.
    --keyfile KEYFILE          Specify private key file for certificate import.
    --keyfilepass kfpassword   Specify password needed to read KEYFILE.
    -m msg, --message msg      Specify a text message to send to clients. 
    -s, --stats                Return FILE or CLIENT stats.
    --savekey                  Save the database encryption password.
    -t sec, --gracetime sec    Specify time in seconds before client is forced
                               to disconnect.
`

var cancelHelpTextTemplate = `Usage: fmcsadmin CANCEL [TYPE]

Description:
    Cancel the currently running operation of specified TYPE.

    Valid operation TYPEs:
        BACKUP          Cancel the currently running backup.
`

var certificateHelpTextTemplate = `Usage: fmcsadmin CERTIFICATE [CERT_OP] [options] [NAME] [FILE]

Description:
    This command lets the administrator manage SSL certificates.

    Valid certificate operations (CERT_OP) are:
        CREATE     Generate an SSL private key and a certificate request
                   to be sent to a certificate authority for signing.
        IMPORT     Import an SSL certificate issued by a certificate authority.
        DELETE     Remove the certificate request, custom certificate, and
                   associated private key.

    For the CREATE operation, a unique NAME for the database server is
    needed.  This is in the form of server name or DNS name. For example
      fmcsadmin certificate create /CN=svr.example.com/C=US --keyfilepass secret

    For the IMPORT operation, the full path of the signed certificate FILE
    from the certificate authority is required, e.g.
      fmcsadmin certificate import /tmp/Signed.cer --keyfilepass secret

Options:
    --keyfile KEYFILE
        Specifies the private key file which is associated with the signed
        certificate file.

    --keyfilepass secret
        Specifies the encryption password used to encrypt and decrypt the
        private key file.

    --intermediateCA intermediateCAfile
        Specifies the file that contains the intermediate CA certificate(s).
        If the certificate was signed by an intermediate certificate authority,
        use this option to IMPORT the intermediateCAFile from the vendor that
        issued the certificate.
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

    -f, --force 
        Forces a database to be closed, immediately disconnecting clients.
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

var getHelpTextTemplate = `Usage: fmcsadmin GET BACKUPTIME [ID]
       fmcsadmin GET [CONFIG_TYPE] [NAME1 NAME2 ...]


Description:
    The GET BACKUPTIME command retrieves the start time of a specified backup 
    schedule when you use the optional ID parameter. If you omit the optional ID
    parameter, the start times of all backup schedules are returned.

    The GET CONFIG_TYPE command retrieves the server or Custom Web Publishing 
    configurations.

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

    Note: Input configuration names are not case sensitive.

    Examples:
      fmcsadmin GET BACKUPTIME
      fmcsadmin GET BACKUPTIME 2
      fmcsadmin GET SERVERCONFIG HOSTEDFILES SCRIPTSESSIONS
      fmcsadmin GET SERVERCONFIG
      fmcsadmin GET CWPCONFIG ENABLEPHP USEFMPHP
      fmcsadmin GET CWPCONFIG
`

var listHelpTextTemplate = `Usage: fmcsadmin LIST [TYPE] [options]

Description: 
    Lists items of the specified TYPE. 

    Valid TYPEs:
        CLIENTS         Lists the connected clients.
        FILES           Lists the hosted databases.
        PLUGINS         List Database Server calculation plug-ins.
                        (for FileMaker Server 19.2.1 or later)
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

    --savekey
        Saves the encryption password provided with the --key option. The
        password is saved on the server for each encrypted database being
        opened. The saved password allows the server to open an encrypted
        database without specifying the --key option every time.
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

var removeHelpTextTemplate = `Usage: fmcsadmin REMOVE [FILE...] [PATH...]

Description:
    Moves a database that has been closed into a "Removed" folder so it will 
    no longer be hosted. Each specified database (FILE) is removed, and all 
    databases in each folder (PATH) are removed. If no FILE or PATH is 
    specified, all closed databases in the hosting area are removed.

Options:
    No command specific options.
`

var restartHelpTextTemplate = `Usage: fmcsadmin RESTART [TYPE]

Description:
    Restarts the server of specified TYPE. This command stops the server 
    TYPE and then starts it after a short delay.

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
    The SET CONFIG_TYPE command changes the server or Custom Web Publishing
    configuration settings.

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

    Note: Input configuration names are not case sensitive.

    Examples:
      fmcsadmin SET SERVERCONFIG CACHESIZE=1024 SECUREFILESONLY=true
      fmcsadmin SET CWPCONFIG ENABLEPHP=true ENCODING=ISO-8859-1 LOCALE=de
`

var startHelpTextTemplate = `Usage: fmcsadmin START [TYPE]

Description:
    Starts the server of specified TYPE.

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

    Valid server TYPE:
        SERVER          Stops the Database Server. By default, all clients
                        are disconnected after 90 seconds. 

Options: (applicable to SERVER only)
    -m message, --message message 
        Specifies a text message to send to the connected clients.
`
