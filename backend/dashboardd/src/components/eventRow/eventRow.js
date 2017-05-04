import React, {Component} from 'react';
import {TableRow, TableRowColumn} from 'material-ui/Table';

export default class EventRow extends Component {
  render() {
    const {entity, check, timestamp, ...other } = this.props;
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
}
