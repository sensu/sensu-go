import React from "react";
import PropTypes from "prop-types";

import { withApollo } from "react-apollo";
import { compose } from "recompose";
import gql from "graphql-tag";
import { withStyles } from "@material-ui/core/styles";

import resolveEvent from "/mutations/resolveEvent";
import deleteEvent from "/mutations/deleteEvent";

import TableList, {
  TableListBody,
  TableListEmptyState,
} from "/components/TableList";

import Loader from "/components/util/Loader";
import ListController from "/components/util/ListController";

import Pagination from "/components/partials/Pagination";
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
});

class EventsContainer extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    client: PropTypes.object.isRequired,
    environment: PropTypes.shape({
      events: PropTypes.object,
    }),
    onChangeQuery: PropTypes.func.isRequired,
    limit: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    offset: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    loading: PropTypes.bool,
  };

  static defaultProps = {
    loading: false,
    environment: null,
    limit: undefined,
    offset: undefined,
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

        events(
          limit: $limit
          offset: $offset
          filter: $filter
          orderBy: $order
        ) @connection(key: "events", filter: ["filter", "orderBy"]) {
          nodes {
            id
            deleted @client
            entity {
              name
            }
            check {
              name
            }

            namespace {
              environment
              organization
            }
            ...EventsListItem_event
          }

          pageInfo {
            ...Pagination_pageInfo
          }
        }

        ...EventsListHeader_environment
      }

      ${EventsListItem.fragments.event}
      ${EventsListHeader.fragments.environment}
      ${Pagination.fragments.pageInfo}
    `,
  };

  state = {
    silence: null,
  };

  resolveEvents = events => {
    const { client } = this.props;
    events.forEach(event => resolveEvent(client, { id: event.id }));
  };

  deleteEvents = events => {
    const { client } = this.props;
    events.forEach(event => deleteEvent(client, { id: event.id }));
  };

  silenceEvents = events => {
    const targets = events.map(event => ({
      ns: {
        environment: event.namespace.environment,
        organization: event.namespace.organization,
      },
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

  silenceEntity = entity => {
    this.setState({
      silence: {
        subscription: `entity:${entity.name}`,
      },
    });
  };

  silenceCheck = check => {
    this.setState({ silence: { check: check.name } });
  };

  renderEmptyState = () => {
    const { loading } = this.props;

    return (
      <TableListEmptyState
        loading={loading}
        primary="No results matched your query."
        secondary="
          Try refining your search query in the search box. The filter buttons
          above are also a helpful way of quickly finding events.
        "
      />
    );
  };

  renderEvent = ({ key, item: event, selected, toggleSelected }) => (
    <EventsListItem
      key={key}
      event={event}
      selected={selected}
      onClickSelect={toggleSelected}
      onClickSilenceEntity={() => this.silenceEntity(event.entity)}
      onClickSilenceCheck={() => this.silenceCheck(event.check)}
    />
  );

  render() {
    const {
      classes,
      environment,
      loading,
      limit,
      offset,
      onChangeQuery,
    } = this.props;

    const items = environment
      ? environment.events.nodes.filter(event => !event.deleted)
      : [];

    return (
      <ListController
        items={items}
        // Event ID includes timestamp and cannot be reliably used to identify
        // an event between refreshes, subscriptions and mutations.
        getItemKey={event => `${event.check.name}:::${event.entity.name}`}
        renderEmptyState={this.renderEmptyState}
        renderItem={this.renderEvent}
      >
        {({ children, selectedItems, toggleSelectedItems }) => (
          <React.Fragment>
            <TableList className={classes.root}>
              <EventsListHeader
                selectedCount={selectedItems.length}
                onClickSelect={toggleSelectedItems}
                onClickSilence={() => this.silenceEvents(selectedItems)}
                onClickResolve={() => this.resolveEvents(selectedItems)}
                onClickDelete={() => this.deleteEvents(selectedItems)}
                environment={environment}
                onChangeQuery={onChangeQuery}
              />
              <Loader loading={loading}>
                <TableListBody>{children}</TableListBody>
              </Loader>
              <Pagination
                limit={limit}
                offset={offset}
                pageInfo={environment && environment.events.pageInfo}
                onChangeQuery={onChangeQuery}
              />
            </TableList>

            {this.state.silence && (
              <SilenceEntryDialog
                values={this.state.silence}
                onClose={() => this.setState({ silence: null })}
              />
            )}
          </React.Fragment>
        )}
      </ListController>
    );
  }
}

export default compose(withStyles(styles), withApollo)(EventsContainer);
