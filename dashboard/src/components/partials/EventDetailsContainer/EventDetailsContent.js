import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import CollapsingMenu from "/components/partials/CollapsingMenu";
import Content from "/components/Content";
import DeleteIcon from "@material-ui/icons/Delete";
import Grid from "@material-ui/core/Grid";
import LiveIcon from "/icons/Live";
import Loader from "/components/util/Loader";
import SmallCheckIcon from "/icons/SmallCheck";
import QueueIcon from "@material-ui/icons/Queue";
import RelatedEntitiesCard from "/components/partials/RelatedEntitiesCard";

import CheckResult from "./EventDetailsCheckResult";
import DeleteAction from "./EventDetailsDeleteAction";
import ResolveAction from "./EventDetailsResolveAction";
import ReRunAction from "./EventDetailsReRunAction";
import Summary from "./EventDetailsSummary";

class EventDetailsContainer extends React.PureComponent {
  static propTypes = {
    event: PropTypes.object,
    loading: PropTypes.bool.isRequired,
    poller: PropTypes.shape({
      running: PropTypes.bool,
      start: PropTypes.func,
      stop: PropTypes.func,
    }).isRequired,
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

  render() {
    const { event, loading, poller } = this.props;

    return (
      <Loader loading={loading} passthrough>
        {event && (
          <React.Fragment>
            <Content bottomMargin>
              <div style={{ flexGrow: 1 }} />
              <CollapsingMenu>
                <ResolveAction event={event}>
                  {({ resolve }) => (
                    <CollapsingMenu.Button
                      title="Resolve"
                      icon={<SmallCheckIcon />}
                      onClick={() => resolve()}
                    />
                  )}
                </ResolveAction>
                <ReRunAction event={event}>
                  {exec => (
                    <CollapsingMenu.Button
                      title="Re-run"
                      icon={<QueueIcon />}
                      onClick={() => exec()}
                    />
                  )}
                </ReRunAction>
                <DeleteAction event={event}>
                  {del => (
                    <CollapsingMenu.Button
                      title="Delete"
                      icon={<DeleteIcon />}
                      onClick={() => del()}
                    />
                  )}
                </DeleteAction>
                <CollapsingMenu.Button
                  pinned
                  title="LIVE"
                  icon={<LiveIcon active={poller.running} />}
                  onClick={() =>
                    poller.running ? poller.stop() : poller.start()
                  }
                />
              </CollapsingMenu>
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
