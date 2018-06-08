import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withRouter } from "react-router-dom";
import Button from "@material-ui/core/Button";
import ConfirmDelete from "/components/partials/ConfirmDelete";
import deleteEvent from "/mutations/deleteEvent";

class EventDetailsDeleteAction extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    event: PropTypes.object,
    history: PropTypes.object.isRequired,
    onRequestStart: PropTypes.func.isRequired,
    onRequestEnd: PropTypes.func.isRequired,
  };

  static defaultProps = {
    event: null,
  };

  static fragments = {
    event: gql`
      fragment EventDetailsDeleteAction_event on Event {
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

  deleteEvent = () => {
    const {
      client,
      event: { id, ns },
      history,
    } = this.props;
    if (this.state.locked) {
      return;
    }

    // Cleanup
    this.requestStart();

    // Send request
    deleteEvent(client, { id }).then(
      () => {
        this.requestEnd();
        history.replace(`/${ns.org}/${ns.env}/events`);
      },
      error => {
        this.requestEnd();
        throw error;
      },
    );
  };

  render() {
    return (
      <ConfirmDelete identifier="this event" onSubmit={this.deleteEvent}>
        {dialog => (
          <Button variant="raised" onClick={dialog.open}>
            Delete
          </Button>
        )}
      </ConfirmDelete>
    );
  }
}

export default withRouter(EventDetailsDeleteAction);
