import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Grid from "material-ui/Grid";

import Content from "/components/Content";
import Loader from "/components/Loader";
import CheckResult from "./EventDetailsCheckResult";
import RelatedEntities from "./EventDetailsRelatedEntities";
import Configuration from "./EventDetailsConfiguration";
import Actions from "./EventDetailsActions";

class EventDetailsContainer extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    event: PropTypes.object,
    loading: PropTypes.bool.isRequired,
  };

  static defaultProps = {
    event: null,
  };

  static fragments = {
    event: gql`
      fragment EventDetailsContainer_event on Event {
        id
        timestamp
        deleted @client
        ...EventDetailsActions_event

        check {
          ...EventDetailsCheckResult_check
          ...EventDetailsConfiguration_check
        }
        entity {
          ...EventDetailsCheckResult_entity
          ...EventDetailsRelatedEntities_entity
          ...EventDetailsConfiguration_entity
        }
      }

      ${CheckResult.fragments.check}
      ${CheckResult.fragments.entity}
      ${RelatedEntities.fragments.entity}
      ${Configuration.fragments.check}
      ${Configuration.fragments.entity}
      ${Actions.fragments.event}
    `,
  };

  state = {
    processing: false,
  };

  handleProcessing = newVal => {
    this.setState({ processing: newVal });
  };

  render() {
    const { client, event, loading } = this.props;
    const { processing } = this.state;

    return (
      <Loader loading={loading || processing} passthrough>
        {event && (
          <React.Fragment>
            <Actions
              client={client}
              event={event}
              onProcessing={this.handleProcessing}
            />
            <Content>
              <Grid container>
                <Grid item xs={12}>
                  <CheckResult check={event.check} entity={event.entity} />
                </Grid>
                <Grid item xs={12} md={6}>
                  <RelatedEntities entity={event.entity} />
                </Grid>
                <Grid item xs={12} md={6}>
                  <Configuration check={event.check} entity={event.entity} />
                </Grid>
              </Grid>
            </Content>
          </React.Fragment>
        )}
      </Loader>
    );
  }
}

export default EventDetailsContainer;
