import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";

import Paper from "@material-ui/core/Paper";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";

import { TableListEmptyState } from "/lib/component/base/TableList";
import Loader from "/lib/component/base/Loader";
import ListController from "/lib/component/controller/ListController";

import ClearSilencesDialog from "/app/component/partial/ClearSilencedEntriesDialog";
import Pagination from "/app/component/partial/Pagination";
import SilenceEntryDialog from "/app/component/partial/SilenceEntryDialog";

import ListHeader from "./SilencesListHeader";
import ListItem from "./SilencesListItem";

class SilencesList extends React.Component {
  static propTypes = {
    editable: PropTypes.bool,
    loading: PropTypes.bool,
    limit: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    namespace: PropTypes.shape({
      silences: PropTypes.shape({
        nodes: PropTypes.array.isRequired,
      }),
    }),
    offset: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    onChangeQuery: PropTypes.func.isRequired,
    order: PropTypes.string.isRequired,
  };

  static defaultProps = {
    editable: false,
    loading: false,
    limit: undefined,
    namespace: null,
    offset: undefined,
  };

  static fragments = {
    namespace: gql`
      fragment SilencesList_namespace on Namespace {
        silences(
          limit: $limit
          offset: $offset
          orderBy: $order
          filter: $filter
        ) @connection(key: "silences", filter: ["filter", "orderBy"]) {
          nodes {
            id
            deleted @client
            ...SilencesListItem_silence
            ...ClearSilencedEntriesDialog_silence
          }

          pageInfo {
            ...Pagination_pageInfo
          }
        }
      }

      ${Pagination.fragments.pageInfo}
      ${ClearSilencesDialog.fragments.silence}
      ${ListItem.fragments.silence}
    `,
  };

  state = {
    silence: null,
    openClearDialog: false,
    selectedItems: [],
  };

  openSilenceDialog = targets => {
    this.setState({ openClearDialog: true });
    this.setState({ selectedItems: targets });
  };

  _handleChangeSort = val => {
    let newVal = val;
    this.props.onChangeQuery(query => {
      // Toggle between ASC & DESC
      const curVal = query.get("order");
      if (curVal === "ID" && newVal === "ID") {
        newVal = "ID_DESC";
      }
      query.set("order", newVal);
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
              Try refining your search query in the search box. The filter
              buttons above are also a helpful way of quickly finding entities.
            "
          />
        </TableCell>
      </TableRow>
    );
  };

  renderItem = ({
    key,
    item,
    hovered,
    setHovered,
    selectedCount,
    selected,
    setSelected,
  }) => (
    <ListItem
      key={key}
      editable={this.props.editable}
      editing={selectedCount > 0}
      silence={item}
      hovered={hovered}
      onHover={setHovered}
      selected={selected}
      onClickClearSilences={() => {
        this.openSilenceDialog([item]);
        setSelected([item]);
      }}
      onClickSelect={setSelected}
    />
  );

  render() {
    const {
      editable,
      loading,
      limit,
      namespace,
      offset,
      order,
      onChangeQuery,
    } = this.props;

    const items = namespace
      ? namespace.silences.nodes.filter(node => !node.deleted)
      : [];

    return (
      <ListController
        items={items}
        getItemKey={check => check.id}
        renderEmptyState={this.renderEmptyState}
        renderItem={this.renderItem}
      >
        {({ children, selectedItems, toggleSelectedItems }) => (
          <React.Fragment>
            <Paper>
              <Loader loading={loading}>
                <ListHeader
                  editable={editable}
                  rowCount={children.length || 0}
                  selectedItems={selectedItems}
                  onChangeQuery={onChangeQuery}
                  onClickSelect={toggleSelectedItems}
                  onClickClearSilences={() =>
                    this.openSilenceDialog(selectedItems)
                  }
                  order={order}
                />

                <Table>
                  <TableBody>{children}</TableBody>
                </Table>

                <Pagination
                  limit={limit}
                  offset={offset}
                  pageInfo={namespace && namespace.silences.pageInfo}
                  onChangeQuery={onChangeQuery}
                />

                {this.state.silence && (
                  <SilenceEntryDialog
                    values={this.state.silence}
                    onClose={() => this.setState({ silence: null })}
                  />
                )}
              </Loader>
            </Paper>
            <ClearSilencesDialog
              silences={this.state.selectedItems}
              open={this.state.openClearDialog}
              close={() => {
                this.setState({ openClearDialog: false });
                toggleSelectedItems();
              }}
              confirmed
              scrollable
            />
          </React.Fragment>
        )}
      </ListController>
    );
  }
}

export default withApollo(SilencesList);
