# Distributed state machine

Distributed state machine is a project to compare the performance of multiple consensus algorithms.

The goal of a distributed state machine is to have multiple instances (servers) holding the same log

The client can query any server (also called node) to get the current version of the log. The client can also request to add something in the log.

The system is supposed to be fault-tolerant, which means that if some nodes happen to crash, the client should still be able to communicate with functionning nodes.

## Context

This project is part of my end-of-studies research project on consensus algorithms.

I started my work with a [bibliographical survey](https://github.com/Nicolascrd/researchProjectConsensus) and I also did before this a fault-tolerant [decentralized calculator](https://github.com/Nicolascrd/decentralized-calculator).

Unfortunately, the decentralized calculator is not a suitable context to implement the [Snowball consensus protocol](https://assets.website-files.com/5d80307810123f5ffbb34d6e/6009805681b416f34dcae012_Avalanche%20Consensus%20Whitepaper.pdf), which I wanted to study.

Let's compare some consensus algorithms, in the context of building a distributed state machine

## Versions

- master branch will implement the [Raft consensus algorithm](https://raft.github.io/raft.pdf).
- slush, snowflake and snowball branches will implement the differents versions of the [Snowball consensus algorithm](https://assets.website-files.com/5d80307810123f5ffbb34d6e/6009805681b416f34dcae012_Avalanche%20Consensus%20Whitepaper.pdf)

Snowball is a protocol to achieve binary consensus. For now, we will assume one side is the query from the client. The other would be any other value (except the empty string). The neutral state is nothing in the log at that position (empty string). 

## Testing

Select your parameters in config.json :
With any of the snowball algorithm, it is VERY important to change your parameters according to the number of nodes that you want to have in your system
```
{
    "sampleSize": size of the sample for each query made by a node (k in the avalanche paper),
    "majorityThreshold": majority threshold of the queries to be considered validated (same color) (alpha in the avalanche paper),
    "counterThreshold": counter threshold above which the color can effectively switch (beta in the avalanche paper)
}
```


Launch the containers with :

```
bash launch.sh <3 4 5 6>
# number of containers to start (min 3 max 99)
```

\
Add one element to the log by querying one container with :

```
curl -X POST localhost:800<1 2 3>/add-log -H 'Content-Type: application/json' \
    -d '{"position":10, "content": "Foo"}' # ask container <1 2 3> to append "Foo" as log number 10
```
\
Request one element to the log by querying one container with :

```
curl -X POST localhost:800<1 2 3>/request-log -H 'Content-Type: application/json' \
    -d '{"position": 10}' # ask container <1 2 3> for the content of log number 10
```
\
\
Check the logs of container <1 2 3> with :

```
docker logs --follow sm-server-<1 2 3>
```
\
You can also try crashing containers with :

```
docker stop sm-server-<1 2 3>
```
\
See what happens in the logs as the consensus algorithm run !
