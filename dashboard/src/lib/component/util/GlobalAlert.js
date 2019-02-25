import React from "react";
import PropTypes from "prop-types";
import { graphql } from "react-apollo";
import gql from "graphql-tag";

import BannerSink from "/lib/component/relocation/BannerSink";
import RetryConnectionBanner from "/lib/component/relocation/RetryConnectionBanner";

class GlobalAlert extends React.PureComponent {
  static propTypes = {
    data: PropTypes.object.isRequired,
  };

  render() {
    const { data } = this.props;

    return (
      <React.Fragment>
        {data.localNetwork && data.localNetwork.offline && (
          <BannerSink>
            <RetryConnectionBanner />
          </BannerSink>
        )}
      </React.Fragment>
    );
  }
}

export default graphql(gql`
  query GlobalAlertQuery {
    localNetwork @client {
      offline
      retry
    }
  }
`)(GlobalAlert);
