import React from "react";
import PropTypes from "prop-types";
import { createFragmentContainer, graphql } from "react-relay";
import moment from "moment";

import { TableRow, TableCell } from "material-ui/Table";
import Checkbox from "material-ui/Checkbox";

class EventRow extends React.Component {
  render() {
    const { event: { entity, check, timestamp }, ...other } = this.props;
    const time = moment(timestamp).fromNow();

    return (
      <TableRow {...other}>
        <TableCell padding="checkbox">
          <Checkbox />
        </TableCell>
        <TableCell>{entity.name}</TableCell>
        <TableCell>{check.config.name}</TableCell>
        <TableCell>{check.config.command}</TableCell>
        <TableCell>{time}</TableCell>
      </TableRow>
    );
  }
}

EventRow.propTypes = {
  event: PropTypes.shape({
    entity: PropTypes.shape({ id: "" }).isRequired,
    check: PropTypes.shape({
      config: PropTypes.shape({ name: "", command: "" }),
    }).isRequired,
    timestamp: PropTypes.string.isRequired,
  }).isRequired,
};

export default createFragmentContainer(
  EventRow,
  graphql`
    fragment EventRow_event on Event {
      ... on Event {
        timestamp
        check {
          config {
            name
            command
          }
        }
        entity {
          name
        }
      }
    }
  `,
);
