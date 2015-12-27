# BoltDB REST API [![GoDoc](https://godoc.org/github.com/marconi/boltapi?status.png)](https://godoc.org/github.com/marconi/boltapi)

Adds restful API on top of [BoltDB](https://github.com/boltdb/bolt).

## Building

1. Install [gpm](https://github.com/pote/gpm)
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

## Endpoints

Exposes the following endpoints:

**Buckets endpoint**
```
/api/v1/buckets

GET  - List buckets
POST - Add bucket
```

**Bucket endpoint**
```
/api/v1/buckets/<name>

GET    - List bucket items
POST   - Add item on the bucket
DELETE - Delete bucket
```

**Bucket item endpoint**
```
/api/v1/buckets/<name>/<key>

GET    - Retrieve item
PUT    - Update item
DELETE - Delete item
```

You can also check the tests for sample usage of these endpoints.
