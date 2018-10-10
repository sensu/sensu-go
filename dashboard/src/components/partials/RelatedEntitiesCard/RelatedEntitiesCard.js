import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import EntityStatusDescriptor from "/components/partials/EntityStatusDescriptor";
import Typography from "@material-ui/core/Typography";
import List from "@material-ui/core/List";
import StatusIcon from "/components/CheckStatusIcon";
import InlineLink from "/components/InlineLink";
import ListItem, {
  ListItemTitle,
  ListItemSubtitle,
} from "/components/DetailedListItem";

class RelatedEntitiesCard extends React.Component {
  static propTypes = {
    entity: PropTypes.object.isRequired,
  };

  static fragments = {
    entity: gql`
      fragment RelatedEntitiesCard_entity on Entity {
        related(limit: 5) {
          ns: namespace {
            org: organization
            env: environment
          }

          id
          name
          status

          ...EntityStatusDescriptor_entity
        }
      }

      ${EntityStatusDescriptor.fragments.entity}
    `,
  };

  _renderItem = entityProp => {
    const { ns, ...entity } = entityProp;

    return (
      <ListItem key={entity.id}>
        <ListItemTitle inset>
          <Typography
            component="span"
            style={{ position: "absolute", left: 0 }}
          >
            <StatusIcon statusCode={entity.status} inline mutedOK small />
          </Typography>
          <InlineLink to={`/${ns.org}/${ns.env}/entities/${entity.name}`}>
            {entity.name}
          </InlineLink>
        </ListItemTitle>
        <ListItemSubtitle inset>
          <EntityStatusDescriptor entity={entity} />
        </ListItemSubtitle>
      </ListItem>
    );
  };

  _renderItems = () => {
    const {
      entity: { related },
    } = this.props;
    if (related.length === 0) {
      return <Typography>None found.</Typography>;
    }
    return related.map(this._renderItem);
  };

  render() {
    return (
      <Card>
        <CardContent>
          <Typography variant="headline" paragraph>
            Related Entities
          </Typography>
          <List disablePadding>{this._renderItems()}</List>
        </CardContent>
      </Card>
    );
  }
}

export default RelatedEntitiesCard;
