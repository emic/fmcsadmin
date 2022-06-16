fmcsadmin [![Build Status](https://github.com/emic/fmcsadmin/actions/workflows/go.yml/badge.svg)](https://github.com/emic/fmcsadmin/actions/workflows/go.yml)
=========
fmcsadmin is a command line tool to administer the Database Server component of Claris FileMaker Server via Claris FileMaker Admin API. fmcsadmin supports remote server administration.

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
- Start a server process (for FileMaker Server)
- Restart a server process (for FileMaker Server)
- Stop a server process (for FileMaker Server)
- Retrieve server or CWP configuration settings
- Change server or CWP configuration settings
- List plug-ins (for FileMaker Server 19.2.1 or later)
- Manage SSL certificates (for FileMaker Server 19.2.1 or later)
- Move databases out of hosted folder (for FileMaker Server 19.3.1 or later)
- View and change the setting for sharing streaming URLs (for FileMaker Server 19.3.2 or later)
- View the server parallel backup setting (for FileMaker Server 19.5.1 or later)

Supported Servers
-----
- Claris FileMaker Server 19 (19.1, 19.2, 19.3, 19.4, 19.5)

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
Note: Handling schedule ID 1 is not supported for FileMaker Server.

System Requirements
-----
- Linux version   : Ubuntu 18.04 LTS, Ubuntu 20.04 LTS or CentOS Linux 7
- macOS version   : macOS Catalina 10.15.7 or later (tested on macOS Monterey 12)
- Windows version : Windows 10 Version 21H1 or later (tested on Windows 11)

Download
-----
Download from the [latest release page](https://github.com/emic/fmcsadmin/releases/latest).

Author
-----
Emic Corporation <https://www.emic.co.jp/>

License
-----
This software is distributed under the [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0), see LICENSE.txt and NOTICE.txt for more information.