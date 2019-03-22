// @flow
/* eslint-disable react/no-unused-prop-types */
/* eslint-disable react/sort-comp */
/* eslint-disable react/no-multi-comp */

import * as React from "react";
import { ApolloError } from "apollo-client";
import { withApollo } from "react-apollo";
import type {
  ObservableQuery,
  ApolloClient,
  WatchQueryOptions,
  NetworkStatus,
  ApolloQueryResult,
} from "react-apollo";
import shallowEqual from "fbjs/lib/shallowEqual";
import gql from "graphql-tag";

import QueryAbortedError from "/lib/error/QueryAbortedError";

type ObservableMethods = {
  fetchMore: $PropertyType<ObservableQuery<mixed>, "fetchMore">,
  refetch: $PropertyType<ObservableQuery<mixed>, "refetch">,
  startPolling: $PropertyType<ObservableQuery<mixed>, "startPolling">,
  stopPolling: $PropertyType<ObservableQuery<mixed>, "stopPolling">,
  subscribeToMore: $PropertyType<ObservableQuery<mixed>, "subscribeToMore">,
  updateQuery: $PropertyType<ObservableQuery<mixed>, "updateQuery">,
};

type Props = {
  client: ApolloClient<mixed>,
  // eslint-disable-next-line no-use-before-define
  children: State => React.Node,
  onError: Error => void,
} & WatchQueryOptions;

type LocalData = { localNetwork: { offline: boolean, retry: boolean } };

type State = {
  aborted: boolean,
  data: mixed | null, // TODO: Infer data type from query
  error: Error | null,
  loading: boolean,
  networkStatus: NetworkStatus,
  observable: ObservableQuery<mixed>,
  localQuery: {
    observable: ObservableQuery<LocalData>,
    data: LocalData,
  },
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

const localQuery = gql`
  query LocalNetworkStatusQuery {
    localNetwork @client {
      offline
      retry
    }
  }
`;

const createQueryObservable = (props: Props) => {
  const observable: ObservableQuery<mixed> = props.client.watchQuery(
    extractQueryOptions(props),
  );

  // retrieve the result of the query from the local cache
  const { data, loading, networkStatus } = observable.currentResult();

  return {
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
};

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
  localSubscription: { unsubscribe(): void } | null = null;

  static getDerivedStateFromProps(props: Props, state: State | null) {
    if (state !== null && state.props === props) {
      return null;
    }

    let nextState: $Shape<State> = { props };

    if (state === null || state.props.client !== props.client) {
      const observable: ObservableQuery<LocalData> = props.client.watchQuery({
        query: localQuery,
      });

      const { data } = observable.currentResult();

      nextState = {
        ...nextState,
        localQuery: {
          observable,
          // flowlint-next-line unclear-type: off
          data: ((data: any): LocalData),
        },
      };
    }

    if (
      state === null ||
      state.props.client !== props.client ||
      state.props.query !== props.query
    ) {
      nextState = {
        ...nextState,
        ...createQueryObservable(props),
      };
    } else if (
      state !== null &&
      modifiableWatchQueryOptionsHaveChanged(state.props, props)
    ) {
      state.observable.setOptions(extractQueryOptions(props));
    }

    // Changes to `metadata` and `context` props are ignored.

    return nextState;
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

  subscribeLocal() {
    if (this.localSubscription) {
      throw new Error("Cannot subscribe. Currently subscribed.");
    }

    this.localSubscription = this.state.localQuery.observable.subscribe({
      next: this.onNextLocal,
      error: this.onErrorLocal,
    });
  }

  unsubscribe() {
    if (!this.subscription) {
      throw new Error("Cannot unsubscribe. Not currently subscribed");
    }

    this.subscription.unsubscribe();
    this.subscription = null;
  }

  unsubscribeLocal() {
    if (!this.localSubscription) {
      throw new Error("Cannot unsubscribe. Not currently subscribed");
    }

    this.localSubscription.unsubscribe();
    this.localSubscription = null;
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

    this.setState({
      aborted: false,
      error,
      data,
      loading,
      networkStatus,
    });

    if (error) {
      this.props.onError(error);
    }
  };

  onNextLocal = ({ data }: ApolloQueryResult<LocalData>) => {
    this.setState((state: State, props: Props) => {
      let nextState = {
        localQuery: {
          ...state.localQuery,
          data,
        },
      };

      if (
        state.localQuery.data.localNetwork.offline &&
        (!nextState.localQuery.data.localNetwork.offline ||
          nextState.localQuery.data.localNetwork.retry)
      ) {
        nextState = {
          ...nextState,
          ...createQueryObservable(props),
        };
      }

      return nextState;
    });
  };

  onError = (error: Error) => {
    // flowlint-next-line unclear-type: off
    if ((error: Object).networkError instanceof QueryAbortedError) {
      this.setState({ aborted: true, error: null });
    } else {
      this.setState({ error });
      this.props.onError(error);
    }
  };

  onErrorLocal = (error: Error) => {
    throw error;
  };

  componentDidMount() {
    this.subscribe();
    this.subscribeLocal();
  }

  componentDidUpdate(previousProps: Props, previousState: State) {
    if (this.state.observable !== previousState.observable) {
      this.unsubscribe();
      this.subscribe();
    }

    if (
      this.state.localQuery.observable !== previousState.localQuery.observable
    ) {
      this.unsubscribeLocal();
      this.subscribeLocal();
    }
  }

  componentWillUnmount() {
    this.unsubscribe();
    this.unsubscribeLocal();
  }

  render() {
    return this.props.children(this.state);
  }
}

export default withApollo(Query);
