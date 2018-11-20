## How to build from sources
- [Install Golang 1.9 or newer](https://golang.org/dl/)
```bash
go get -d -u github.com/cloudradar-monitoring/frontman
go build -o frontman -ldflags="-X main.version=$(git --git-dir=src/github.com/cloudradar-monitoring/frontman/.git describe --always --long --dirty --tag)" github.com/cloudradar-monitoring/frontman/cmd/frontman
```

## Run the example

```bash
./frontman -i src/github.com/cloudradar-monitoring/frontman/example.json -o result.out
```
Use `ctrl-c` to stop it

## What kind of checks frontman can perform
* [ICMP ping](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L53)
* [TCP – check connection on port](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L68)
* [TCP – service check (check the connection and the common output pattern)](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L77)
     * HTTP(S)
     * FTP(S)
     * FTP(S)
     * SMTP(S)
     * POP3(S)
     * SSH
     * NNTP
     * LDAP
* [SSL – check the certificate validity and expiration date](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L119)
* HTTP web checks
     * [Check status](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L4)
     * [Match raw HTML pattern](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L31)
     * [Match extracted text patter](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L28)
## Command line Usage
```
Usage of frontman:
  -c  config file path (default depends on OS)
  -d  daemonize – run the proccess in background
  -i  JSON file to read the list (required if no hub_url specified in the config)
  -o  file to write the results (default ./results.out)
  -r  one run only – perform checks once and exit. Overwrites output file
  -s  username to install and start the system service
  -u  stop and uninstall the system service
  -p  print the active config
  -v  log level – overrides the level in config file (values "error","info","debug") (default "error")
  -version  show the frontman version
```
## Configuration
On first run frontman will automatically create the config file and fill it with the default value.

Default config locations:
* Mac OS: `~/.frontman/frontman.conf`
* Windows: `./frontman.conf`
* UNIX: `/etc/frontman/frontman.conf`

If some of the fields are missing in the config `frontman` will use the defaults.
To print the active config you can use `frontman -p`

Also you may want to check the [example config](https://github.com/cloudradar-monitoring/frontman/blob/master/example.config.toml) that contains comments on each field.

## Logs location
* Mac OS: `~/.frontman/frontman.log`
* Windows: `./frontman.log`
* UNIX: `/etc/frontman/frontman.conf`

## Build binaries and deb/rpm packages
– Install [goreleaser](https://goreleaser.com/introduction/)
```bash
goreleaser --snapshot
```

## Build MSI package
- Should be done on Windows machine
- Open command prompt(cmd.exe)
- Go to frontman directory `cd path_to_directory`
- Run `goreleaser --snapshot` to build binaries
- Run `build-win.bat`
