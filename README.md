# Memento backup system

Link: https://github.com/mementobackup/

## Description:

Memento is a backup system for remote machines. It is similar of other systems
(e.g. rsnapshot) but it has some differences

## Licence:

Memento is released under GPLv2 (see GPL.txt for details)

## Features:

It use an agent for check and download data.
It manage four distinct datasets (hour, day, week, month).
It save space with hard link creation.
It save data attribute (owner, group, permission, ACL) into database.

## Dependencies:

 * PostgreSQL

## Building:
```
mkdir mserver && cd mserver
export GOPATH=`pwd`
git clone git@github.com:mementobackup/server.git .
go get github.com/gaal/go-options/options
go get github.com/go-ini/ini
go get github.com/lib/pq
go get github.com/op/go-logging
go get github.com/mementobackup/common/src/common
go build mserver.go
```

## Installation:

 - Create database on PostgreSQL;
 - Edit the configuration file (see USAGE);
 - Execute server.

## Usage:

Usage is simpliest:
```
mserver --backup --cfg=<cfgfile> -H # hour backup
mserver --backup --cfg=<cfgfile> -D # day backup
mserver --backup --cfg=<cfgfile> -W # week backup
mserver --backup --cfg=<cfgfile> -M # month backup
```

Where `<cfgfile>` is a file structured like the backup.cfg reported in the
archive. For other options, use -h switch. Some notes:

 - It is possible to have multiple configuration files, where each file has
   different parameters.
 
 - While database, dataset and and general sections are global, it is possible
   to set any number of sections, one per client.

This is an example of configuration file:
```
[general]
repository = /full/path/to/store/backups
log_file = memento.log
log_level = INFO

[database]
host = postgresql-host
port = 5432
user = postgresql-user
password = postgresql-password
dbname = postgresql-database

[dataset]
hour = 24
day = 6
week = 4
month = 12

[a_server]
type = file
host = localhost
port = 4444
ssl = true # true or false
sslcert = ssl.crt
sslkey = ssl.key
path = /full/path/to/backup
acl = true # true or false
compress = true # true or false
pre_command = ""
post_command = ""
```

## SSL:

If you want use the SSL connection, you need:

 - Create SSL certificate with these commands:
    ```
    openssl genrsa -des3 -out memento.key 2048
    openssl rsa -in memento.key -out memento.key
    openssl req -new -key memento.key -out memento.csr
    openssl x509 -req -days 365 -in memento.csr -signkey memento.key -out memento.crt
    ```
    
   In particular, be sure to add the hostname of the client machine in the CN field     
  - Configure server for use SSL with the same certificate.

## Caveats:

Because memento use hard links to store its dataset, its use is guaranteed on
linux or unix environments.

