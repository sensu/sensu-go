import React from "react";
import gql from "graphql-tag";
import Placeholder from "../../../PlaceholderCard";

class EventDetailsConfiguration extends React.Component {
  static fragments = {
    entity: gql`
      fragment EventDetailsConfiguration_entity on Entity {
        class
        system {
          platform
        }
        lastSeen
        subscriptions
      }
    `,
    check: gql`
      fragment EventDetailsConfiguration_check on Check {
        command
        interval
        subscriptions
        # timeout
        # TTL
      }
    `,
  };

  render() {
    return <Placeholder tall />;
  }
}

export default EventDetailsConfiguration;
