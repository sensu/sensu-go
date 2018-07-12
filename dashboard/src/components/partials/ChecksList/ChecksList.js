import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Button from "@material-ui/core/Button";
import DropdownArrow from "@material-ui/icons/ArrowDropDown";
import Paper from "@material-ui/core/Paper";
import RootRef from "@material-ui/core/RootRef";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableRow from "@material-ui/core/TableRow";
import Typography from "@material-ui/core/Typography";

import { TableListEmptyState } from "/components/TableList";
import ButtonSet from "/components/ButtonSet";

import Loader from "/components/util/Loader";

import MenuController from "/components/controller/MenuController";
import ListController from "/components/controller/ListController";

import Pagination from "/components/partials/Pagination";
import ListHeader from "/components/partials/ListHeader";
import SilenceEntryDialog from "/components/partials/SilenceEntryDialog";
import ListSortMenu from "/components/partials/ListSortMenu";
import IconButton from "/components/partials/IconButton";

import executeCheck from "/mutations/executeCheck";
import { withApollo } from "react-apollo";

import ChecksListItem from "./ChecksListItem";

class ChecksList extends React.Component {
  static propTypes = {
    client: PropTypes.object.isRequired,
    environment: PropTypes.shape({
      checks: PropTypes.shape({
        nodes: PropTypes.array.isRequired,
      }),
    }),
    loading: PropTypes.bool,
    onChangeQuery: PropTypes.func.isRequired,
    limit: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    offset: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
  };

  static defaultProps = {
    environment: null,
    loading: false,
    limit: undefined,
    offset: undefined,
  };

  static fragments = {
    environment: gql`
      fragment ChecksList_environment on Environment {
        checks(
          limit: $limit
          offset: $offset
          orderBy: $order
          filter: $filter
        ) @connection(key: "checks", filter: ["filter", "orderBy"]) {
          nodes {
            id
            name
            namespace {
              environment
              organization
            }

            ...ChecksListItem_check
          }

          pageInfo {
            ...Pagination_pageInfo
          }
        }
      }

      ${Pagination.fragments.pageInfo}
      ${ChecksListItem.fragments.check}
    `,
  };

  state = {
    silence: null,
  };

  silenceChecks = checks => {
    const targets = checks.map(check => ({
      ns: {
        environment: check.namespace.environment,
        organization: check.namespace.organization,
      },
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

  silenceCheck = check => {
    this.silenceChecks([check]);
  };

  executeChecks = checks => {
    checks.forEach(({ id }) => executeCheck(this.props.client, { id }));
  };

  executeCheck = check => {
    this.executeChecks([check]);
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

  renderCheck = ({ key, item: check, selected, setSelected }) => (
    <ChecksListItem
      key={key}
      check={check}
      selected={selected}
      onChangeSelected={setSelected}
      onClickSilence={() => this.silenceCheck(check)}
      onClickExecute={() => this.executeCheck(check)}
    />
  );

  render() {
    const { environment, loading, limit, offset, onChangeQuery } = this.props;
    const items = environment ? environment.checks.nodes : [];

    return (
      <ListController
        items={items}
        getItemKey={check => check.id}
        renderEmptyState={this.renderEmptyState}
        renderItem={this.renderCheck}
      >
        {({ children, selectedItems, toggleSelectedItems }) => (
          <Paper>
            <Loader loading={loading}>
              <ListHeader
                sticky
                selectedCount={selectedItems.length}
                rowCount={children.length || 0}
                onClickSelect={toggleSelectedItems}
                renderBulkActions={() => (
                  <ButtonSet>
                    <Button onClick={() => this.silenceChecks(selectedItems)}>
                      <Typography variant="button">Silence</Typography>
                    </Button>
                    <Button onClick={() => this.executeChecks(selectedItems)}>
                      <Typography variant="button">Execute</Typography>
                    </Button>
                  </ButtonSet>
                )}
                renderActions={() => (
                  <ButtonSet>
                    <MenuController
                      renderMenu={({ anchorEl, close }) => (
                        <ListSortMenu
                          anchorEl={anchorEl}
                          onClose={close}
                          options={["NAME"]}
                          onChangeQuery={onChangeQuery}
                        />
                      )}
                    >
                      {({ open, ref }) => (
                        <RootRef rootRef={ref}>
                          <IconButton onClick={open} icon={<DropdownArrow />}>
                            Sort
                          </IconButton>
                        </RootRef>
                      )}
                    </MenuController>
                  </ButtonSet>
                )}
              />

              <Table>
                <TableBody>{children}</TableBody>
              </Table>

              <Pagination
                limit={limit}
                offset={offset}
                pageInfo={environment && environment.checks.pageInfo}
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

export default withApollo(ChecksList);
