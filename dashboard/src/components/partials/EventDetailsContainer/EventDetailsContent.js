import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Grid from "material-ui/Grid";

import Content from "/components/Content";
import Loader from "/components/Loader";
import ButtonSet from "/components/ButtonSet";
import CheckResult from "./EventDetailsCheckResult";
import RelatedEntities from "./EventDetailsRelatedEntities";
import Configuration from "./EventDetailsConfiguration";
import DeleteAction from "./EventDetailsDeleteAction";
import ResolveAction from "./EventDetailsResolveAction";

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
        ...EventDetailsDeleteAction_event
        ...EventDetailsResolveAction_event

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
      ${DeleteAction.fragments.event}
      ${ResolveAction.fragments.event}
    `,
  };

  state = {
    pendingRequests: 0,
  };

  handleRequestStart = () => {
    this.setState(({ pendingRequests }) => ({
      pendingRequests: pendingRequests + 1,
    }));
  };

  handleRequestEnd = () => {
    this.setState(({ pendingRequests }) => ({
      pendingRequests: pendingRequests - 1,
    }));
  };

  render() {
    const { client, event, loading } = this.props;
    const { pendingRequests } = this.state;
    const hasPendingRequests = pendingRequests > 0;

    return (
      <Loader loading={loading || hasPendingRequests} passthrough>
        {event && (
          <React.Fragment>
            <Content marginBottom>
              <div style={{ flexGrow: 1 }} />
              <ButtonSet>
                <ResolveAction client={client} event={event} />
                <DeleteAction
                  client={client}
                  event={event}
                  onRequestStart={this.handleRequestStart}
                  onRequestEnd={this.handleRequestEnd}
                />
              </ButtonSet>
            </Content>
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
