import React from "react";
import PropTypes from "prop-types";

import { withApollo } from "react-apollo";
import { compose } from "lodash/fp";
import gql from "graphql-tag";
import { withStyles } from "@material-ui/core/styles";

import resolveEvent from "/mutations/resolveEvent";

import TableList, {
  TableListBody,
  TableListEmptyState,
} from "/components/TableList";

import Loader from "/components/util/Loader";
import ListController from "/components/util/ListController";

import SilenceEntryDialog from "/components/partials/SilenceEntryDialog";

import EventsListHeader from "./EventsListHeader";
import EventsListItem from "./EventsListItem";

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
      fragment EventsList_environment on Environment {
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
            namespace {
              environment
              organization
            }
            ...EventsListItem_event
          }
        }

        ...EventsListHeader_environment
      }

      ${EventsListItem.fragments.event}
      ${EventsListHeader.fragments.environment}
    `,
  };

  state = {
    filters: [],
  };

  resolveEvents = events =>
    events.forEach(event => {
      resolveEvent(this.props.client, { id: event.id });
    });

  silenceEvents = events => {
    const targets = events.map(event => ({
      ns: event.namespace,
      subscription: `entity:${event.entity.name}`,
      check: event.check.name,
    }));

    if (targets.length === 1) {
      this.setState({
        silence: {
          props: {},
          ...targets[0],
        },
      });
    } else if (targets.length) {
      this.setState({
        silence: { props: {}, targets },
      });
    }
  };

  renderTable() {
    const { classes, environment, loading, onQueryChange } = this.props;

    return (
      <ListController
        items={environment ? environment.events.nodes : []}
        getItemKey={item => item.id}
        renderEmptyState={() =>
          !loading && (
            <TableListEmptyState
              primary="No results matched your query."
              secondary="
                Try refining your search query in the search box.
                The filter buttons above are also a helpful way of quickly
                finding events.
              "
            />
          )
        }
        renderItem={({ key, item, selected, toggleSelected }) => (
          <EventsListItem
            key={key}
            event={item}
            selected={selected}
            onClickSelect={toggleSelected}
            onClickSilenceEntity={() => {
              this.setState({
                silence: {
                  subscription: `entity:${item.entity.name}`,
                },
              });
            }}
            onClickSilenceCheck={() => {
              this.setState({ silence: { check: item.check.name } });
            }}
          />
        )}
      >
        {({ children, selectedItems, toggleSelectedItems }) => (
          <TableList className={classes.root}>
            <EventsListHeader
              selectedCount={selectedItems.length}
              onClickSelect={toggleSelectedItems}
              onClickSilence={() => this.silenceEvents(selectedItems)}
              onClickResolve={() => this.resolveEvents(selectedItems)}
              environment={environment}
              onQueryChange={onQueryChange}
            />
            <Loader loading={loading}>
              <TableListBody className={classes.tableBody}>
                {children}
              </TableListBody>
            </Loader>
          </TableList>
        )}
      </ListController>
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
          />
        )}
      </React.Fragment>
    );
  }
}

export default compose(withStyles(styles), withApollo)(EventsContainer);
