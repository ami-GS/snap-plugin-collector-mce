[![Build Status](https://travis-ci.org/ami-GS/snap-plugin-collector-mce.svg?branch=master)](https://travis-ci.org/ami-GS/snap-plugin-collector-mce)

# snap-plugin-collector-mce

This plugin collects machine check error metrics from /var/log/mcelog.

## Getting Started
### System Requirements
* [golang 1.7+](https://golang.org/dl/) - needed only for building
* [mcelog](http://www.mcelog.org/) - needed for running

### Operating systems
All OSs currently supported by plugin:
* Linux/amd64

## Installation
### build from source

```
$ go get github.com/ami-GS/snap-plugin-collector-mce
$ $GOPATH/bin/snap-plugin-collector-mce
```
or

```
$ git clone https://github.com/ami-GS/snap-plugin-collector-mcecollector
$ cd snap-plugin-collector-mcecollector
$ go get -t ./...
$ go build
$ ./snap-plugin-collector-mce
```

### Configuration and Usage
* If your mcelog is not using default logging directory (/var/log/mcelog), you should specify the directory in main.go (snap seems not to be able to pass user defined argument), or will update more flexibly

As part of snapteld global config

```yaml
TBD
```

Or as part of the task manifest

```json
{
...
    "workflow": {
        "collect": {
            "metrics": {
              "/ami-GS/mce/CPU" : {}
              "/ami-GS/mce/Corrected" : {}
            },
            "config": {
              "/ami-GS/mce/" : {
                "logpath" : "/my/custom/path"
              }
            },
        ...
        },
    },
...
```

### Usage
#### Run
```bash
$ snapteld -t 0 -l 1
# different terminal bellow
$ wget  http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/linux/x86_64/snap-plugin-publisher-file
$ snaptel plugin load snap-plugin-publisher-file
$ snaptel plugin load snap-plugin-collector-mcecollector
$ snaptel task create -t mce-file.json
```

Create a task manifest file

```json
TBD
```

## Contributing
I appreciate Any PR, any feature request!

## License
[Snap](http://github.com:intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements
* Author: [Daiki AMINAKA](https://github.com/ami-GS)
