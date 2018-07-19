import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withRouter } from "react-router-dom";
import Button from "@material-ui/core/Button";

import ConfirmDelete from "/components/partials/ConfirmDelete";
import deleteCheck from "/mutations/deleteCheck";

class CheckDetailsDeleteAction extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    check: PropTypes.object,
    history: PropTypes.object.isRequired,
    onRequestStart: PropTypes.func.isRequired,
    onRequestEnd: PropTypes.func.isRequired,
  };

  static defaultProps = {
    check: null,
  };

  static fragments = {
    checkConfig: gql`
      fragment CheckDetailsDeleteAction_checkConfig on CheckConfig {
        id
        ns: namespace {
          org: organization
          env: environment
        }
      }
    `,
  };

  state = {
    locked: false,
  };

  requestStart() {
    this.props.onRequestStart();
    this.setState({ locked: true });
  }

  requestEnd() {
    this.props.onRequestEnd();
    this.setState({ locked: false });
  }

  deleteCheck = () => {
    const {
      client,
      check: { id, ns },
      history,
    } = this.props;

    if (this.state.locked) {
      return;
    }

    // Lock
    this.requestStart();

    // Send request
    deleteCheck(client, { id }).then(
      () => {
        this.requestEnd();
      },
      error => {
        this.requestEnd();
        throw error;
      },
    );

    // Optimistically redirect
    history.replace(`/${ns.org}/${ns.env}/checks`);
  };

  render() {
    return (
      <ConfirmDelete identifier="this check" onSubmit={this.deleteCheck}>
        {dialog => (
          <Button variant="raised" onClick={dialog.open}>
            Delete
          </Button>
        )}
      </ConfirmDelete>
    );
  }
}

export default withRouter(CheckDetailsDeleteAction);
