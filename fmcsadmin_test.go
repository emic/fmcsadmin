package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}
	var args []string
	var status int

	args = strings.Split("fmcsadmin", " ")
	status = cli.Run(args)
	assert.Equal(t, 0, status)

	args = strings.Split("fmcsadmin -V", " ")
	status = cli.Run(args)
	assert.Equal(t, 248, status)

	args = strings.Split("fmcsadmin certificate", " ")
	status = cli.Run(args)
	assert.Equal(t, 248, status)

	args = strings.Split("fmcsadmin close -b", " ")
	status = cli.Run(args)
	assert.Equal(t, 249, status)

	args = strings.Split("fmcsadmin close --unknown", " ")
	status = cli.Run(args)
	assert.Equal(t, 249, status)

	args = strings.Split("fmcsadmin disconnect unknown", " ")
	status = cli.Run(args)
	assert.Equal(t, 248, status)

	args = strings.Split("fmcsadmin disconnect 0", " ")
	status = cli.Run(args)
	assert.Equal(t, 248, status)

	args = strings.Split("fmcsadmin list", " ")
	status = cli.Run(args)
	assert.Equal(t, 248, status)

	args = strings.Split("fmcsadmin list unknown", " ")
	status = cli.Run(args)
	assert.Equal(t, 248, status)

	args = strings.Split("fmcsadmin get", " ")
	status = cli.Run(args)
	assert.Equal(t, 248, status)

	args = strings.Split("fmcsadmin get cwpconfig invalidparameter", " ")
	status = cli.Run(args)
	assert.Equal(t, 10001, status)

	args = strings.Split("fmcsadmin get cwpconfig invalidparameter", " ")
	_ = cli.Run(args)
	expected := "Invalid configuration name: invalidparameter"
	assert.Contains(t, outStream.String(), expected)

	args = strings.Split("fmcsadmin get serverconfig invalidparameter", " ")
	status = cli.Run(args)
	assert.Equal(t, 10001, status)

	args = strings.Split("fmcsadmin get serverprefs invalidparameter", " ")
	status = cli.Run(args)
	assert.Equal(t, 3, status)

	args = strings.Split("fmcsadmin get unknown", " ")
	status = cli.Run(args)
	assert.Equal(t, 248, status)

	args = strings.Split("fmcsadmin restart", " ")
	status = cli.Run(args)
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		assert.Equal(t, 248, status)
	} else {
		assert.Equal(t, 248, status)
	}

	args = strings.Split("fmcsadmin restart unknown -y", " ")
	status = cli.Run(args)
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		assert.Equal(t, 23, status)
	} else {
		assert.Equal(t, 23, status)
	}

	args = strings.Split("fmcsadmin run unknown", " ")
	status = cli.Run(args)
	assert.Equal(t, 248, status)

	args = strings.Split("fmcsadmin set", " ")
	status = cli.Run(args)
	assert.Equal(t, 248, status)

	args = strings.Split("fmcsadmin set cwpconfig", " ")
	status = cli.Run(args)
	assert.Equal(t, 10001, status)

	args = strings.Split("fmcsadmin set cwpconfig enablephp", " ")
	status = cli.Run(args)
	assert.Equal(t, 10001, status)

	args = strings.Split("fmcsadmin set cwpconfig invalidparameter=true", " ")
	status = cli.Run(args)
	assert.Equal(t, 10001, status)

	args = strings.Split("fmcsadmin set cwpconfig invalidparameter=true", " ")
	_ = cli.Run(args)
	expected = "Invalid configuration name: invalidparameter"
	assert.Contains(t, outStream.String(), expected)

	args = strings.Split("fmcsadmin set serverconfig", " ")
	status = cli.Run(args)
	assert.Equal(t, 10001, status)

	args = strings.Split("fmcsadmin set serverconfig invalidparameter", " ")
	status = cli.Run(args)
	assert.Equal(t, 10001, status)

	args = strings.Split("fmcsadmin set serverconfig invalidparameter=true", " ")
	status = cli.Run(args)
	assert.Equal(t, 10001, status)

	args = strings.Split("fmcsadmin set serverprefs", " ")
	status = cli.Run(args)
	assert.Equal(t, 10001, status)

	args = strings.Split("fmcsadmin set serverprefs invalidparameter", " ")
	status = cli.Run(args)
	assert.Equal(t, 10001, status)

	args = strings.Split("fmcsadmin set serverprefs invalidparameter=true", " ")
	status = cli.Run(args)
	assert.Equal(t, 3, status)

	args = strings.Split("fmcsadmin set unknown", " ")
	status = cli.Run(args)
	assert.Equal(t, 248, status)

	args = strings.Split("fmcsadmin start", " ")
	status = cli.Run(args)
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		assert.Equal(t, 248, status)
	} else {
		assert.Equal(t, 248, status)
	}

	args = strings.Split("fmcsadmin start unknown", " ")
	status = cli.Run(args)
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		assert.Equal(t, 23, status)
	} else {
		assert.Equal(t, 23, status)
	}

	args = strings.Split("fmcsadmin status unknown", " ")
	status = cli.Run(args)
	assert.Equal(t, 248, status)

	args = strings.Split("fmcsadmin stop", " ")
	status = cli.Run(args)
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		assert.Equal(t, 248, status)
	} else {
		assert.Equal(t, 248, status)
	}

	args = strings.Split("fmcsadmin stop unknown -y", " ")
	status = cli.Run(args)
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		assert.Equal(t, 23, status)
	} else {
		assert.Equal(t, 23, status)
	}
}

func TestRunInvalidCommand(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin invalidcommand", " ")
	status := cli.Run(args)
	assert.Equal(t, 248, status)
	expected := "Usage: fmcsadmin [options] [COMMAND]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunWithHelpOption1(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin -h", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin [options] [COMMAND]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunWithHelpOption2(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin --help", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin [options] [COMMAND]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunWithIdentityFileOption(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin -i notexist.pub list files", " ")
	status := cli.Run(args)
	assert.Equal(t, 20405, status)
}

func TestRunWithVersionOption1(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin -v", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "fmcsadmin"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunWithVersionOption2(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin --version", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "fmcsadmin"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunCancelCommand(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin cancel", " ")
	status := cli.Run(args)
	assert.Equal(t, 248, status)
	expected := "Error: 11000 (Invalid command)"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunDeleteCommand1(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin delete", " ")
	status := cli.Run(args)
	assert.Equal(t, 248, status)
	expected := "Error: 11000 (Invalid command)"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunDeleteCommand2(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin delete schedule", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "fmcsadmin: really delete a schedule? (y, n)"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunDisableCommand1(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin disable", " ")
	status := cli.Run(args)
	assert.Equal(t, 248, status)
	expected := "Error: 11000 (Invalid command)"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunDisableCommand2(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin disable schedule", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "fmcsadmin: really disable schedule(s)? (y, n) "
	assert.Contains(t, outStream.String(), expected)
}

func TestRunDisableCommand3(t *testing.T) {
	running := true
	url := "http://127.0.0.1:16001/fmi/admin/api/v2/user/auth"
	_, err := http.Get(url)
	if err != nil {
		running = false
	}

	if running == true {
		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli := &cli{outStream: outStream, errStream: errStream}
		args := strings.Split("fmcsadmin -u USERNAME -p PASSWORD -y disable schedule", " ")
		status := cli.Run(args)
		assert.Equal(t, 10600, status)
		expected := "Error: 10600 (Schedule at specified index does not exist)"
		assert.Contains(t, outStream.String(), expected)
	}
}

func TestRunDisconnectCommand1(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin disconnect", " ")
	status := cli.Run(args)
	assert.Equal(t, 248, status)
	expected := "Error: 11000 (Invalid command)"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunableEnbleCommand1(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin enable", " ")
	status := cli.Run(args)
	assert.Equal(t, 248, status)
	expected := "Error: 11000 (Invalid command)"
	assert.Contains(t, outStream.String(), expected)
}

/*
func TestRunEnableCommand2(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin -y enable schedule", " ")
	status := cli.Run(args)
	assert.Equal(t, 104, status)
	expected := "Error: 10600 (Schedule at specified index does not exist)"
	assert.Contains(t, outStream.String(), expected)
}
*/

func TestRunHelpCommand1(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin [options] [COMMAND]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunHelpCommand2(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help commands", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "fmcsadmin commands are:"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunHelpCommand3(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help options", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Many fmcsadmin commands take options and parameters."
	assert.Contains(t, outStream.String(), expected)
}

func TestRunHelpCommand4(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help invalidoption", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin [options] [COMMAND]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunRunCommand1(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin run", " ")
	status := cli.Run(args)
	assert.Equal(t, 248, status)
	expected := "Error: 11000 (Invalid command)"
	assert.Contains(t, outStream.String(), expected)
}

/*
func TestRunRunCommand2(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin -y run schedule", " ")
	status := cli.Run(args)
	assert.Equal(t, 104, status)
	expected := "Error: 10600 (Schedule at specified index does not exist)"
	assert.Contains(t, outStream.String(), expected)
}
*/

func TestRunShowCertificateCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help certificate", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin CERTIFICATE [CERT_OP] [options] [NAME] [FILE]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowCloseCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help close", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin CLOSE [FILE...] [PATH...] [options]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowDeleteCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help delete", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin DELETE [TYPE] [SCHEDULE_NUMBER]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowDisableCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help disable", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin DISABLE [TYPE] [SCHEDULE_NUMBER]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowDisconnectCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help disconnect", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin DISCONNECT CLIENT [CLIENT_NUMBER] [options]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowEnableCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help enable", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin ENABLE [TYPE] [SCHEDULE_NUMBER]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowListCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help list", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin LIST [TYPE] [options]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowOpenCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help open", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin OPEN [options] [FILE...] [PATH...]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowPauseCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help pause", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin PAUSE [FILE...] [PATH...]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowRestartCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help restart", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin RESTART [TYPE]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowResumeCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help resume", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin RESUME [FILE...] [PATH...]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowRunCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help run", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin RUN SCHEDULE [SCHEDULE_NUMBER]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowSendCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help send", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin SEND [options] [CLIENT_NUMBER] [FILE...] [PATH...]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowStartCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help start", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin START [TYPE]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowStatusCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help status", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin STATUS [TYPE] [CLIENT_NUMBER] [FILE...]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunShowStopCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help stop", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin STOP [TYPE] [options]"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunCloseCommand1(t *testing.T) {
	running := true
	url := "http://127.0.0.1:16001/fmi/admin/api/v2/user/auth"
	_, err := http.Get(url)
	if err != nil {
		running = false
	}

	if running == false {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "{\"response\": {\"token\": \"ACCESSTOKEN\", \"totalDBCount\": 1, \"clients\": [], \"databases\": [{\"id\": \"1\", \"filename\": \"TestDB.fmp12\", \"status\": \"NORMAL\", \"folder\": \"filemac:/Macintosh HD/Library/FileMaker Server/Data/Databases/Sample/\", \"decryptHint\": \"\"}]}, \"messages\": [{\"code\": \"0\"}]}")
			if strings.Contains(r.URL.Path, "/fmi/admin/api/v2/databases/") {
				request, _ := io.ReadAll(r.Body)
				if strings.Contains(string([]byte(request)), "\"status\":\"CLOSED\"") {
					assert.Equal(t, "{\"status\":\"CLOSED\",\"messageText\":\"TESTMESSAGE\",\"force\":false}", string([]byte(request)))
				}
			}
		})

		address := "127.0.0.1:16001"
		l, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatal(err)
		}
		ts := httptest.Server{
			Listener: l,
			Config:   &http.Server{Handler: handler},
		}
		ts.Start()
		defer ts.Close()
	}

	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}
	args := strings.Split("fmcsadmin close TestDB -y -u USERNAME -p PASSWORD -m TESTMESSAGE", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Closing: TestDB.fmp12"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunOpenCommand1(t *testing.T) {
	running := true
	url := "http://127.0.0.1:16001/fmi/admin/api/v2/user/auth"
	_, err := http.Get(url)
	if err != nil {
		running = false
	}

	if running == false {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "{\"response\": {\"token\": \"ACCESSTOKEN\", \"totalDBCount\": 1, \"clients\": [], \"databases\": [{\"id\": \"1\", \"filename\": \"TestDB.fmp12\", \"status\": \"CLOSED\", \"folder\": \"filemac:/Macintosh HD/Library/FileMaker Server/Data/Databases/Sample/\", \"decryptHint\": \"\"}]}, \"messages\": [{\"code\": \"0\"}]}")
			if strings.Contains(r.URL.Path, "/fmi/admin/api/v2/databases/") {
				request, _ := io.ReadAll(r.Body)
				if strings.Contains(string([]byte(request)), "\"status\":\"OPENED\"") {
					assert.Equal(t, "{\"status\":\"OPENED\",\"key\":\"\",\"saveKey\":false}", string([]byte(request)))
				}
			}
		})

		address := "127.0.0.1:16001"
		l, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatal(err)
		}
		ts := httptest.Server{
			Listener: l,
			Config:   &http.Server{Handler: handler},
		}
		ts.Start()
		defer ts.Close()
	}

	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}
	args := strings.Split("fmcsadmin open TestDB -u USERNAME -p PASSWORD", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "File Opening: TestDB.fmp12"
	assert.Contains(t, outStream.String(), expected)
}
func TestRunStatusCommand1(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin status", " ")
	status := cli.Run(args)
	assert.Equal(t, 248, status)
	expected := "Error: 11000 (Invalid command)"
	assert.Contains(t, outStream.String(), expected)
}

func TestGetFlags(t *testing.T) {
	var expected []string
	var args []string
	var resultFlags commandOptions
	var cmdArgs []string

	flags := commandOptions{}
	flags.helpFlag = false
	flags.versionFlag = false
	flags.yesFlag = false
	flags.statsFlag = false
	flags.fqdn = ""
	flags.username = ""
	flags.password = ""
	flags.key = ""
	flags.message = ""
	flags.clientID = -1
	flags.graceTime = 90

	/*
	 * cancel
	 * Usage: fmcsadmin CANCEL [TYPE]
	 *
	 * fmcsadmin cancel backup
	 * fmcsadmin cancel -y backup
	 * fmcsadmin --fqdn example.jp cancel backup
	 * fmcsadmin --fqdn example.jp cancel -y backup
	 * fmcsadmin --fqdn example.jp -u USERNAME cancel backup
	 * fmcsadmin --fqdn example.jp -u USERNAME cancel -y backup
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD cancel backup
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD cancel -y backup
	 */
	expected = []string{"cancel", "backup"}
	args = strings.Split("fmcsadmin cancel backup", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	/*
	 * certificate
	 * Usage: fmcsadmin CERTIFICATE [CERT_OP] [options] [NAME] [FILE]
	 *
	 * fmcsadmin certificate create "/CN=fms.example.com/C=US" --keyfilepass secret
	 * fmcsadmin -u USERNAME certificate create "/CN=fms.example.com/C=US" --keyfilepass secret
	 * fmcsadmin -p PASSWORD certificate create "/CN=fms.example.com/C=US" --keyfilepass secret
	 * fmcsadmin -u USERNAME -p PASSWORD certificate create "/CN=fms.example.com/C=US" --keyfilepass secret
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD certificate create "/CN=fms.example.com/C=US" --keyfilepass secret
	 * fmcsadmin certificate import --keyfile KEYFILE --keyfilepass secret --intermediateCA intermediateCAfile CRTFILE
	 * fmcsadmin certificate import --keyfile KEYFILE --keyfilepass secret --intermediateCA intermediateCAfile -y CRTFILE
	 * fmcsadmin certificate import -u USERNAME --keyfile KEYFILE --keyfilepass secret --intermediateCA intermediateCAfile CRTFILE
	 * fmcsadmin certificate import -p PASSWORD --keyfile KEYFILE --keyfilepass secret --intermediateCA intermediateCAfile CRTFILE
	 * fmcsadmin certificate import -u USERNAME -p PASSWORD --keyfile KEYFILE --keyfilepass secret --intermediateCA intermediateCAfile CRTFILE
	 * fmcsadmin certificate import --fqdn example.jp -u USERNAME -p PASSWORD --keyfile KEYFILE --keyfilepass secret --intermediateCA intermediateCAfile CRTFILE
	 * fmcsadmin certificate import --fqdn example.jp -u USERNAME -p PASSWORD --keyfile KEYFILE --keyfilepass secret --intermediateCA intermediateCAfile -y CRTFILE
	 * fmcsadmin certificate delete
	 * fmcsadmin certificate delete -y
	 * fmcsadmin -u USERNAME certificate delete
	 * fmcsadmin -p PASSWORD certificate delete
	 * fmcsadmin -u USERNAME -p PASSWORD certificate delete
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD certificate delete
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD certificate delete -y
	 * etc.
	 */

	/*
	 * close
	 * Usage: fmcsadmin CLOSE [FILE...] [PATH...] [options]
	 *
	 * fmcsadmin close
	 * fmcsadmin close 1
	 * fmcsadmin close 1 2
	 * fmcsadmin close TestDB
	 * fmcsadmin close TestDB.fmp12
	 * fmcsadmin close "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"
	 * fmcsadmin close "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"
	 * fmcsadmin close "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"
	 * fmcsadmin close "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"
	 * fmcsadmin close "/opt/FileMaker/FileMaker Server/Data/Databases/"
	 * fmcsadmin close "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/"
	 * fmcsadmin close TestDB FMServer_Sample
	 * fmcsadmin --fqdn example.jp close TestDB
	 * fmcsadmin close -y TestDB
	 * fmcsadmin close -u USERNAME TestDB
	 * fmcsadmin close -p PASSWORD TestDB
	 * fmcsadmin close -u USERNAME -p PASSWORD TestDB
	 * fmcsadmin close -u USERNAME -p PASSWORD -y TestDB
	 * fmcsadmin close -m "Test Message" TestDB
	 * fmcsadmin close -m "Test Message" -y TestDB
	 * etc.
	 */
	expected = []string{"close", "1"}
	args = strings.Split("fmcsadmin close 1", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "1", "2"}
	args = strings.Split("fmcsadmin close 1 2", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB"}
	args = strings.Split("fmcsadmin close TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, false, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB", "FMServer_Sample"}
	args = strings.Split("fmcsadmin close TestDB FMServer_Sample", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, false, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB"}
	args = strings.Split("fmcsadmin --fqdn example.jp close TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB"}
	args = strings.Split("fmcsadmin -y close TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB", "FMServer_Sample"}
	args = strings.Split("fmcsadmin -y close TestDB FMServer_Sample", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB"}
	args = strings.Split("fmcsadmin close TestDB -y", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB", "FMServer_Sample"}
	args = strings.Split("fmcsadmin close TestDB FMServer_Sample -y", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB"}
	args = strings.Split("fmcsadmin close -y TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB", "FMServer_Sample"}
	args = strings.Split("fmcsadmin close -y TestDB FMServer_Sample", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB"}
	args = strings.Split("fmcsadmin close -u USERNAME TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "USERNAME", resultFlags.username)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB"}
	args = strings.Split("fmcsadmin close -p PASSWORD TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "PASSWORD", resultFlags.password)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB"}
	args = strings.Split("fmcsadmin close -u USERNAME -p PASSWORD TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "USERNAME", resultFlags.username)
	assert.Equal(t, "PASSWORD", resultFlags.password)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB"}
	args = strings.Split("fmcsadmin close -u USERNAME -p PASSWORD -y TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "USERNAME", resultFlags.username)
	assert.Equal(t, "PASSWORD", resultFlags.password)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB"}
	args = strings.Split("fmcsadmin close -m Message TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "Message", resultFlags.message)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"close", "TestDB"}
	args = strings.Split("fmcsadmin close -m Message -y TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "Message", resultFlags.message)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	/*
	 * delete
	 * Usage: fmcsadmin DELETE [TYPE] [SCHEDULE_NUMBER]
	 *
	 * fmcsadmin delete schedule 2
	 * fmcsadmin delete -y schedule 2
	 * fmcsadmin --fqdn example.jp delete schedule 2
	 * fmcsadmin --fqdn example.jp delete -y schedule 2
	 * fmcsadmin --fqdn example.jp -u USERNAME delete schedule 2
	 * fmcsadmin --fqdn example.jp -u USERNAME delete -y schedule 2
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD delete schedule 2
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD delete -y schedule 2
	 */
	expected = []string{"delete", "schedule", "2"}
	args = strings.Split("fmcsadmin delete schedule 2", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"delete", "schedule", "2"}
	args = strings.Split("fmcsadmin delete -y schedule 2", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"delete", "schedule", "2"}
	args = strings.Split("fmcsadmin --fqdn example.jp delete schedule 2", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"delete", "schedule", "2"}
	args = strings.Split("fmcsadmin --fqdn example.jp delete -y schedule 2", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"delete", "schedule", "2"}
	args = strings.Split("fmcsadmin --fqdn example.jp -u USERNAME delete schedule 2", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, "USERNAME", resultFlags.username)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"delete", "schedule", "2"}
	args = strings.Split("fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD delete schedule 2", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, "USERNAME", resultFlags.username)
	assert.Equal(t, "PASSWORD", resultFlags.password)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"delete", "schedule", "2"}
	args = strings.Split("fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD -y delete schedule 2", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, "USERNAME", resultFlags.username)
	assert.Equal(t, "PASSWORD", resultFlags.password)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	/*
	 * disable
	 * Usage: fmcsadmin DISABLE [TYPE] [SCHEDULE_NUMBER]
	 *
	 * fmcsadmin disable schedule 1
	 * fmcsadmin disable -y schedule 1
	 * fmcsadmin --fqdn example.jp disable schedule 1
	 * fmcsadmin --fqdn example.jp disable -y schedule 1
	 * fmcsadmin --fqdn example.jp -u USERNAME disable schedule 1
	 * fmcsadmin --fqdn example.jp -u USERNAME disable -y schedule 1
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD disable schedule 1
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD disable -y schedule 1
	 */
	expected = []string{"disable", "schedule", "1"}
	args = strings.Split("fmcsadmin disable schedule 1", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disable", "schedule", "1"}
	args = strings.Split("fmcsadmin disable -y schedule 1", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disable", "schedule", "1"}
	args = strings.Split("fmcsadmin --fqdn example.jp disable schedule 1", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disable", "schedule", "1"}
	args = strings.Split("fmcsadmin --fqdn example.jp disable -y schedule 1", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disable", "schedule", "1"}
	args = strings.Split("fmcsadmin --fqdn example.jp -u USERNAME disable schedule 1", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, "USERNAME", resultFlags.username)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disable", "schedule", "1"}
	args = strings.Split("fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD disable schedule 1", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, "USERNAME", resultFlags.username)
	assert.Equal(t, "PASSWORD", resultFlags.password)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disable", "schedule", "1"}
	args = strings.Split("fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD -y disable schedule 1", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, "USERNAME", resultFlags.username)
	assert.Equal(t, "PASSWORD", resultFlags.password)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	/*
	 * disconnect
	 * Usage: fmcsadmin DISCONNECT CLIENT [CLIENT_NUMBER] [options]
	 *
	 * fmcsadmin disconnect client
	 * fmcsadmin disconnect client -y
	 * fmcsadmin disconnect client -m "Message"
	 * fmcsadmin disconnect client -m "Message" -y
	 * fmcsadmin disconnect client -t 90
	 * fmcsadmin disconnect client -t 90 -y
	 * fmcsadmin disconnect client -m "Message" -t 90
	 * fmcsadmin disconnect client -m "Message" -t 90 -y
	 * fmcsadmin disconnect client 1
	 * fmcsadmin disconnect client 1 -y
	 * fmcsadmin disconnect client 1 -m "Message"
	 * fmcsadmin disconnect client 1 -m "Message" -y
	 * fmcsadmin disconnect client 1 -t 90
	 * fmcsadmin disconnect client 1 -t 90 -y
	 * fmcsadmin disconnect client 1 -m "Message" -t 90
	 * fmcsadmin disconnect client 1 -m "Message" -t 90 -y
	 * fmcsadmin --fqdn example.jp disconnect client
	 * fmcsadmin --fqdn example.jp disconnect -y client
	 * fmcsadmin --fqdn example.jp -u USERNAME disconnect client
	 * fmcsadmin --fqdn example.jp -u USERNAME disconnect -y client
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD disconnect client
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD disconnect -y client
	 * etc.
	 */
	expected = []string{"disconnect", "client"}
	args = strings.Split("fmcsadmin disconnect client", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disconnect", "client"}
	args = strings.Split("fmcsadmin disconnect client -y", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disconnect", "client"}
	args = strings.Split("fmcsadmin disconnect client -m Message", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "Message", resultFlags.message)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disconnect", "client"}
	args = strings.Split("fmcsadmin disconnect client -m Message -y", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "Message", resultFlags.message)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disconnect", "client"}
	args = strings.Split("fmcsadmin disconnect client -m Message -t 90", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "Message", resultFlags.message)
	assert.Equal(t, 90, resultFlags.graceTime)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disconnect", "client"}
	args = strings.Split("fmcsadmin disconnect client -m Message -t 90 -y", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "Message", resultFlags.message)
	assert.Equal(t, 90, resultFlags.graceTime)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disconnect", "client", "1"}
	args = strings.Split("fmcsadmin disconnect client 1", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disconnect", "client", "1"}
	args = strings.Split("fmcsadmin disconnect client 1 -y", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disconnect", "client", "1"}
	args = strings.Split("fmcsadmin disconnect client 1 -m Message", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "Message", resultFlags.message)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disconnect", "client", "1"}
	args = strings.Split("fmcsadmin disconnect client 1 -m Message -y", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "Message", resultFlags.message)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disconnect", "client", "1"}
	args = strings.Split("fmcsadmin disconnect client 1 -m Message -t 90", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "Message", resultFlags.message)
	assert.Equal(t, 90, resultFlags.graceTime)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"disconnect", "client", "1"}
	args = strings.Split("fmcsadmin disconnect client 1 -m Message -t 90 -y", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "Message", resultFlags.message)
	assert.Equal(t, 90, resultFlags.graceTime)
	assert.Equal(t, true, resultFlags.yesFlag)
	assert.Equal(t, expected, cmdArgs)

	/*
	 * enable
	 * Usage: fmcsadmin ENABLE [TYPE] [SCHEDULE_NUMBER]
	 *
	 * fmcsadmin enable schedule 2
	 * fmcsadmin -u USERNAME enable schedule 2
	 * fmcsadmin -u USERNAME -p PASSWORD enable schedule 2
	 * fmcsadmin --fqdn example.jp enable schedule 2
	 * fmcsadmin --fqdn example.jp -u USERNAME enable schedule 2
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD enable schedule 2
	 */
	expected = []string{"enable", "schedule", "2"}
	args = strings.Split("fmcsadmin enable schedule 2", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"enable", "schedule", "2"}
	args = strings.Split("fmcsadmin --fqdn example.jp enable schedule 2", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, expected, cmdArgs)

	/*
	 * get
	 * Usage: fmcsadmin GET BACKUPTIME [ID]
	 *        fmcsadmin GET CONFIG_TYPE [NAME1 NAME2 ...]
	 *
	 * fmcsadmin get backuptime
	 * fmcsadmin get backuptime 2
	 * fmcsadmin get serverconfig
	 * fmcsadmin get serverconfig hostedfiles scriptsessions
	 * fmcsadmin get serverconfig scriptsessions hostedfiles
	 * fmcsadmin get cwpconfig
	 * fmcsadmin get cwpconfig enablexml
	 * fmcsadmin get cwpconfig enablephp usefmphp
	 * fmcsadmin get cwpconfig usefmphp enablephp
	 * fmcsadmin get serverprefs
	 * fmcsadmin get serverprefs maxguests maxfiles
	 * fmcsadmin get serverprefs maxfiles maxguests
	 * fmcsadmin get serverprefs AuthenticatedStream
	 * fmcsadmin get serverprefs ParallelBackupEnabled
	 * fmcsadmin get serverprefs PersistCacheEnabled
	 * fmcsadmin get serverprefs SyncPersistCache
	 * fmcsadmin --fqdn example.jp get backuptime
	 * fmcsadmin --fqdn example.jp -u USERNAME get backuptime
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD get backuptime
	 * etc.
	 */

	/*
	 * help
	 * Usage: fmcsadmin HELP COMMANDS
	 *        fmcsadmin HELP [COMMAND]
	 *        fmcsadmin HELP OPTIONS
	 *
	 * fmcsadmin help
	 * fmcsadmin help commands
	 * fmcsadmin help options
	 * fmcsadmin help certificate
	 * fmcsadmin help close
	 * fmcsadmin help delete
	 * fmcsadmin help disable
	 * fmcsadmin help disconnect
	 * fmcsadmin help enable
	 * fmcsadmin help get
	 * fmcsadmin help help
	 * fmcsadmin help list
	 * fmcsadmin help open
	 * fmcsadmin help pause
	 * fmcsadmin help run
	 * fmcsadmin help send
	 * fmcsadmin help set
	 * fmcsadmin help status
	 */
	expected = []string{"help"}
	args = strings.Split("fmcsadmin help", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	/*
	 * list
	 * Usage: fmcsadmin LIST [TYPE] [options]
	 *
	 * fmcsadmin list clients
	 * fmcsadmin list clients -s
	 * fmcsadmin list files
	 * fmcsadmin list files -s
	 * fmcsadmin list plugins
	 * fmcsadmin list plugins -s
	 * fmcsadmin list schedules
	 * fmcsadmin list schedules -s
	 * fmcsadmin --fqdn example.jp list clients
	 * fmcsadmin --fqdn example.jp list clients -s
	 * fmcsadmin --fqdn example.jp list files
	 * fmcsadmin --fqdn example.jp list files -s
	 * fmcsadmin --fqdn example.jp list plugins
	 * fmcsadmin --fqdn example.jp list plugins -s
	 * fmcsadmin --fqdn example.jp list schedules
	 * fmcsadmin --fqdn example.jp list schedules -s
	 * fmcsadmin -u USERNAME list clients
	 * fmcsadmin -u USERNAME list clients -s
	 * fmcsadmin -u USERNAME list files
	 * fmcsadmin -u USERNAME list files -s
	 * fmcsadmin -u USERNAME list plugins
	 * fmcsadmin -u USERNAME list plugins -s
	 * fmcsadmin -u USERNAME list schedules
	 * fmcsadmin -u USERNAME list schedules -s
	 * fmcsadmin -p PASSWORD list clients
	 * fmcsadmin -p PASSWORD list clients -s
	 * fmcsadmin -p PASSWORD list files
	 * fmcsadmin -p PASSWORD list files -s
	 * fmcsadmin -p PASSWORD list plugins
	 * fmcsadmin -p PASSWORD list plugins -s
	 * fmcsadmin -p PASSWORD list schedules
	 * fmcsadmin -p PASSWORD list schedules -s
	 * fmcsadmin -u USERNAME -p PASSWORD list clients
	 * fmcsadmin -u USERNAME -p PASSWORD list clients -s
	 * fmcsadmin -u USERNAME -p PASSWORD list files
	 * fmcsadmin -u USERNAME -p PASSWORD list files -s
	 * fmcsadmin -u USERNAME -p PASSWORD list plugins
	 * fmcsadmin -u USERNAME -p PASSWORD list plugins -s
	 * fmcsadmin -u USERNAME -p PASSWORD list schedules
	 * fmcsadmin -u USERNAME -p PASSWORD list schedules -s
	 */
	// list clients
	expected = []string{"list", "clients"}
	args = strings.Split("fmcsadmin list clients", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, false, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "clients"}
	args = strings.Split("fmcsadmin -s list clients", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "clients"}
	args = strings.Split("fmcsadmin --stats list clients", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "clients"}
	args = strings.Split("fmcsadmin list clients -s", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "clients"}
	args = strings.Split("fmcsadmin list clients --stats", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "clients"}
	args = strings.Split("fmcsadmin list -s clients", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "clients"}
	args = strings.Split("fmcsadmin list --stats clients", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	// list files
	expected = []string{"list", "files"}
	args = strings.Split("fmcsadmin list files", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, false, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "files"}
	args = strings.Split("fmcsadmin -s list files", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "files"}
	args = strings.Split("fmcsadmin --stats list files", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "files"}
	args = strings.Split("fmcsadmin list files -s", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "files"}
	args = strings.Split("fmcsadmin list files --stats", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "files"}
	args = strings.Split("fmcsadmin list -s files", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "files"}
	args = strings.Split("fmcsadmin list --stats files", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	// list plugins
	expected = []string{"list", "plugins"}
	args = strings.Split("fmcsadmin list plugins", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, false, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "plugins"}
	args = strings.Split("fmcsadmin -s list plugins", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "plugins"}
	args = strings.Split("fmcsadmin --stats list plugins", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "plugins"}
	args = strings.Split("fmcsadmin list plugins -s", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "plugins"}
	args = strings.Split("fmcsadmin list plugins --stats", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "plugins"}
	args = strings.Split("fmcsadmin list -s plugins", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "plugins"}
	args = strings.Split("fmcsadmin list --stats plugins", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	// list schedules
	expected = []string{"list", "schedules"}
	args = strings.Split("fmcsadmin list schedules", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, false, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "schedules"}
	args = strings.Split("fmcsadmin -s list schedules", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "schedules"}
	args = strings.Split("fmcsadmin --stats list schedules", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "schedules"}
	args = strings.Split("fmcsadmin list schedules -s", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "schedules"}
	args = strings.Split("fmcsadmin list schedules --stats", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "schedules"}
	args = strings.Split("fmcsadmin list -s schedules", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"list", "schedules"}
	args = strings.Split("fmcsadmin list --stats schedules", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, true, resultFlags.statsFlag)
	assert.Equal(t, expected, cmdArgs)

	/*
	 * open
	 * Usage: fmcsadmin OPEN [options] [FILE...] [PATH...]
	 *
	 * fmcsadmin open
	 * fmcsadmin open 1
	 * fmcsadmin open 1 2
	 * fmcsadmin open TestDB
	 * fmcsadmin open TestDB.fmp12
	 * fmcsadmin open "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"
	 * fmcsadmin open "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"
	 * fmcsadmin open "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"
	 * fmcsadmin open "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"
	 * fmcsadmin open "/opt/FileMaker/FileMaker Server/Data/Databases/"
	 * fmcsadmin open "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/"
	 * fmcsadmin open TestDB FMServer_Sample
	 * fmcsadmin --fqdn example.jp open TestDB
	 * fmcsadmin open -y TestDB
	 * fmcsadmin open -u USERNAME TestDB
	 * fmcsadmin open -p PASSWORD TestDB
	 * fmcsadmin open -u USERNAME -p PASSWORD TestDB
	 * fmcsadmin open --key ENCRYPTPASS TestDB
	 * fmcsadmin open --key ENCRYPTPASS --savekey TestDB
	 */
	expected = []string{"open", "TestDB"}
	args = strings.Split("fmcsadmin open TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "", resultFlags.key)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"open", "TestDB"}
	args = strings.Split("fmcsadmin --key KEY open TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "KEY", resultFlags.key)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"open", "TestDB"}
	args = strings.Split("fmcsadmin open TestDB --key KEY", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "KEY", resultFlags.key)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"open", "TestDB"}
	args = strings.Split("fmcsadmin open --key KEY TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "KEY", resultFlags.key)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"open", "TestDB"}
	args = strings.Split("fmcsadmin open --key KEY TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "KEY", resultFlags.key)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"open"}
	args = strings.Split("fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD open --key KEY", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, "USERNAME", resultFlags.username)
	assert.Equal(t, "PASSWORD", resultFlags.password)
	assert.Equal(t, "KEY", resultFlags.key)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"open"}
	args = strings.Split("fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD open --key KEY --savekey", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, "USERNAME", resultFlags.username)
	assert.Equal(t, "PASSWORD", resultFlags.password)
	assert.Equal(t, "KEY", resultFlags.key)
	assert.Equal(t, expected, cmdArgs)

	/*
	 * pause
	 * Usage: fmcsadmin PAUSE [FILE...] [PATH...]
	 *
	 * fmcsadmin pause
	 * fmcsadmin pause 1
	 * fmcsadmin pause 1 2
	 * fmcsadmin pause TestDB
	 * fmcsadmin pause TestDB.fmp12
	 * fmcsadmin pause "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"
	 * fmcsadmin pause "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"
	 * fmcsadmin pause "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"
	 * fmcsadmin pause "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"
	 * fmcsadmin pause "/opt/FileMaker/FileMaker Server/Data/Databases/"
	 * fmcsadmin pause "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/"
	 * fmcsadmin pause TestDB FMServer_Sample
	 * fmcsadmin --fqdn example.jp pause TestDB
	 * fmcsadmin pause -u USERNAME TestDB
	 * fmcsadmin pause -p PASSWORD TestDB
	 * fmcsadmin pause -u USERNAME -p PASSWORD TestDB
	 */
	expected = []string{"pause", "TestDB"}
	args = strings.Split("fmcsadmin pause TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"pause", "TestDB", "FMServer_Sample"}
	args = strings.Split("fmcsadmin pause TestDB FMServer_Sample", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	/*
	 * remove
	 * Usage: fmcsadmin REMOVE [FILE...] [PATH...]
	 *
	 * fmcsadmin remove
	 * fmcsadmin remove 1
	 * fmcsadmin remove 1 2
	 * fmcsadmin remove TestDB
	 * fmcsadmin remove TestDB.fmp12
	 * fmcsadmin remove 11.fmp12
	 * fmcsadmin remove "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"
	 * fmcsadmin remove "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"
	 * fmcsadmin remove "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"
	 * fmcsadmin remove "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"
	 * fmcsadmin remove "/opt/FileMaker/FileMaker Server/Data/Databases/"
	 * fmcsadmin remove "filemac:/Macintosh HD/Library/FileMaker Server/Data/Databases/"
	 * fmcsadmin remove "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/"
	 * fmcsadmin remove TestDB FMServer_Sample
	 * fmcsadmin --fqdn example.jp remove TestDB
	 * fmcsadmin remove -u USERNAME TestDB
	 * fmcsadmin remove -p PASSWORD TestDB
	 * fmcsadmin remove -u USERNAME -p PASSWORD TestDB
	 * fmcsadmin remove -u USERNAME -p PASSWORD TestDB -y
	 */
	expected = []string{"remove", "TestDB"}
	args = strings.Split("fmcsadmin remove TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"remove", "TestDB", "FMServer_Sample"}
	args = strings.Split("fmcsadmin remove TestDB FMServer_Sample", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	/*
	 * resume
	 * Usage: fmcsadmin RESUME [FILE...] [PATH...]
	 *
	 * fmcsadmin resume
	 * fmcsadmin resume 1
	 * fmcsadmin resume 1 2
	 * fmcsadmin resume TestDB
	 * fmcsadmin resume TestDB.fmp12
	 * fmcsadmin resume "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"
	 * fmcsadmin resume "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"
	 * fmcsadmin resume "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"
	 * fmcsadmin resume "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"
	 * fmcsadmin resume "/opt/FileMaker/FileMaker Server/Data/Databases/"
	 * fmcsadmin resume "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/"
	 * fmcsadmin resume TestDB FMServer_Sample
	 * fmcsadmin --fqdn example.jp resume TestDB
	 * fmcsadmin resume -u USERNAME TestDB
	 * fmcsadmin resume -p PASSWORD TestDB
	 * fmcsadmin resume -u USERNAME -p PASSWORD TestDB
	 */
	expected = []string{"resume", "TestDB"}
	args = strings.Split("fmcsadmin resume TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"resume", "TestDB", "FMServer_Sample"}
	args = strings.Split("fmcsadmin resume TestDB FMServer_Sample", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	/*
	 * run
	 * Usage: fmcsadmin RUN SCHEDULE [SCHEDULE_NUMBER]
	 *
	 * fmcsadmin run schedule 2
	 * fmcsadmin --fqdn example.jp run schedule 2
	 * fmcsadmin --fqdn example.jp -u USERNAME run schedule 2
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD run schedule 2
	 */
	expected = []string{"run", "schedule", "2"}
	args = strings.Split("fmcsadmin run schedule 2", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	/*
	 * send
	 * Usage: fmcsadmin SEND [options] [CLIENT_NUMBER] [FILE...] [PATH...]
	 *
	 * fmcsadmin send -m "This is a test message"
	 * fmcsadmin send -c 2 -m "This is a test message"
	 * etc.
	 */

	/*
	 * set
	 * Usage: fmcsadmin SET CONFIG_TYPE [NAME1=VALUE1 NAME2=VALUE2 ...]
	 *
	 * fmcsadmin set serverconfig hostedfiles=125 scriptsessions=100
	 * fmcsadmin set cwpconfig enablephp=true
	 * fmcsadmin set cwpconfig enablexml=true
	 * fmcsadmin set cwpconfig encoding=UTF-8
	 * fmcsadmin set cwpconfig encoding=ISO-8859-1
	 * fmcsadmin set cwpconfig locale=en
	 * fmcsadmin set cwpconfig locale=ja
	 * fmcsadmin set cwpconfig prevalidation=false
	 * fmcsadmin set cwpconfig enablephp=true usefmphp=false
	 * fmcsadmin set cwpconfig enablephp=true usefmphp=true
	 * fmcsadmin --fqdn example.jp set cwpconfig enablephp=true usefmphp=true
	 * fmcsadmin --fqdn example.jp -u USERNAME set cwpconfig enablephp=true usefmphp=true
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD set cwpconfig enablephp=true usefmphp=true
	 * fmcsadmin set serverprefs maxguests=125 maxfiles=125
	 * fmcsadmin set serverprefs AuthenticatedStream=1
	 * fmcsadmin set serverprefs AuthenticatedStream=2
	 * fmcsadmin set serverprefs ParallelBackupEnabled=false
	 * fmcsadmin set serverprefs ParallelBackupEnabled=true
	 * fmcsadmin --fqdn example.jp set serverconfig hostedfiles=125 scriptsessions=100
	 * fmcsadmin --fqdn example.jp -u USERNAME set serverconfig hostedfiles=125 scriptsessions=100
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD set serverconfig hostedfiles=125 scriptsessions=100
	 * etc.
	 */

	/*
	 * status
	 * Usage: fmcsadmin STATUS [TYPE] [CLIENT_NUMBER] [FILE...]
	 *
	 * fmcsadmin status client 1
	 * fmcsadmin status file TestDB
	 * fmcsadmin status file "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"
	 * fmcsadmin status file "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"
	 * fmcsadmin status file "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"
	 * fmcsadmin status file "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"
	 * fmcsadmin status file TestDB FMServer_Sample
	 */
	expected = []string{"status", "client", "1"}
	args = strings.Split("fmcsadmin status client 1", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"status", "file", "TestDB"}
	args = strings.Split("fmcsadmin status file TestDB", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)
}

func TestOutputInvalidCommandErrorMessage(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}
	status := outputInvalidCommandErrorMessage(cli)
	assert.Equal(t, 248, status)
}

func TestGetBaseURI(t *testing.T) {
	if runtime.GOOS == "linux" {
		assert.Equal(t, "http://127.0.0.1:16001", getBaseURI(""))
	} else if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		assert.Equal(t, "http://127.0.0.1:16001", getBaseURI(""))
	}
	assert.Equal(t, "https://example.jp", getBaseURI("example.jp"))
	assert.Equal(t, "https://example.jp", getBaseURI("example.jp "))
	assert.Equal(t, "https://example.jp", getBaseURI(" example.jp"))
}

func TestGetAPIBasePath(t *testing.T) {
	assert.Equal(t, "/fmi/admin/api/v2", getAPIBasePath("http://127.0.0.1:16001"))
	assert.Equal(t, "/fmi/admin/api/v2", getAPIBasePath("https://example.jp"))
}

func TestGetServerVersionAsFloat(t *testing.T) {
	version, _ := getServerVersionAsFloat("19.3.1")
	assert.Equal(t, 19.3, version)
	version, _ = getServerVersionAsFloat("19.3.2")
	assert.Equal(t, 19.3, version)
}

func TestComparePath(t *testing.T) {
	assert.Equal(t, false, comparePath("TestDB", "TestDB2"))

	assert.Equal(t, true, comparePath("TestDB", "TestDB"))
	assert.Equal(t, true, comparePath("TestDB.fmp12", "TestDB.fmp12"))
	assert.Equal(t, true, comparePath("TestDB", "TestDB.fmp12"))
	assert.Equal(t, true, comparePath("TestDB.fmp12", "TestDB"))

	assert.Equal(t, true, comparePath("filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/", "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/"))

	assert.Equal(t, true, comparePath("filemac:/opt/FileMaker/FileMaker Server/Data/Databases/", "filemac:/opt/FileMaker/FileMaker Server/Data/Databases/"))

	assert.Equal(t, true, comparePath("filewin:/opt/FileMaker/FileMaker Server/Data/Databases/", "filewin:/opt/FileMaker/FileMaker Server/Data/Databases/"))

	assert.Equal(t, true, comparePath("filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))
	assert.Equal(t, true, comparePath("filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))

	assert.Equal(t, true, comparePath("filemac:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filemac:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))
	assert.Equal(t, true, comparePath("filemac:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "filemac:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("filemac:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filemac:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("filemac:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "filemac:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))

	assert.Equal(t, true, comparePath("filemac:/Macintosh HD/Library/FileMaker Server/Data/Databases/test/TestDB", "/Volumes/Macintosh HD/Library/FileMaker Server/Data/Databases/test/TestDB"))
	assert.Equal(t, true, comparePath("filemac:/Macintosh HD/Library/FileMaker Server/Data/Databases/test/TestDB", "/Volumes/Macintosh HD/Library/FileMaker Server/Data/Databases/test/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("filemac:/Macintosh HD/Library/FileMaker Server/Data/Databases/test/TestDB.fmp12", "/Volumes/Macintosh HD/Library/FileMaker Server/Data/Databases/test/TestDB"))
	assert.Equal(t, true, comparePath("filemac:/Macintosh HD/Library/FileMaker Server/Data/Databases/test/TestDB.fmp12", "/Volumes/Macintosh HD/Library/FileMaker Server/Data/Databases/test/TestDB.fmp12"))

	assert.Equal(t, true, comparePath("filewin:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filewin:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))
	assert.Equal(t, true, comparePath("filewin:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "filewin:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("filewin:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filewin:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("filewin:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "filewin:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))

	assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))
	assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))

	assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))
	assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))
	assert.Equal(t, true, comparePath("filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))

	if runtime.GOOS == "darwin" {
		assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filemac:/"+getVolumeName()+"/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))
		assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filemac:/"+getVolumeName()+"/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
		assert.Equal(t, true, comparePath("filemac:/"+getVolumeName()+"/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))
		assert.Equal(t, true, comparePath("filemac:/"+getVolumeName()+"/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	}

	assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filewin:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))
	assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filewin:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("filewin:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))
	assert.Equal(t, true, comparePath("filewin:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
}

func TestGetErrorDescription(t *testing.T) {
	assert.Equal(t, "", getErrorDescription(0))
	assert.Equal(t, "Unknown error", getErrorDescription(-1))
	assert.Equal(t, "Unavailable command", getErrorDescription(3))
	assert.Equal(t, "Command is unknown", getErrorDescription(4))
	assert.Equal(t, "Empty result", getErrorDescription(8))
	assert.Equal(t, "Access denied", getErrorDescription(9))
	assert.Equal(t, "Not Supported", getErrorDescription(21))
	assert.Equal(t, "Invalid user account and/or password; please try again", getErrorDescription(212))
	assert.Equal(t, "Unable to open the file", getErrorDescription(802))
	assert.Equal(t, "Parameter missing", getErrorDescription(958))
	assert.Equal(t, "Parameter is invalid", getErrorDescription(960))
	assert.Equal(t, "Parameter value is invalid", getErrorDescription(1708))
	assert.Equal(t, "Service already running", getErrorDescription(10006))
	assert.Equal(t, "Schedule at specified index does not exist", getErrorDescription(10600))
	assert.Equal(t, "Schedule is misconfigured; invalid taskType or run status", getErrorDescription(10601))
	assert.Equal(t, "Schedule can't be created or duplicated", getErrorDescription(10603))
	assert.Equal(t, "Cannot enable schedule", getErrorDescription(10604))
	assert.Equal(t, "No schedules created in configuration file", getErrorDescription(10610))
	assert.Equal(t, "Schedule name is already used", getErrorDescription(10611))
	assert.Equal(t, "No applicable files for this operation", getErrorDescription(10904))
	assert.Equal(t, "Script is missing", getErrorDescription(10906))
	assert.Equal(t, "System script aborted", getErrorDescription(10908))
	assert.Equal(t, "Invalid command", getErrorDescription(11000))
	assert.Equal(t, "Unable to create command", getErrorDescription(11002))
	assert.Equal(t, "Disconnect Client invalid ID", getErrorDescription(11005))
	assert.Equal(t, "File permission error", getErrorDescription(20402))
	assert.Equal(t, "File not found or not accessible.", getErrorDescription(20405))
	assert.Equal(t, "File already exists", getErrorDescription(20406))
	assert.Equal(t, "File read error", getErrorDescription(20408))
	assert.Equal(t, "SSL certificate expired", getErrorDescription(20630))
	assert.Equal(t, "SSL certificate verification error", getErrorDescription(20632))
	assert.Equal(t, "Parameters are invalid", getErrorDescription(25004))
	assert.Equal(t, "Invalid session error", getErrorDescription(25006))
}

func TestGetDateTimeStringOfCurrentTimeZone(t *testing.T) {
	const location = "Asia/Tokyo"
	loc, err := time.LoadLocation(location)
	if err != nil {
		loc = time.FixedZone(location, 9*60*60)
	}
	time.Local = loc
	assert.Equal(t, "", getDateTimeStringOfCurrentTimeZone("", "2006/01/02 15:04:05", false))
	assert.Equal(t, "", getDateTimeStringOfCurrentTimeZone("0000-00-00 00:00:00", "2006/01/02 15:04:05", false))
	assert.Equal(t, "", getDateTimeStringOfCurrentTimeZone("0000-00-00 00:00:00 GMT", "2006/01/02 15:04:05", false))

	//assert.Equal(t, "2006/01/03 00:04", getDateTimeStringOfCurrentTimeZone("2006-01-02T15:04:05"))
	assert.Equal(t, "2006/01/02 15:04", getDateTimeStringOfCurrentTimeZone("2006-01-02T15:04:05", "2006/01/02 15:04", false))
	assert.Equal(t, "2006/01/03 00:04", getDateTimeStringOfCurrentTimeZone("2006-01-02T15:04:05", "2006/01/02 15:04", true))
	assert.Equal(t, "2006/01/02 15:04:05", getDateTimeStringOfCurrentTimeZone("2006-01-02T15:04:05", "2006/01/02 15:04:05", false))

	//assert.Equal(t, "2006/01/03 00:04", getDateTimeStringOfCurrentTimeZone("2006-01-02 15:04:05 GMT"))
	assert.Equal(t, "2006/01/02 15:04", getDateTimeStringOfCurrentTimeZone("2006-01-02 15:04:05 GMT", "2006/01/02 15:04", false))
	assert.Equal(t, "2006/01/03 00:04", getDateTimeStringOfCurrentTimeZone("2006-01-02 15:04:05 GMT", "2006/01/02 15:04", true))
	assert.Equal(t, "2006/01/02 15:04:05", getDateTimeStringOfCurrentTimeZone("2006-01-02 15:04:05 GMT", "2006/01/02 15:04:05", false))
}
