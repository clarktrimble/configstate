
# notes

## consul, etc

```bash
docker run -d --name consul01 -p 8500:8500 hashicorp/consul:1.17
docker logs -f consul01

curl localhost:8500/v1/kv/bargle -XPUT -d'{"ima":"pc"}'
curl localhost:8500/v1/kv/bargle
curl -v localhost:8500/v1/kv/bargle?raw
curl "localhost:8500/v1/kv/bargle?index=59&wait=5s"

curl -s localhost:8500/v1/kv/svc/24 -XPUT -d@service24.json | jq
curl -s localhost:8500/v1/kv/svc/05 -XPUT -d@service05.json | jq
curl -s "localhost:8500/v1/kv/svc?recurse"
curl -s "localhost:8500/v1/kv/svc?recurse" | jq

curl -s localhost:8500/v1/kv/services-test -XPUT -d@services.json | jq
curl -s localhost:8500/v1/kv/services-test?raw | jq

go run cmd/discover/main.go -h
. etc/demo.sh
go run cmd/discover/main.go -c
go run cmd/discover/main.go

curl localhost:8081/config | jq
curl localhost:8081/services | jq

make
bin/discover | jq
```

http://tartu:8500/ui/dc1/kv/bargle/edit



## configstate

If you think about it hard enough, you can blur the line between configuration and program state; kind of like pointing your index fingers together just a few inches in front of your eyes.
Go ahead, give it a try ...
Can you see the "sausage"?
Lol, anyway, let's take the example of discoverables.

Suppose we have a PhotoBook service and one of it's many features it to resize photos.
This sort of thing can be labor-intensive and we have broken resize out into its own separatly scalable service.
We're strapped for time/money in the beginning and its a reasonable shortcut to simply configure PhotoBook so it knows about available resize services.


Long poll feels like a good fit here.
 - yes change in configstate is an event
 - but need a good "backstop" and on startup can do the same thing
