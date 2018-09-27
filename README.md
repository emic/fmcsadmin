fmcsadmin [![Build Status](https://travis-ci.org/emic/fmcsadmin.svg?branch=master)](https://travis-ci.org/emic/fmcsadmin)
=========
fmcsadmin is the command line tool to administer the Database Server component of FileMaker Cloud and FileMaker Server via FileMaker Admin API. FileMaker is a trademark of FileMaker, Inc., registered in the U.S. and other countries.

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
- Start a server process (for FileMaker Server 17)
- Restart a server process (for FileMaker Server 17)
- Stop a server process (for FileMaker Server 17)
- Retrieve server or CWP configuration settings (for FileMaker Server 17)
- Change server or CWP configuration settings (for FileMaker Server 17)

Supported Servers
-----
- FileMaker Cloud 1.17 (FileMaker Admin API (Trial) in FileMaker Cloud 1.17 will expire on September 27, 2019)
- FileMaker Server 17 (FileMaker Admin API (Trial) in FileMaker Server 17 will expire on September 27, 2019)

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
Note: "--fqdn" option and handling schedule ID 1 is not supported for FileMaker Server 17.

Author
-----
Emic Corporation <https://www.emic.co.jp/>


License
-----
This software is distributed under the
[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0),
see LICENSE.txt and NOTICE.txt for more information.
