import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";

import Paper from "@material-ui/core/Paper";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableRow from "@material-ui/core/TableRow";
import TableCell from "@material-ui/core/TableCell";

import { TableListEmptyState } from "/components/TableList";

import deleteEntity from "/mutations/deleteEntity";

import Loader from "/components/util/Loader";
import ListController from "/components/controller/ListController";
import Pagination from "/components/partials/Pagination";
import SilenceEntryDialog from "/components/partials/SilenceEntryDialog";
import ClearSilencesDialog from "/components/partials/ClearSilencedEntriesDialog";

import EntitiesListHeader from "./EntitiesListHeader";
import EntitiesListItem from "./EntitiesListItem";

class EntitiesList extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    editable: PropTypes.bool,
    loading: PropTypes.bool,
    namespace: PropTypes.object,
    order: PropTypes.string.isRequired,
    onChangeQuery: PropTypes.func.isRequired,
    limit: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    offset: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    refetch: PropTypes.func,
  };

  static defaultProps = {
    editable: false,
    loading: false,
    limit: undefined,
    namespace: null,
    offset: undefined,
    refetch: () => null,
  };

  static fragments = {
    namespace: gql`
      fragment EntitiesList_namespace on Namespace {
        entities(
          limit: $limit
          offset: $offset
          filter: $filter
          orderBy: $order
        ) @connection(key: "entities", filter: ["filter", "orderBy"]) {
          nodes {
            id
            namespace
            deleted @client

            silences {
              ...ClearSilencedEntriesDialog_silence
            }

            ...EntitiesListItem_entity
          }

          pageInfo {
            ...Pagination_pageInfo
          }
        }
        ...EntitiesListHeader_namespace
      }

      ${ClearSilencesDialog.fragments.silence}
      ${EntitiesListHeader.fragments.namespace}
      ${EntitiesListItem.fragments.entity}
      ${Pagination.fragments.pageInfo}
    `,
  };

  state = {
    silence: null,
    unsilence: null,
  };

  deleteEntities = entities => {
    const { client } = this.props;
    entities.forEach(entity => deleteEntity(client, { id: entity.id }));
  };

  silenceItems = entities => {
    const targets = entities
      .filter(entity => entity.silences.length === 0)
      .map(entity => ({
        namespace: entity.namespace,
        check: "*",
        subscription: `entity:${entity.name}`,
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

  clearSilences = items => {
    this.setState({
      unsilence: items.reduce((memo, item) => [...memo, ...item.silences], []),
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
          above are also a helpful way of quickly finding entities.
        "
          />
        </TableCell>
      </TableRow>
    );
  };

  renderEntity = ({
    key,
    item: entity,
    hovered,
    setHovered,
    selectedCount,
    selected,
    setSelected,
  }) => (
    <EntitiesListItem
      key={key}
      editable={this.props.editable}
      editing={selectedCount > 0}
      entity={entity}
      hovered={hovered}
      onHover={this.props.editable ? setHovered : () => null}
      selected={selected}
      onChangeSelected={setSelected}
      onClickDelete={() => this.deleteEntities([entity])}
      onClickSilence={() => this.silenceItems([entity])}
      onClickClearSilence={() => this.clearSilences([entity])}
    />
  );

  render() {
    const { silence, unsilence } = this.state;
    const {
      editable,
      loading,
      onChangeQuery,
      limit,
      namespace,
      offset,
      refetch,
      order,
    } = this.props;

    const items = namespace
      ? namespace.entities.nodes.filter(entity => !entity.deleted)
      : [];

    return (
      <ListController
        items={items}
        getItemKey={entity => entity.id}
        renderEmptyState={this.renderEmptyState}
        renderItem={this.renderEntity}
      >
        {({
          children,
          selectedItems,
          setSelectedItems,
          toggleSelectedItems,
        }) => (
          <Paper>
            <Loader loading={loading}>
              <EntitiesListHeader
                editable={editable}
                selectedItems={selectedItems}
                rowCount={children.length || 0}
                onClickSelect={toggleSelectedItems}
                onClickDelete={() => this.deleteEntities(selectedItems)}
                onClickSilence={() => this.silenceItems(selectedItems)}
                onClickClearSilences={() => this.clearSilences(selectedItems)}
                onChangeQuery={onChangeQuery}
                namespace={namespace}
                order={order}
              />

              <Table>
                <TableBody>{children}</TableBody>
              </Table>

              <Pagination
                limit={limit}
                offset={offset}
                pageInfo={namespace && namespace.entities.pageInfo}
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

export default withApollo(EntitiesList);
