import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import CollapsingMenu from "/components/CollapsingMenu";
import Content from "/components/Content";
import Grid from "@material-ui/core/Grid";
import DeleteIcon from "@material-ui/icons/Delete";
import LiveIcon from "/icons/Live";

import RelatedEntitiesCard from "/components/partials/RelatedEntitiesCard";
import EntityDetailsEvents from "./EntityDetailsEvents";
import EntityDetailsInformation from "./EntityDetailsInformation";
import DeleteAction from "./EntityDetailsDeleteAction";

class EntityDetailsContainer extends React.PureComponent {
  static propTypes = {
    entity: PropTypes.object.isRequired,
    poller: PropTypes.shape({
      running: PropTypes.bool,
      start: PropTypes.func,
      stop: PropTypes.func,
    }).isRequired,
  };

  static fragments = {
    entity: gql`
      fragment EntityDetailsContainer_entity on Entity {
        id
        events {
          ...EntityDetailsEvents_event
        }

        ...RelatedEntitiesCard_entity
        ...EntityDetailsInformation_entity
        ...EntityDetailsDeleteAction_entity
      }

      ${RelatedEntitiesCard.fragments.entity}
      ${EntityDetailsEvents.fragments.event}
      ${EntityDetailsInformation.fragments.entity}
      ${DeleteAction.fragments.entity}
    `,
  };

  render() {
    const { entity, poller } = this.props;

    return (
      <React.Fragment>
        <Content bottomMargin>
          <div style={{ flexGrow: 1 }} />
          <CollapsingMenu>
            <DeleteAction entity={entity}>
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
              onClick={() => (poller.running ? poller.stop() : poller.start())}
            />
          </CollapsingMenu>
        </Content>
        <Content>
          <Grid container spacing={16}>
            <Grid item xs={12}>
              <EntityDetailsInformation entity={entity} />
            </Grid>
            <Grid item xs={12} md={6}>
              <RelatedEntitiesCard entity={entity} />
            </Grid>
            <Grid item xs={12} md={6}>
              <EntityDetailsEvents events={entity.events} />
            </Grid>
          </Grid>
        </Content>
      </React.Fragment>
    );
  }
}

export default EntityDetailsContainer;
