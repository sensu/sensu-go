import React from "react";
import PropTypes from "prop-types";
import { ApolloError } from "apollo-client";
import { Query as BaseQuery } from "react-apollo";

import QueryAbortedError from "/errors/QueryAbortedError";

class Query extends React.PureComponent {
  static propTypes = {
    onError: PropTypes.func.isRequired,
    children: PropTypes.func.isRequired,
  };

  static defaultProps = {
    onError(error) {
      throw error;
    },
  };

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
    const { onError, children, ...props } = this.props;

    return (
      <BaseQuery ref={this.queryRef} {...props}>
        {queryResult => {
          const { error, ...rest } = queryResult;
          if (error && error.networkError instanceof QueryAbortedError) {
            return children({ aborted: true, ...rest });
          }
          return children({ aborted: false, ...queryResult });
        }}
      </BaseQuery>
    );
  }
}

export default Query;
