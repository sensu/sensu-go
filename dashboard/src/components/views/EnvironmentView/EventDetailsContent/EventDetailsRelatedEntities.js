import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Card, { CardContent } from "material-ui/Card";
import Typography from "material-ui/Typography";
import List from "material-ui/List";
import RelativeDate from "/components/RelativeDate";
import StatusIcon from "/components/CheckStatusIcon";
import ListItem, {
  ListItemTitle,
  ListItemSubtitle,
} from "/components/DetailedListItem";

class EventDetailsRelatedEntities extends React.Component {
  static propTypes = {
    entity: PropTypes.object.isRequired,
  };

  static fragments = {
    entity: gql`
      fragment EventDetailsRelatedEntities_entity on Entity {
        related(limit: 5) {
          id
          name
          lastSeen
          status
        }
      }
    `,
  };

  render() {
    const { entity: { related } } = this.props;

    const listItems = related.map(entity => (
      <ListItem key={entity.id}>
        <ListItemTitle inset>
          <Typography
            component="span"
            style={{ position: "absolute", left: 0 }}
          >
            <StatusIcon statusCode={entity.status} inline mutedOK small />
          </Typography>
          {entity.name}
        </ListItemTitle>
        <ListItemSubtitle inset>
          Last seen <RelativeDate dateTime={entity.lastSeen} />
        </ListItemSubtitle>
      </ListItem>
    ));

    return (
      <Card>
        <CardContent>
          <Typography variant="headline" paragraph>
            Related Entities
          </Typography>
          <List disablePadding>{listItems}</List>
        </CardContent>
      </Card>
    );
  }
}

export default EventDetailsRelatedEntities;
