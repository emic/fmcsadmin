fmcsadmin [![Build Status](https://github.com/emic/fmcsadmin/actions/workflows/go.yml/badge.svg)](https://github.com/emic/fmcsadmin/actions/workflows/go.yml)
=========
fmcsadmin is a command line tool to administer the Database Server component of Claris FileMaker Server via Claris FileMaker Admin API. fmcsadmin supports public key authentication and remote server administration.

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
- Start a server process
- Restart a server process
- Stop a server process
- Retrieve server or CWP configuration settings
- Change server or CWP configuration settings
- List plug-ins
- Manage SSL certificates
- Move databases out of hosted folder
- View and change the setting for sharing streaming URLs
- Cancel the currently running backup
- View and change the setting for parallel backup
- FileMaker Admin API PKI Authentication
- View and change the settings for the persistent cache (for FileMaker Server 2024)
- View and change the setting for blocking new users (for FileMaker Server 2024)
- View the setting for opening the databases that were opened when the database server was last shut down (for FileMaker Server 2024 (21.1))

Supported Servers
-----
Please see details: https://support.claris.com/s/article/Claris-support-policy?language=en_US
- Claris FileMaker Server 2024 (until June 2026)
- Claris FileMaker Server 2023 (until Apr 2025)

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

Noteworthy Options
-----
- --fqdn (for remote server administration)
- -i (for PKI authentication)

```
    fmcsadmin --fqdn fms.example.com -i /path/to/IDENTITYFILE list files
```

System Requirements
-----
- Linux version   : Ubuntu 20.04 LTS, Ubuntu 22.04 LTS or Ubuntu 22.04 LTS for ARM
- macOS version   : macOS Ventura 13, macOS Sonoma 14 or macOS Sequoia 15
- Windows version : Windows Server 2019, Windows Server 2022, Windows 10 Version 22H2, Windows 11 Version 22H2 or later

Download
-----
Download from the [latest release page](https://github.com/emic/fmcsadmin/releases/latest).

Author
-----
Emic Corporation <https://www.emic.co.jp/>

License
-----
This software is distributed under the [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0), see LICENSE.txt and NOTICE.txt for more information.
