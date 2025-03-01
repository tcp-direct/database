module github.com/tcp-direct/database

go 1.21

toolchain go1.22.4

require (
	git.mills.io/prologic/bitcask v1.0.2
	git.tcp.direct/kayos/common v0.9.9
	github.com/akrylysov/pogreb v0.10.2
	github.com/davecgh/go-spew v1.1.1
)

require (
	github.com/abcum/lcp v0.0.0-20201209214815-7a3f3840be81 // indirect
	github.com/gofrs/flock v0.8.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/plar/go-adaptive-radix-tree v1.0.4 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	golang.org/x/exp v0.0.0-20200228211341-fcea875c7e85 // indirect
	golang.org/x/sys v0.21.0 // indirect
	nullprogram.com/x/rng v1.1.0 // indirect
)

retract (
	v0.5.5 // missed some spots
	v0.5.0 // riddled
	v0.4.7 // with
	v0.4.6 // pre
	v0.4.5 // v1
	v0.4.4 // bugs
	//
	v0.4.3 // deadlock
	v0.4.0 // broken metadata system
	v0.3.0 // doesn't pass go vet
)
