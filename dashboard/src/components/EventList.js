import React from "react";
import PropTypes from "prop-types";
import map from "lodash/map";
import get from "lodash/get";
import { createFragmentContainer, graphql } from "react-relay";

import Table, {
  TableBody,
  TableHead,
  TableCell,
  TableRow,
} from "material-ui/Table";
import Checkbox from "material-ui/Checkbox";
import EventRow from "./EventRow";

class EventsList extends React.Component {
  static propTypes = {
    viewer: PropTypes.shape({ checkEvents: PropTypes.object }).isRequired,
  };

  render() {
    const { viewer } = this.props;
    const events = get(viewer, "checkEvents.edges", []);

    return (
      <Table>
        <TableHead>
          <TableRow>
            <TableCell padding="checkbox">
              <Checkbox />
            </TableCell>
            <TableCell>Entity</TableCell>
            <TableCell>Check</TableCell>
            <TableCell>Command</TableCell>
            <TableCell>Occurred At</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {map(events, (event, i) => <EventRow key={i} event={event.node} />)}
        </TableBody>
      </Table>
    );
  }
}

export default createFragmentContainer(
  EventsList,
  graphql`
    fragment EventList_viewer on Viewer {
      events(first: 1000) {
        edges {
          node {
            ...EventRow_event
          }
        }
        pageInfo {
          hasNextPage
        }
      }
    }
  `,
);
