
# NATS flavored configstate!

## TLDR;

It works :)
Which is to say the discover service watches a key in NATS and reflects via "/services" endpoint.

The local `nats` package implements `Poller` to be injectable into `Discover`.
This is the tiniest bit awkward, but I like that the nats-io package does not escape.

For dynamic config:
 - convert `Discover` to `Dynafig` (or an equally fun name) to expose config rather than services
 - think through any sub-service restart scenarios that arise
 - think through operational aspects of updating dynamic config in NATS
   - version control
   - blue/green
   - roll back
 - ???

## Dockerize

For demo, we'll need a NATS running Jetstream:

```bash
$ docker run -d --name nats01 -p 4222:4222 nats:2.10.9-alpine3.19 -js
$ docker logs -f nats01 ## as needed
```

### Setup a bucket and tryout key-val

```bash
$ sudo apt install ./nats-0.1.1-amd64.deb
$ nats kv add dynafig
$ nats kv put dynafig TestoKey '{"thing":"one"}'
$ nats kv get --raw dynafig TestoKey | jq
```

```json
{
  "thing": "one"
}
```

That was easy; thanks NATS peeps!

## Tryout discover demo, NATS flavor

### Test and Build Project

```bash
~/proj/configstate$ make
go generate ./...
golangci-lint run ./...
go test -count 1 configstate/chi configstate/consul configstate/discover configstate/entity configstate/nats
ok      configstate/chi 0.010s
ok      configstate/consul      0.021s
?       configstate/entity      [no test files]
?       configstate/nats        [no test files]
ok      configstate/discover    0.027s
rm -rf bin
:: Building discover
go build -ldflags '-X main.version=nats-flav.14.e3fc909' -o bin/discover cmd/discover/main.go
:: Building nats-discover
go build -ldflags '-X main.version=nats-flav.14.e3fc909' -o bin/nats-discover cmd/nats-discover/main.go
:: Done
```

### Have a look at configurables

```bash
~/proj/configstate$ bin/nats-discover -h

'nats-discover' demonstrates service discovery, via NATS

The following environment variables are available for configuration:

KEY                   TYPE        DEFAULT    REQUIRED    DESCRIPTION
DSC_LOGGER_MAXLEN     Integer                            maximum length that will be logged for any field
DSC_NATS_URL          String                 true        nats server url
DSC_NATS_BUCKET       String                 true        bucket be watched
DSC_NATS_KEY          String                 true        key to be watched
DSC_SERVER_HOST       String                             hostname or ip for which to bind
DSC_SERVER_PORT       Integer                true        port on which to listen
DSC_SERVER_TIMEOUT    Duration    10s                    characteristic timeout
```

### Load config and double check

```bash
~/proj/configstate$ . etc/nats-demo.sh
~/proj/configstate$ bin/nats-discover -c
```

```json
{
  "version": "nats-flav.14.e3fc909",
  "logger": {
    "max_len": 999
  },
  "nats": {
    "url": "localhost",
    "bucket": "dynafig",
    "key": "TestoKey"
  },
  "http_server": {
    "host": "",
    "port": 8081,
    "timeout": 10000000000
  }
}
```

### Ok, ok, run the darn thing

```bash
~/proj/configstate$ bin/nats-discover
{"app_id":"dsc-demo","config":"{\"version\":\"nats-flav.14.e3fc909\",\"logger\":{\"max_len\":999},\"nats\":{\"url\":\"localhost\",\"bucket\":\"dynafig\",\"key\":\"TestoKey\"},\"http_server\":{\"host\":\"\",\"port\":8081,\"timeout\":10000000000}}","level":"info","msg":"starting up","run_id":"vslYfGx","ts":"2024-01-23T22:15:23.523304332Z"}
{"app_id":"dsc-demo","level":"info","msg":"worker starting","name":"discovery","run_id":"vslYfGx","ts":"2024-01-23T22:15:23.527994608Z","worker_id":"YZAR37K"}
{"app_id":"dsc-demo","level":"info","msg":"starting http service","run_id":"vslYfGx","ts":"2024-01-23T22:15:23.528071682Z"}
{"app_id":"dsc-demo","error":"json: cannot unmarshal object into Go value of type []entity.Service\nfailed to unmarshal services from: {\"thing\":\"one\"}\nconfigstate/entity.DecodeServices\n\t/home/trimble/proj/configstate/entity/service.go:29\nconfigstate/discover.(*Discover).work\n\t/home/trimble/proj/configstate/discover/discover.go:98\nruntime.goexit\n\t/home/trimble/go1211/src/runtime/asm_amd64.s:1650","level":"error","msg":"failed to watch","run_id":"vslYfGx","ts":"2024-01-23T22:15:23.528280792Z","worker_id":"YZAR37K"}
...
...
```

### Oops, forgot to load sensible json into NATS key

```bash
~/proj/configstate$ nats kv put dynafig TestoKey < test/data/services.json
```

### Woot, discover picked it up

```json
...
{"app_id":"dsc-demo","level":"info","msg":"updating services","run_id":"vslYfGx","ts":"2024-01-23T22:17:07.636266401Z","worker_id":"YZAR37K"}
...
```

### Request via endpont
```bash
$ curl -s localhost:8081/services | jq
```

```json
{
  "services": [
    {
      "uri": "http://pool04.boxworld.org/api/v2",
      "capabilities": [
        {
          "name": "resize",
          "capacity": 23
        }
      ]
    },
    {
      "uri": "http://pool24.boxworld.org/api/v2",
      "capabilities": [
        {
          "name": "resize",
          "capacity": 5
        }
      ]
    }
  ]
}
```

### Change in NATS kv, check that discover is hip

```bash
~/proj/configstate$ nats kv put dynafig TestoKey < test/data/services-too.json
~/proj/configstate$ curl -s localhost:8081/services | jq
```

```json
{
  "services": [
    {
      "uri": "http://pool04.boxworld.org/api/v2",
      "capabilities": [
        {
          "name": "resize",
          "capacity": 23
        }
      ]
    },
    {
      "uri": "http://pool24.boxworld.org/api/v2",
      "capabilities": [
        {
          "name": "resize",
          "capacity": 55
        }
      ]
    }
  ]
}
```

Hurrah!

