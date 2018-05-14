/*

Package statsd implements functionality for creating servers compatible with the statsd protocol.
See https://github.com/etsy/statsd/blob/master/docs/metric_types.md for a description of the protocol.

The main components of the library are Receiver, Dispatcher, Aggregator and Flusher.
Receiver is responsible for receiving metrics from the socket.
Dispatcher dispatches received metrics among several Aggregators, which do
aggregation based on type of the metric. At every FlushInterval Flusher flushes metrics via
associated Backend objects.

Currently the library implements just a few types of Backend, one compatible with Graphite
(http://graphite.wikidot.org), one for Datadog and one just for stdout, but any object implementing the Backend
interface can be used with the library. See available backends at
https://github.com/atlassian/gostatsd/tree/master/backend/backends.

As with the original etsy statsd, multiple backends can be used simultaneously.
*/
package statsd
