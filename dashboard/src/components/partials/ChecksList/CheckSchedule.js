import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import CronDescriptor from "/components/partials/CronDescriptor";

class CheckSchedule extends React.Component {
  static propTypes = {
    check: PropTypes.object.isRequired,
  };

  static fragments = {
    check: gql`
      fragment CheckSchedule_check on CheckConfig {
        subscriptions
        interval
        cron
        publish
      }
    `,
  };

  render() {
    const { check } = this.props;

    return (
      <React.Fragment>
        {check.publish && (
          <span>
            Published: Scheduled{" "}
            <strong>
              {check.interval ? (
                `
                every
                ${check.interval}
                ${check.interval === 1 ? "second" : "seconds"}
              `
              ) : (
                <CronDescriptor expression={check.cron} />
              )}
            </strong>{" "}
            by{" "}
            <strong>
              {check.subscriptions.length}{" "}
              {check.subscriptions.length === 1
                ? "subscription"
                : "subscriptions"}
            </strong>
            .
          </span>
        )}
        {!check.publish && (
          <span>Unpublished: This check is not scheduled.</span>
        )}
      </React.Fragment>
    );
  }
}

export default CheckSchedule;
