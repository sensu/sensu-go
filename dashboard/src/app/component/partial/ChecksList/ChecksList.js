import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";

import deleteCheck from "/lib/mutation/deleteCheck";
import executeCheck from "/lib/mutation/executeCheck";
import setCheckPublish from "/lib/mutation/setCheckPublish";
import ClearSilencesDialog from "/app/component/partial/ClearSilencedEntriesDialog";
import ListController from "/lib/component/controller/ListController";
import Loader from "/lib/component/base/Loader";
import Paper from "@material-ui/core/Paper";
import Pagination from "/app/component/partial/Pagination";
import SilenceEntryDialog from "/app/component/partial/SilenceEntryDialog";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import { TableListEmptyState } from "/lib/component/base/TableList";
import ExecuteCheckStatusToast from "/lib/component/relocation/ExecuteCheckStatusToast";
import PublishCheckStatusToast from "/lib/component/relocation/PublishCheckStatusToast";

import ChecksListHeader from "./ChecksListHeader";
import ChecksListItem from "./ChecksListItem";

class ChecksList extends React.Component {
  static propTypes = {
    client: PropTypes.object.isRequired,
    editable: PropTypes.bool,
    namespace: PropTypes.shape({
      checks: PropTypes.shape({
        nodes: PropTypes.array.isRequired,
      }),
    }),
    loading: PropTypes.bool,
    onChangeQuery: PropTypes.func.isRequired,
    limit: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    offset: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    order: PropTypes.string.isRequired,
    refetch: PropTypes.func.isRequired,
    addToast: PropTypes.func.isRequired,
  };

  static defaultProps = {
    editable: true,
    namespace: null,
    loading: false,
    limit: undefined,
    offset: undefined,
  };

  static fragments = {
    namespace: gql`
      fragment ChecksList_namespace on Namespace {
        checks(
          limit: $limit
          offset: $offset
          orderBy: $order
          filter: $filter
        ) @connection(key: "checks", filter: ["filter", "orderBy"]) {
          nodes {
            id
            deleted @client
            name
            namespace
            silences {
              name
              ...ClearSilencedEntriesDialog_silence
            }

            ...ChecksListItem_check
          }

          pageInfo {
            ...Pagination_pageInfo
          }
        }

        ...ChecksListHeader_namespace
      }

      ${ChecksListHeader.fragments.namespace}
      ${ChecksListItem.fragments.check}
      ${ClearSilencesDialog.fragments.silence}
      ${Pagination.fragments.pageInfo}
    `,
  };

  state = {
    silence: null,
    unsilence: null,
  };

  setChecksPublish = (checks, publish = true) => {
    checks.forEach(check => {
      const promise = setCheckPublish(this.props.client, {
        id: check.id,
        publish,
      });
      this.props.addToast(({ remove }) => (
        <PublishCheckStatusToast
          onClose={remove}
          mutation={promise}
          checkName={check.name}
          publish={publish}
        />
      ));
    });
  };

  silenceChecks = checks => {
    const targets = checks
      .filter(check => check.silences.length === 0)
      .map(check => ({
        namespace: check.namespace,
        check: check.name,
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

  clearSilences = checks => {
    this.setState({
      unsilence: checks.reduce((memo, ch) => [...memo, ...ch.silences], []),
    });
  };

  executeChecks = checks => {
    checks.forEach(({ id, name, namespace }) => {
      const promise = executeCheck(this.props.client, { id });

      this.props.addToast(({ remove }) => (
        <ExecuteCheckStatusToast
          onClose={remove}
          mutation={promise}
          checkName={name}
          namespace={namespace}
        />
      ));
    });
  };

  deleteChecks = checks => {
    checks.forEach(({ id }) => deleteCheck(this.props.client, { id }));
  };

  _handleChangeSort = val => {
    let newVal = val;
    this.props.onChangeQuery(query => {
      // Toggle between ASC & DESC
      const curVal = query.get("order");
      if (curVal === "NAME" && newVal === "NAME") {
        newVal = "NAME_DESC";
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

  renderCheck = ({
    key,
    item: check,
    hovered,
    setHovered,
    selectedCount,
    selected,
    setSelected,
  }) => (
    <ChecksListItem
      key={key}
      editable={this.props.editable}
      editing={selectedCount > 0}
      check={check}
      hovered={hovered}
      onHover={this.props.editable ? setHovered : () => null}
      selected={selected}
      onChangeSelected={setSelected}
      onClickClearSilences={() => this.clearSilences([check])}
      onClickDelete={() => this.deleteChecks([check])}
      onClickExecute={() => this.executeChecks([check])}
      onClickSetPublish={publish => this.setChecksPublish([check], publish)}
      onClickSilence={() => this.silenceChecks([check])}
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
      order,
      onChangeQuery,
      refetch,
    } = this.props;

    const items = namespace
      ? namespace.checks.nodes.filter(ch => !ch.deleted)
      : [];

    return (
      <ListController
        items={items}
        getItemKey={check => check.id}
        renderEmptyState={this.renderEmptyState}
        renderItem={this.renderCheck}
      >
        {({
          children,
          selectedItems,
          setSelectedItems,
          toggleSelectedItems,
        }) => (
          <Paper>
            <Loader loading={loading}>
              <ChecksListHeader
                editable={editable}
                namespace={namespace}
                onChangeQuery={onChangeQuery}
                onClickClearSilences={() => this.clearSilences(selectedItems)}
                onClickDelete={() => {
                  this.deleteChecks(selectedItems);
                  setSelectedItems([]);
                }}
                onClickExecute={() => {
                  this.executeChecks(selectedItems);
                  setSelectedItems([]);
                }}
                onClickSetPublish={publish => {
                  this.setChecksPublish(selectedItems, publish);
                  setSelectedItems([]);
                }}
                onClickSilence={() => this.silenceChecks(selectedItems)}
                order={order}
                rowCount={items.length}
                selectedItems={selectedItems}
                toggleSelectedItems={toggleSelectedItems}
              />

              <Table>
                <TableBody>{children}</TableBody>
              </Table>

              <Pagination
                limit={limit}
                offset={offset}
                pageInfo={namespace && namespace.checks.pageInfo}
                onChangeQuery={onChangeQuery}
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

              <ClearSilencesDialog
                silences={unsilence}
                open={!!unsilence}
                close={() => {
                  this.setState({ unsilence: null });
                  setSelectedItems([]);
                  refetch();
                }}
              />
            </Loader>
          </Paper>
        )}
      </ListController>
    );
  }
}

export default withApollo(ChecksList);
