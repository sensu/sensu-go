import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { TableRow, TableCell } from "material-ui/Table";
import Checkbox from "material-ui/Checkbox";

class CheckRow extends React.Component {
  static fragments = {
    check: gql`
      fragment CheckRow_check on CheckConfig {
        name
        command
        subscriptions
        interval
        namespace {
          organization
          environment
        }
      }
    `,
  };

  render() {
    const { check, ...other } = this.props;

    return (
      <TableRow {...other}>
        <TableCell padding="checkbox">
          <Checkbox />
        </TableCell>
        <TableCell>{check.name}</TableCell>
        <TableCell>{check.command}</TableCell>
        <TableCell>{check.subscriptions}</TableCell>
        <TableCell>{check.interval}</TableCell>
      </TableRow>
    );
  }
}

CheckRow.propTypes = {
  check: PropTypes.shape({
    name: PropTypes.string.isRequired,
    command: PropTypes.string.isRequired,
    subscriptions: PropTypes.arrayOf(PropTypes.string).isRequired,
    interval: PropTypes.number.isRequired,
  }).isRequired,
};

export default CheckRow;
