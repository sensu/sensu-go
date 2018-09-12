import React from "react";
import PropTypes from "prop-types";
import capitalize from "lodash/capitalize";

import ArrowUp from "@material-ui/icons/ArrowUpward";
import ArrowDown from "@material-ui/icons/ArrowDownward";
import ListItemText from "@material-ui/core/ListItemText";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import MenuBase from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";

function strEndsWith(str, suffix) {
  return str.indexOf(suffix, str.length - suffix.length) !== -1;
}

class Menu extends React.PureComponent {
  static displayName = "ListSortSelector.Menu";

  static propTypes = {
    options: PropTypes.arrayOf(
      PropTypes.shape({
        label: PropTypes.string,
        value: PropTypes.string,
      }),
    ).isRequired,
    queryKey: PropTypes.string,
    onChangeQuery: PropTypes.func.isRequired,
    anchorEl: PropTypes.instanceOf(Element).isRequired,
    onClose: PropTypes.func.isRequired,
    value: PropTypes.string,
  };

  static defaultProps = {
    queryKey: "order",
    value: "",
  };

  renderOption = ({ label, value }) => {
    const { onChangeQuery, onClose, queryKey, value: valueProp } = this.props;

    let icon;
    if (valueProp === value || valueProp === `${value}_DESC`) {
      icon = (
        <ListItemIcon style={{ transform: "scale(0.77)" }}>
          {strEndsWith(valueProp, "_DESC") ? <ArrowDown /> : <ArrowUp />}
        </ListItemIcon>
      );
    }

    const onClick = () => {
      onChangeQuery(query => {
        query.set(
          queryKey,
          query.get(queryKey) === value ? `${value}_DESC` : value,
        );
      });
      onClose();
    };

    return (
      <MenuItem key={value} value={value} onClick={onClick}>
        {icon}
        <ListItemText inset>{label || capitalize(value)}</ListItemText>
      </MenuItem>
    );
  };

  render() {
    const { anchorEl, options, onClose } = this.props;

    return (
      <MenuBase open anchorEl={anchorEl} onClose={onClose}>
        {options.map(this.renderOption)}
      </MenuBase>
    );
  }
}

export default Menu;
