import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { compose } from "recompose";
import { withRouter } from "react-router-dom";
import { withApollo } from "react-apollo";

import ConfirmDelete from "/components/partials/ConfirmDelete";
import deleteEvent from "/mutations/deleteEvent";

class EventDetailsDeleteAction extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
    client: PropTypes.object.isRequired,
    event: PropTypes.object,
    history: PropTypes.object.isRequired,
  };

  static defaultProps = {
    event: null,
  };

  static fragments = {
    event: gql`
      fragment EventDetailsDeleteAction_event on Event {
        id
        namespace
      }
    `,
  };

  deleteEvent = () => {
    const { client, event, history } = this.props;
    const { id, namespace } = event;

    // Send request
    deleteEvent(client, { id });

    // Optimistically redirect
    history.replace(`/${namespace}/events`);
  };

  render() {
    return (
      <ConfirmDelete identifier="this event" onSubmit={this.deleteEvent}>
        {dialog => this.props.children(dialog.open)}
      </ConfirmDelete>
    );
  }
}

const enhancer = compose(
  withRouter,
  withApollo,
);
export default enhancer(EventDetailsDeleteAction);
