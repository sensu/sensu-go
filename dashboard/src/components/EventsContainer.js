import React from "react";
import PropTypes from "prop-types";

import { withRouter, routerShape } from "found";
import { map, get, every, some, reduce } from "lodash";
import { createFragmentContainer, graphql } from "react-relay";
import { withStyles } from "material-ui/styles";
import Paper from "material-ui/Paper";
import Button from "material-ui/Button";
import Typography from "material-ui/Typography";

import Checkbox from "material-ui/Checkbox";

import EventsListItem from "./EventsListItem";
import EventsContainerMenu from "./EventsContainerMenu";
import ResolveEventMutation from "../mutations/ResolveEventMutation";

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
  altMenuButton: {
    color: theme.palette.primary.contrastText,
    padding: "0 0 1px",
    minHeight: 20,
    "&:hover": { backgroundColor: "inherit" },
  },
  hidden: {
    display: "none",
  },
});

class EventsContainer extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    relay: PropTypes.shape({ environment: PropTypes.object }).isRequired,
    viewer: PropTypes.shape({
      checks: PropTypes.object,
      entities: PropTypes.object,
    }).isRequired,
    environment: PropTypes.shape({
      events: PropTypes.object,
    }).isRequired,
    router: routerShape.isRequired,
  };

  state = {
    rowState: {},
    filters: [],
  };

  // click checkbox for all items in list
  selectAll = () => {
    const keys = map(this.props.environment.events.edges, edge => edge.node.id);
    // if every state is false or undefined, switch the header
    const newState = !this.eventsSelected();
    this.setState({
      rowState: reduce(
        keys,
        (acc, key) => Object.assign(acc, { [key]: newState }),
        this.state.rowState,
      ),
    });
  };

  // click single checkbox
  selectCheckbox = id => () => {
    this.state.rowState[id] = !this.state.rowState[id];
    this.setState({ rowState: this.state.rowState });
  };

  eventsSelected = () => some(this.state.rowState, Boolean);

  allEventsSelected = () => {
    const { rowState } = this.state;
    return (
      this.props.environment.events.edges.length ===
        Object.keys(rowState).length && every(rowState, Boolean)
    );
  };

  resolve = () => {
    const selectedKeys = reduce(
      this.state.rowState,
      (selected, val, key) => (val ? [...selected, key] : selected),
      [],
    );

    selectedKeys.forEach(key => {
      ResolveEventMutation.commit(this.props.relay.environment, key, {
        onCompleted: () => {
          this.setState(({ rowState }) =>
            Object.assign(rowState, { [key]: false }),
          );
        },
      });
    });
  };

  silence = () => {
    // silence each item that is true in rowState
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
    const { classes, viewer, environment } = this.props;
    const { rowState } = this.state;

    // TODO maybe revisit for pagination issues
    const events = get(environment, "events.edges", []);
    const entities = get(viewer, "entities.edges", []);
    const entityNames = map(entities, edge => edge.node.name);
    const checks = get(viewer, "checks.edges", []);
    const checkNames = [...map(checks, edge => edge.node.name), "keepalive"];
    const statuses = [0, 1, 2, 3];
    const someEventsSelected = this.eventsSelected();

    return (
      <Paper className={classes.eventsContainer}>
        <div className={classes.tableHeader}>
          <span className={classes.tableHeaderButton}>
            <Checkbox
              color="secondary"
              className={classes.checkbox}
              onClick={this.selectAll}
              checked={false}
              indeterminate={someEventsSelected}
            />
          </span>
          <div style={someEventsSelected ? {} : { display: "none" }}>
            <Button className={classes.altMenuButton} onClick={this.silence}>
              <Typography type="button">Silence</Typography>
            </Button>
            <Button className={classes.altMenuButton} onClick={this.resolve}>
              <Typography type="button">Resolve</Typography>
            </Button>
          </div>
          <div style={someEventsSelected ? { display: "none" } : {}}>
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
        </div>
        {/* TODO pass in resolve and silence functions to reuse for single actions
            the silence dialog is the same, just maybe some prefilled options for list */}
        {events.map(event => (
          <EventsListItem
            key={event.node.id}
            event={event.node}
            onChange={this.selectCheckbox(event.node.id)}
            checked={Boolean(rowState[event.node.id])}
          />
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
