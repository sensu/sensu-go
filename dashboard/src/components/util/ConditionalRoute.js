import React from "react";
import PropTypes from "prop-types";
import { Route } from "react-router-dom";

class ConditionalRoute extends React.PureComponent {
  static propTypes = {
    ...Route.propTypes,
    component: PropTypes.func,
    render: PropTypes.func,
    active: PropTypes.bool,

    fallbackRender: PropTypes.func,
    fallbackComponent: PropTypes.func,
  };

  static defaultProps = {
    component: undefined,
    render: undefined,
    active: false,
    redirectPath: "/",
    fallbackRender: undefined,
    fallbackComponent: undefined,
  };

  render() {
    const {
      active,
      redirectPath,
      render,
      component,
      fallbackComponent,
      fallbackRender,
      ...props
    } = this.props;

    if (active) {
      return <Route {...props} render={render} component={component} />;
    }

    if (fallbackComponent || fallbackRender) {
      return (
        <Route
          {...props}
          component={fallbackComponent}
          render={fallbackRender}
        />
      );
    }

    return null;
  }
}

export default ConditionalRoute;
