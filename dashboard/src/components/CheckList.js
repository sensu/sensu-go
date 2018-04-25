import React from "react";
import PropTypes from "prop-types";
import map from "lodash/map";
import get from "lodash/get";
import gql from "graphql-tag";

import Table, {
  TableBody,
  TableHead,
  TableCell,
  TableRow,
} from "material-ui/Table";
import Checkbox from "material-ui/Checkbox";
import Row from "/components/CheckRow";

import Loader from "/components/Loader";

class CheckList extends React.Component {
  static propTypes = {
    environment: PropTypes.shape({
      checks: PropTypes.shape({
        edges: PropTypes.array.isRequired,
      }),
    }),
    loading: PropTypes.bool,
  };

  static defaultProps = {
    environment: null,
    loading: false,
  };

  static fragments = {
    environment: gql`
      fragment CheckList_environment on Environment {
        checks(first: 1500) {
          edges {
            node {
              ...CheckRow_check
            }
            cursor
          }
          pageInfo {
            hasNextPage
          }
        }
      }

      ${Row.fragments.check}
    `,
  };

  render() {
    const { environment, loading } = this.props;
    const checks = get(environment, "checks.edges", []);

    return (
      <Loader loading={loading}>
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
          <TableBody style={{ minHeight: 200 }}>
            {map(checks, edge => <Row key={edge.cursor} check={edge.node} />)}
          </TableBody>
        </Table>
      </Loader>
    );
  }
}

export default CheckList;
