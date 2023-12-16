
# ConfigState

Demonstrating an approach to a discovery worker from within a Golang service.

## Blogses

This project is a companion to a [post](https://clarktrimble.online/blog/configstate/) about Golang discovery, etc.
Head on over for blather! :)

## Generalizable Highlights

Both discover and consul packages here are off to a good start and likely already re-usable depending on particular features demanded by the use case.

## Tactical Code Re-Use

- [delish](https://github.com/clarktrimble/delish) http json api server, middlewares, etc.
- [giant](https://github.com/clarktrimble/giant) http json api client and tripperwares
- [launch](https://github.com/clarktrimble/launch) envconfig, etc. helpers targeting main.go
- [sabot](https://github.com/clarktrimble/sabot) structured, contextual logging

Are featured prominently here and inching ever closer to a v1.0.0 release!
