// Package broker provides task broker implementations for use in the GoFlow framework.
// It offers several types of brokers to manage task queues, allowing users to plug in
// different queueing mechanisms as per their needs.
//
// Two main broker implementations are provided:
// 1. ChannelBroker: A broker that wraps a buffered Go channel.
// 2. RedisBroker: A broker that uses Redis for the underlying queues.
//
// These implement the Broker interface defined in the main application, enabling
// flexibility for users to choose their preferred queueing backend (in-memory or Redis)
// for task handling.
package broker
