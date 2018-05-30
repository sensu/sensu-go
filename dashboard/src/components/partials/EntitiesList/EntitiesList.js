import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import TableList, {
  TableListBody,
  TableListEmptyState,
} from "/components/TableList";

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
    // eslint-disable-next-line react/no-unused-prop-types
    environment: PropTypes.object,
    loading: PropTypes.bool,
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

  componentWillReceiveProps(nextProps) {
    this.setState({
      selectedIds: trimIds(this.state.selectedIds, nextProps),
    });
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

  render() {
    const entities = getEntities(this.props);

    return (
      <TableList>
        <EntitiesListHeader
          onClickSelect={this._handleClickHeaderSelect}
          selectedCount={this.state.selectedIds.length}
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
              />
            ))}
          </TableListBody>
        </Loader>
      </TableList>
    );
  }
}

export default EntitiesList;
