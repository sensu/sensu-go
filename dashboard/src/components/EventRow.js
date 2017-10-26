import React from "react";
import PropTypes from "prop-types";
import { createFragmentContainer, graphql } from "react-relay";
import moment from "moment";

import { TableRow, TableCell } from "material-ui/Table";
import Checkbox from "material-ui/Checkbox";

class EventRow extends React.Component {
  render() {
    const { event: { entity, config, timestamp }, ...other } = this.props;
    const time = moment(timestamp).fromNow();

    return (
      <TableRow {...other}>
        <TableCell padding="checkbox">
          <Checkbox />
        </TableCell>
        <TableCell>{entity.uid}</TableCell>
        <TableCell>{config.name}</TableCell>
        <TableCell>{config.command}</TableCell>
        <TableCell>{time}</TableCell>
      </TableRow>
    );
  }
}

EventRow.propTypes = {
  event: PropTypes.shape({
    entity: PropTypes.shape({ id: "" }).isRequired,
    config: PropTypes.shape({ name: "", command: "" }).isRequired,
    timestamp: PropTypes.string.isRequired,
  }).isRequired,
};

export default createFragmentContainer(
  EventRow,
  graphql`
    fragment EventRow_event on CheckEvent {
      ... on CheckEvent {
        timestamp
        config {
          name
          command
        }
        entity {
          uid: entityId
        }
      }
    }
  `,
);
