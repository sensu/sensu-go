import React from "react";
import PropTypes from "prop-types";
import { compose } from "recompose";
import { withStyles } from "material-ui/styles";

import { withNamespace, namespaceShape } from "./NamespaceLink";
import NamespaceLabelBuilder from "./NamespaceLabelBuilder";

const styles = {};

class NamespaceLabel extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    currentNamespace: namespaceShape.isRequired,
  };

  render() {
    const { currentNamespace, ...props } = this.props;

    /* TODO use global variables or something for this */
    return (
      <NamespaceLabelBuilder
        org={currentNamespace.organization}
        env={currentNamespace.environment}
        icon="HalfHeart"
        iconColor="#FA8072"
        {...props}
      />
    );
  }
}

export default compose(withNamespace, withStyles(styles))(NamespaceLabel);
