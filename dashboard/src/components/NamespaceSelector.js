import React from "react";
import PropTypes from "prop-types";

import { compose } from "recompose";
import { createFragmentContainer, graphql } from "react-relay";
import { withStyles } from "material-ui/styles";
import Button from "material-ui/ButtonBase";

import NamespaceSelectorBuilder from "./NamespaceSelectorBuilder";
import NamespaceSelectorMenu from "./NamespaceSelectorMenu";
import { withNamespace, namespaceShape } from "./NamespaceLink";

const styles = {
  button: {
    width: "100%",
    padding: "8px 16px 8px 16px",
    display: "block",
    textAlign: "left",
  },
};

class NamespaceSelector extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    currentNamespace: namespaceShape.isRequired,
    viewer: PropTypes.objectOf(PropTypes.any).isRequired,
  };

  state = {
    anchorEl: null,
  };

  onClose = () => {
    this.setState({ anchorEl: null });
  };

  handleClick = event => {
    this.setState({ anchorEl: event.currentTarget });
  };

  render() {
    const { viewer, currentNamespace, classes, ...props } = this.props;
    const { anchorEl } = this.state;

    /* TODO use global variables or something for this */
    return (
      <div {...props}>
        <Button
          aria-owns="test"
          className={classes.button}
          onClick={this.handleClick}
        >
          <NamespaceSelectorBuilder
            org={currentNamespace.organization}
            env={currentNamespace.environment}
          />
        </Button>
        <NamespaceSelectorMenu
          viewer={viewer}
          anchorEl={anchorEl}
          open={Boolean(anchorEl)}
          onClose={this.onClose}
          id="drawer-selector-menu"
        />
      </div>
    );
  }
}

export default createFragmentContainer(
  compose(withStyles(styles), withNamespace)(NamespaceSelector),
  graphql`
    fragment NamespaceSelector_viewer on Viewer {
      ...NamespaceSelectorMenu_viewer
    }
  `,
);
