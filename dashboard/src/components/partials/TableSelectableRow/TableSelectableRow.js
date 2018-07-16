import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

import TableRow from "@material-ui/core/TableRow";

const styles = theme => ({
  root: {
    verticalAlign: "top",

    // hover
    // https://material.io/guidelines/components/data-tables.html#data-tables-interaction
    "&:hover": {
      backgroundColor: theme.palette.action.selected,
    },
  },
  // selected
  // https://material.io/guidelines/components/data-tables.html#data-tables-interaction
  selected: {
    backgroundColor: theme.palette.action.hover,
  },
});

class TableSelectableRow extends React.PureComponent {
  static propTypes = {
    selected: PropTypes.bool.isRequired,
    children: PropTypes.node.isRequired,
    classes: PropTypes.object.isRequired,
  };

  render() {
    const { classes, selected, children } = this.props;

    return (
      <TableRow
        className={classnames(classes.root, { [classes.selected]: selected })}
      >
        {children}
      </TableRow>
    );
  }
}

export default withStyles(styles)(TableSelectableRow);
