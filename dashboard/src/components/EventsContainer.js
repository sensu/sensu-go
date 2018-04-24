import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";

import { withApollo } from "react-apollo";
import { every, filter, reduce, capitalize } from "lodash";
import { compose } from "lodash/fp";
import { map, join } from "ramda";
import gql from "graphql-tag";
import { withStyles } from "material-ui/styles";

import Typography from "material-ui/Typography";
import { MenuItem } from "material-ui/Menu";
import { ListItemText, ListItemIcon } from "material-ui/List";
import Checkbox from "material-ui/Checkbox";

import EventsListItem from "./EventsListItem";
import EventStatus from "./EventStatus";
import ResolveEventMutation from "../mutations/ResolveEventMutation";
import TableList, {
  TableListHeader,
  TableListBody,
  TableListSelect,
  TableListEmptyState,
  TableListButton as Button,
} from "./TableList";

import Loader from "./Loader";

const styles = theme => {
  const toolbar = theme.mixins.toolbar;
  const xsBrk = `${theme.breakpoints.up("xs")} and (orientation: landscape)`;
  const smBrk = theme.breakpoints.up("sm");
  const calcTopWithFallback = size => ({
    top: `calc(${size}px + env(safe-area-inset-top))`,
    fallbacks: [{ top: size }],
  });

  return {
    root: {
      marginTop: 16,
      marginBottom: 16,
    },
    header: {
      position: "sticky",
      ...calcTopWithFallback(toolbar.minHeight),
      [xsBrk]: {
        ...calcTopWithFallback(toolbar[xsBrk].minHeight),
      },
      [smBrk]: {
        ...calcTopWithFallback(toolbar[smBrk].minHeight),
      },
      color: theme.palette.primary.contrastText,
    },
    headerButton: {
      marginLeft: theme.spacing.unit / 2,
      "&:first-child": {
        marginLeft: theme.spacing.unit,
      },
    },
    filterActions: {
      display: "none",
      [theme.breakpoints.up("sm")]: {
        display: "flex",
      },
    },
    // Remove padding from button container
    checkbox: {
      marginLeft: -12,
      color: theme.palette.primary.contrastText,
    },
    hidden: {
      display: "none",
    },
    grow: {
      flex: "1 1 auto",
    },

    tableBody: {
      minHeight: 200,
    },
  };
};

class EventsContainer extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    client: PropTypes.object.isRequired,
    environment: PropTypes.shape({
      events: PropTypes.object,
      checks: PropTypes.object,
      entities: PropTypes.object,
    }),
    onQueryChange: PropTypes.func.isRequired,
    loading: PropTypes.bool,
  };

  static defaultProps = {
    loading: false,
    environment: null,
  };

  static fragments = {
    environment: gql`
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

      ${EventsListItem.fragments.event}
    `,
  };

  state = {
    rowState: {},
    filters: [],
  };

  // click checkbox for all items in list
  selectAll = () => {
    const { environment } = this.props;

    const keys = map(
      edge => edge.node.id,
      environment ? environment.events.edges : [],
    );
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

  selectedEvents = () => filter(this.state.rowState, Boolean);
  eventsSelected = () => this.selectedEvents().length > 0;

  allEventsSelected = () => {
    const { environment } = this.props;

    const { rowState } = this.state;
    return (
      (environment ? environment.events.edges : []).length ===
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
      ResolveEventMutation.commit(this.props.client, key).then(
        () => {
          this.setState(({ rowState }) =>
            Object.assign(rowState, { [key]: false }),
          );
        },
        error => {
          // eslint-disable-next-line no-console
          console.error(error);
        },
      );
    });
  };

  silence = () => {
    // silence each item that is true in rowState
  };

  requeryEntity = newValue => {
    this.props.onQueryChange("filter", `Entity.ID=='${newValue}'`);
  };

  requeryCheck = newValue => {
    this.props.onQueryChange("filter", `Check.Name=='${newValue}'`);
  };

  requeryStatus = newValue => {
    if (Array.isArray(newValue)) {
      if (newValue.length === 1) {
        this.props.onQueryChange("filter", `Check.Status==${newValue}`);
      } else {
        const val = join(",", newValue);
        this.props.onQueryChange("filter", `Check.Status IN (${val})`);
      }
    } else {
      this.props.onQueryChange("filter", newValue);
    }
  };

  requerySort = newValue => {
    this.props.onQueryChange("order", newValue);
  };

  render() {
    const { classes, environment, loading } = this.props;
    const { rowState } = this.state;

    const entityNames = environment
      ? map(edge => edge.node.name, environment.entities.edges)
      : [];

    const checkNames = [
      ...(environment
        ? map(edge => edge.node.name, environment.checks.edges)
        : []),
      "keepalive",
    ];

    const events =
      (environment && environment.events && environment.events.edges) || [];
    const eventsSelected = this.selectedEvents();
    const someEventsSelected = eventsSelected.length > 0;
    const hiddenIf = hide => classnames({ [classes.hidden]: hide });

    return (
      <TableList className={classes.root}>
        <TableListHeader className={classes.header} active={someEventsSelected}>
          <Checkbox
            component="button"
            className={classes.checkbox}
            onClick={this.selectAll}
            checked={false}
            indeterminate={someEventsSelected}
          />
          <div className={hiddenIf(!someEventsSelected)}>
            {eventsSelected.length} Selected
          </div>
          <div className={classes.grow} />
          <div className={hiddenIf(!someEventsSelected)}>
            <Button className={classes.headerButton} onClick={this.silence}>
              <Typography variant="button">Silence</Typography>
            </Button>
            <Button className={classes.headerButton} onClick={this.resolve}>
              <Typography variant="button">Resolve</Typography>
            </Button>
          </div>
          <div
            className={classnames(
              classes.filterActions,
              hiddenIf(someEventsSelected),
            )}
          >
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
              <MenuItem key="incident" value={"HasCheck && IsIncident"}>
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
        <Loader loading={loading}>
          <TableListBody className={classes.tableBody}>
            {!loading &&
              events.length === 0 && (
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
          </TableListBody>
        </Loader>
      </TableList>
    );
  }
}

export default compose(withStyles(styles), withApollo)(EventsContainer);
