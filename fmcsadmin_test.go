package main

import (
	"bytes"
	"strings"
	"testing"

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

	args = strings.Split("fmcsadmin list", " ")
	status = cli.Run(args)
	assert.Equal(t, 248, status)

	args = strings.Split("fmcsadmin list unknwon", " ")
	status = cli.Run(args)
	assert.Equal(t, 248, status)
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

func TestRunWithVersionOption1(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin -v", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "fmcsadmin 0.9.0"
	assert.Contains(t, outStream.String(), expected)
}

func TestRunWithVersionOption2(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin --version", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "fmcsadmin 0.9.0"
	assert.Contains(t, outStream.String(), expected)
}

/*
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
*/

func TestRunDisableCommand1(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin disable", " ")
	status := cli.Run(args)
	assert.Equal(t, 248, status)
	expected := "Error: 11000 (Invalid command)"
	assert.Contains(t, outStream.String(), expected)
}

/*
func TestRunDisableCommand2(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin -y disable schedule", " ")
	status := cli.Run(args)
	assert.Equal(t, 104, status)
	expected := "Error: 10600 (Schedule at specified index no longer exists)"
	assert.Contains(t, outStream.String(), expected)
}
*/

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
	expected := "Error: 10600 (Schedule at specified index no longer exists)"
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
	expected := "Error: 10600 (Schedule at specified index no longer exists)"
	assert.Contains(t, outStream.String(), expected)
}
*/

func TestRunShowCloseCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help close", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin CLOSE [FILE...] [PATH...] [options]"
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

func TestRunShowStatusCommandHelp(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &cli{outStream: outStream, errStream: errStream}

	args := strings.Split("fmcsadmin help status", " ")
	status := cli.Run(args)
	assert.Equal(t, 0, status)
	expected := "Usage: fmcsadmin STATUS [TYPE] [CLIENT_NUMBER] [FILE...]"
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
	 * (Not Implemented)
	 */

	/*
	 * disable
	 * Usage: fmcsadmin DISABLE [TYPE] [SCHEDULE_NUMBER]
	 *
	 * fmcsadmin disable schedule 1
	 * fmcsadmin disable -y schedule 1
	 * fmcsadmin --fqdn example.jp disable schedule 1
	 * fmcsadmin --fqdn example.jp disable -y schedule 1
	 * fmcsadmin --fqdn example.jp -u USERNAME disable schedule 1
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD disable schedule 1
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD -y disable schedule 1
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
	 * [WIP] --fqdn
	 * fmcsadmin disconnect client 1
	 * fmcsadmin disconnect client 1 -y
	 * fmcsadmin disconnect client 1 -m "Message"
	 * fmcsadmin disconnect client 1 -m "Message" -y
	 * fmcsadmin disconnect client 1 -t 90
	 * fmcsadmin disconnect client 1 -t 90 -y
	 * fmcsadmin disconnect client 1 -m "Message" -t 90
	 * fmcsadmin disconnect client 1 -m "Message" -t 90 -y
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
	 * fmcsadmin enable schedule 1
	 * fmcsadmin --fqdn example.jp enable schedule 1
	 * fmcsadmin --fqdn example.jp -u USERNAME enable schedule 1
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD enable schedule 1
	 */
	expected = []string{"enable", "schedule", "1"}
	args = strings.Split("fmcsadmin enable schedule 1", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	expected = []string{"enable", "schedule", "1"}
	args = strings.Split("fmcsadmin --fqdn example.jp enable schedule 1", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, "example.jp", resultFlags.fqdn)
	assert.Equal(t, expected, cmdArgs)
	/* [WIP] */

	/*
	 * help
	 * Usage: fmcsadmin HELP COMMANDS
	 *        fmcsadmin HELP [COMMAND]
	 *        fmcsadmin HELP OPTIONS
	 *
	 * fmcsadmin help
	 * fmcsadmin help commands
	 * fmcsadmin help options
	 * fmcsadmin help close
	 * (Not Implemented) fmcsadmin help delete
	 * fmcsadmin help disable
	 * fmcsadmin help disconnect
	 * fmcsadmin help enable
	 * fmcsadmin help help
	 * fmcsadmin help list
	 * fmcsadmin help open
	 * fmcsadmin help pause
	 * fmcsadmin help run
	 * fmcsadmin help send
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
	 * fmcsadmin list schedules
	 * fmcsadmin list schedules -s
	 * fmcsadmin --fqdn example.jp list clients
	 * fmcsadmin --fqdn example.jp list clients -s
	 * fmcsadmin --fqdn example.jp list files
	 * fmcsadmin --fqdn example.jp list files -s
	 * fmcsadmin --fqdn example.jp list schedules
	 * fmcsadmin --fqdn example.jp list schedules -s
	 * fmcsadmin open -u USERNAME list clients
	 * fmcsadmin open -u USERNAME list clients -s
	 * fmcsadmin open -u USERNAME list files
	 * fmcsadmin open -u USERNAME list files -s
	 * fmcsadmin open -u USERNAME list schedules
	 * fmcsadmin open -u USERNAME list schedules -s
	 * fmcsadmin open -p PASSWORD list clients
	 * fmcsadmin open -p PASSWORD list clients -s
	 * fmcsadmin open -p PASSWORD list files
	 * fmcsadmin open -p PASSWORD list files -s
	 * fmcsadmin open -p PASSWORD list schedules
	 * fmcsadmin open -p PASSWORD list schedules -s
	 * fmcsadmin open -u USERNAME -p PASSWORD list clients
	 * fmcsadmin open -u USERNAME -p PASSWORD list clients -s
	 * fmcsadmin open -u USERNAME -p PASSWORD list files
	 * fmcsadmin open -u USERNAME -p PASSWORD list files -s
	 * fmcsadmin open -u USERNAME -p PASSWORD list schedules
	 * fmcsadmin open -u USERNAME -p PASSWORD list schedules -s
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
	 * fmcsadmin run schedule 1
	 * fmcsadmin --fqdn example.jp run schedule 1
	 * fmcsadmin --fqdn example.jp -u USERNAME run schedule 1
	 * fmcsadmin --fqdn example.jp -u USERNAME -p PASSWORD run schedule 1
	 */
	expected = []string{"run", "schedule", "1"}
	args = strings.Split("fmcsadmin run schedule 1", " ")
	cmdArgs, resultFlags, _ = getFlags(args, flags)
	assert.Equal(t, expected, cmdArgs)

	/*
	 * send
	 * Usage: fmcsadmin SEND [options] [CLIENT_NUMBER] [FILE...] [PATH...]
	 *
	 * fmcsadmin send -m "This is a test message"
	 * fmcsadmin send -c 2 -m "This is a test message"
	 * [WIP] ...
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

func TestGetHostName(t *testing.T) {
	assert.Equal(t, "http://127.0.0.1:8080", getHostName(""))
	assert.Equal(t, "https://example.jp", getHostName("example.jp"))
}

func TestGetAPIBasePath(t *testing.T) {
	assert.Equal(t, "/admin/api/v1", getAPIBasePath())
}

func TestComparePath(t *testing.T) {
	assert.Equal(t, false, comparePath("TestDB", "TestDB2"))

	assert.Equal(t, true, comparePath("TestDB", "TestDB"))
	assert.Equal(t, true, comparePath("TestDB.fmp12", "TestDB.fmp12"))
	assert.Equal(t, true, comparePath("TestDB", "TestDB.fmp12"))
	assert.Equal(t, true, comparePath("TestDB.fmp12", "TestDB"))

	assert.Equal(t, true, comparePath("filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/", "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/"))

	assert.Equal(t, true, comparePath("filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))
	assert.Equal(t, true, comparePath("filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))

	assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))
	assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))

	assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))
	assert.Equal(t, true, comparePath("/opt/FileMaker/FileMaker Server/Data/Databases/TestDB", "filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
	assert.Equal(t, true, comparePath("filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB"))
	assert.Equal(t, true, comparePath("filelinux:/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12", "/opt/FileMaker/FileMaker Server/Data/Databases/TestDB.fmp12"))
}

func TestGetErrorDescription(t *testing.T) {
	assert.Equal(t, "", getErrorDescription(0))
	assert.Equal(t, "Empty result", getErrorDescription(8))
	assert.Equal(t, "Insufficient privileges", getErrorDescription(9))
	assert.Equal(t, "Invalid user account and/or password; please try again", getErrorDescription(212))
	assert.Equal(t, "Unable to open the file", getErrorDescription(802))
	assert.Equal(t, "Parameter missing", getErrorDescription(958))
	assert.Equal(t, "Parameter is invalid", getErrorDescription(960))
	assert.Equal(t, "Schedule at specified index no longer exists", getErrorDescription(10600))
	assert.Equal(t, "Schedule is misconfigured; invalid taskType or run status", getErrorDescription(10601))
	assert.Equal(t, "Schedule can't be created or duplicated", getErrorDescription(10603))
	assert.Equal(t, "Cannot enable schedule", getErrorDescription(10604))
	assert.Equal(t, "No schedules created in configuration file", getErrorDescription(10610))
	assert.Equal(t, "Schedule name is already used", getErrorDescription(10611))
	assert.Equal(t, "No applicable files for this operation", getErrorDescription(10904))
	assert.Equal(t, "Script is missing", getErrorDescription(10906))
	assert.Equal(t, "Disconnect client invalid ID", getErrorDescription(11005))
	assert.Equal(t, "Parameters are invalid", getErrorDescription(25004))
	assert.Equal(t, "Invalid session error", getErrorDescription(25006))
}