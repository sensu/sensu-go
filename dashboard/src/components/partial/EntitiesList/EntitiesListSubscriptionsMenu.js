import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { MenuItem } from "material-ui/Menu";
import { ListItemText } from "material-ui/List";

import { TableListSelect } from "/components/TableList";

export const arrayMerge = (arr1, arr2) =>
  arr2.reduce(
    (result, val) => {
      if (result.indexOf(val) === -1) {
        result.push(val);
      }
      return result;
    },
    [...arr1],
  );

class EntitiesListSubscriptionsMenu extends React.PureComponent {
  static propTypes = {
    entities: PropTypes.object.isRequired,
    onChange: PropTypes.func,
  };

  static defaultProps = {
    values: [],
    onChange: undefined,
  };

  static fragments = {
    entityConnection: gql`
      fragment EntitiesListSubscriptionsMenu_entityConnection on EntityConnection {
        nodes {
          subscriptions
        }
      }
    `,
  };

  render() {
    const { entities } = this.props;

    const subscriptions = entities.nodes.reduce(
      (result, node) => arrayMerge(result, node.subscriptions),
      [],
    );

    return (
      <TableListSelect label="Subscriptions" onChange={this.props.onChange}>
        {subscriptions.map(subscription => (
          <MenuItem key={subscription} value={subscription}>
            <ListItemText primary={subscription} />
          </MenuItem>
        ))}
      </TableListSelect>
    );
  }
}

export default EntitiesListSubscriptionsMenu;
