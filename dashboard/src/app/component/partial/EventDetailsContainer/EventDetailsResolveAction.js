import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";
import resolveEvent from "/lib/mutation/resolveEvent";

class EventDetailsResolveAction extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    children: PropTypes.func.isRequired,
    event: PropTypes.object,
  };

  static defaultProps = {
    event: null,
  };

  static fragments = {
    event: gql`
      fragment EventDetailsResolveAction_event on Event {
        id
        check {
          status
        }
      }
    `,
  };

  resolveEvent = () => {
    const { client, event } = this.props;
    if (event.check.status === 0) {
      return;
    }

    resolveEvent(client, event);
  };

  render() {
    const canResolve = this.props.event && this.props.event.check.status > 0;
    const childProps = {
      canResolve,
      resolve: this.resolveEvent,
    };

    return this.props.children(childProps);
  }
}

export default withApollo(EventDetailsResolveAction);
