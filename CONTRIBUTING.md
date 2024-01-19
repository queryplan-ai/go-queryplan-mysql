# Contributing to go-queryplan-mysql

Design considerations:

1. Minimize the number of dependencies. We are trying to avoid pulling in a lot of deps that could cause conflicts. Make it easy to consume.

2. Performance is key. Remember to make sure that everything performs sufficiently well to run in high-throughput environments.

3. Expect the unexpected. An example here is the ring buffer we've added instead of an unbounded array. Sometimes a system might be down, and we don't want to be the cause of a cascading failure. That will destroy any confidence.

