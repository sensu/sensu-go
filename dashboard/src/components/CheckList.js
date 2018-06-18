import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableHead from "@material-ui/core/TableHead";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import Checkbox from "@material-ui/core/Checkbox";
import Row from "/components/CheckRow";

import Loader from "/components/util/Loader";
import Pagination from "/components/partials/Pagination";

class CheckList extends React.Component {
  static propTypes = {
    environment: PropTypes.shape({
      checks: PropTypes.shape({
        nodes: PropTypes.array.isRequired,
      }),
    }),
    loading: PropTypes.bool,
    onChangeQuery: PropTypes.func.isRequired,
    limit: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    offset: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
  };

  static defaultProps = {
    environment: null,
    loading: false,
    limit: undefined,
    offset: undefined,
  };

  static fragments = {
    environment: gql`
      fragment CheckList_environment on Environment {
        checks(limit: $limit, offset: $offset) {
          nodes {
            id
            ...CheckRow_check
          }

          pageInfo {
            ...Pagination_pageInfo
          }
        }
      }

      ${Row.fragments.check}
      ${Pagination.fragments.pageInfo}
    `,
  };

  render() {
    const { environment, loading, limit, offset, onChangeQuery } = this.props;
    const checks = environment ? environment.checks.nodes : [];

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
            {checks.map(node => <Row key={node.id} check={node} />)}
          </TableBody>
        </Table>

        <Pagination
          limit={limit}
          offset={offset}
          pageInfo={environment && environment.checks.pageInfo}
          onChangeQuery={onChangeQuery}
        />
      </Loader>
    );
  }
}

export default CheckList;
