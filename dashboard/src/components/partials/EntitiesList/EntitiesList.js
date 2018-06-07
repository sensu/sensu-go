import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import TableList, {
  TableListBody,
  TableListEmptyState,
} from "/components/TableList";

import Loader from "/components/util/Loader";
import ListController from "/components/util/ListController";

import EntitiesListHeader from "./EntitiesListHeader";
import EntitiesListItem from "./EntitiesListItem";

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

  render() {
    const { environment, loading } = this.props;

    return (
      <ListController
        items={environment ? environment.entities.nodes : []}
        getItemKey={item => item.id}
        renderEmptyState={() =>
          !loading && (
            <TableListEmptyState
              primary="No results matched your query."
              secondary="
                Try refining your search query in the search box. The filter
                buttons above are also a helpful way of quickly finding
                entities.
              "
            />
          )
        }
        renderItem={({ key, item, selected, setSelected }) => (
          <EntitiesListItem
            key={key}
            entity={item}
            selected={selected}
            onClickSelect={() => setSelected(!selected)}
          />
        )}
      >
        {({ children, selectedItems, toggleSelectedItems }) => (
          <TableList>
            <EntitiesListHeader
              onClickSelect={toggleSelectedItems}
              selectedCount={selectedItems.length}
            />
            <Loader loading={loading}>
              <TableListBody>{children}</TableListBody>
            </Loader>
          </TableList>
        )}
      </ListController>
    );
  }
}

export default EntitiesList;
