import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { RelativeToCurrentDate } from "/components/RelativeDate";

class EventStatusDescriptor extends React.PureComponent {
  static propTypes = {
    check: PropTypes.object,
    compact: PropTypes.bool,
    event: PropTypes.object.isRequired,
  };

  static defaultProps = {
    check: null,
    compact: false,
  };

  static fragments = {
    check: gql`
      fragment EventStatusDescriptor_check on Check {
        lastOK
        status
        occurrences
      }
    `,
    event: gql`
      fragment EventStatusDescriptor_event on Event {
        timestamp
      }
    `,
  };

  numFormatter = new Intl.NumberFormat();

  // Metric received X minutes ago.
  renderMetricDescription() {
    const { event } = this.props;

    return (
      <React.Fragment>
        Metric received
        <strong>
          <RelativeToCurrentDate dateTime={event.timestamp} />
        </strong>.
      </React.Fragment>
    );
  }

  // Last executed X minutes ago.
  renderLastExecution() {
    const { event } = this.props;

    return (
      <React.Fragment>
        Last executed{" "}
        <strong>
          <RelativeToCurrentDate dateTime={event.timestamp} />
        </strong>.
      </React.Fragment>
    );
  }

  // Incident started X minutes ago and has continued for Y consecutive
  // executions.
  renderIncidentDescription() {
    const { check, compact } = this.props;

    return (
      <React.Fragment>
        Incident started{" "}
        <strong>
          <RelativeToCurrentDate dateTime={check.lastOK} />
        </strong>
        {!compact && (
          <React.Fragment>
            {" "}
            and has continued for{" "}
            <strong>{this.numFormatter.format(check.occurrences)}</strong>{" "}
            consecutive executions
          </React.Fragment>
        )}
        .
      </React.Fragment>
    );
  }

  // Event began returning unknown status {status} Y minutes ago.
  // Starting X minutes ago, began returning unknown status code Y.
  renderUnknownStatusDescription() {
    const { check } = this.props;

    return (
      <React.Fragment>
        Starting{" "}
        <strong>
          <RelativeToCurrentDate dateTime={check.lastOK} />
        </strong>
        {", "}
        event began returning unknown status code{" "}
        <strong>{check.status}</strong>.
      </React.Fragment>
    );
  }

  render() {
    const { check } = this.props;

    if (check === null) {
      return this.renderMetricDescription();
    }

    if (check.status === 0 || check.lastOK === null) {
      return this.renderLastExecution();
    }

    if (check.status > 2) {
      return this.renderUnknownStatusDescription();
    }

    return this.renderIncidentDescription();
  }
}

export default EventStatusDescriptor;
