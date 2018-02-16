import React from "react";
import PropTypes from "prop-types";

import { withStyles } from "material-ui/styles";

import warningIcon from "material-ui-icons/Warning";
import criticalIcon from "material-ui-icons/Error";
import unknownIcon from "material-ui-icons/Help";
import passingIcon from "material-ui-icons/CheckCircle";

const styles = {};

class EventStatus extends React.Component {
  static propTypes = {
    status: PropTypes.number.isRequired,
  };

  static defaultProps = {};

  render() {
    const { status } = this.props;
    let style = {};
    let Icon;

    // TODO make these nicer colours
    switch (status) {
      case 0:
        Icon = passingIcon;
        style = { color: "green" };
        break;
      case 1:
        Icon = warningIcon;
        style = { color: "orange" };
        break;
      case 2:
        Icon = criticalIcon;
        style = { color: "red" };
        break;
      default:
        Icon = unknownIcon;
        style = { color: "grey" };
    }

    return <Icon style={style} />;
  }
}

export default withStyles(styles)(EventStatus);
