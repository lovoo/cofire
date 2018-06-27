# Pixelreco

A weird recommender that maps the x axis of an image as the users, the y axis as the products and the gray level of the pixel as the score. Any image can be used as input.

The idea of recommending pixels of an image was taken from
[here](https://ruivieira.github.io/a-streaming-als-implementation.html).

## How to start?

- run locally: `go run local/main.go -input image.jpg`
- start Kafka and ZooKeeper, for example, using this [Makefile](https://github.com/lovoo/goka/blob/master/examples/Makefile)
- run with Kafka: `go run kafka/main.go -input image.jpg`
- see help message for details: `go run local/main.go -h` and `go run kafka/main.go -h`
