# BoltDB REST API

Adds restful API on top of [BoltDB](https://github.com/boltdb/bolt).

## Building

1. Install [gpm](https://github.com/pote/gpm) and [gpm-bootstrap](https://github.com/pote/gpm-bootstrap)
2. Install dependencies:

```bash
$ git clone https://github.com/marconi/boltapi && cd boltapi
$ gpm
$ make build && make install
```

### Running

```bash
$ boltapi -dbpath=./app.db
```

You can change what port the API listens with `-port` param.
