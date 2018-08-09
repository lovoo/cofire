# Cofire [![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause) [![GoDoc](https://godoc.org/github.com/lovoo/cofire?status.svg)](https://godoc.org/github.com/lovoo/cofire)

Cofire is a stream-based collaborative filtering implementation for recommendation engines solely running on [Apache Kafka].
Leveraging the [Goka] stream processing library, Cofire continously learns user-product ratings arriving from Kafka event streams and maintains its model up-to-date in Kafka tables.
Cofire implements streaming matrix factorization employing stochastic gradient descent (SGD).

> We assume you understand the SGD algorithm, so this README focus on how Cofire works on top of [Goka].
> See [this blog post](https://ruivieira.github.io/a-streaming-als-implementation.html) for a great introduction on SGD for streams.

## Components

Cofire has two types of stream processors:

- The *learner* consumes ratings from an input topic and applies SGD to learn the latent features of users and products.
Here is where machine learning happens.
Learners are stateful and store the latent features in Kafka in a log-compacted topic.
- The *refeeder* reemits already learnt ratings into the learner's input topic after a predefined delay.
The refeeder effectively implements training iterations. By default, the number of iterations is configured to be 1, so the refeeder is optional.

Besides these processors, two other components are necessary to get the system running:

- At least one *producer* that writes the ratings into the learner's input topic.
- At least one *predictor* that performs predictions using the learnt model.
The predictor does not need to be co-located with the learner, it can simply keep a local view of the model using Goka.

## Preparation

Before starting any processors, one has to do the following:

1. Choose a group name, eg, "cofire-app".
2. Create topics for the processors (if not auto created):
   - `<group>-input` as input for the learner, eg, "cofire-app-input"
   - `<group>-loop` to loopback messages among learner instances, eg, "cofire-app-loop"
   - `<group>-table` to store the learnt model, eg, "cofire-app-table"
   - `<group>-update` to overwrite the learner model if desired , eg, "cofire-app-update"
   - `<group>-refeed` to send ratings from the learner to the refeeder, eg, "cofire-app-refeed"
3. Ensure all topics have the same number of partitions.
4. Ensure `<group>-table` is configured with log compaction.


See the [examples](examples) directory for detailed examples.

## How it works

Learners (as any Goka processor) have processing partitions, which match the number of partitions of the inputs and state.
These processing partitions may be distributed over multiple processor instances, ie, multiple program instances running in multiple hosts.

Ratings and other messages are assigned to partitions via the key used when emitting the message.

> To simplify the explanation, we refer to keys as if they would be active entities.
> For example, "we send a message to a key k" means that we send a message using key k and the learner partition responsible for the key k receives that message;
> also, "a key k processes a message" means the learner partition responsible for key k processes a message.


### Producing ratings

A rating is defined as follows.

```
message Rating {
  string user_id    = 1;
  string product_id = 2;
  double score      = 3;
}
```

A producer sends ratings to the learner instances via `<group>-input` topic.
The key of each rating message is the `user_id`.


### Learning

Cofire can be configured with these parameters:

```go
type Parameters struct {
  // Rank is the number of latent factors (features).
  Rank int
  // Gamma is the learning step of SGD.
  Gamma float64
  // Lambda is the regularization parameter of SGD.
  Lambda float64
  // Iterations is the number of times the data will be used for training.
  Iterations int
}
```

Learners have state, which is partitioned and stored in Kafka in the `<group>-table` topic.
Each key has an entry in learner state defined as follows.
```
message Entry {
  Features u = 1;
  Features p = 2;
}
```

The algorithm for one rating has 3 steps:
1. When `user_id` receives a rating via the input topic, it retrieves the U features and sends (rating,U) to the `product_id` via the `<group>-loop` topic.
2. When `product_id` receives (rating,U), it retrieves the P features, applies SGD, and sends (rating,P) back to `user_id`.
3. When `user_id` receives (rating,P), it retrieves the U features, applies SGD, and sends the rating to the refeeder via `<group>-refeed`.

With that the iteration for this rating is finished.

In ASCII-art, one iteration would look like this
```
    Rating
      |
      v
     USER
      |
      * Entry                     PRODUCT
      |        (Rating, U)           |
      +----------------------------->|
      |                              |
      |                              * Update P
      |        (Rating, P)           |
      |<-----------------------------+
      |
      * Update U
      |
      v
   REFEEDER
```

### Iterating

If the algorithm is configured to run multiple iterations, the refeeder sends the rating back to the `user_id` to retrain the rating.
That happens for the configured number of iterations.
Since the stream never ends, the refeeder creates iterations by delaying the `<group>-refeed` topic by a configurable duration.
Note that the retention time configured for the topic has to be longer than the delay duration of the refeeder, otherwise ratings will be lost.


In ASCII-art, the complete flow is as follows.
Here we see the three components: producer, learner and refeeder.
The **producer** sends a rating to the **learner**.
The *user* key receives the rating and sends the rating plus its U vector to the *product* key in the learner (by using Goka's loopback).
The product updates P, sends it back to the user, which updates U and sends the rating to the **refeeder**.
The refeeder sends the rating back to the learner/user-key after a delay, and the number of remaining iterations is decremented.
The *user* receives the rating and the next iteration starts.

```
    PRODUCER             ,..LEARNER..,               REFEEDER
      |             USER'             'PRODUCT          .
      |   Rating     .                   .              .
      +------------->+                   .              .
      .              |                   .              .
      .              |                   .              .
      .              * Entry             .              .
      .              |                   .              .
      .              |    (Rating, U)    .              .
      .              +------------------>+              .
      .              .                   |              .
      .              .                   |              .
      .              .                   * Update P     .
      .              .                   |              .
      .              .    (Rating, P)    |              .
      .              +<------------------+              .
      .              |                   .              .
      .              |                   .              .
      .              * Update U          .              .
      .              |                   .              .
      .              |      (Rating, #iterations)       .
      .              +--------------------------------->+
      .              .                   .              |
      .              .                   .              |
      .              .                   .              * Delay
      .              .                   .              |
      .              .      (Rating, #iterations--)     |
      .              +<---------------------------------+
      .              |                   .              .
      .              |                   .              .
      .              * Entry             .              .
      .              |                   .              .
      .              |   next iteration  .              .
      .              +------------------>+              .
```


### Predicting

Every update of U or P in a learner produces an update of `<group>-table`.
To perform predictions, one simply creates a Goka view of the `<group>-table>` and gets the entries for the desired user and product. For example:

```go
view, _ := goka.NewView(brokers, goka.GroupTable(group), new(cofire.EntryCodec))

user, _ := view.Get("user")
u := user.(*cofire.Entry).U
product, _ := view.Get("product")
p := ep.(*cofire.Entry).P

prediction := u.Predict(p, bias)
```

### Global bias

The global bias of SGD is not stored anywhere in the state, only in memory. So to apply predictions, one needs to compute the bias manually.
However, if one is simply creating product recommendations for a user, bias can be set to 0 since that won't affect the sorted order of the scored products.

## How to contribute

Contributions are always welcome.
Please fork the repo, create a pull request against master, and be sure tests pass.
See the [GitHub Flow] for details.

[Apache Kafka]: https://kafka.apache.org/
[Goka]: https://github.com/lovoo/goka
[GoDoc]: https://godoc.org/github.com/lovoo/cofire
[GitHub Flow]: https://guides.github.com/introduction/flow
