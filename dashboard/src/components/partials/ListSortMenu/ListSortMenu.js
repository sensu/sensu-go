import React from "react";
import PropTypes from "prop-types";
import capitalize from "lodash/capitalize";

import ListItemText from "@material-ui/core/ListItemText";
import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";

class ListSortMenu extends React.PureComponent {
  static propTypes = {
    options: PropTypes.arrayOf(PropTypes.string).isRequired,
    queryKey: PropTypes.string,
    onChangeQuery: PropTypes.func.isRequired,
    anchorEl: PropTypes.instanceOf(Element).isRequired,
    onClose: PropTypes.func.isRequired,
  };

  static defaultProps = {
    queryKey: "order",
  };

  render() {
    const { options, anchorEl, onClose, onChangeQuery, queryKey } = this.props;

    return (
      <Menu open anchorEl={anchorEl} onClose={onClose}>
        {options.map(option => (
          <MenuItem
            key={option}
            value={option}
            onClick={() => {
              onChangeQuery(query => {
                query.set(
                  queryKey,
                  query.get(queryKey) === option ? `${option}_DESC` : option,
                );
              });
              onClose();
            }}
          >
            <ListItemText>{capitalize(option)}</ListItemText>
          </MenuItem>
        ))}
      </Menu>
    );
  }
}

export default ListSortMenu;
