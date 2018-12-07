import React from "react";
import PropTypes from "prop-types";
import { withApollo } from "react-apollo";
import gql from "graphql-tag";

import Paper from "@material-ui/core/Paper";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";

import deleteEvent from "/mutations/deleteEvent";
import executeCheck from "/mutations/executeCheck";
import resolveEvent from "/mutations/resolveEvent";

import Loader from "/components/util/Loader";
import ListController from "/components/controller/ListController";

import Pagination from "/components/partials/Pagination";
import SilenceEntryDialog from "/components/partials/SilenceEntryDialog";
import ClearSilencesDialog from "/components/partials/ClearSilencedEntriesDialog";
import ExecuteCheckStatusToast from "/components/relocation/ExecuteCheckStatusToast";

import { TableListEmptyState } from "/components/TableList";

import EventsListHeader from "./EventsListHeader";
import EventsListItem from "./EventsListItem";

class EventsContainer extends React.Component {
  static propTypes = {
    addToast: PropTypes.func.isRequired,
    client: PropTypes.object.isRequired,
    editable: PropTypes.bool,
    namespace: PropTypes.shape({
      checks: PropTypes.object,
      entities: PropTypes.object,
    }),
    onChangeQuery: PropTypes.func.isRequired,
    limit: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    offset: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    loading: PropTypes.bool,
    refetch: PropTypes.func.isRequired,
  };

  static defaultProps = {
    loading: false,
    editable: true,
    namespace: null,
    limit: undefined,
    offset: undefined,
  };

  static fragments = {
    namespace: gql`
      fragment EventsList_namespace on Namespace {
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
            namespace
            deleted @client

            entity {
              name
            }

            check {
              nodeId
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

        ...EventsListHeader_namespace
      }

      ${ClearSilencesDialog.fragments.silence}
      ${EventsListHeader.fragments.namespace}
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

  executeCheck = events => {
    const { client } = this.props;

    events.forEach(({ check, entity }) => {
      const promise = executeCheck(client, {
        id: check.nodeId,
        subscriptions: [`entity:${entity.name}`],
      });

      this.props.addToast(({ remove }) => (
        <ExecuteCheckStatusToast
          onClose={remove}
          mutation={promise}
          checkName={check.name}
          namespace={check.namespace}
        />
      ));
    });
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
      namespace: event.namespace,
      subscription: `entity:${event.entity.name}`,
      check: event.check.name,
    }));

    if (targets.length === 1) {
      this.setState({
        silence: {
          ...targets[0],
          props: {
            begin: null,
          },
        },
      });
    } else if (targets.length) {
      this.setState({
        silence: {
          props: {
            begin: null,
          },
          targets,
        },
      });
    }
  };

  silenceEntity = event => {
    this.setState({
      silence: {
        namespace: event.namespace,
        check: "*",
        subscription: `entity:${event.entity.name}`,
        props: {
          begin: null,
        },
      },
    });
  };

  silenceCheck = event => {
    this.setState({
      silence: {
        namespace: event.namespace,
        check: event.check.name,
        subscription: "*",
        props: {
          begin: null,
        },
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

  renderEvent = ({
    key,
    item,
    selectedCount,
    hovered,
    setHovered,
    selected,
    setSelected,
  }) => (
    <EventsListItem
      key={key}
      event={item}
      editable={this.props.editable}
      editing={selectedCount > 0}
      selected={selected}
      onChangeSelected={setSelected}
      hovered={hovered}
      onHover={this.props.editable ? setHovered : () => null}
      onClickClearSilences={() => this.clearSilences([item])}
      onClickSilencePair={() => this.silenceEvents([item])}
      onClickSilenceEntity={() => this.silenceEntity(item)}
      onClickSilenceCheck={() => this.silenceCheck(item)}
      onClickResolve={() => this.resolveEvents([item])}
      onClickRerun={() => this.executeCheck([item])}
      onClickDelete={() => this.deleteEvents([item])}
    />
  );

  render() {
    const { silence, unsilence } = this.state;
    const {
      editable,
      loading,
      limit,
      namespace,
      offset,
      onChangeQuery,
      refetch,
    } = this.props;

    const items = namespace
      ? namespace.events.nodes.filter(event => !event.deleted)
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
                editable={editable}
                namespace={namespace}
                onClickSelect={toggleSelectedItems}
                onClickClearSilences={() => this.clearSilences(selectedItems)}
                onClickSilence={() => this.silenceEvents(selectedItems)}
                onClickResolve={() => {
                  this.resolveEvents(selectedItems);
                  setSelectedItems([]);
                }}
                onClickRerun={() => {
                  this.executeCheck(selectedItems);
                  setSelectedItems([]);
                }}
                onClickDelete={() => {
                  this.deleteEvents(selectedItems);
                  setSelectedItems([]);
                }}
                onChangeQuery={onChangeQuery}
                rowCount={children.length || 0}
                selectedItems={selectedItems}
              />

              <Table>
                <TableBody>{children}</TableBody>
              </Table>

              <Pagination
                limit={limit}
                offset={offset}
                pageInfo={namespace && namespace.events.pageInfo}
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
