import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import gql from "graphql-tag";

import { Route } from "react-router-dom";
import { withApollo } from "react-apollo";
import { reduce, capitalize } from "lodash";
import { compose } from "lodash/fp";
import { withStyles } from "@material-ui/core/styles";
import MenuItem from "@material-ui/core/MenuItem";
import ListItemText from "@material-ui/core/ListItemText";
import Checkbox from "@material-ui/core/Checkbox";

import EventsListItem from "/components/EventsListItem";
import TableList, {
  TableListHeader,
  TableListBody,
  TableListSelect,
  TableListEmptyState,
  TableListButton as Button,
} from "/components/TableList";
import ButtonSet from "/components/ButtonSet";
import Loader from "/components/util/Loader";
import StatusMenu from "/components/partials/StatusMenu";
import SilenceEntryDialog from "/components/partials/SilenceEntryDialog";
import Typography from "@material-ui/core/Typography";
import ConfirmDelete from "/components/partials/ConfirmDelete";
import resolveEvent from "/mutations/resolveEvent";
import deleteEvent from "/mutations/deleteEvent";

// Event ID includes timestamp and cannot be reliably used to identify an event
// between refreshes, subscriptions and mutations.
const makeKey = ev => `${ev.check.name}:::${ev.entity.name}`;

// Safely retrieve the events from component's props;
const getEvents = props =>
  (props.environment && props.environment.events.nodes) || [];

const styles = theme => ({
  root: {
    marginTop: 16,
    marginBottom: 16,
  },
  filterActions: {
    display: "none",
    [theme.breakpoints.up("sm")]: {
      display: "flex",
    },
  },
  // Remove padding from button container
  checkbox: {
    marginLeft: -11,
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
});

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
        checks(limit: 1000) {
          nodes {
            name
          }
        }

        entities(limit: 1000) {
          nodes {
            name
          }
        }

        events(limit: 100, filter: $filter, orderBy: $order)
          @connection(key: "events", filter: ["filter", "orderBy"]) {
          nodes {
            id
            entity {
              name
            }
            check {
              name
            }

            ...EventsListItem_event
          }
        }
      }

      ${EventsListItem.fragments.event}
    `,
  };

  static getDerivedStateFromProps(props, state) {
    const events = getEvents(props);
    const rowState = events.reduce((acc, ev) => {
      const key = makeKey(ev);
      return { ...acc, [key]: state.rowState[key] };
    }, {});

    return {
      rowState,
    };
  }

  state = {
    rowState: {},
    silence: null,
  };

  // Retrieve those events associated with the current selection set.
  selectedEvents = () => {
    const events = getEvents(this.props);
    return events.filter(ev => this.state.rowState[makeKey(ev)]);
  };

  // click checkbox for all items in list
  selectAll = () => {
    const keys = getEvents(this.props).map(makeKey);
    const events = this.selectedEvents();

    // if every state is false or undefined, switch the header
    const newState = events.length === 0;
    this.setState({
      rowState: reduce(
        keys,
        (acc, key) => ({ ...acc, [key]: newState }),
        this.state.rowState,
      ),
    });
  };

  clearSelectionSet = () => this.setState({ rowState: {} });

  // click single checkbox
  selectCheckbox = key => () => {
    this.setState(state => {
      const curValue = state.rowState[key];
      const rowState = { ...state.rowState, [key]: !curValue };
      return { rowState };
    });
  };

  _handleBulkResolve = () => {
    // Delete selected events
    const events = this.selectedEvents();
    events.forEach(ev => resolveEvent(this.props.client, ev));

    // Clear selection set
    this.clearSelectionSet();
  };

  _handleBulkDelete = () => {
    // Delete selected events
    const events = this.selectedEvents();
    events.forEach(ev => deleteEvent(this.props.client, ev));

    // Clear selection set
    this.clearSelectionSet();
  };

  silenceSelectedEvents = ({ environment, organization }) => {
    const events = this.selectedEvents();
    const targets = events.map(ev => ({
      subscription: `entity:${ev.entity.name}`,
      check: ev.check.name,
    }));

    if (targets.length === 1) {
      this.setState({
        silence: {
          ns: { environment, organization },
          props: {},
          ...targets[0],
        },
      });
    } else {
      this.setState({
        silence: { ns: { environment, organization }, props: {}, targets },
      });
    }
  };

  requeryEntity = newValue => {
    this.props.onQueryChange({ filter: `Entity.ID == '${newValue}'` });
  };

  requeryCheck = newValue => {
    this.props.onQueryChange({ filter: `Check.Name == '${newValue}'` });
  };

  requeryStatus = newValue => {
    if (Array.isArray(newValue)) {
      if (newValue.length === 1) {
        this.props.onQueryChange({ filter: `Check.Status == ${newValue}` });
      } else {
        const val = newValue.join(",");
        this.props.onQueryChange({ filter: `Check.Status IN (${val})` });
      }
    } else {
      this.props.onQueryChange(query => query.delete("filter"));
    }
  };

  requerySort = newValue => {
    this.props.onQueryChange({ order: newValue });
  };

  renderTable(params) {
    const { classes, environment, loading } = this.props;
    const { rowState } = this.state;

    const entityNames = environment
      ? environment.entities.nodes.map(node => node.name)
      : [];

    const checkNames = [
      ...(environment ? environment.checks.nodes.map(node => node.name) : []),
      "keepalive",
    ];

    const events = (environment && environment.events.nodes) || [];
    const selected = this.selectedEvents();
    const selectedLen = selected.length;
    const someSelected = selectedLen > 0;
    const hiddenIf = hide => classnames({ [classes.hidden]: hide });

    return (
      <TableList className={classes.root}>
        <TableListHeader sticky active={someSelected}>
          <Checkbox
            component="button"
            className={classes.checkbox}
            onClick={this.selectAll}
            checked={false}
            indeterminate={someSelected}
          />
          <div className={hiddenIf(!someSelected)}>{selectedLen} Selected</div>
          <div className={classes.grow} />
          <div className={hiddenIf(!someSelected)}>
            <ButtonSet>
              <ConfirmDelete
                identifier={`${selectedLen} ${
                  selectedLen === 1 ? "event" : "events"
                }`}
                onSubmit={this._handleBulkDelete}
              >
                {confirm => (
                  <Button onClick={confirm.open}>
                    <Typography variant="button">Delete</Typography>
                  </Button>
                )}
              </ConfirmDelete>
              <Button onClick={this.silenceSelectedEvents(params)}>
                <Typography variant="button">Silence</Typography>
              </Button>
              <Button onClick={this._handleBulkResolve}>
                <Typography variant="button">Resolve</Typography>
              </Button>
            </ButtonSet>
          </div>
          <div
            className={classnames(
              classes.filterActions,
              hiddenIf(someSelected),
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
            <StatusMenu
              className={classes.headerButton}
              onChange={this.requeryStatus}
            />
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
                  secondary="
                    Try refining your search query in the search box.
                    The filter buttons above are also a helpful way of quickly
                    finding events.
                  "
                />
              )}
            {events.map(event => (
              <EventsListItem
                key={event.id}
                event={event}
                onClickSelect={this.selectCheckbox(makeKey(event))}
                onClickSilenceEntity={() => {
                  this.setState({
                    silence: {
                      subscription: `entity:${event.entity.name}`,
                    },
                  });
                }}
                onClickSilenceCheck={() => {
                  this.setState({ silence: { check: event.check.name } });
                }}
                checked={Boolean(rowState[makeKey(event)])}
              />
            ))}
          </TableListBody>
        </Loader>
      </TableList>
    );
  }

  render() {
    return (
      <Route
        path="/:organization/:environment"
        render={({ match: { params } }) => (
          <React.Fragment>
            {this.renderTable(params)}
            {this.state.silence && (
              <SilenceEntryDialog
                values={this.state.silence}
                onClose={() => this.setState({ silence: null })}
              />
            )}
          </React.Fragment>
        )}
      />
    );
  }
}

export default compose(withStyles(styles), withApollo)(EventsContainer);
