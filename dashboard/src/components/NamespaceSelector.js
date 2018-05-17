import React from "react";
import PropTypes from "prop-types";
import { Route } from "react-router-dom";

import gql from "graphql-tag";
import { withStyles } from "@material-ui/core/styles";
import Button from "@material-ui/core/ButtonBase";

import NamespaceSelectorBuilder from "/components/NamespaceSelectorBuilder";
import NamespaceSelectorMenu from "/components/NamespaceSelectorMenu";

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
    environment: PropTypes.object,
    loading: PropTypes.bool,
    onChange: PropTypes.func.isRequired,
    viewer: PropTypes.object,
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
    const {
      classes,
      environment,
      loading,
      onChange,
      viewer,
      ...props
    } = this.props;
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
