fmcsadmin [![Build Status](https://travis-ci.org/emic/fmcsadmin.svg?branch=master)](https://travis-ci.org/emic/fmcsadmin)
=========
fmcsadmin is the command line tool to administer the Database Server component of Claris FileMaker Server and Claris FileMaker Cloud for AWS 1.18 via Claris FileMaker Admin API. fmcsadmin supports remote server administration.

Features
-----
- Close databases
- Delete a schedule
- Disable schedules
- Disconnect clients
- Enable schedules
- List clients, databases or schedules
- Open databases
- Temporarily stop database access
- Make paused databases available
- Run a schedule
- Send a message to clients
- Start a server process (for Claris FileMaker Server)
- Restart a server process (for Claris FileMaker Server)
- Stop a server process (for Claris FileMaker Server)
- Retrieve server or CWP configuration settings (for Claris FileMaker Server)
- Change server or CWP configuration settings (for Claris FileMaker Server)
- List plug-ins (for Claris FileMaker Server 19.2.1 or later)
- Manage SSL certificates (for Claris FileMaker Server 19.2.1 or later)
- Move databases out of hosted folder (for Claris FileMaker Server 19.3.1 or later)

Supported Servers
-----
- Claris FileMaker Server 19 (19.0, 19.1, 19.2, 19.3)
- Claris FileMaker Server 18
- Claris FileMaker Cloud for AWS 1.18

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
Note: Handling schedule ID 1 is not supported for Claris FileMaker Server.

System Requirements
-----
- Ubuntu 18.04 LTS
- CentOS Linux 7 or later
- macOS Mojave 10.14.6 or later
- Windows 10

Download
-----
Download from the [latest release page](https://github.com/emic/fmcsadmin/releases/latest).

Installing Source Code
-----
```
go get github.com/emic/fmcsadmin
```
Note: You need to install Go (not "Claris FileMaker Go").

Author
-----
Emic Corporation <https://www.emic.co.jp/>

License
-----
This software is distributed under the MIT License, see LICENSE.txt and NOTICE.txt for more information.