import React from "react";
import PropTypes from "prop-types";
import { Route } from "react-router-dom";

import gql from "graphql-tag";
import { withStyles } from "@material-ui/core/styles";
import Button from "@material-ui/core/ButtonBase";

import NamespaceSelectorBuilder from "./NamespaceSelectorBuilder";
import NamespaceSelectorMenu from "./NamespaceSelectorMenu";

const styles = () => ({
  button: {
    width: "100%",
    padding: "8px 16px 8px 16px",
    display: "block",
    textAlign: "left",
  },
});

class NamespaceSelector extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    namespace: PropTypes.object,
    loading: PropTypes.bool,
    onChange: PropTypes.func.isRequired,
    viewer: PropTypes.object,
  };

  static defaultProps = {
    viewer: null,
    namespace: null,
    loading: false,
  };

  static fragments = {
    viewer: gql`
      fragment NamespaceSelector_viewer on Viewer {
        ...NamespaceSelectorMenu_viewer
      }

      ${NamespaceSelectorMenu.fragments.viewer}
    `,

    namespace: gql`
      fragment NamespaceSelector_namespace on Namespace {
        name
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
    const {
      classes,
      namespace,
      loading,
      onChange,
      viewer,
      ...props
    } = this.props;
    const { anchorEl } = this.state;

    return (
      <Route
        path="/:namespace"
        render={({ match: { params } }) => (
          <div {...props}>
            <Button
              aria-owns="drawer-selector-menu"
              className={classes.button}
              onClick={this.handleClick}
            >
              <NamespaceSelectorBuilder namespace={namespace} />
            </Button>
            <NamespaceSelectorMenu
              anchorEl={anchorEl}
              id="drawer-selector-menu"
              viewer={viewer}
              org={params.organization}
              open={Boolean(anchorEl)}
              onClose={this.onClose}
              onClick={ev => {
                this.onClose();
                onChange(ev);
              }}
            />
          </div>
        )}
      />
    );
  }
}

export default withStyles(styles)(NamespaceSelector);
