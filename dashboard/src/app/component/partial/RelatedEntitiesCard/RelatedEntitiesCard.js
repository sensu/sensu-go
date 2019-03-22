import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Typography from "@material-ui/core/Typography";
import List from "@material-ui/core/List";

import StatusIcon from "/lib/component/base/CheckStatusIcon";
import InlineLink from "/lib/component/base/InlineLink";
import ListItem, {
  ListItemTitle,
  ListItemSubtitle,
} from "/lib/component/base/DetailedListItem";

import EntityStatusDescriptor from "/app/component/partial/EntityStatusDescriptor";

class RelatedEntitiesCard extends React.Component {
  static propTypes = {
    entity: PropTypes.object.isRequired,
  };

  static fragments = {
    entity: gql`
      fragment RelatedEntitiesCard_entity on Entity {
        related(limit: 5) {
          id
          namespace
          name
          status

          ...EntityStatusDescriptor_entity
        }
      }

      ${EntityStatusDescriptor.fragments.entity}
    `,
  };

  renderItem = entityProp => {
    const { namespace, ...entity } = entityProp;

    return (
      <ListItem key={entity.id}>
        <ListItemTitle inset>
          <Typography
            component="span"
            style={{ position: "absolute", left: 0 }}
          >
            <StatusIcon statusCode={entity.status} inline mutedOK small />
          </Typography>
          <InlineLink to={`/${namespace}/entities/${entity.name}`}>
            {entity.name}
          </InlineLink>
        </ListItemTitle>
        <ListItemSubtitle inset>
          <EntityStatusDescriptor entity={entity} />
        </ListItemSubtitle>
      </ListItem>
    );
  };

  renderItems = () => {
    const { entity } = this.props;

    if (entity.related.length === 0) {
      return <Typography>None found.</Typography>;
    }
    return entity.related.map(this.renderItem);
  };

  render() {
    return (
      <Card>
        <CardContent>
          <Typography variant="h5" paragraph>
            Related Entities
          </Typography>
          <List disablePadding>{this.renderItems()}</List>
        </CardContent>
      </Card>
    );
  }
}

export default RelatedEntitiesCard;
