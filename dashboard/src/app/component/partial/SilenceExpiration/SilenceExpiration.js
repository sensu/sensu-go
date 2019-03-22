import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { RelativeToCurrentDate } from "/lib/component/base/RelativeDate";

class SilenceExpiration extends React.Component {
  static propTypes = {
    silence: PropTypes.object.isRequired,
  };

  static fragments = {
    silence: gql`
      fragment SilenceExpiration_silence on Silenced {
        expireOnResolve
        expires
      }
    `,
  };

  render() {
    const { expires, expireOnResolve } = this.props.silence;

    if (expires && expireOnResolve) {
      return (
        <React.Fragment>
          Expires when <strong>resolved</strong> or{" "}
          <strong>
            <RelativeToCurrentDate dateTime={expires} />
          </strong>
          .
        </React.Fragment>
      );
    } else if (expireOnResolve) {
      return (
        <React.Fragment>
          Expires when <strong>resolved</strong>.
        </React.Fragment>
      );
    } else if (expires) {
      return (
        <React.Fragment>
          Expires{" "}
          <strong>
            <RelativeToCurrentDate dateTime={expires} />
          </strong>
          .
        </React.Fragment>
      );
    }
    return "Does not expire.";
  }
}

export default SilenceExpiration;
