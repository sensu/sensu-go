import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";

const styles = theme => ({
  root: {
    display: "flex",
    width: "100%",
    border: 0,
    borderBottomWidth: 1,
    borderColor: theme.palette.divider,
    borderStyle: "solid",
    color: theme.palette.text.primary,
    padding: "0 16px",
    transition: theme.transitions.create("background-color", {
      easing: theme.transitions.easing.easeIn,
      duration: theme.transitions.duration.shortest,
    }),

    // remove border from last item
    "&:last-child": {
      borderBottom: "none",
    },

    // hover
    // https://material.io/guidelines/components/data-tables.html#data-tables-interaction
    "&:hover": {
      backgroundColor: theme.palette.action.hover,
    },
  },

  // selected
  // https://material.io/guidelines/components/data-tables.html#data-tables-interaction
  selected: {
    backgroundColor: theme.palette.action.selected,
  },
});

export class TableListItem extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    children: PropTypes.node.isRequired,
    selected: PropTypes.bool,
  };

  static defaultProps = {
    className: "",
    selected: false,
  };

  render() {
    const {
      classes,
      className: classNameProp,
      children,
      selected,
    } = this.props;
    const className = classnames(classes.root, classNameProp, {
      [classes.selected]: selected,
    });

    return (
      <Typography component="div" className={className}>
        {children}
      </Typography>
    );
  }
}

export default withStyles(styles)(TableListItem);
