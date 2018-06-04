import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";

import TableList, {
  TableListBody,
  TableListEmptyState,
  TableListSelect as Select,
} from "/components/TableList";

import Button from "@material-ui/core/Button";
import ButtonSet from "/components/ButtonSet";
import MenuItem from "@material-ui/core/MenuItem";
import ListItemText from "@material-ui/core/ListItemText";

import ConfirmDelete from "/components/partials/ConfirmDelete";
import deleteEntity from "/mutations/deleteEntity";

import Loader from "/components/util/Loader";

import EntitiesListHeader from "./EntitiesListHeader";
import EntitiesListItem from "./EntitiesListItem";

const arrayRemove = (arr, val) => {
  const index = arr.indexOf(val);
  return index === -1 ? arr : arr.slice(0, index).concat(arr.slice(index + 1));
};

const arrayAdd = (arr, val) =>
  arr.indexOf(val) === -1 ? arr.concat([val]) : arr;

const arrayIntersect = (arr1, arr2) =>
  arr1.filter(val => arr2.indexOf(val) !== -1);

const getEntities = props => {
  const { environment } = props;
  return environment ? environment.entities : { nodes: [] };
};

const trimIds = (selectedIds, props) => {
  const ids = getEntities(props).nodes.map(node => node.id);
  return arrayIntersect(selectedIds, ids);
};

class EntitiesList extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    // eslint-disable-next-line react/no-unused-prop-types
    environment: PropTypes.object,
    loading: PropTypes.bool,
    onQueryChange: PropTypes.func.isRequired,
  };

  static defaultProps = {
    environment: null,
    loading: false,
  };

  static fragments = {
    environment: gql`
      fragment EntitiesList_environment on Environment {
        entities(limit: 1000, filter: $filter, orderBy: $order)
          @connection(key: "entities", filter: ["filter", "orderBy"]) {
          nodes {
            ...EntitiesListItem_entity
          }
        }
      }

      ${EntitiesListItem.fragments.entity}
    `,
  };

  state = {
    selectedIds: [],
  };

  static getDerivedStateFromProps(props, state) {
    return {
      ...state,
      selectedIds: trimIds(state.selectedIds, props),
    };
  }

  _handleClickHeaderSelect = () => {
    const entities = getEntities(this.props);

    if (this.state.selectedIds.length < entities.nodes.length) {
      const ids = entities.nodes.map(node => node.id);

      this.setState({ selectedIds: ids });
    } else {
      this.setState({ selectedIds: [] });
    }
  };

  _handleClickItemSelect = id => (event, selected) =>
    this.setState((previousState, props) => {
      const nextSelectedIds = selected
        ? arrayAdd(previousState.selectedIds, id)
        : arrayRemove(previousState.selectedIds, id);

      return {
        selectedIds: trimIds(nextSelectedIds, props),
      };
    });

  _handleDeleteItems = () => {
    const { selectedIds } = this.state;
    const { client } = this.props;

    // Fire delete operations
    Promise.all(selectedIds.map(id => deleteEntity(client, { id })));

    // Clear selected items
    this.setState({ selectedIds: [] });
  };

  _handleDeleteItem = id => () => {
    deleteEntity(this.props.client, { id });
  };

  _handleSort = val => {
    let newVal = val;
    this.props.onQueryChange(query => {
      // Toggle between ASC & DESC
      const curVal = query.get("order");
      if (curVal === "ID" && newVal === "ID") {
        newVal = "ID_DESC";
      }
      query.set("order", newVal);
    });
  };

  render() {
    const entities = getEntities(this.props);
    const selectLen = this.state.selectedIds.length;

    return (
      <TableList>
        <EntitiesListHeader
          onClickSelect={this._handleClickHeaderSelect}
          selectedCount={this.state.selectedIds.length}
          actions={
            <ButtonSet>
              <Select label="Sort" onChange={this._handleSort}>
                <MenuItem key="ID" value="ID">
                  <ListItemText>Name</ListItemText>
                </MenuItem>
                <MenuItem key="LASTSEEN" value="LASTSEEN">
                  <ListItemText>Last Seen</ListItemText>
                </MenuItem>
              </Select>
            </ButtonSet>
          }
          bulkActions={
            <ButtonSet>
              <ConfirmDelete
                identifier={`${selectLen} ${
                  selectLen === 1 ? "entity" : "entities"
                }`}
                onSubmit={() => this._handleDeleteItems()}
              >
                {confirm => (
                  <Button onClick={() => confirm.open()}>Delete</Button>
                )}
              </ConfirmDelete>
            </ButtonSet>
          }
        />
        <Loader loading={this.props.loading}>
          <TableListBody>
            {!this.props.loading &&
              entities.nodes.length === 0 && (
                <TableListEmptyState
                  primary="No results matched your query."
                  secondary="
                  Try refining your search query in the search box. The filter
                  buttons above are also a helpful way of quickly finding
                  entities.
                "
                />
              )}
            {entities.nodes.map(node => (
              <EntitiesListItem
                key={node.id}
                entity={node}
                selected={this.state.selectedIds.indexOf(node.id) !== -1}
                onClickSelect={this._handleClickItemSelect(node.id)}
                onClickDelete={this._handleDeleteItem(node.id)}
              />
            ))}
          </TableListBody>
        </Loader>
      </TableList>
    );
  }
}

export default withApollo(EntitiesList);
