# go-check-graphite

This library is *nearly* a drop-in replacement for
[nagios_graphite](https://github.com/SegFaultAX/nagios_graphite). Use it as a
better alternative to that.

## Installation

1. Make sure you have Go installed first. If you're on a Mac, you can install
it via Homebrew: `brew install go`
2. Once installed, run `go install github.com/segfaultax/go-check-graphite`
3. You should now be able to run `go-check-graphite`. If the executable can't
be found, make sure that the Go bin directory is in your path. If it isn't, 
add the following line to your profile:
`export PATH=$PATH:”$(go env GOPATH)/bin”`

## Usage

**Important**: This library supports all of the same features as
nagios_graphite, however some of the switches have changed name (both
long and short) so pay close attention when migrating.

```shell
go-check-graphite -g graphite.example.com -m 'my.metric' -w 10 -c 100
```

## Contributors

* Michael-Keith Bernard (@segfaultax)

## License

This project is released under the MIT license. See LICENSE for details.
