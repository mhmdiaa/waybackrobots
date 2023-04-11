# waybackrobots
Enumerate old versions of robots.txt paths using Wayback Machine for content discovery.

## Usage
Pass a list of domains/URLs through stdin and waybackrobots will write the full URLs to stdout.
```sh
$ cat targets.txt | waybackrobots

Enumerating https://google.com/robots.txt versions... 100% |███████████████████████████████████████████| (50/50, 10 it/s)

https://google.com/analytics/reporting/
https://google.com/ebooks?*q=related:*
https://google.com/compare/*/apply*
...
```

## Command-line options

| Option   | Description                                                    | Default |
|----------|----------------------------------------------------------------|---------|
| -limit   | Limit the number of crawled snapshots. Use -1 for unlimited.   | 50       |
| -recent  | Use the most recent snapshots without evenly distributing them | false   |

## Snapshot Distribution
By default, `waybackrobots` evenly distributes the snapshots it analyzes across the file's history when a limit is set. This is done to diversify the results and get a broader view of the `robots.txt` file over time.

For example, if you set the limit to 5 and there are 10 snapshots, waybackrobots will analyze every other snapshot starting from the latest one. This means it will analyze the first, third, fifth, seventh, and ninth most recent snapshots.

This default behavior can be changed with the `-recent` option, which tells `waybackrobots` to use only the most recent snapshots.

```sh
$ echo google.com | waybackrobots | wc
     422     422   13973
$ echo google.com | waybackrobots -recent | wc
     277     277    9100
```

## Installation
### Binary
Check out the [latest release](https://github.com/mhmdiaa/waybackrobots/releases/latest).

### Go install
```
go install github.com/mhmdiaa/waybackrobots@latest
```

## References
- This tool is an improved and updated version of [waybackrobots.py](https://gist.github.com/mhmdiaa/2742c5e147d49a804b408bfed3d32d07).
- If you need a more customizable tool for working with Wayback Machine data, check out [chronos](https://github.com/mhmdiaa/chronos).
