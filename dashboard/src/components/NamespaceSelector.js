import React from "react";
import PropTypes from "prop-types";
import { Route } from "react-router-dom";

import gql from "graphql-tag";
import { withStyles } from "material-ui/styles";
import Button from "material-ui/ButtonBase";

import NamespaceSelectorBuilder from "./NamespaceSelectorBuilder";
import NamespaceSelectorMenu from "./NamespaceSelectorMenu";

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
    classes: PropTypes.object.isRequired,
    viewer: PropTypes.object,
    environment: PropTypes.object,
    loading: PropTypes.bool,
  };

  static defaultProps = {
    viewer: null,
    environment: null,
    loading: false,
  };

  static fragments = {
    viewer: gql`
      fragment NamespaceSelector_viewer on Viewer {
        ...NamespaceSelectorMenu_viewer
      }

      ${NamespaceSelectorMenu.fragments.viewer}
    `,

    environment: gql`
      fragment NamespaceSelector_environment on Environment {
        name
        organization {
          name
        }
      }
    `,
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
    const { loading, environment, viewer, classes, ...props } = this.props;
    const { anchorEl } = this.state;

    return (
      <Route
        path="/:organization/:environment"
        render={({ match: { params } }) => (
          <div {...props}>
            <Button
              aria-owns="drawer-selector-menu"
              className={classes.button}
              onClick={this.handleClick}
            >
              <NamespaceSelectorBuilder environment={environment} />
            </Button>
            <NamespaceSelectorMenu
              viewer={viewer}
              anchorEl={anchorEl}
              open={Boolean(anchorEl)}
              onClose={this.onClose}
              id="drawer-selector-menu"
              org={params.organization}
            />
          </div>
        )}
      />
    );
  }
}

export default withStyles(styles)(NamespaceSelector);
