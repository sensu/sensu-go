import React from "react";
import PropTypes from "prop-types";

import { withStyles } from "material-ui/styles";

import NamespaceSelectorBuilder from "./NamespaceSelectorBuilder";

const styles = {};

class NamespaceSelector extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
  };

  render() {
    const { ...props } = this.props;

    /* TODO use global variables or something for this */
    return (
      <NamespaceSelectorBuilder
        org="Test Org, Inc"
        env="Production"
        {...props}
      />
    );
  }
}

export default withStyles(styles)(NamespaceSelector);
