import React from 'react';
import PropTypes from 'prop-types';
import { Table, TableBody, TableHeader, TableHeaderColumn, TableRow } from 'material-ui/Table';
import map from 'lodash/map';

import EventRow from 'components/eventRow';

const styles = require('./eventsList.css');

function EventsList({ events }) {
  return (
    <Table className={styles.table}>
      <TableHeader>
        <TableRow>
          <TableHeaderColumn>Entity</TableHeaderColumn>
          <TableHeaderColumn>Check</TableHeaderColumn>
          <TableHeaderColumn>Command</TableHeaderColumn>
          <TableHeaderColumn>Timestamp</TableHeaderColumn>
        </TableRow>
      </TableHeader>
      <TableBody>
        {map(events, (event, i) => (
          <EventRow key={i} {...event} />
        ))}
      </TableBody>
    </Table>
  );
}

EventsList.propTypes = {
  events: PropTypes.arrayOf(PropTypes.object).isRequired,
};

export default EventsList;
