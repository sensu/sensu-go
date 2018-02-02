import React from "react";
import PropTypes from "prop-types";

import { withStyles } from "material-ui/styles";

import NamespaceLabelBuilder from "./NamespaceLabelBuilder";

const styles = {};

class NamespaceLabel extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
  };

  render() {
    const { ...props } = this.props;

    /* TODO use global variables or something for this */
    return (
      <NamespaceLabelBuilder
        org="Test Org, Inc"
        env="Production"
        icon="HalfHeart"
        iconColor="#FA8072"
        {...props}
      />
    );
  }
}

export default withStyles(styles)(NamespaceLabel);
