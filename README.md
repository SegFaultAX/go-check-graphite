# go-check-graphite

This library is *nearly* a drop-in replacement for
[nagios_graphite](https://github.com/SegFaultAX/nagios_graphite). Use it as a
better alternative to that.

## Installation

`go install github.com/segfaultax/go-check-graphite`

## Usage

**Important**: This library supports all of the same features as
nagios_graphite, however some of the switches have changed name (both
long and short) so pay close attention when migrating.

```shell
check-graphite -g graphite.example.com -m 'my.metric' -w 10 -c 100
```

## Contributors

* Michael-Keith Bernard (@segfaultax)

## License

This project is released under the MIT license. See LICENSE for details.
