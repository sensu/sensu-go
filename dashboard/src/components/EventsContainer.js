import React from "react";
import PropTypes from "prop-types";

import { withRouter, routerShape } from "found";
import map from "lodash/map";
import get from "lodash/get";
import { createFragmentContainer, graphql } from "react-relay";
import { withStyles } from "material-ui/styles";
import Paper from "material-ui/Paper";

import checkboxIcon from "material-ui/Checkbox";

import EventsListItem from "./EventsListItem";
import EventsContainerMenu from "./EventsContainerMenu";

const styles = theme => ({
  eventsContainer: {
    marginTop: 16,
    marginBottom: 16,
  },
  tableHeader: {
    padding: "20px 0 16px",
    backgroundColor: theme.palette.primary.light,
    color: theme.palette.primary.contrastText,
    display: "flex",
    alignItems: "center",
  },
  tableHeaderButton: {
    marginLeft: 16,
    display: "flex",
  },
  checkbox: {
    marginTop: -4,
    width: 24,
    height: 24,
    color: theme.palette.primary.contrastText,
  },
});

class EventsContainer extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    viewer: PropTypes.shape({
      checks: PropTypes.object,
      entities: PropTypes.object,
    }).isRequired,
    environment: PropTypes.shape({
      events: PropTypes.object,
    }).isRequired,
    router: routerShape.isRequired,
    Checkbox: PropTypes.func.isRequired,
  };

  static defaultProps = {
    Checkbox: checkboxIcon,
  };

  state = {
    filters: [],
  };

  // TODO revist this later
  requeryEntity = newValue => {
    this.props.router.push(
      `${window.location.pathname}?filter=event.Entity.ID=='${newValue}'`,
    );
  };
  requeryCheck = newValue => {
    this.props.router.push(
      `${window.location.pathname}?filter=event.Check.Name=='${newValue}'`,
    );
  };

  requeryStatus = newValue => {
    this.props.router.push(
      `${window.location.pathname}?filter=event.Check.Status==${newValue}`,
    );
  };

  render() {
    const { classes, viewer, environment, Checkbox } = this.props;

    // TODO maybe revisit for pagination issues
    const events = get(environment, "events.edges", []);
    const entities = get(viewer, "entities.edges", []);
    const entityNames = map(entities, edge => edge.node.name);
    const checks = get(viewer, "checks.edges", []);
    const checkNames = [...map(checks, edge => edge.node.name), "keepalive"];
    const statuses = [0, 1, 2, 3];

    return (
      <Paper className={classes.eventsContainer}>
        <div className={classes.tableHeader}>
          <span className={classes.tableHeaderButton}>
            <Checkbox color="secondary" className={classes.checkbox} />
          </span>
          <EventsContainerMenu
            onSelectValue={this.requeryEntity}
            label="Entity"
            contents={entityNames}
          />
          <EventsContainerMenu
            onSelectValue={this.requeryCheck}
            label="Check"
            contents={checkNames}
          />
          <EventsContainerMenu
            onSelectValue={this.requeryStatus}
            label="Status"
            contents={statuses}
            icons
          />
        </div>
        {events.map(event => (
          <EventsListItem key={event.node.id} event={event.node} />
        ))}
      </Paper>
    );
  }
}

export default createFragmentContainer(
  withStyles(styles)(withRouter(EventsContainer)),
  graphql`
    fragment EventsContainer_viewer on Viewer {
      entities(first: 1000) {
        edges {
          node {
            name
          }
        }
      }
      checks(first: 1000) {
        edges {
          node {
            name
          }
        }
      }
    }

    fragment EventsContainer_environment on Environment {
      events(first: 100, filter: $filter) {
        edges {
          node {
            id
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
