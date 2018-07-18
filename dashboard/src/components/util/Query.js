import React from "react";
import PropTypes from "prop-types";
import { ApolloError } from "apollo-client";
import { Query as BaseQuery } from "react-apollo";

import QueryAbortedError from "/errors/QueryAbortedError";

class Query extends React.PureComponent {
  static propTypes = {
    onError: PropTypes.func.isRequired,
    children: PropTypes.func.isRequired,
    pollInterval: PropTypes.number,
  };

  static defaultProps = {
    pollInterval: 0,
    onError(error) {
      throw error;
    },
  };

  constructor(props) {
    super(props);
    this.state = {
      pollInterval: props.pollInterval,
    };
  }

  componentDidMount() {
    const { onError } = this.props;
    this.querySubscription = this.queryRef.current.queryObservable.subscribe({
      next(result) {
        if (result.errors && result.errors.length > 0) {
          throw new ApolloError({ graphQLErrors: result.errors });
        }
      },
      error(error) {
        if (!(error.networkError instanceof QueryAbortedError)) {
          onError(error);
        }
      },
    });
  }

  componentWillUnmount() {
    if (this.querySubscription) {
      this.querySubscription.unsubscribe();
    }
  }

  queryRef = React.createRef();

  render() {
    const {
      onError,
      children,
      pollInterval: pollIntervalProp,
      ...props
    } = this.props;

    // Overwrites Apollo's existing start / stop methods and includes
    // `isPolling` prop so that the UI can reflect the current state.
    const { pollInterval } = this.state;
    const childProps = {
      isPolling: pollInterval > 0,
      startPolling: i => this.setState({ pollInterval: i || pollIntervalProp }),
      stopPolling: () => this.setState({ pollInterval: 0 }),
    };

    return (
      <BaseQuery ref={this.queryRef} pollInterval={pollInterval} {...props}>
        {queryResult => {
          const { error, ...rest } = queryResult;
          if (error && error.networkError instanceof QueryAbortedError) {
            return children({ aborted: true, ...rest, ...childProps });
          }
          return children({
            aborted: false,
            ...queryResult,
            ...childProps,
          });
        }}
      </BaseQuery>
    );
  }
}

export default Query;
