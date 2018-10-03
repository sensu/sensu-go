import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Content from "/components/Content";
import Grid from "@material-ui/core/Grid";
import Loader from "/components/util/Loader";
<<<<<<< HEAD
=======
import Maybe from "/components/Maybe";
import Monospaced from "/components/Monospaced";
import SilencedIcon from "/icons/Silence";
import Typography from "@material-ui/core/Typography";
import Tooltip from "@material-ui/core/Tooltip";
import QueueIcon from "@material-ui/icons/Queue";
import PublishIcon from "@material-ui/icons/Publish";
import UnpublishIcon from "/icons/Unpublish";
>>>>>>> Add buttons for publish and unpublish

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
