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
import Row from "./CheckRow";
import AppContent from "./AppContent";

const fakeCheckData = [
  {
    edge: "xxx",
    node: {
      name: "evil-check",
      command: "rm -rf /",
      subscriptions: ["unix"],
      interval: 30,
    },
  },
  {
    edge: "yyy",
    node: {
      name: "disk-full",
      command: "df -hs /var",
      subscriptions: ["mysql"],
      interval: 10,
    },
  },
  {
    edge: "zzz",
    node: {
      name: "is-google-up",
      command: "curl google.com",
      subscriptions: ["dunno", "unix"],
      interval: 5,
    },
  },
];

class CheckList extends React.Component {
  static propTypes = {
    viewer: PropTypes.shape({
      checks: PropTypes.shape({
        edges: PropTypes.array.isRequired,
      }),
    }).isRequired,
  };

  render() {
    const { viewer } = this.props;
    const checks = get(viewer, "checks.edges", fakeCheckData);

    return (
      <AppContent>
        <Paper>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell padding="checkbox">
                  <Checkbox />
                </TableCell>
                <TableCell>Check</TableCell>
                <TableCell>Command</TableCell>
                <TableCell>Subscribers</TableCell>
                <TableCell>Interval</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {map(checks, edge => <Row key={edge.cursor} check={edge.node} />)}
            </TableBody>
          </Table>
        </Paper>
      </AppContent>
    );
  }
}

export default CheckList;
