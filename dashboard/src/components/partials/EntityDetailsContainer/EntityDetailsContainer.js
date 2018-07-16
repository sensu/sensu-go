import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Grid from "@material-ui/core/Grid";
import Content from "/components/Content";
import ButtonSet from "/components/ButtonSet";
import RelatedEntitiesCard from "/components/partials/RelatedEntitiesCard";
import EntityDetailsEvents from "./EntityDetailsEvents";
import EntityDetailsInformation from "./EntityDetailsInformation";
import DeleteAction from "./EntityDetailsDeleteAction";

class EntityDetailsContainer extends React.PureComponent {
  static propTypes = {
    entity: PropTypes.object.isRequired,
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
    const { entity } = this.props;
    return (
      <React.Fragment>
        <Content bottomMargin>
          <div style={{ flexGrow: 1 }} />
          <ButtonSet>
            <DeleteAction entity={entity} />
          </ButtonSet>
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
