fmcsadmin [![Build Status](https://travis-ci.org/emic/fmcsadmin.svg?branch=master)](https://travis-ci.org/emic/fmcsadmin)
=========
fmcsadmin is the command line tool to administer the Database Server component of FileMaker Cloud via FileMaker Admin API. FileMaker is a trademark of FileMaker, Inc., registered in the U.S. and other countries.

Installing Source Code
-----
```
go get github.com/emic/fmcsadmin
```
Note: You need to install Go (not "FileMaker Go").

Features
-----
- Close a database
- Disable a schedule
- Disconnect clients
- Enable a schedule
- List databases
- Get schedules
- Open a database
- Pause a database
- Resume a database
- Run a schedule
- Send a message to clients

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

Author
-----
Emic Corporation <https://www.emic.co.jp/>


License
-----
This software is distributed under the
[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0),
see LICENSE.txt and NOTICE.txt for more information.
