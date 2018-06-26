import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Grid from "@material-ui/core/Grid";

import Loader from "/components/util/Loader";

import Content from "/components/Content";
import ButtonSet from "/components/ButtonSet";
import RelatedEntitiesCard from "/components/partials/RelatedEntitiesCard";
import CheckResult from "./EventDetailsCheckResult";
import Summary from "./EventDetailsSummary";
import DeleteAction from "./EventDetailsDeleteAction";
import ResolveAction from "./EventDetailsResolveAction";
import ReRunAction from "./EventDetailsReRunAction";

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
        ...EventDetailsReRunAction_event

        check {
          ...EventDetailsCheckResult_check
          ...EventDetailsSummary_check
        }
        entity {
          ...EventDetailsCheckResult_entity
          ...RelatedEntitiesCard_entity
          ...EventDetailsSummary_entity
        }
      }

      ${CheckResult.fragments.check}
      ${CheckResult.fragments.entity}
      ${RelatedEntitiesCard.fragments.entity}
      ${Summary.fragments.check}
      ${Summary.fragments.entity}
      ${DeleteAction.fragments.event}
      ${ResolveAction.fragments.event}
      ${ReRunAction.fragments.event}
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
            <Content bottomMargin>
              <div style={{ flexGrow: 1 }} />
              <ButtonSet>
                <ResolveAction client={client} event={event} />
                <ReRunAction client={client} event={event} />
                <DeleteAction
                  client={client}
                  event={event}
                  onRequestStart={this.handleRequestStart}
                  onRequestEnd={this.handleRequestEnd}
                />
              </ButtonSet>
            </Content>
            <Content>
              <Grid container spacing={16}>
                <Grid item xs={12}>
                  <CheckResult check={event.check} entity={event.entity} />
                </Grid>
                <Grid item xs={12} md={6}>
                  <RelatedEntitiesCard entity={event.entity} />
                </Grid>
                <Grid item xs={12} md={6}>
                  <Summary check={event.check} entity={event.entity} />
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
