/* eslint-disable react/no-unused-prop-types */

import React from "react";
import PropTypes from "prop-types";
import { ApolloError } from "apollo-client";
import { withApollo } from "react-apollo";
import shallowEqual from "fbjs/lib/shallowEqual";
import equal from "lodash/isEqual";

import QueryAbortedError from "/errors/QueryAbortedError";

const extractQueryOpts = props => {
  const {
    variables,
    pollInterval,
    fetchPolicy,
    errorPolicy,
    notifyOnNetworkStatusChange,
    query,
  } = props;

  return {
    variables,
    pollInterval,
    query,
    fetchPolicy,
    errorPolicy,
    notifyOnNetworkStatusChange,
    context: {},
  };
};

class Poller {
  constructor(observable, interval) {
    this._observable = observable;
    this._interval = interval;
    this._running = interval > 0;
  }

  isRunning() {
    return this._running;
  }

  getInterval() {
    return this._interval;
  }

  setInterval(newInterval) {
    if (newInterval === this._interval) {
      return;
    }

    if (newInterval > 0) {
      this.start(newInterval);
    } else {
      this._interval = newInterval;
      this.stop();
    }
  }

  start(newInterval = null) {
    if (this._running && newInterval) {
      this.setInterval(newInterval);
    } else if (!this._running) {
      this._interval = newInterval || this._interval;
      this._running = true;

      this._observable.startPolling(this._interval);
    }
  }

  stop() {
    this._running = false;
    this._observable.stopPolling();
  }

  toggle(newInterval = null) {
    if (this._running) {
      this.stop();
    } else {
      this.start(newInterval);
    }
  }
}

//
// Drop in replacement for react-apollo's Query component.
//
// Differs from official implementation by:
//
// - reflecting the value last emitted by the configured watcher instead of
//   firing a forceUpdate and relying the observer's currentResult. By doing so,
//   given data should be more consistent (see apollographql/apollo-client#3947)
//   and should be slightly more efficient by not having to reselect the same
//   data from the cache on each render.
// - allows us to handle error's however we perfer.
//
class Query extends React.PureComponent {
  static propTypes = {
    // Provided by withApollo
    client: PropTypes.object.isRequired,
    children: PropTypes.func.isRequired,
    errorPolicy: PropTypes.oneOf(["none", "ignore", "all"]),
    fetchPolicy: PropTypes.oneOf([
      "cache-first",
      "cache-and-network",
      "network-only",
      "cache-only",
      "no-cache",
      "standby",
    ]),
    notifyOnNetworkStatusChange: PropTypes.bool,
    onComplete: PropTypes.func,
    onError: PropTypes.func.isRequired,
    pollInterval: PropTypes.number,
    query: PropTypes.object.isRequired,
    variables: PropTypes.object,
  };

  static defaultProps = {
    errorPolicy: "all",
    fetchPolicy: "cache-and-network",
    notifyOnNetworkStatusChange: false,
    onComplete: () => null,
    onError(error) {
      throw error;
    },
    pollInterval: 0,
    variables: {},
  };

  static getDerivedStateFromProps(props, state) {
    const nextState = {};

    const queryOpts = extractQueryOpts(props);
    if (!shallowEqual(queryOpts, state.queryOpts)) {
      nextState.queryOpts = queryOpts;
    }

    // If client changed we need to tear everything and observe new client
    if (props.client !== state.client) {
      nextState.observable = props.client.watchQuery(queryOpts);
      nextState.poller = new Poller(nextState.observable, props.pollInterval);
      return nextState;
    }

    // If query options change by client did not apply new set of options
    if (nextState.queryOpts) {
      state.observable.setOptions(queryOpts);
    }

    // If the polling interval was updated do update poller.
    if (props.pollInterval !== state.poller.getInterval()) {
      state.poller.setInterval(props.pollInterval);
    }

    if (Object.keys(nextState).length === 0) {
      return null;
    }
    return nextState;
  }

  constructor(props) {
    super(props);

    // Setup query observable
    const queryOpts = extractQueryOpts(props);
    const observable = props.client.watchQuery(queryOpts);
    const poller = new Poller(observable, props.pollInterval);

    // Fetch current state of query
    const { data, loading, networkStatus } = observable.currentResult();

    // Add initial data and observable to the component's state
    this.state = {
      client: props.client,
      data,
      loading,
      observable,
      networkStatus,
      poller,
      queryOpts,
    };
  }

  componentDidMount() {
    this.subscribe();
  }

  componentDidUpdate() {
    this.subscribe();
  }

  componentWillUnmount() {
    this.disconnect();
  }

  subscribe = () => {
    const { observable } = this.state;

    // We only want to subscribe if a subscription has not already been setup
    if (this.query && this.query.observable === observable) {
      return;
    }

    // Ensure that the old subscription is torn down
    if (this.query) {
      this.query.subscription.unsubscribe();
    }

    const subscription = observable.subscribe({
      next: result => {
        // if any errors are present in the response throw
        if (result.errors && result.errors.length > 0) {
          throw new ApolloError({ graphQLErrors: result.errors });
        }

        this.setState(state => {
          let nextState = {
            error: result.error,
            loading: result.loading,
            networkStatus: result.networkStatus,
          };
          const prevState = {
            error: state.error,
            loading: state.loading,
            networkStatus: state.networkStatus,
          };

          if (shallowEqual(nextState, prevState)) {
            nextState = {};
          }

          // Only update reference if deep compare fails; can enable
          // optimizations in downstream components.
          if (!equal(result.data, state.data)) {
            // NOTE: maybe this behaviour should be opt-out?
            nextState.data = result.data;
          }

          if (Object.keys(nextState).length === 0) {
            return null;
          }
          return nextState;
        });
      },
      error: error => {
        this.resubscribeToQuery();

        // Reimplemented from react-apollo
        // https://github.com/apollographql/react-apollo/blob/master/src/Query.tsx#L320-L321
        // eslint-disable-next-line no-prototype-builtins
        if (!error.hasOwnProperty("graphQLErrors")) {
          throw error;
        } else if (!(error.networkError instanceof QueryAbortedError)) {
          this.props.onError(error);
        }
      },
    });

    this.query = {
      observable,
      subscription,
    };
  };

  resubscribeToQuery = () => {
    this.disconnect();

    //
    // Reimplemented from react-apollo
    // https://github.com/apollographql/react-apollo/blob/master/src/Query.tsx#L340-L343
    //
    // If lastError is set, the observable will immediately
    // send it, causing the stream to terminate on initialization.
    // We clear everything here and restore it afterward to
    // make sure the new subscription sticks.
    //
    const lastError = this.query.observable.getLastError();
    const lastResult = this.query.observable.getLastResult();
    this.query.observable.resetLastResults();
    this.subscribe();
    Object.assign(this.query.observable, { lastError, lastResult });
  };

  disconnect = () => {
    if (this.query) {
      this.query.subscription.unsubscribe();
    }
  };

  render() {
    const { client } = this.props;
    const { data, loading, networkStatus, observable, poller } = this.state;

    let error = this.state.error;
    let aborted = false;
    if (error && error.networkError instanceof QueryAbortedError) {
      error = undefined;
      aborted = true;
    }

    return this.props.children({
      client,
      data,
      loading,
      networkStatus,
      poller,
      error,
      aborted,

      // TODO: Move into state?
      variables: observable.variables,
      refetch: observable.refetch.bind(observable),
      fetchMore: observable.fetchMore.bind(observable),
      updateQuery: observable.updateQuery.bind(observable),
      startPolling: observable.startPolling.bind(observable),
      stopPolling: observable.stopPolling.bind(observable),
      subscribeToMore: observable.subscribeToMore.bind(observable),
    });
  }
}

export default withApollo(Query);
