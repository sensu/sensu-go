import React from "react";
import PropTypes from "prop-types";
import hoistStatics from "hoist-non-react-statics";
import QueryParams from "./QueryParams";

const withQueryParams = keys => Component => {
  const C = props => {
    const { wrappedComponentRef, ...restProps } = props;
    return (
      <QueryParams keys={keys}>
        {(query, setQuery) => (
          <Component
            {...restProps}
            queryParams={query}
            setQueryParams={setQuery}
            ref={wrappedComponentRef}
          />
        )}
      </QueryParams>
    );
  };

  C.displayName = `withQueryParams(${Component.displayName || Component.name})`;
  C.WrappedComponent = Component;
  C.propTypes = { wrappedComponentRef: PropTypes.func };
  C.defaultProps = { wrappedComponentRef: null };

  return hoistStatics(C, Component);
};

export default withQueryParams;
