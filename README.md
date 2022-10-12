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
- snowball branch will implement the [Snowball consensus algorithm](https://assets.website-files.com/5d80307810123f5ffbb34d6e/6009805681b416f34dcae012_Avalanche%20Consensus%20Whitepaper.pdf)