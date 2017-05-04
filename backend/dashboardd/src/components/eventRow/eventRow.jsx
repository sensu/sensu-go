import React from 'react';
import PropTypes from 'prop-types';
import { TableRow, TableRowColumn } from 'material-ui/Table';

function EventRow({ entity, check, timestamp, ...other }) {
  return (
    <TableRow {...other}>
      {other.children[0] /* checkbox passed down from TableBody*/}
      <TableRowColumn>{entity.id}</TableRowColumn>
      <TableRowColumn>{check.name}</TableRowColumn>
      <TableRowColumn>{check.command}</TableRowColumn>
      <TableRowColumn>{timestamp}</TableRowColumn>
    </TableRow>
  );
}

EventRow.propTypes = {
  entity: PropTypes.shape({ id: '' }).isRequired,
  check: PropTypes.shape({ name: '', command: '' }).isRequired,
  timestamp: PropTypes.number.isRequired,
};

export default EventRow;
