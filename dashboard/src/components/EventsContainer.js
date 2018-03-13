import React from "react";
import PropTypes from "prop-types";

import { withRouter, routerShape, matchShape } from "found";
import { every, some, reduce, capitalize } from "lodash";
import { compose } from "lodash/fp";
import { map, join } from "ramda";
import { createFragmentContainer, graphql } from "react-relay";
import { withStyles } from "material-ui/styles";

import Button from "material-ui/Button";
import Typography from "material-ui/Typography";
import { MenuItem } from "material-ui/Menu";
import { ListItemText, ListItemIcon } from "material-ui/List";
import Checkbox from "material-ui/Checkbox";

import EventsListItem from "./EventsListItem";
import EventStatus from "./EventStatus";
import ResolveEventMutation from "../mutations/ResolveEventMutation";
import TableList, {
  TableListHeader,
  TableListSelect,
  TableListEmptyState,
} from "./TableList";

const styles = theme => ({
  root: {
    marginTop: 16,
    marginBottom: 16,
  },
  headerButton: {
    marginLeft: theme.spacing.unit / 2,
    "&:first-child": {
      marginLeft: theme.spacing.unit,
    },
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
    environment: PropTypes.shape({
      events: PropTypes.object,
      checks: PropTypes.object,
      entities: PropTypes.object,
    }).isRequired,
    router: routerShape.isRequired,
    match: matchShape.isRequired,
  };

  state = {
    rowState: {},
    filters: [],
  };

  // click checkbox for all items in list
  selectAll = () => {
    const keys = map(edge => edge.node.id, this.props.environment.events.edges);
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

  changeQuery = (key, val) => {
    const { match, router } = this.props;
    const query = new URLSearchParams(match.location.query);

    query.set(key, val);
    router.push(`${match.location.pathname}?${query.toString()}`);
  };

  requeryEntity = newValue => {
    this.changeQuery("filter", `Entity.ID=='${newValue}'`);
  };

  requeryCheck = newValue => {
    this.changeQuery("filter", `Check.Name=='${newValue}'`);
  };

  requeryStatus = newValue => {
    if (newValue.length === 1) {
      this.changeQuery("filter", `Check.Status==${newValue}`);
      return;
    }
    const val = join(",", newValue);
    this.changeQuery("filter", `Check.Status IN (${val})`);
  };

  requerySort = newValue => {
    this.changeQuery("order", newValue);
  };

  render() {
    const { classes, environment } = this.props;
    const { rowState } = this.state;

    const entityNames = map(edge => edge.node.name, environment.entities.edges);
    const checkNames = [
      ...map(edge => edge.node.name, environment.checks.edges),
      "keepalive",
    ];

    const events = (environment.events && environment.events.edges) || [];
    const someEventsSelected = this.eventsSelected();

    return (
      <TableList className={classes.root}>
        <TableListHeader active={someEventsSelected}>
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
            <TableListSelect
              className={classes.headerButton}
              label="Entity"
              onChange={this.requeryEntity}
            >
              {entityNames.map(name => (
                <MenuItem key={name} value={name}>
                  <ListItemText primary={name} />
                </MenuItem>
              ))}
            </TableListSelect>
            <TableListSelect
              className={classes.headerButton}
              label="Check"
              onChange={this.requeryCheck}
            >
              {checkNames.map(name => (
                <MenuItem key={name} value={name}>
                  <ListItemText primary={name} />
                </MenuItem>
              ))}
            </TableListSelect>
            <TableListSelect
              className={classes.headerButton}
              label="Status"
              onChange={this.requeryStatus}
            >
              <MenuItem key="incident" value={[1, 2, 3]}>
                <ListItemText primary="Incident" style={{ paddingLeft: 40 }} />
              </MenuItem>
              <MenuItem key="warning" value={[1]}>
                <ListItemIcon>
                  <EventStatus status={1} />
                </ListItemIcon>
                <ListItemText primary="Warning" />
              </MenuItem>
              <MenuItem key="critical" value={[2]}>
                <ListItemIcon>
                  <EventStatus status={2} />
                </ListItemIcon>
                <ListItemText primary="Critical" />
              </MenuItem>
              <MenuItem key="unknown" value={[3]}>
                <ListItemIcon>
                  <EventStatus status={3} />
                </ListItemIcon>
                <ListItemText primary="Unknown" />
              </MenuItem>
              <MenuItem key="passing" value={[0]}>
                <ListItemIcon>
                  <EventStatus status={0} />
                </ListItemIcon>
                <ListItemText primary="Passing" />
              </MenuItem>
            </TableListSelect>
            <TableListSelect
              className={classes.headerButton}
              label="Sort"
              onChange={this.requerySort}
            >
              {["SEVERITY", "NEWEST", "OLDEST"].map(name => (
                <MenuItem key={name} value={name}>
                  <ListItemText primary={capitalize(name)} />
                </MenuItem>
              ))}
            </TableListSelect>
          </div>
        </TableListHeader>
        {events.length === 0 && (
          <TableListEmptyState
            primary="No results matched your query."
            secondary="Try refining your search query in the search box. The filter buttons above are also a helpful way of quickly finding events."
          />
        )}
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
      </TableList>
    );
  }
}

const enhance = compose(withStyles(styles), withRouter);
export default createFragmentContainer(
  enhance(EventsContainer),
  graphql`
    fragment EventsContainer_environment on Environment {
      checks(first: 1000) {
        edges {
          node {
            name
          }
        }
      }

      entities(first: 1000) {
        edges {
          node {
            name
          }
        }
      }

      events(first: 100, filter: $filter, orderBy: $order) {
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
