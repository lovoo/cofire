# Movie Lens example

- download Movie Lens dataset, for example, [the smallest](http://files.grouplens.org/datasets/movielens/ml-latest-small.zip)
- uncompress zip file and copy the `ratings` file to `/tmp/`
- start Kafka and ZooKeeper, for example, using Goka's example [Makefile](https://github.com/lovoo/goka/blob/master/examples/Makefile)
- run the example: `go run cmd/movie-lens/main.go`
