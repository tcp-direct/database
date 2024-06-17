module git.tcp.direct/tcp.direct/database

go 1.18

require (
	git.tcp.direct/Mirrors/bitcask-mirror v0.0.0-20220228092422-1ec4297c7e34
	git.tcp.direct/kayos/common v0.9.3
	github.com/akrylysov/pogreb v0.10.1
	github.com/davecgh/go-spew v1.1.1
	github.com/hashicorp/go-multierror v1.0.0
)

require (
	github.com/abcum/lcp v0.0.0-20201209214815-7a3f3840be81 // indirect
	github.com/gofrs/flock v0.8.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/plar/go-adaptive-radix-tree v1.0.4 // indirect
	github.com/rs/zerolog v1.26.1 // indirect
	golang.org/x/exp v0.0.0-20200228211341-fcea875c7e85 // indirect
	golang.org/x/sys v0.13.0 // indirect
	nullprogram.com/x/rng v1.1.0 // indirect
)

retract (
	v0.4.0 // broken metadata system
	v0.3.0 // doesn't pass go vet
)
