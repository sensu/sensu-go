import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";

import Paper from "@material-ui/core/Paper";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import { TableListEmptyState } from "/components/TableList";
import Loader from "/components/util/Loader";
import ListController from "/components/controller/ListController";
import Pagination from "/components/partials/Pagination";
import SilenceEntryDialog from "/components/partials/SilenceEntryDialog";
import deleteSilence from "/mutations/deleteSilence";

import ListHeader from "./SilencesListHeader";
import ListItem from "./SilencesListItem";

class SilencesList extends React.Component {
  static propTypes = {
    client: PropTypes.object.isRequired,
    namespace: PropTypes.shape({
      silences: PropTypes.shape({
        nodes: PropTypes.array.isRequired,
      }),
    }),
    loading: PropTypes.bool,
    onChangeQuery: PropTypes.func.isRequired,
    limit: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    offset: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    order: PropTypes.string.isRequired,
  };

  static defaultProps = {
    namespace: null,
    loading: false,
    limit: undefined,
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
          }

          pageInfo {
            ...Pagination_pageInfo
          }
        }
      }

      ${Pagination.fragments.pageInfo}
      ${ListItem.fragments.silence}
    `,
  };

  state = {
    silence: null,
  };

  // TODO, create something that exists similarily to
  // the silencing on ChecksList. This method can set a dialog
  // to appear, right now we just call the dialog in the
  // children themselves.
  deleteItem = item => {
    this.deleteItems([item]);
  };

  deleteItems = items => {
    const { client } = this.props;
    items.forEach(item => deleteSilence(client, item));
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

  renderItem = ({ key, item, selected, setSelected }) => (
    <ListItem
      key={key}
      silence={item}
      selected={selected}
      onClickSelect={setSelected}
    />
  );

  render() {
    const {
      namespace,
      loading,
      limit,
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
          <Paper>
            <Loader loading={loading}>
              <ListHeader
                rowCount={children.length || 0}
                selectedItems={selectedItems}
                onChangeQuery={onChangeQuery}
                onClickSelect={toggleSelectedItems}
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
        )}
      </ListController>
    );
  }
}

export default withApollo(SilencesList);
