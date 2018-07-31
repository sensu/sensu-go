import React from "react";
import PropTypes from "prop-types";
import cronstrue from "cronstrue";
import Tooltip from "@material-ui/core/Tooltip";

class CronDescriptor extends React.PureComponent {
  static propTypes = {
    expression: PropTypes.string.isRequired,
    component: PropTypes.oneOfType([PropTypes.func, PropTypes.string]),
  };

  static defaultProps = {
    component: "span",
  };

  render() {
    const { expression, component: Component } = this.props;
    const statement = cronstrue.toString(expression);

    return (
      <Tooltip title={expression}>
        <Component>{statement}</Component>
      </Tooltip>
    );
  }
}

export default CronDescriptor;
