import React from "react";
import PropTypes from "prop-types";
import map from "lodash/map";
import gql from "graphql-tag";

import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableHead from "@material-ui/core/TableHead";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import Checkbox from "@material-ui/core/Checkbox";
import Row from "/components/CheckRow";

import Loader from "/components/util/Loader";

class CheckList extends React.Component {
  static propTypes = {
    environment: PropTypes.shape({
      checks: PropTypes.shape({
        nodes: PropTypes.array.isRequired,
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
        checks(limit: $limit) {
          nodes {
            id
            ...CheckRow_check
          }
        }
      }

      ${Row.fragments.check}
    `,
  };

  render() {
    const { environment, loading } = this.props;
    const checks = environment && (environment.checks.nodes || []);

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
            {map(checks, node => <Row key={node.id} check={node} />)}
          </TableBody>
        </Table>
      </Loader>
    );
  }
}

export default CheckList;
