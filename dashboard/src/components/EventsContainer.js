import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";

import { withApollo } from "react-apollo";
import { every, filter, reduce, capitalize } from "lodash";
import { compose } from "lodash/fp";
import { map, join } from "ramda";
import gql from "graphql-tag";
import { withStyles } from "@material-ui/core/styles";

import Typography from "@material-ui/core/Typography";
import MenuItem from "@material-ui/core/MenuItem";
import ListItemText from "@material-ui/core/ListItemText";
import Checkbox from "@material-ui/core/Checkbox";

import EventsListItem from "/components/EventsListItem";
import resolveEvent from "/mutations/resolveEvent";
import TableList, {
  TableListHeader,
  TableListBody,
  TableListSelect,
  TableListEmptyState,
  TableListButton as Button,
} from "/components/TableList";

import Loader from "/components/util/Loader";

import StatusMenu from "/components/partials/StatusMenu";
import SilenceEntryDialog from "/components/partials/SilenceEntryDialog";

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
            ...EventsListItem_event
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
      node => node.id,
      environment ? environment.events.nodes : [],
    );
    // if every state is false or undefined, switch the header
    const newState = !this.eventsSelected();
    this.setState({
      rowState: reduce(
        keys,
        (acc, key) => ({ ...acc, [key]: newState }),
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
      (environment ? environment.events.nodes : []).length ===
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
      resolveEvent(this.props.client, { id: key }).then(() => {
        this.setState(({ rowState }) => ({ ...rowState, [key]: false }));
      });
    });
  };

  silenceSelectedEvents = () => {
    const events = this.props.environment.events.nodes.filter(
      node => this.state.rowState[node.id],
    );

    const targets = events.map(event => ({
      subscription: `entity:${event.entity.name}`,
      check: event.check.name,
    }));

    if (targets.length === 1) {
      this.setState({ silence: targets[0] });
    } else {
      this.setState({ silence: { targets } });
    }
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

  renderTable() {
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
    const eventsSelected = this.selectedEvents();
    const someEventsSelected = eventsSelected.length > 0;
    const hiddenIf = hide => classnames({ [classes.hidden]: hide });

    return (
      <TableList className={classes.root}>
        <TableListHeader sticky active={someEventsSelected}>
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
            <Button
              className={classes.headerButton}
              onClick={this.silenceSelectedEvents}
            >
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
                  secondary="Try refining your search query in the search box. The filter buttons above are also a helpful way of quickly finding events."
                />
              )}
            {/* TODO pass in resolve and silence functions to reuse for single actions
              the silence dialog is the same, just maybe some prefilled options for list */}
            {events.map(event => (
              <EventsListItem
                key={event.id}
                event={event}
                onClickSelect={this.selectCheckbox(event.id)}
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
                checked={Boolean(rowState[event.id])}
              />
            ))}
          </TableListBody>
        </Loader>
      </TableList>
    );
  }

  render() {
    return (
      <React.Fragment>
        {this.renderTable()}
        {this.state.silence && (
          <SilenceEntryDialog
            values={this.state.silence}
            onClose={() => this.setState({ silence: null })}
            onSave={result => {
              this.setState({ silence: null });
              // eslint-disable-next-line no-console
              console.log("persist silence entry", result);
            }}
          />
        )}
      </React.Fragment>
    );
  }
}

export default compose(withStyles(styles), withApollo)(EventsContainer);
