import React from "react";
import PropTypes from "prop-types";
import Checkbox from "@material-ui/core/Checkbox";
import { withStyles } from "@material-ui/core/styles";
import { TableListHeader } from "/components/TableList";

const styles = theme => ({
  filterActions: {
    display: "none",
    [theme.breakpoints.up("sm")]: {
      display: "flex",
    },
  },
  // Remove padding from button container
  checkbox: {
    marginLeft: -11,
    color: theme.palette.primary.contrastText,
  },
  grow: {
    flex: "1 1 auto",
  },
});

class EntitiesListHeader extends React.PureComponent {
  static propTypes = {
    actions: PropTypes.node.isRequired,
    bulkActions: PropTypes.node.isRequired,
    classes: PropTypes.object.isRequired,
    onClickSelect: PropTypes.func,
    selectedCount: PropTypes.number,
  };

  static defaultProps = {
    onClickSelect: () => {},
    selectedCount: 0,
  };

  render() {
    const {
      actions,
      bulkActions,
      classes,
      selectedCount,
      onClickSelect,
    } = this.props;

    return (
      <TableListHeader sticky active={selectedCount > 0}>
        <Checkbox
          component="button"
          className={classes.checkbox}
          onClick={onClickSelect}
          checked={false}
          indeterminate={selectedCount > 0}
        />
        {selectedCount > 0 && <div>{selectedCount} Selected</div>}
        <div className={classes.grow} />
        {selectedCount > 0 ? bulkActions : actions}
      </TableListHeader>
    );
  }
}

export default withStyles(styles)(EntitiesListHeader);
