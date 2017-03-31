# Gover - Go dependencies versions manager

##### gover.yaml

```yaml
name: test
version: 1
dependencies:
  - package: github.com/gocraft/dbr
    version: sha:c4f8f99b25c74a821ffa41ca2e339cc24a6b4a67
    url: https://github.com/gocraft/dbr.git
  - package: github.com/gorilla/mux
    version: v1.3.0
    url: https://github.com/gorilla/mux.git
```

commands
```gover get <dogs.yaml> ```
```gover list```
```gover help <command>```

## Installation

```
go get github.com/bgaifullin/gover
```
