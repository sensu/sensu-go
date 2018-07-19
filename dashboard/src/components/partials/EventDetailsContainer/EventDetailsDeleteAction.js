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
        ns: namespace {
          org: organization
          env: environment
        }
      }
    `,
  };

  deleteEvent = () => {
    const {
      client,
      event: { id, ns },
      history,
    } = this.props;

    // Send request
    deleteEvent(client, { id }).then(() =>
      history.replace(`/${ns.org}/${ns.env}/events`),
    );
  };

  render() {
    return (
      <ConfirmDelete identifier="this event" onSubmit={this.deleteEvent}>
        {dialog => this.props.children(dialog.open)}
      </ConfirmDelete>
    );
  }
}

const enhancer = compose(withRouter, withApollo);
export default enhancer(EventDetailsDeleteAction);
