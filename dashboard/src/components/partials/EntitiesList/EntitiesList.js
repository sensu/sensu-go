import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";
import deleteEntity from "/mutations/deleteEntity";
import Loader from "/components/util/Loader";
import TableList, {
  TableListBody,
  TableListEmptyState,
} from "/components/TableList";
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
    environment: PropTypes.object,
    loading: PropTypes.bool,
    onChangeQuery: PropTypes.func.isRequired,
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
        subscriptions(orderBy: OCCURRENCES, omitEntity: true) {
          ...EntitiesListHeader_subscriptions
        }
      }

      ${EntitiesListItem.fragments.entity}
      ${EntitiesListHeader.fragments.subscriptions}
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

  _handleChangeFilter = (filter, val) => {
    switch (filter) {
      case "subscription":
        this.props.onChangeQuery({ filter: `'${val}' IN Subscriptions` });
        break;
      default:
        throw new Error(`unexpected filter '${filter}'`);
    }
  };

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
    this.props.onChangeQuery(query => {
      // Toggle between ASC & DESC
      const curVal = query.get("order");
      if (curVal === "ID" && newVal === "ID") {
        newVal = "ID_DESC";
      }
      query.set("order", newVal);
    });
  };

  _getSubscriptions = () =>
    this.props.environment && this.props.environment.subscriptions;

  render() {
    const entities = getEntities(this.props);

    return (
      <TableList>
        <EntitiesListHeader
          onChangeFilter={this._handleChangeFilter}
          onClickSelect={this._handleClickHeaderSelect}
          onChangeSort={this._handleSort}
          onSubmitDelete={this._handleDeleteItems}
          selectedCount={this.state.selectedIds.length}
          subscriptions={this._getSubscriptions()}
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
