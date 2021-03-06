## Frontman - at a glance
Frontman is a general-purpose monitoring proxy that performs checks on foreign hosts. 
The main goal is to check services and performs other checks, where no logon-rights are needed for.
It's expected to be used in any kind of monitoring solutions, where agent-less checks must be executed over the network.
Frontman is a robust check executor that can work in the background of any monitoring solution. It enables developers to build custom solutions. Frontman can be controlled by local Json files or via Json-based HTTP APIs.

## Credits
<table border="0">
  <tr>
    <td>
      <img alt="European Regional Development Fund" src="https://efre.brandenburg.de/sixcms/media.php/9/EFRE%20Logo_unten_oweb_en_rgb.jpg" height="200" width="210" />
    </td>
    <td>
      Project co-financed by the European Regional Development Fund<br/>under the Innovative Economy Operational Programme. Innovation grants.<br/>We invest in your future.
    </td>
  </tr>
</table>


## What kind of checks frontman can perform
* [ICMP ping](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L53)
* [TCP/UDP – check connection on port](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L68)
* [TCP/UDP – service check (check the connection and the common output pattern)](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L77)
     * HTTP(S)
     * FTP(S)
     * FTP(S)
     * SMTP(S)
     * POP3(S)
     * SSH
     * NNTP
     * LDAP
     * [SIP](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L140)
     * [IAX2](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L143)
* [SSL – check the certificate validity and expiration date](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L119)
* HTTP web checks
     * [Check status](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L4)
     * [Match raw HTML pattern](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L31)
     * [Match extracted text patter](https://github.com/cloudradar-monitoring/frontman/blob/master/example.json#L28)

     
## Run the example

```bash
./frontman -i src/github.com/cloudradar-monitoring/frontman/example.json -o result.out
```
Use `ctrl-c` to stop it    

## Command line Usage
```
Usage of frontman:
  -c  config file path (default depends on OS)
  -d  daemonize – run the process in background
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

## How to build from sources
- [Install Golang 1.9 or newer](https://golang.org/dl/)
```bash
go get -d -u github.com/cloudradar-monitoring/frontman
go build -o frontman -ldflags="-X main.version=$(git --git-dir=src/github.com/cloudradar-monitoring/frontman/.git describe --always --long --dirty --tag)" github.com/cloudradar-monitoring/frontman/cmd/frontman
```

## Build binaries and deb/rpm packages
– Install [goreleaser](https://goreleaser.com/introduction/)
```bash
make goreleaser-snapshot
```

## Build MSI package
Should be done on Windows machine
- [Download go-msi](https://github.com/cloudradar-monitoring/go-msi/releases) and put it in the `C:\Program Files\go-msi`
- Open command prompt(cmd.exe or powershell)
- Go to frontman directory `cd path_to_directory`
- Run `make goreleaser-snapshot` to build binaries
- Run `build-win.bat`

## Versioning model
Frontman uses `<major>.<minor>.<buildnumber>` model for compatibility with a maximum number of package managers.

Starting from version 1.2.0 packages with even `<minor>` number are considered stable.

## Running as a docker container
Check [dockerhub](https://cloud.docker.com/u/cloudradario/repository/docker/cloudradario/frontman) for available images.

### Passing credentials
Username and password need to be configured via environment variables. You can pass them using the `-e` flags.
`docker run -d -e FRONTMAN_HUB_USER=YOUR_USERNAME -e FRONTMAN_HUB_PASSWORD=YOUR_PASS cloudradario/frontman:1.0.7`
