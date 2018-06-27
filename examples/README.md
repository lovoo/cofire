# Examples

This directory contains a few examples of how to use Cofire.

`examples.go` contains examples of how to start 4 component types:

1. The [learner](examples.go#L13) consumes ratings from an input topic and applies SGD to learn the latent features of users and products (ie, factorizes the ratings matrix). It also periodically prints to stdout the RMSE of its predictions. 
2. The [producer](examples.go#L43) simply takes a list of ratings and emits them into the learner's input topic.
3. The [refeeder](examples.go#L71) reemits already learnt ratings into the learner's input topic after a predefined delay and for a number of iterations.
4. The [validator](examples.go#L98) takes a set of ratings (eg, a test set) and calculates the RMSE using a view on the learnt model (technically, a view of the learner's group table).

Besides these generic starters, there are two concrete examples of recommendation:

- [movielens](movielens): the classic Movie Lens example, where users are watchers, products are movies and scores are the number of stars each user gives to each movie.
- [pixelreco](pixelreco): a funny example that maps the x axis of an image as the users, the y axis as the products and the gray level of the pixel as the score. Any image can be used as input.

Both examples contain the Kafka-based implementation as well as a simple local runner.
