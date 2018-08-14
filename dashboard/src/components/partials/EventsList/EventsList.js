import React from "react";
import PropTypes from "prop-types";
import { withApollo } from "react-apollo";
import gql from "graphql-tag";

import Paper from "@material-ui/core/Paper";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";

import resolveEvent from "/mutations/resolveEvent";
import deleteEvent from "/mutations/deleteEvent";

import Loader from "/components/util/Loader";
import ListController from "/components/controller/ListController";

import Pagination from "/components/partials/Pagination";
import SilenceEntryDialog from "/components/partials/SilenceEntryDialog";
import ClearSilencesDialog from "/components/partials/ClearSilencedEntriesDialog";

import { TableListEmptyState } from "/components/TableList";

import EventsListHeader from "./EventsListHeader";
import EventsListItem from "./EventsListItem";

class EventsContainer extends React.Component {
  static propTypes = {
    client: PropTypes.object.isRequired,
    environment: PropTypes.shape({
      events: PropTypes.object,
    }),
    onChangeQuery: PropTypes.func.isRequired,
    limit: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    offset: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    loading: PropTypes.bool,
    refetch: PropTypes.func.isRequired,
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

            namespace {
              environment
              organization
            }

            entity {
              name
            }

            check {
              name
              silences {
                ...ClearSilencedEntriesDialog_silence
              }
            }

            ...EventsListHeader_event
            ...EventsListItem_event
          }

          pageInfo {
            ...Pagination_pageInfo
          }
        }

        ...EventsListHeader_environment
      }

      ${ClearSilencesDialog.fragments.silence}
      ${EventsListHeader.fragments.environment}
      ${EventsListHeader.fragments.event}
      ${EventsListItem.fragments.event}
      ${Pagination.fragments.pageInfo}
    `,
  };

  state = {
    silence: null,
    unsilence: null,
  };

  resolveEvents = events => {
    const { client } = this.props;
    events.forEach(event => resolveEvent(client, { id: event.id }));
  };

  deleteEvents = events => {
    const { client } = this.props;
    events.forEach(event => deleteEvent(client, { id: event.id }));
  };

  clearSilences = items => {
    this.setState({
      unsilence: items
        .filter(item => item.check.silences.length > 0)
        .reduce((memo, item) => [...memo, ...item.check.silences], []),
    });
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
        check: "*",
        subscription: `entity:${entity.name}`,
        props: {},
      },
    });
  };

  silenceCheck = check => {
    this.setState({
      silence: {
        check: check.name,
        subscription: "*",
        props: {},
      },
    });
  };

  renderEmptyState = () => {
    const { loading } = this.props;

    return (
      <TableRow>
        <TableCell>
          <TableListEmptyState
            loading={loading}
            primary="No results matched your query."
            secondary="
          Try refining your search query in the search box. The filter buttons
          above are also a helpful way of quickly finding events.
        "
          />
        </TableCell>
      </TableRow>
    );
  };

  renderEvent = ({ key, item: event, selected, setSelected }) => (
    <EventsListItem
      key={key}
      event={event}
      selected={selected}
      onChangeSelected={setSelected}
      onClickClearSilences={() => this.clearSilences([event])}
      onClickSilencePair={() => this.silenceEvents([event])}
      onClickSilenceEntity={() => this.silenceEntity(event.entity)}
      onClickSilenceCheck={() => this.silenceCheck(event.check)}
      onClickResolve={() => this.resolveEvents([event])}
    />
  );

  render() {
    const { silence, unsilence } = this.state;
    const {
      environment,
      loading,
      limit,
      offset,
      onChangeQuery,
      refetch,
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
        {({
          children,
          selectedItems,
          setSelectedItems,
          toggleSelectedItems,
        }) => (
          <Paper>
            <Loader loading={loading}>
              <EventsListHeader
                selectedItems={selectedItems}
                rowCount={children.length || 0}
                onClickSelect={toggleSelectedItems}
                onClickClearSilences={() => this.clearSilences(selectedItems)}
                onClickSilence={() => this.silenceEvents(selectedItems)}
                onClickResolve={() => this.resolveEvents(selectedItems)}
                onClickDelete={() => this.deleteEvents(selectedItems)}
                environment={environment}
                onChangeQuery={onChangeQuery}
              />
              <Table>
                <TableBody>{children}</TableBody>
              </Table>

              <Pagination
                limit={limit}
                offset={offset}
                pageInfo={environment && environment.events.pageInfo}
                onChangeQuery={onChangeQuery}
              />

              <ClearSilencesDialog
                silences={unsilence}
                open={!!unsilence}
                close={() => {
                  this.setState({ unsilence: null });
                  setSelectedItems([]);
                  refetch();
                }}
              />

              {silence && (
                <SilenceEntryDialog
                  values={silence}
                  onClose={() => {
                    this.setState({ silence: null });
                    setSelectedItems([]);
                    refetch();
                  }}
                />
              )}
            </Loader>
          </Paper>
        )}
      </ListController>
    );
  }
}

export default withApollo(EventsContainer);
