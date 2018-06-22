# Cofire [![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause) [![GoDoc](https://godoc.org/github.com/lovoo/cofire?status.svg)](https://godoc.org/github.com/lovoo/cofire)

Cofire is a stream-based collaborative filtering implementation for recommendation engines solely running on [Apache Kafka].
Leveraging the [Goka] stream processing library, Cofire continously learns user-product ratings arriving from Kafka event streams and maintains its model up-to-date in Kafka tables.
Cofire implements streaming matrix factorization employing stochastic gradient descent to learn latent factors for users and products.
The streaming matrix factorization follows the lines of [this blog post](https://ruivieira.github.io/a-streaming-als-implementation.html).

## How does it work?


## How to set it up?

1. create topics
2. adapt input streams
3. embed model table in application

## How to contribute

Contributions are always welcome.
Please fork the repo, create a pull request against master, and be sure tests pass.
See the [GitHub Flow] for details.

[Apache Kafka]: https://kafka.apache.org/
[Goka]: https://github.com/lovoo/goka
[GoDoc]: https://godoc.org/github.com/lovoo/cofire
[GitHub Flow]: https://guides.github.com/introduction/flow
