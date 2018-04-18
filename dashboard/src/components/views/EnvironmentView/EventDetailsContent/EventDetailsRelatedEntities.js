import React from "react";
import gql from "graphql-tag";
import Placeholder from "../../../PlaceholderCard";

class EventDetailsRelatedEntities extends React.Component {
  static fragments = {
    entity: gql`
      fragment EventDetailsRelatedEntities_entity on Entity {
        related {
          name
          lastSeen
        }
      }
    `,
  };

  render() {
    return <Placeholder />;
  }
}

export default EventDetailsRelatedEntities;
