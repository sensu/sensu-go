import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Content from "/lib/component/base/Content";
import Grid from "@material-ui/core/Grid";
import RelatedEntitiesCard from "/app/component/partial/RelatedEntitiesCard";

import EntityDetailsEvents from "./EntityDetailsEvents";
import EntityDetailsInformation from "./EntityDetailsInformation";
import EntityDetailsToolbar from "./EntityDetailsToolbar";

class EntityDetailsContainer extends React.PureComponent {
  static propTypes = {
    entity: PropTypes.object.isRequired,
    refetch: PropTypes.func.isRequired,
  };

  static fragments = {
    entity: gql`
      fragment EntityDetailsContainer_entity on Entity {
        events(orderBy: LASTOK) {
          ...EntityDetailsEvents_event
        }

        ...RelatedEntitiesCard_entity
        ...EntityDetailsInformation_entity
        ...EntityDetailsToolbar_entity
      }

      ${RelatedEntitiesCard.fragments.entity}
      ${EntityDetailsEvents.fragments.event}
      ${EntityDetailsInformation.fragments.entity}
      ${EntityDetailsToolbar.fragments.entity}
    `,
  };

  render() {
    const { entity, refetch } = this.props;

    return (
      <React.Fragment>
        <Content marginBottom>
          <EntityDetailsToolbar entity={entity} refetch={refetch} />
        </Content>

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
      </React.Fragment>
    );
  }
}

export default EntityDetailsContainer;
