fmcsadmin [![Build Status](https://travis-ci.org/emic/fmcsadmin.svg?branch=master)](https://travis-ci.org/emic/fmcsadmin)
=========
fmcsadmin is the command line tool to administer the Database Server component of FileMaker Cloud for AWS and FileMaker Server via FileMaker Admin API.

Installing Source Code
-----
```
go get github.com/emic/fmcsadmin
```
Note: You need to install Go (not "FileMaker Go").

Features
-----
- Close databases
- Disable schedules
- Disconnect clients
- Enable schedules
- List clients, databases, or schedules
- Open databases
- Temporarily stop database access
- Make paused databases available
- Run a schedule
- Send a message to clients
- Start a server process (for FileMaker Server 18)
- Restart a server process (for FileMaker Server 18)
- Stop a server process (for FileMaker Server 18)
- Retrieve server or CWP configuration settings (for FileMaker Server 18)
- Change server or CWP configuration settings (for FileMaker Server 18)

Supported Servers
-----
- FileMaker Cloud for AWS 1.18
- FileMaker Server 18

Usage
-----
You can script many tasks with fmcsadmin by using a scripting language that allows execution of shell or terminal commands.

```
    fmcsadmin HELP COMMANDS
       Lists available commands

    fmcsadmin HELP [COMMAND]
       Displays help on the specified COMMAND

    fmcsadmin HELP OPTIONS
       Lists available options
```
Note: "--fqdn" option and handling schedule ID 1 is not supported for FileMaker Server.

Author
-----
Emic Corporation <https://www.emic.co.jp/>


License
-----
This software is distributed under the MIT License, see LICENSE.txt and NOTICE.txt for more information.