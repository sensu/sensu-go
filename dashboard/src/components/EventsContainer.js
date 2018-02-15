import React from "react";
import PropTypes from "prop-types";

import map from "lodash/map";
import get from "lodash/get";
import { createFragmentContainer, graphql } from "react-relay";
import { withStyles } from "material-ui/styles";

import checkboxIcon from "material-ui/Checkbox";

import EventsListItem from "./EventsListItem";
import EventsContainerMenu from "./EventsContainerMenu";

const styles = theme => ({
  eventsContainer: {
    border: "1px solid",
    borderRadius: 2,
    borderColor: theme.palette.divider,
  },
  tableHeader: {
    padding: "20px 0 16px",
    backgroundColor: theme.palette.primary.light,
    color: theme.palette.primary.contrastText,
    // TODO revist with typography
    display: "flex",
    alignItems: "center",
  },
  tableHeaderButton: {
    marginLeft: 16,
    display: "flex",
  },
  checkbox: { marginTop: -4, width: 24, height: 24 },
});

class EventsContainer extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    viewer: PropTypes.shape({ checkEvents: PropTypes.object }).isRequired,
    Checkbox: PropTypes.func.isRequired,
  };

  static defaultProps = {
    Checkbox: checkboxIcon,
  };

  render() {
    const { classes, viewer, Checkbox } = this.props;
    const events = get(viewer, "events.edges", []);
    const entityNames = map(events, edge => edge.node.entity.name);
    const checkNames = map(events, edge => edge.node.check.config.name);
    const statuses = map(events, edge => edge.node.check.status);

    return (
      <div className={classes.eventsContainer}>
        <div className={classes.tableHeader}>
          <span className={classes.tableHeaderButton}>
            <Checkbox className={classes.checkbox} />
          </span>
          <EventsContainerMenu label="Entity" contents={entityNames} />
          <EventsContainerMenu label="Check" contents={checkNames} />
          <EventsContainerMenu label="Status" contents={statuses} icons />
        </div>
        {map(events, (event, i) => (
          <EventsListItem key={i} event={event.node} />
        ))}
      </div>
    );
  }
}

export default createFragmentContainer(
  withStyles(styles)(EventsContainer),
  graphql`
    fragment EventsContainer_viewer on Viewer {
      events(first: 1000) {
        edges {
          node {
            entity {
              name
            }
            check {
              status
              config {
                name
              }
            }
            ...EventsListItem_event
          }
        }
        pageInfo {
          hasNextPage
        }
      }
    }
  `,
);
