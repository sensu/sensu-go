// @flow
/* eslint-disable react/no-unused-prop-types */
/* eslint-disable react/sort-comp */
/* eslint-disable react/no-multi-comp */

import * as React from "react";
import { ApolloError, withApollo } from "react-apollo";
import type {
  ObservableQuery,
  ApolloClient,
  WatchQueryOptions,
  NetworkStatus,
  ApolloQueryResult,
} from "react-apollo";
import shallowEqual from "fbjs/lib/shallowEqual";

import QueryAbortedError from "/errors/QueryAbortedError";

type ObservableMethods = {
  fetchMore: () => void,
  refetch: () => void,
  startPolling: () => void,
  stopPolling: () => void,
  subscribeToMore: () => void,
  updateQuery: () => void,
};

type Props = {
  client: ApolloClient<mixed>,
  // eslint-disable-next-line no-use-before-define
  children: State => React.Node,
  onError: Error => void,
} & WatchQueryOptions;

type State = {
  aborted: boolean,
  data: mixed | null, // TODO: Infer data type from query
  error: Error | null,
  loading: boolean,
  networkStatus: NetworkStatus,
  observable: ObservableQuery<mixed>,
  props: Props,
} & ObservableMethods;

const modifiableWatchQueryOptionsHaveChanged = (
  a: WatchQueryOptions,
  b: WatchQueryOptions,
) =>
  a.pollInterval !== b.pollInterval ||
  a.fetchPolicy !== b.fetchPolicy ||
  a.errorPolicy !== b.errorPolicy ||
  a.fetchResults !== b.fetchResults ||
  a.notifyOnNetworkStatusChange !== b.notifyOnNetworkStatusChange ||
  !shallowEqual(a.variables, b.variables);

const extractQueryOptions = (
  props: WatchQueryOptions,
): $Shape<WatchQueryOptions> => ({
  variables: props.variables,
  pollInterval: props.pollInterval,
  query: props.query,
  fetchPolicy: props.fetchPolicy,
  errorPolicy: props.errorPolicy,
  notifyOnNetworkStatusChange: props.notifyOnNetworkStatusChange,
});

class Query extends React.PureComponent<Props, State> {
  static defaultProps = {
    variables: {},
    pollInterval: 0,
    children: () => {},
    onError: (error: Error) => {
      throw error;
    },
  };

  subscription: { unsubscribe(): void } | null = null;

  static getDerivedStateFromProps(props: Props, state: State | null) {
    if (state !== null && state.props === props) {
      return null;
    }

    if (
      state === null ||
      state.props.client !== props.client ||
      state.props.query !== props.query
    ) {
      const observable: ObservableQuery<mixed> = props.client.watchQuery(
        extractQueryOptions(props),
      );

      // retrieve the result of the query from the local cache
      const { data, loading, networkStatus } = observable.currentResult();

      return {
        props,
        observable,
        data,
        loading,
        networkStatus,
        refetch: observable.refetch.bind(observable),
        fetchMore: observable.fetchMore.bind(observable),
        updateQuery: observable.updateQuery.bind(observable),
        startPolling: observable.startPolling.bind(observable),
        stopPolling: observable.stopPolling.bind(observable),
        subscribeToMore: observable.subscribeToMore.bind(observable),
      };
    }

    if (state && modifiableWatchQueryOptionsHaveChanged(state.props, props)) {
      state.observable.setOptions(extractQueryOptions(props));
    }

    // Changes to `metadata` and `context` props are ignored.

    return { props };
  }

  subscribe() {
    if (this.subscription) {
      throw new Error("Cannot subscribe. Currently subscribed.");
    }

    this.subscription = this.state.observable.subscribe({
      next: this.onNext,
      error: this.onError,
    });
  }

  unsubscribe() {
    if (!this.subscription) {
      throw new Error("Cannot unsubscribe. Not currently subscribed");
    }

    this.subscription.unsubscribe();
    this.subscription = null;
  }

  onNext = ({
    data,
    errors,
    loading,
    networkStatus /* stale */,
  }: ApolloQueryResult<mixed>) => {
    let error = null;

    if (errors && errors.length > 0) {
      error = new ApolloError({ graphQLErrors: errors });
    }

    this.setState((state: State) => ({
      ...state,
      aborted: false,
      error,
      data,
      loading,
      networkStatus,
    }));

    if (error) {
      this.props.onError(error);
    }
  };

  onError = (error: Error) => {
    // flowlint-next-line unclear-type: off
    if (!((error: Object).networkError instanceof QueryAbortedError)) {
      this.setState({ error });
      this.props.onError(error);
    } else {
      this.setState({ aborted: true, error: null });
    }
  };

  componentDidMount() {
    this.subscribe();
  }

  componentDidUpdate(previousProps: Props, previousState: State) {
    if (this.state.observable !== previousState.observable) {
      this.unsubscribe();
      this.subscribe();
    }
  }

  componentWillUnmount() {
    this.unsubscribe();
  }

  render() {
    return this.props.children(this.state);
  }
}

export default withApollo(Query);
