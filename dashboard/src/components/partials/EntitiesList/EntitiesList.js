import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";
import TableList, {
  TableListBody,
  TableListEmptyState,
} from "/components/TableList";

import deleteEntity from "/mutations/deleteEntity";

import Loader from "/components/util/Loader";
import ListController from "/components/util/ListController";
import Pagination from "/components/partials/Pagination";

import EntitiesListHeader from "./EntitiesListHeader";
import EntitiesListItem from "./EntitiesListItem";

class EntitiesList extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    environment: PropTypes.object,
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
      fragment EntitiesList_environment on Environment {
        entities(
          limit: $limit
          offset: $offset
          filter: $filter
          orderBy: $order
        ) @connection(key: "entities", filter: ["filter", "orderBy"]) {
          nodes {
            id
            deleted @client
            ...EntitiesListItem_entity
          }

          pageInfo {
            ...Pagination_pageInfo
          }
        }
        ...EntitiesListHeader_environment
      }

      ${EntitiesListItem.fragments.entity}
      ${EntitiesListHeader.fragments.environment}
      ${Pagination.fragments.pageInfo}
    `,
  };

  deleteEntities = entities => {
    const { client } = this.props;
    entities.forEach(entity => deleteEntity(client, { id: entity.id }));
  };

  renderEmptyState = () => {
    const { loading } = this.props;

    return (
      <TableListEmptyState
        loading={loading}
        primary="No results matched your query."
        secondary="
          Try refining your search query in the search box. The filter buttons
          above are also a helpful way of quickly finding entities.
        "
      />
    );
  };

  renderEntity = ({ key, item: entity, selected, setSelected }) => (
    <EntitiesListItem
      key={key}
      entity={entity}
      selected={selected}
      onClickSelect={() => setSelected(!selected)}
      onClickDelete={() => this.deleteEntities([entity])}
    />
  );

  render() {
    const { environment, loading, onChangeQuery, limit, offset } = this.props;

    const items = environment
      ? environment.entities.nodes.filter(entity => !entity.deleted)
      : [];

    return (
      <ListController
        items={items}
        getItemKey={entity => entity.id}
        renderEmptyState={this.renderEmptyState}
        renderItem={this.renderEntity}
      >
        {({ children, selectedItems, toggleSelectedItems }) => (
          <TableList>
            <EntitiesListHeader
              selectedCount={selectedItems.length}
              onClickSelect={toggleSelectedItems}
              onClickDelete={() => this.deleteEntities(selectedItems)}
              onChangeQuery={onChangeQuery}
              environment={environment}
            />
            <Loader loading={loading}>
              <TableListBody>{children}</TableListBody>
            </Loader>
            <Pagination
              limit={limit}
              offset={offset}
              pageInfo={environment && environment.entities.pageInfo}
              onChangeQuery={onChangeQuery}
            />
          </TableList>
        )}
      </ListController>
    );
  }
}

export default withApollo(EntitiesList);
