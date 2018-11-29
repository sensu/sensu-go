import React from "react";
import PropTypes from "prop-types";
import { graphql } from "react-apollo";
import gql from "graphql-tag";

import BannerSink from "/components/relocation/BannerSink";
import Banner from "/components/relocation/Banner";

class GlobalAlert extends React.PureComponent {
  static propTypes = {
    data: PropTypes.object.isRequired,
  };

  static defaultProps = {
    data: {
      localNetwork: {
        offline: true,
      },
    },
  };

  render() {
    const { data } = this.props;

    return (
      <React.Fragment>
        {data.localNetwork.offline && (
          <BannerSink>
            <Banner variant="warning" message="offline" />
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
