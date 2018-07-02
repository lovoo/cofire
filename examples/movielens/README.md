# Movie Lens example

## How to start?

- download Movie Lens dataset, for example, [the smallest](http://files.grouplens.org/datasets/movielens/ml-latest-small.zip)
- uncompress zip file and copy the `ratings` file to `/tmp/`
- run locally: `go run local/main.go -input /tmp/ratings`
- start Kafka and ZooKeeper, for example, using this [Makefile](https://github.com/lovoo/goka/blob/master/examples/Makefile)
- run with Kafka: `go run kafka/main.go -input /tmp/ratings`
- see help message for details: `go run local/main.go -h` and `go run kafka/main.go -h`

