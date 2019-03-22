import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

import TableRow from "@material-ui/core/TableRow";

const styles = theme => {
  const transitionIn = `
    background-color
    ${theme.transitions.duration.shortest}ms
    ${theme.transitions.easing.sharp}
  `;
  const transitionOut = `
    background-color
    ${theme.transitions.duration.complex}ms
    ${theme.transitions.easing.easeOut}
  `;

  return {
    root: {
      verticalAlign: "top",
      transition: transitionOut,
    },
    // selected
    // https://material.io/guidelines/components/data-tables.html#data-tables-interaction
    selected: {
      backgroundColor: theme.palette.action.hover,
      transition: transitionIn,
    },
    highlight: {
      animation: `
        ${theme.transitions.duration.complex}ms
        ${theme.transitions.easing.easeInOut}
        0s
        alternate
        2
        selectable-row-highlight
      `,
    },
    "@keyframes selectable-row-highlight": {
      "0%": {
        backgroundColor: "inherit",
      },
      "100%": {
        backgroundColor: theme.palette.action.hover,
      },
    },
  };
};

class TableSelectableRow extends React.PureComponent {
  static propTypes = {
    selected: PropTypes.bool.isRequired,
    children: PropTypes.node.isRequired,
    classes: PropTypes.object.isRequired,
    highlight: PropTypes.bool,
  };

  static defaultProps = {
    highlight: false,
  };

  render() {
    const { classes, selected, children, highlight, ...props } = this.props;
    const className = classnames(classes.root, {
      [classes.selected]: selected,
      [classes.highlight]: highlight,
    });

    return (
      <TableRow className={className} {...props}>
        {children}
      </TableRow>
    );
  }
}

export default withStyles(styles)(TableSelectableRow);
