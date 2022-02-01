### tiny DNS resolver

This is a command line program that makes DNS queries. There's a version in bash and a version in Go.

The main disclaimer is that it doesn't work on all domain names (for example it
can't resolve `maths.ox.ac.uk`). So you should definitely not consider this to
be a reference implementation of How All DNS Resolvers Work. Real DNS resolvers
are actually 

It's intended more as a basic starting point.

### how to run it


The bash version:

```
bash resolve.sh example.com
```

The go version:

```
go run resolve.go example.com
```
