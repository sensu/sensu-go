import React from "react";
import PropTypes from "prop-types";
import map from "lodash/map";
import get from "lodash/get";

import Table, {
  TableBody,
  TableHead,
  TableCell,
  TableRow,
} from "material-ui/Table";
import Checkbox from "material-ui/Checkbox";
import Paper from "material-ui/Paper";
import EventRow from "./EventRow";
import AppContent from "./AppContent";

class EventsList extends React.Component {
  static propTypes = {
    viewer: PropTypes.shape({
      events: PropTypes.array,
    }).isRequired,
  };

  render() {
    const { viewer } = this.props;
    const events = get(viewer, "events.edges", []);

    return (
      <AppContent>
        <Paper>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell checkbox>
                  <Checkbox />
                </TableCell>
                <TableCell>Entity</TableCell>
                <TableCell>Check</TableCell>
                <TableCell>Command</TableCell>
                <TableCell>Timestamp</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {map(events, (event, i) => <EventRow key={i} event={event} />)}
            </TableBody>
          </Table>
        </Paper>
      </AppContent>
    );
  }
}

export default EventsList;
