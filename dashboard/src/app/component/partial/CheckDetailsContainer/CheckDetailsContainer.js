import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Content from "/lib/component/base/Content";
import Grid from "@material-ui/core/Grid";
import Loader from "/lib/component/base/Loader";

import Configuration from "./CheckDetailsConfiguration";
import Toolbar from "./CheckDetailsToolbar";

class CheckDetailsContainer extends React.PureComponent {
  static propTypes = {
    check: PropTypes.object,
    loading: PropTypes.bool.isRequired,
    refetch: PropTypes.func,
  };

  static defaultProps = {
    check: null,
    refetch: () => null,
  };

  static fragments = {
    check: gql`
      fragment CheckDetailsContainer_check on CheckConfig {
        id
        deleted @client

        ...CheckDetailsToolbar_check
        ...CheckDetailsConfiguration_check
      }

      ${Toolbar.fragments.check}
      ${Configuration.fragments.check}
    `,
  };

  render() {
    const { check, loading, refetch } = this.props;

    return (
      <Loader loading={loading} passthrough>
        {check && (
          <React.Fragment>
            <Content marginBottom>
              <Toolbar check={check} refetch={refetch} />
            </Content>

            <Grid container spacing={16}>
              <Grid item xs={12}>
                <Configuration check={check} />
              </Grid>
            </Grid>
          </React.Fragment>
        )}
      </Loader>
    );
  }
}

export default CheckDetailsContainer;
