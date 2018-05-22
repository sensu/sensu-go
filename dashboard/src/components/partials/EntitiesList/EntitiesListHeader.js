import React from "react";
import PropTypes from "prop-types";

import Checkbox from "@material-ui/core/Checkbox";
import { withStyles } from "@material-ui/core/styles";

import { TableListHeader } from "/components/TableList";

const styles = theme => ({
  headerButton: {
    marginLeft: theme.spacing.unit / 2,
    "&:first-child": {
      marginLeft: theme.spacing.unit,
    },
  },
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
    onClickSelect: PropTypes.func,
    selectedCount: PropTypes.number,
    classes: PropTypes.object.isRequired,
  };

  static defaultProps = {
    onClickSelect: () => {},
    selectedCount: 0,
  };

  render() {
    const { selectedCount, classes, onClickSelect } = this.props;

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
      </TableListHeader>
    );
  }
}

export default withStyles(styles)(EntitiesListHeader);
